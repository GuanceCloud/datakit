// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package io

import (
	"fmt"
	"strings"

	"gitlab.jiagouyun.com/cloudcare-tools/cliutils/logger"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io/point"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/pipeline/script"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/pipeline/stats"
)

var plLogger = logger.DefaultSLogger("pipeline")

func runPl(category string, pts []*point.Point, opt *Option) (ret []*point.Point, retErr error) {
	defer func() {
		if err := recover(); err != nil {
			retErr = fmt.Errorf("run pl: %s", err)
		}
	}()

	var scriptMap map[string]string
	var plOpt *script.Option
	if opt != nil {
		scriptMap = opt.PlScript
		plOpt = opt.PlOption
	}
	ret = []*point.Point{}
	for _, pt := range pts {
		tags := pt.Tags()
		fields, err := pt.Fields()
		if err != nil {
			plLogger.Debug(err)
			continue
		}

		scriptName, ok := scriptName(category, pt.Name(), tags, fields, scriptMap)
		if !ok {
			ret = append(ret, pt)
			continue
		}

		script, ok := script.QueryScript(category, scriptName)
		if !ok { // script not found
			ret = append(ret, pt)
			continue
		}

		out, drop, err := script.Run(pt.Point.Name(), tags, fields, "message", pt.Point.Time(), plOpt)
		if err != nil {
			plLogger.Debug(err)
			ret = append(ret, pt)
			continue
		}

		if drop { // drop
			continue
		}

		ptOpt := &point.PointOption{
			DisableGlobalTags: true,
			Category:          category,
			Time:              out.Time,
		}

		if plOpt != nil {
			ptOpt.MaxFieldValueLen = plOpt.MaxFieldValLen
		}
		if p, err := point.NewPoint(out.Measurement, out.Tags, out.Fields, ptOpt); err != nil {
			plLogger.Error(err)
			stats.WriteScriptStats(script.Category(), script.NS(), script.Name(), 0, 0, 1, 0, err)
		} else {
			pt = p
		}
		ret = append(ret, pt)
	}

	return ret, nil
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
