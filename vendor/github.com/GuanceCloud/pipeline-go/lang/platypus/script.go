// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2025-present Guance, Inc.

// Package platypus use to parse platypus script
package platypus

import (
	"fmt"
	"time"

	"github.com/GuanceCloud/cliutils/point"
	plengine "github.com/GuanceCloud/platypus/pkg/engine"
	plruntime "github.com/GuanceCloud/platypus/pkg/engine/runtime"

	"github.com/GuanceCloud/pipeline-go/lang"
	"github.com/GuanceCloud/pipeline-go/ptinput"
	"github.com/GuanceCloud/pipeline-go/ptinput/funcs"
	"github.com/GuanceCloud/pipeline-go/ptinput/plcache"
	"github.com/GuanceCloud/pipeline-go/ptinput/plmap"
	"github.com/GuanceCloud/pipeline-go/ptinput/ptwindow"
	"github.com/GuanceCloud/pipeline-go/stats"
)

type Opt struct {
	Meta      map[string]string
	Cat       point.Category
	Namespace string

	Bucket   *plmap.AggBuckets
	PtWindow *ptwindow.WindowPool
	Cache    *plcache.Cache
}

type PlScript struct {
	name    string
	content string
	proc    *plruntime.Script

	opt Opt

	updateTS int64
}

// NewScripts parse platypus script
func NewScripts(scripts map[string]string, opts ...lang.Option) (
	map[string]*PlScript, map[string]error,
) {
	opt := lang.Opt{}
	for _, o := range opts {
		if o != nil {
			o(&opt)
		}
	}

	switch opt.Cat { //nolint:exhaustive
	case point.MetricDeprecated:
		opt.Cat = point.Metric
	}

	if !opt.CustomFnSet {
		opt.FnCall = funcs.FuncsMap
		opt.FnCheck = funcs.FuncsCheckMap
	}

	ret, retErr := plengine.ParseScript(scripts,
		opt.FnCall, opt.FnCheck)

	retScipt := map[string]*PlScript{}

	for name, ng := range ret {
		tags := map[string]string{
			"category":  opt.Cat.String(),
			"name":      name,
			"namespace": opt.Namespace,

			"lang": "platypus",
		}

		for k, v := range opt.Meta {
			if _, ok := tags[k]; !ok {
				tags[k] = v
			}
		}

		s := &PlScript{
			name:     name,
			content:  scripts[name],
			proc:     ng,
			updateTS: time.Now().UnixNano(),
			opt: Opt{
				Meta:      tags,
				Cat:       opt.Cat,
				Namespace: opt.Namespace,
			},
		}

		if opt.Bucket != nil {
			s.opt.Bucket = opt.Bucket()
		}
		if opt.PtWindow != nil {
			s.opt.PtWindow = opt.PtWindow()
		}
		if opt.Cache != nil {
			s.opt.Cache = opt.Cache()
		}

		retScipt[name] = s
	}

	return retScipt, retErr
}

func (script *PlScript) Engine() *plruntime.Script {
	return script.proc
}

func (script *PlScript) Run(plpt ptinput.PlInputPt, signal plruntime.Signal, opt *lang.LogOption,
) error {
	startTime := time.Now()
	if script.proc == nil {
		return fmt.Errorf("no script")
	}

	if plpt == nil {
		return fmt.Errorf("no data")
	}

	plpt.SetAggBuckets(script.opt.Bucket)
	plpt.SetCache(script.opt.Cache)
	plpt.SetPtWinPool(script.opt.PtWindow)

	err := script.proc.Run(plpt, signal)
	if err != nil {
		stats.WriteMetric(script.opt.Meta, 1, 0, 1, time.Since(startTime))
		return err
	}

	if script.opt.Cat == point.Logging {
		var disable bool
		var ignore []string

		if opt != nil {
			disable = opt.DisableAddStatusField
			ignore = opt.IgnoreStatus
		}

		if plpt.GetStatusMapping() {
			lang.ProcLoggingStatus(plpt, disable, ignore)
		}
	}

	if plpt.Dropped() {
		stats.WriteMetric(script.opt.Meta, 1, 1, 0, time.Since(startTime))
	} else {
		stats.WriteMetric(script.opt.Meta, 1, 0, 0, time.Since(startTime))
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
	return script.opt.Cat
}

func (script *PlScript) NS() string {
	return script.opt.Namespace
}

func (script *PlScript) Meta() map[string]string {
	return script.opt.Meta
}

func (script *PlScript) Content() string {
	return script.content
}

// Cleanup is used to clean up resources
func (script *PlScript) Cleanup() {
	if script.opt.Bucket != nil {
		script.opt.Bucket.StopAllBukScanner()
	}
	if script.opt.Cache != nil {
		script.opt.Cache.Stop()
	}
	if script.opt.PtWindow != nil {
		script.opt.PtWindow.Deprecated()
	}
}
