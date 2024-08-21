// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

// Package manager for managing pipeline scripts
package manager

import (
	"fmt"
	"time"

	"github.com/GuanceCloud/cliutils/point"
	plengine "github.com/GuanceCloud/platypus/pkg/engine"
	plruntime "github.com/GuanceCloud/platypus/pkg/engine/runtime"

	"github.com/GuanceCloud/cliutils/pipeline/ptinput"
	"github.com/GuanceCloud/cliutils/pipeline/ptinput/funcs"
	"github.com/GuanceCloud/cliutils/pipeline/ptinput/plcache"
	"github.com/GuanceCloud/cliutils/pipeline/ptinput/plmap"
	"github.com/GuanceCloud/cliutils/pipeline/ptinput/ptwindow"
	"github.com/GuanceCloud/cliutils/pipeline/stats"
)

type Option struct {
	MaxFieldValLen        int // deprecated
	DisableAddStatusField bool
	IgnoreStatus          []string
	ScriptMap             map[string]string
}

type PlScript struct {
	name     string // script name
	filePath string
	script   string // script content

	ns       string // script 所属 namespace
	category point.Category

	proc *plruntime.Script

	plBuks *plmap.AggBuckets

	ptWindow *ptwindow.WindowPool

	updateTS int64

	tags  map[string]string
	cache *plcache.Cache
}

func NewScripts(scripts, scriptPath, scriptTags map[string]string, ns string, cat point.Category,
	buks ...*plmap.AggBuckets,
) (map[string]*PlScript, map[string]error) {
	var plbuks *plmap.AggBuckets
	if len(buks) > 0 {
		plbuks = buks[0]
	}

	switch cat {
	case point.Metric:
	case point.MetricDeprecated:
		cat = point.Metric
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
			retErr[k] = fmt.Errorf("unsupported category: %s", cat)
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
		cache, _ := plcache.NewCache(time.Second, 100)
		ptWin := ptwindow.NewManager()

		sTags := map[string]string{
			"category":  cat.String(),
			"name":      name,
			"namespace": ns,

			"lang": "platypus",
		}

		for k, v := range scriptTags {
			if _, ok := sTags[k]; !ok {
				sTags[k] = v
			}
		}

		retScipt[name] = &PlScript{
			script:   scripts[name],
			name:     name,
			filePath: sPath,
			ns:       ns,
			category: cat,
			proc:     ng,
			updateTS: time.Now().UnixNano(),
			plBuks:   plbuks,
			tags:     sTags,
			cache:    cache,
			ptWindow: ptWin,
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
	plpt.SetCache(script.cache)
	plpt.SetPtWinPool(script.ptWindow)

	err := script.proc.Run(plpt, signal)
	if err != nil {
		stats.WriteMetric(script.tags, 1, 0, 1, time.Since(startTime))
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
		stats.WriteMetric(script.tags, 1, 1, 0, time.Since(startTime))
	} else {
		stats.WriteMetric(script.tags, 1, 0, 0, time.Since(startTime))
	}

	plpt.KeyTime2Time()

	for _, v := range plpt.GetSubPoint() {
		v.KeyTime2Time()
	}

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
