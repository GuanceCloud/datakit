// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package pipeline

import (
	"fmt"
	"strings"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io/point"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/pipeline/ptinput"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/pipeline/relation"
	plscript "gitlab.jiagouyun.com/cloudcare-tools/datakit/pipeline/script"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/pipeline/stats"
)

func RunPl(category string, pts []*point.Point, plOpt *plscript.Option, scriptMap map[string]string) (ret []*point.Point, retErr error) {
	defer func() {
		if err := recover(); err != nil {
			retErr = fmt.Errorf("run pl: %s", err)
		}
	}()

	if plscript.ScriptCount(category) < 1 {
		return pts, nil
	}

	ret = []*point.Point{}
	ptOpt := &point.PointOption{
		DisableGlobalTags: true,
		Category:          category,
	}
	if plOpt != nil {
		ptOpt.MaxFieldValueLen = plOpt.MaxFieldValLen
	}

	plPt := ptinput.GetPoint()
	defer ptinput.PutPoint(plPt)

	var ok bool
	var script *plscript.PlScript
	for _, pt := range pts {
		// 这里将清理 plPt 并填充 point 到 plPt,
		// plPt 在函数运行结束后尽量放回对象池
		script, plPt, ok = getScriptAndFillPlPt(category, pt, scriptMap, plPt)

		if !ok || script == nil {
			ret = append(ret, pt)
			continue
		}

		err := script.Run(plPt, nil, plOpt)
		if err != nil {
			l.Warn(err)
			ret = append(ret, pt)
			continue
		}

		if plPt.Drop { // drop
			continue
		}

		if p, err := point.NewPoint(plPt.Name, plPt.Tags, plPt.Fields, ptOpt); err != nil {
			l.Error(err)
			stats.WriteScriptStats(script.Category(), script.NS(), script.Name(), 0, 0, 1, 0, err)
		} else {
			pt = p
		}

		ret = append(ret, pt)
	}

	return ret, nil
}

func getScriptAndFillPlPt(category string, pt *point.Point, scriptMap map[string]string, plpt *ptinput.Point) (
	*plscript.PlScript, *ptinput.Point, bool,
) {
	if pt == nil {
		return nil, plpt, false
	}
	if plpt == nil {
		plpt = &ptinput.Point{}
	}

	switch category {
	case datakit.RUM, datakit.Security, datakit.Tracing, datakit.Profiling:
		fields, err := pt.Fields()
		if err != nil {
			l.Debug(err)
			break
		}
		ptTime := pt.Time()
		name := pt.Name()
		tags := pt.Tags()

		scriptName, ok := scriptName(category, name, tags, fields, scriptMap)
		if !ok {
			break
		}
		if s, ok := plscript.QueryScript(category, scriptName); ok {
			fields, err := pt.Fields()
			if err != nil {
				l.Errorf("Fields: %s", err)
				break
			}

			plpt = ptinput.InitPt(plpt, name, tags, fields, ptTime)

			return s, plpt, true
		}
	default:
		name := pt.Name()
		scriptName, ok := scriptName(category, name, nil, nil, scriptMap)
		if !ok {
			break
		}
		// 未查询到脚本时条过解析 Point
		s, ok := plscript.QueryScript(category, scriptName)
		if !ok {
			break
		}

		fields, err := pt.Fields()
		if err != nil {
			l.Errorf("Fields: %s", err)
			break
		}

		ptTime := pt.Time()
		tags := pt.Tags()
		plpt = ptinput.InitPt(plpt, name, tags, fields, ptTime)
		return s, plpt, true
	}

	return nil, plpt, false
}

func scriptName(category string, name string, tags map[string]string, fields map[string]interface{},
	scriptMap map[string]string,
) (string, bool) {
	var scriptName string

	// tag 优先，key 唯一
	switch category {
	case datakit.RUM:
		if id, ok := tags["app_id"]; ok {
			scriptName = joinName(id, name)
		} else if id, ok := fields["app_id"]; ok {
			switch id := id.(type) {
			case string:
				scriptName = joinName(id, name)
			default:
			}
		}
	case datakit.Security:
		if scheckCategory, ok := tags["category"]; ok {
			scriptName = scheckCategory
		} else if scheckCategory, ok := fields["category"]; ok {
			switch scheckCategory := scheckCategory.(type) {
			case string:
				scriptName = scheckCategory
			default:
			}
		}
	case datakit.Tracing, datakit.Profiling:
		if svc, ok := tags["service"]; ok {
			scriptName = svc
		} else if svc, ok := fields["service"]; ok {
			switch svc := svc.(type) {
			case string:
				scriptName = svc
			default:
			}
		}
	default:
		scriptName = name
	}

	if scriptName == "" {
		return "", false
	}

	// 查找，值 `-` 禁用
	if sName, ok := scriptMap[scriptName]; ok {
		switch sName {
		case "-":
			return "", false
		case "":
		default:
			return sName, ok
		}
	}

	if sName, ok := relation.QueryRemoteRelation(category, scriptName); ok {
		return sName, ok
	}

	return scriptName + ".p", true
}

func joinName(name ...string) string {
	return strings.Join(name, "_")
}
