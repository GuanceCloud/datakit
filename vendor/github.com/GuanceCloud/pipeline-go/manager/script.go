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

	"github.com/GuanceCloud/pipeline-go/ptinput"
	"github.com/GuanceCloud/pipeline-go/ptinput/funcs"
	"github.com/GuanceCloud/pipeline-go/ptinput/plcache"
	"github.com/GuanceCloud/pipeline-go/ptinput/plmap"
	"github.com/GuanceCloud/pipeline-go/ptinput/ptwindow"
	"github.com/GuanceCloud/pipeline-go/stats"
)

type Option struct {
	MaxFieldValLen        int // deprecated
	DisableAddStatusField bool
	IgnoreStatus          []string
	ScriptMap             map[string]string
}

type PlScript struct {
	name   string // script name
	script string // script content
	ns     string // script 所属 namespace

	proc     *plruntime.Script
	plBuks   *plmap.AggBuckets
	ptWindow *ptwindow.WindowPool
	cache    *plcache.Cache

	tags     map[string]string
	updateTS int64

	category point.Category
}

func NewScripts(scripts, scriptTags map[string]string, ns string, cat point.Category,
	buks ...*plmap.AggBuckets,
) (map[string]*PlScript, map[string]error) {
	var plbuks *plmap.AggBuckets
	if len(buks) > 0 {
		plbuks = buks[0]
	}

	switch cat { //nolint:exhaustive
	case point.MetricDeprecated:
		cat = point.Metric
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
		}

		if plpt.GetStatusMapping() {
			ProcLoggingStatus(plpt, disable, ignore)
		}
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

func (script *PlScript) Category() point.Category {
	return script.category
}

func (script *PlScript) NS() string {
	return script.ns
}
