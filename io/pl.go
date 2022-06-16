// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package io

import (
	"fmt"

	"gitlab.jiagouyun.com/cloudcare-tools/cliutils/logger"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/pipeline/script"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/pipeline/stats"
)

var plLogger = logger.DefaultSLogger("pipeline")

func runPl(category string, pts []*Point, opt *Option) (ret []*Point, retErr error) {
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
	ret = []*Point{}
	for _, pt := range pts {
		scriptName, ok := scriptName(category, pt, scriptMap)
		if !ok {
			ret = append(ret, pt)
			continue
		}

		script, ok := script.QueryScript(category, scriptName)
		if !ok { // script not found
			ret = append(ret, pt)
			continue
		}

		fields, err := pt.Fields()
		if err != nil {
			plLogger.Debug(err)
			continue
		}
		// run pl

		cntKey := ""
		switch category {
		case datakit.Logging:
			if _, ok := fields["message"]; ok {
				cntKey = "message"
				break
			}
			if _, ok := fields["message@json"]; ok {
				cntKey = "message@json"
			}
		}

		out, drop, err := script.Run(pt.Point.Name(), pt.Point.Tags(), fields, cntKey, pt.Point.Time(), plOpt)
		if err != nil {
			plLogger.Debug(err)
			ret = append(ret, pt)
			continue
		}

		if drop { // drop
			continue
		}

		ptOpt := &PointOption{
			DisableGlobalTags: true,
			Category:          category,
			Time:              out.Time,
		}

		if plOpt != nil {
			ptOpt.MaxFieldValueLen = plOpt.MaxFieldValLen
		}
		if p, err := NewPoint(out.Measurement, out.Tags, out.Fields, ptOpt); err != nil {
			plLogger.Error(err)
			stats.WriteScriptStats(script.Category(), script.NS(), script.Name(), 0, 0, 1, 0, err)
		} else {
			pt = p
		}
		ret = append(ret, pt)
	}

	return ret, nil
}

func scriptName(category string, pt *Point, scriptMap map[string]string) (string, bool) {
	if pt == nil {
		return "", false
	}
	var scriptName string
	switch category {
	case datakit.Tracing:
		svc, ok := pt.Point.Tags()["service"]
		if ok {
			scriptName = scriptMap[svc]
			if scriptName == "" {
				scriptName = svc + ".p"
			}
		} else {
			return "", false
		}
	default:
		scriptName = scriptMap[pt.Point.Name()]
		if scriptName == "" {
			scriptName = pt.Point.Name() + ".p"
		}
	}
	if scriptName == "-" {
		return "", false
	}
	return scriptName, true
}
