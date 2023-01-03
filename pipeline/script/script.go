// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

// Package script for managing pipeline scripts
package script

import (
	"fmt"
	"time"

	plengine "github.com/GuanceCloud/platypus/pkg/engine"
	plruntime "github.com/GuanceCloud/platypus/pkg/engine/runtime"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/pipeline/ptinput"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/pipeline/ptinput/funcs"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/pipeline/stats"
)

type Option struct {
	MaxFieldValLen        int // deprecated
	DisableAddStatusField bool
	IgnoreStatus          []string
}

type PlScript struct {
	name     string // script name
	filePath string
	script   string // script content

	ns       string // script 所属 namespace
	category string

	proc *plruntime.Script

	updateTS int64
}

func NewScripts(scripts map[string]string, scriptPath map[string]string, ns, category string) (map[string]*PlScript, map[string]error) {
	switch category {
	case datakit.Metric:
	case datakit.MetricDeprecated:
		category = datakit.Metric
	case datakit.Network:
	case datakit.KeyEvent:
	case datakit.Object:
	case datakit.CustomObject:
	case datakit.Tracing:
	case datakit.RUM:
	case datakit.Security:
	case datakit.Logging:
	case datakit.Profiling:
	default:
		retErr := map[string]error{}
		for k := range scripts {
			retErr[k] = fmt.Errorf("unsupported category: %s", category)
		}
		return nil, retErr
	}
	ret, retErr := plengine.ParseScript(scripts, funcs.FuncsMap, funcs.FuncsCheckMap)

	retScipt := map[string]*PlScript{}

	for name, ng := range ret {
		var sPath string
		if len(scriptPath) > 0 {
			sPath = scriptPath[name]
		}

		retScipt[name] = &PlScript{
			script:   scripts[name],
			name:     name,
			filePath: sPath,
			ns:       ns,
			category: category,
			proc:     ng,
			updateTS: time.Now().UnixNano(),
		}
	}

	return retScipt, retErr
}

func (script *PlScript) Engine() *plruntime.Script {
	return script.proc
}

func (script *PlScript) Run(plpt *ptinput.Point, signal plruntime.Signal, opt *Option,
) error {
	startTime := time.Now()
	if script.proc == nil {
		return fmt.Errorf("no script")
	}

	err := plengine.RunScriptWithRMapIn(script.proc, plpt, signal)
	if err != nil {
		stats.WriteScriptStats(script.category, script.ns, script.name, 1, 0, 1, int64(time.Since(startTime)), err)
		return err
	}

	switch script.category {
	case datakit.Logging:
		var disable bool
		var ignore []string

		if opt != nil {
			disable = opt.DisableAddStatusField
			ignore = opt.IgnoreStatus
			// spiltLen = opt.MaxFieldValLen
		}

		plpt.Tags, plpt.Fields, plpt.Drop = ProcLoggingStatus(plpt.Tags, plpt.Fields, plpt.Drop, disable, ignore)
	default:
	}

	if plpt.Drop {
		stats.WriteScriptStats(script.category, script.ns, script.name, 1, 1, 0, int64(time.Since(startTime)), nil)
	} else {
		stats.WriteScriptStats(script.category, script.ns, script.name, 1, 0, 0, int64(time.Since(startTime)), nil)
	}

	plpt.KeyTime2Time()

	return nil
}

func (script *PlScript) Name() string {
	return script.name
}

func (script PlScript) FilePath() string {
	return script.filePath
}

func (script *PlScript) Category() string {
	return script.category
}

func (script *PlScript) NS() string {
	return script.ns
}
