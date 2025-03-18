// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package pipeline

import (
	"fmt"
	"time"

	"github.com/GuanceCloud/cliutils/point"
	"github.com/GuanceCloud/pipeline-go/constants"
	"github.com/GuanceCloud/pipeline-go/lang"
	"github.com/GuanceCloud/pipeline-go/lang/platypus"
	plmanager "github.com/GuanceCloud/pipeline-go/manager"
	"github.com/GuanceCloud/pipeline-go/ptinput"
	"github.com/GuanceCloud/platypus/pkg/ast"
	plval "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/pipeline/plval"
)

const (
	plTagName    = "_pl_script"
	plTagService = "_pl_service"
	plTagNS      = "_pl_ns"
	plStatus     = "_pl_status"
	plFieldCost  = "_pl_cost" // data type: float64, unit: second

	svcName = "datakit"

	sOk     = "ok"
	sFailed = "failed"
)

type ScriptResult struct {
	pts        []*point.Point
	ptsOffload []*point.Point
	ptsCreated map[point.Category][]*point.Point
}

func (r *ScriptResult) Pts() []*point.Point {
	return r.pts
}

func (r *ScriptResult) PtsOffload() []*point.Point {
	return r.ptsOffload
}

func (r *ScriptResult) PtsCreated() map[point.Category][]*point.Point {
	return r.ptsCreated
}

func RunPl(category point.Category, pts []*point.Point,
	plOpt *lang.LogOption,
) (reslt *ScriptResult, retErr error) {
	defer func() {
		if err := recover(); err != nil {
			retErr = fmt.Errorf("run pl: %s", err)
		}
	}()

	if sManager, ok := plval.GetManager(); ok && sManager != nil {
		if sManager.ScriptCount(category) < 1 {
			return &ScriptResult{
				pts: pts,
			}, nil
		}
	} else {
		return nil, fmt.Errorf("script manager not ready")
	}

	ret := []*point.Point{}
	ptsOffload := []*point.Point{}

	subPt := make(map[point.Category][]*point.Point)
	for _, pt := range pts {
		var sMap map[string]string
		if plOpt != nil {
			sMap = plOpt.ScriptMap
		}
		script, ok := searchScript(category, pt, sMap)

		if !ok || script == nil {
			ret = append(ret, pt)
			continue
		}

		if v, ok := plval.GetOffload(); ok && v != nil &&
			script.NS() == constants.NSRemote &&
			category == point.Logging {
			ptsOffload = append(ptsOffload, pt)
			continue
		}

		startTime := time.Now()
		if plval.EnableAppendRunInfo() {
			pt.AddTag(plTagName, script.Name())
			pt.AddTag(plTagService, svcName)
			pt.AddTag(plTagNS, script.NS())
		}

		inputData := ptinput.PtWrap(category, pt)

		if v, ok := plval.GetRefTb(); ok {
			inputData.SetPlReferTables(v.Tables())
		}
		if v, ok := plval.GetIPDB(); ok {
			inputData.SetIPDB(v)
		}

		// run pl srcipt
		err := script.Run(inputData, nil, plOpt)
		if err != nil {
			l.Warn(err)
			if plval.EnableAppendRunInfo() {
				plCost := time.Since(startTime)
				pt.AddTag(plStatus, sFailed)
				pt.Add(plFieldCost, float64(plCost)/float64(time.Second))
			}
			ret = append(ret, pt)
			continue
		}

		if pts := inputData.GetSubPoint(); len(pts) > 0 {
			for _, pt := range pts {
				if !pt.Dropped() {
					subPt[pt.Category()] = append(subPt[pt.Category()], pt.Point())
				}
			}
		}

		if plval.EnableAppendRunInfo() {
			plCost := time.Since(startTime)
			_ = inputData.SetTag(plStatus, sOk, ast.String)
			_ = inputData.Set(plFieldCost, float64(plCost)/float64(time.Second), ast.Float)
		}

		if ctxPts := inputData.CallbackPtWinMove(); len(ctxPts) > 0 {
			subPt[category] = append(subPt[category], ctxPts...)
		}

		if inputData.Dropped() { // drop
			continue
		}

		ret = append(ret, inputData.Point())
	}

	return &ScriptResult{
		pts:        ret,
		ptsOffload: ptsOffload,
		ptsCreated: subPt,
	}, nil
}

func searchScript(cat point.Category,
	pt *point.Point, scriptMap map[string]string,
) (*platypus.PlScript, bool) {
	if pt == nil {
		return nil, false
	}
	center, ok := plval.GetManager()
	if !ok || center == nil {
		return nil, false
	}

	relat := center.GetScriptRelation()
	scriptName, ok := plmanager.ScriptName(relat, cat, pt, scriptMap)
	if !ok {
		return nil, false
	}

	sc, ok := center.QueryScript(cat, scriptName)
	if ok {
		return sc, true
	} else {
		return nil, false
	}
}
