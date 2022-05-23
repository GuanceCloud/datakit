// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

// Package script used to create pipeline script
package script

import (
	"fmt"
	"time"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/pipeline/funcs"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/pipeline/parser"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/pipeline/stats"
)

// ES value can be at most 32766 bytes long.
const maxFieldsLength = 32766

type Option struct {
	MaxFieldValLen        int
	DisableAddStatusField bool
	IgnoreStatus          []string
}

type PlScript struct {
	name   string // script name
	script string // script content

	ns       string // script 所属 namespace
	category string

	ng *parser.Engine

	updateTS int64
}

func NewScript(name, script, ns, category string) (*PlScript, error) {
	switch category {
	case datakit.Metric, datakit.MetricDeprecated:
	case datakit.Network:
	case datakit.KeyEvent:
	case datakit.Object:
	case datakit.CustomObject:
	case datakit.Tracing:
	case datakit.RUM:
	case datakit.Security:
	case datakit.HeartBeat:
	case datakit.Logging:
	default:
		return nil, fmt.Errorf("unsupported category: %s", category)
	}
	ng, err := parser.NewEngine(script, funcs.FuncsMap, funcs.FuncsCheckMap, false)
	if err != nil {
		return nil, err
	}

	return &PlScript{
		script:   script,
		name:     name,
		ns:       ns,
		category: category,
		ng:       ng,
		updateTS: time.Now().UnixNano(),
	}, nil
}

func (script *PlScript) Engine() *parser.Engine {
	return script.ng
}

func (script *PlScript) Run(measurement string, tags map[string]string, fields map[string]interface{},
	contentKey string, t time.Time, opt *Option) (*parser.Output, bool, error) {
	startTime := time.Now()
	if script == nil || script.ng == nil {
		return nil, false, fmt.Errorf("no engine")
	}
	out, err := script.ng.Run(measurement, tags, fields, contentKey, t)
	if err != nil {
		stats.WriteScriptStats(script.category, script.ns, script.name, 1, 0, 1, int64(time.Since(startTime)), err)
		return nil, false, err
	}

	switch script.category {
	case datakit.Metric, datakit.MetricDeprecated:
	case datakit.Network:
	case datakit.KeyEvent:
	case datakit.Object:
	case datakit.CustomObject:
	case datakit.Tracing:
	case datakit.RUM:
	case datakit.Security:
	case datakit.HeartBeat:
	case datakit.Logging:
		var disable bool
		var ignore []string

		var spiltLen int

		if opt != nil {
			disable = opt.DisableAddStatusField
			ignore = opt.IgnoreStatus
			spiltLen = opt.MaxFieldValLen
		}
		if spiltLen <= 0 { // 当初始化 task 时没有注入最大长度则使用默认值
			spiltLen = maxFieldsLength
		}
		out = ProcLoggingStatus(out, disable, ignore, spiltLen)
	default:
		return nil, false, fmt.Errorf("unsupported category: %s", script.category)
	}

	if out.Drop {
		stats.WriteScriptStats(script.category, script.ns, script.name, 1, 1, 0, int64(time.Since(startTime)), nil)
	} else {
		stats.WriteScriptStats(script.category, script.ns, script.name, 1, 0, 0, int64(time.Since(startTime)), nil)
	}

	return out, out.Drop, nil
}

func (script *PlScript) Name() string {
	return script.name
}

func (script *PlScript) Category() string {
	return script.category
}

func (script *PlScript) NS() string {
	return script.ns
}
