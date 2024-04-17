// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package pipeline

import (
	"fmt"

	plmanager "github.com/GuanceCloud/cliutils/pipeline/manager"
	"github.com/GuanceCloud/cliutils/pipeline/manager/relation"
	"github.com/GuanceCloud/cliutils/pipeline/ptinput"
	"github.com/GuanceCloud/cliutils/point"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/datakit"
	plval "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/pipeline/plval"
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
	plOpt *plmanager.Option,
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
			script.NS() == plmanager.RemoteScriptNS &&
			category == point.Logging {
			ptsOffload = append(ptsOffload, pt)
			continue
		}

		inputData := ptinput.WrapPoint(category, pt)

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

		// oldPt will next be replaced or dropped
		// put the old point back into the pool
		datakit.PutbackPoints(pt)

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
) (*plmanager.PlScript, bool) {
	if pt == nil {
		return nil, false
	}
	center, ok := plval.GetManager()
	if !ok || center == nil {
		return nil, false
	}

	relat := center.GetScriptRelation()
	scriptName, ok := relation.ScriptName(relat, cat, pt, scriptMap)
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
