package io

import (
	"fmt"

	"gitlab.jiagouyun.com/cloudcare-tools/cliutils/logger"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/pipeline/script"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/pipeline/stats"
)

var plLogger = logger.DefaultSLogger("pipeline")

func runPl(category string, pts []*Point, scriptMap map[string]string, plOpt *script.Option) (ret []*Point, retErr error) {
	defer func() {
		if err := recover(); err != nil {
			retErr = fmt.Errorf("run pl: %s", err)
		}
	}()

	if len(scriptMap) == 0 {
		return pts, nil
	}

	ret = []*Point{}

	for _, pt := range pts {
		scriptName, ok := scriptMap[pt.Name()]
		if !ok || scriptName == "-" { // skip
			ret = append(ret, pt)
			continue
		}

		if scriptName == "" {
			scriptName = pt.Name() + ".p"
		}

		script, ok := script.QueryScript(category, scriptName)
		if !ok { // script not found
			ret = append(ret, pt)
			plLogger.Warnf("category: %s, name: %s : script not found", category, scriptName)
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

		out, drop, err := script.Run(pt.Name(), pt.Tags(), fields, cntKey, pt.Time(), plOpt)
		if err != nil {
			plLogger.Debug(err)
			ret = append(ret, pt)
			continue
		}

		if drop { // drop
			continue
		}

		ptOpt := &PointOption{
			Category: category,
			Time:     out.Time,
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
