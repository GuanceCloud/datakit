// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package pipeline

import (
	"fmt"

	"github.com/GuanceCloud/cliutils/point"
	dkpt "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/io/point"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/pipeline/offload"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/pipeline/ptinput"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/pipeline/relation"
	plscript "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/pipeline/script"
)

type ScriptResult struct {
	pts        []*dkpt.Point
	ptsOffload []*dkpt.Point
	ptsCreated map[point.Category][]*dkpt.Point
}

func (r *ScriptResult) Pts() []*dkpt.Point {
	return r.pts
}

func (r *ScriptResult) PtsOffload() []*dkpt.Point {
	return r.ptsOffload
}

func (r *ScriptResult) PtsCreated() map[point.Category][]*dkpt.Point {
	return r.ptsCreated
}

func RunPl(category point.Category, pts []*dkpt.Point,
	plOpt *plscript.Option, scriptMap map[string]string,
) (reslt *ScriptResult, retErr error) {
	defer func() {
		if err := recover(); err != nil {
			retErr = fmt.Errorf("run pl: %s", err)
		}
	}()

	if plscript.ScriptCount(category) < 1 {
		return &ScriptResult{
			pts: pts,
		}, nil
	}

	ret := []*dkpt.Point{}
	offl := []*dkpt.Point{}
	ptOpt := &dkpt.PointOption{
		DisableGlobalTags: true,
		Category:          category.URL(),
	}

	if plOpt != nil {
		ptOpt.MaxFieldValueLen = plOpt.MaxFieldValLen
	}

	subPt := make(map[point.Category][]*dkpt.Point)
	for _, pt := range pts {
		script, inputData, ok := searchScript(category, pt, scriptMap)

		if !ok || script == nil || inputData == nil {
			ret = append(ret, pt)
			continue
		}

		if offload.Enabled() &&
			script.NS() == plscript.RemoteScriptNS &&
			category == point.Logging {
			offl = append(offl, pt)
			continue
		}

		err := script.Run(inputData, nil, plOpt)
		if err != nil {
			l.Warn(err)
			ret = append(ret, pt)
			continue
		}

		if pts := inputData.GetSubPoint(); len(pts) > 0 {
			for _, pt := range pts {
				if !pt.Dropped() {
					if dkpt, err := pt.DkPoint(); err != nil {
						subPt[pt.Category()] = append(subPt[pt.Category()], dkpt)
					}
				}
			}
		}

		if inputData.Dropped() { // drop
			continue
		}

		if dkpt, err := inputData.DkPoint(); err != nil {
			ret = append(ret, pt)
		} else {
			ret = append(ret, dkpt)
		}
	}

	return &ScriptResult{
		pts:        ret,
		ptsOffload: offl,
		ptsCreated: subPt,
	}, nil
}

func searchScript(cat point.Category, pt *dkpt.Point, scriptMap map[string]string) (*plscript.PlScript, ptinput.PlInputPt, bool) {
	if pt == nil {
		return nil, nil, false
	}
	scriptName, plpt, ok := scriptName(cat, pt, scriptMap)
	if !ok {
		return nil, nil, false
	}

	sc, ok := plscript.QueryScript(cat, scriptName)
	if ok {
		if plpt == nil {
			var err error
			plpt, err = ptinput.WrapDeprecatedPoint(cat, pt)
			if err != nil {
				return nil, nil, false
			}
		}
		return sc, plpt, true
	} else {
		return nil, nil, false
	}
}

func scriptName(cat point.Category, pt *dkpt.Point, scriptMap map[string]string) (string, ptinput.PlInputPt, bool) {
	if pt == nil {
		return "", nil, false
	}

	var scriptName string
	var plpt ptinput.PlInputPt
	var err error

	// built-in rules last
	switch cat { //nolint:exhaustive
	case point.RUM:
		plpt, err = ptinput.WrapDeprecatedPoint(cat, pt)
		if err != nil {
			return "", nil, false
		}
		scriptName = _rumSName(plpt)
	case point.Security:
		plpt, err = ptinput.WrapDeprecatedPoint(cat, pt)
		if err != nil {
			return "", nil, false
		}
		scriptName = _securitySName(plpt)
	case point.Tracing, point.Profiling:
		plpt, err = ptinput.WrapDeprecatedPoint(cat, pt)
		if err != nil {
			return "", nil, false
		}
		scriptName = _apmSName(plpt)
	default:
		scriptName = _defaultCatSName(pt)
	}

	if scriptName == "" {
		return "", plpt, false
	}

	// configuration first
	if sName, ok := scriptMap[scriptName]; ok {
		switch sName {
		case "-":
			return "", nil, false
		case "":
		default:
			return sName, plpt, true
		}
	}

	// remote relation sencond
	if sName, ok := relation.QueryRemoteRelation(cat, scriptName); ok {
		return sName, plpt, true
	}

	return scriptName + ".p", plpt, true
}
