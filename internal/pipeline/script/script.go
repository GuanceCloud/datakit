// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

// Package script for managing pipeline scripts
package script

import (
	"fmt"
	"time"

	"github.com/GuanceCloud/cliutils/point"
	plengine "github.com/GuanceCloud/platypus/pkg/engine"
	plruntime "github.com/GuanceCloud/platypus/pkg/engine/runtime"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/pipeline/plmap"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/pipeline/ptinput"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/pipeline/ptinput/funcs"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/pipeline/stats"
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
	category point.Category

	proc *plruntime.Script

	plBuks *plmap.AggBuckets

	updateTS int64
}

func CategoryList() (map[point.Category]struct{}, map[point.Category]struct{}) {
	return map[point.Category]struct{}{
			point.Metric:       {},
			point.Network:      {},
			point.KeyEvent:     {},
			point.Object:       {},
			point.CustomObject: {},
			point.Logging:      {},
			point.Tracing:      {},
			point.RUM:          {},
			point.Security:     {},
			point.Profiling:    {},
		}, map[point.Category]struct{}{
			point.MetricDeprecated: {},
		}
}

func NewScripts(scripts map[string]string, scriptPath map[string]string, ns string, category point.Category,
	buks ...*plmap.AggBuckets,
) (map[string]*PlScript, map[string]error) {
	var plbuks *plmap.AggBuckets
	if len(buks) > 0 {
		plbuks = buks[0]
	}

	switch category {
	case point.Metric:
	case point.MetricDeprecated:
		category = point.Metric
	case point.Network:
	case point.KeyEvent:
	case point.Object:
	case point.CustomObject:
	case point.Tracing:
	case point.RUM:
	case point.Security:
	case point.Logging:
	case point.Profiling:
	case point.UnknownCategory, point.DynamicDWCategory:
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
			plBuks:   plbuks,
		}
	}

	return retScipt, retErr
}

func (script *PlScript) Engine() *plruntime.Script {
	return script.proc
}

func (script *PlScript) SetAggBuks(buks *plmap.AggBuckets) {
	script.plBuks = buks
}

func (script *PlScript) Run(plpt ptinput.PlInputPt, signal plruntime.Signal, opt *Option,
) error {
	startTime := time.Now()
	if script.proc == nil {
		return fmt.Errorf("no script")
	}

	if plpt == nil {
		return fmt.Errorf("no data")
	}

	plpt.SetAggBuckets(script.plBuks)

	err := plengine.RunScriptWithRMapIn(script.proc, plpt, signal)
	if err != nil {
		stats.WriteScriptStats(script.category, script.ns, script.name, 1, 0, 1, int64(time.Since(startTime)), err)
		return err
	}

	if script.category == point.Logging {
		var disable bool
		var ignore []string

		if opt != nil {
			disable = opt.DisableAddStatusField
			ignore = opt.IgnoreStatus
			// spiltLen = opt.MaxFieldValLen
		}

		ProcLoggingStatus(plpt, disable, ignore)
	}

	if plpt.Dropped() {
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

func (script *PlScript) Category() point.Category {
	return script.category
}

func (script *PlScript) NS() string {
	return script.ns
}
