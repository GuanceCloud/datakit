// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package io

import (
	"fmt"
	"strings"
	"time"

	"gitlab.jiagouyun.com/cloudcare-tools/cliutils/logger"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io/point"
	plscript "gitlab.jiagouyun.com/cloudcare-tools/datakit/pipeline/script"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/pipeline/stats"
)

var plLogger = logger.DefaultSLogger("pipeline")

func runPl(category string, pts []*point.Point, opt *Option) (ret []*point.Point, retErr error) {
	defer func() {
		if err := recover(); err != nil {
			retErr = fmt.Errorf("run pl: %s", err)
		}
	}()

	if plscript.ScriptCount(category) < 1 {
		return pts, nil
	}

	var scriptMap map[string]string
	var plOpt *plscript.Option
	if opt != nil {
		scriptMap = opt.PlScript
		plOpt = opt.PlOption
	}
	ret = []*point.Point{}
	ptOpt := &point.PointOption{
		DisableGlobalTags: true,
		Category:          category,
	}
	if plOpt != nil {
		ptOpt.MaxFieldValueLen = plOpt.MaxFieldValLen
	}

	for _, pt := range pts {
		script, ptName, tags, fields, ptTime := getScript(category, pt, scriptMap)

		if script == nil {
			ret = append(ret, pt)
			continue
		}

		name, tags, fields, tn, drop, err := script.Run(ptName,
			tags, fields, *ptTime, nil, plOpt)
		if err != nil {
			plLogger.Debug(err)
			ret = append(ret, pt)
			continue
		}

		if drop { // drop
			continue
		}

		ptOpt.Time = tn

		if p, err := point.NewPoint(name, tags, fields, ptOpt); err != nil {
			plLogger.Error(err)
			stats.WriteScriptStats(script.Category(), script.NS(), script.Name(), 0, 0, 1, 0, err)
		} else {
			pt = p
		}

		ret = append(ret, pt)
	}

	return ret, nil
}

func getScript(category string, pt *point.Point, scriptMap map[string]string) (
	*plscript.PlScript, string, map[string]string, map[string]interface{}, *time.Time,
) {
	switch category {
	case datakit.RUM, datakit.Security, datakit.Tracing, datakit.Profiling:
		fields, err := pt.Fields()
		if err != nil {
			plLogger.Debug(err)
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
			return s, name, tags, fields, &ptTime
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
			plLogger.Errorf("Fields: %s", err)
			break
		}

		ptTime := pt.Time()
		tags := pt.Tags()
		return s, name, tags, fields, &ptTime
	}

	return nil, "", nil, nil, nil
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

	return scriptName + ".p", true
}

func joinName(name ...string) string {
	return strings.Join(name, "_")
}
