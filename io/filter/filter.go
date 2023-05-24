// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

// Package filter contains filter logic.
package filter

import (
	"fmt"
	"sync"
	"time"

	"github.com/GuanceCloud/cliutils/logger"
	"github.com/GuanceCloud/cliutils/point"
	"github.com/GuanceCloud/cliutils/system/rtpanic"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io/parser"
	dkpt "gitlab.jiagouyun.com/cloudcare-tools/datakit/io/point"
)

const packageName = "filter"

var (
	l             = logger.DefaultSLogger(packageName)
	pullInterval  = time.Second * 30
	defaultFilter *filter
)

type IPuller interface {
	Pull(what string) ([]byte, error)
}

type FilterConditions []string

type filter struct {
	// where filters/conditions come from(local/remote)
	source string

	conditions    map[string]parser.WhereConditions
	rawConditions map[string]string

	puller IPuller
	md5    string

	dumpDir string

	// Mutex to R/W on rules: rules are updated(Write) from remote center, or
	// applied to(Read) filter points
	mtx *sync.RWMutex

	pullInterval time.Duration
	tick         *time.Ticker
}

func (f *filter) pull(what string) {
	var (
		body []byte
		err  error
	)

	switch f.source {
	case sourceLocal:
		body, err = f.puller.Pull("")
		if err != nil {
			l.Warnf("local filter: %s, ignored", err)
			return // ignore
		}
	default: // go down

		body, err = f.remotePull(what)
		if err != nil {
			l.Warnf("remote pull: %s, ignored", err)
			return
		}
	}

	if err := f.refresh(body); err != nil {
		l.Warnf("refresh filters failed: %s, ignored", err)
	}
}

// GetConds returns Filter's Parser Conditions and error.
func GetConds(filterArr []string) (parser.WhereConditions, error) {
	var conds parser.WhereConditions
	for _, v := range filterArr {
		cond := parser.GetConds(v)
		if cond == nil {
			return nil, fmt.Errorf("condition empty")
		}
		conds = append(conds, cond...)
	}
	return conds, nil
}

// CheckPointFiltered returns whether the point matches the fitler rule.
// If returns true means they are matched.
func CheckPointFiltered(conds parser.WhereConditions, category string, pt *dkpt.Point) (bool, error) {
	tags := pt.Tags()
	fields, err := pt.Fields()
	if err != nil {
		return false, err
	}

	// Before checks, should adjust tags under some conditions.
	// Must stay the same 'switch' logic with kodo project function named 'getSourceValue' in source file apis/esFields.go.
	switch category {
	case datakit.Logging, datakit.Network, datakit.KeyEvent, datakit.RUM:
		tags["source"] = pt.Point.Name() // set measurement name as tag `source'
	case datakit.Tracing, datakit.Security, datakit.Profiling:
		// using measurement name as tag `service'.
	case datakit.Metric:
		tags["measurement"] = pt.Point.Name() // set measurement name as tag `measurement'
	case datakit.Object, datakit.CustomObject:
		tags["class"] = pt.Point.Name() // set measurement name as tag `class'
	default:
		l.Warnf("unsupport category: %s", category)
		return false, fmt.Errorf("unsupport category: %s", category)
	}

	return filtered(conds, tags, fields), nil
}

func filtered(conds parser.WhereConditions, tags map[string]string, fields map[string]interface{}) bool {
	return conds.Eval(tags, fields)
}

func (f *filter) doFilter(category string, pts []*dkpt.Point) ([]*dkpt.Point, int) {
	l.Debugf("doFilter: %+#v", f)

	start := time.Now()

	f.mtx.RLock()
	defer f.mtx.RUnlock()

	// "/v1/write/metric" => "metric"
	catStr := point.CatURL(category).String()

	conds, ok := f.conditions[catStr]
	if !ok || conds == nil {
		l.Debugf("no condition filter for %s", catStr)
		return pts, 0
	}

	var after []*dkpt.Point

	defer func() {
		filterPtsVec.WithLabelValues(catStr, f.rawConditions[catStr], f.source).Add(float64(len(pts)))
		filterDroppedPtsVec.WithLabelValues(catStr, f.rawConditions[catStr], f.source).Add(float64(len(pts) - len(after)))
		filterLatencyVec.WithLabelValues(catStr, f.rawConditions[catStr], f.source).Observe(float64(time.Since(start)) / float64(time.Second))
	}()

	for _, pt := range pts {
		isFiltered, err := CheckPointFiltered(conds, category, pt)
		if err != nil {
			l.Errorf("pt.Fields: %s, ignored", err.Error())
			continue // filter it!
		}
		if !isFiltered { // Pick those points that not matched filter rules.
			after = append(after, pt)
		} else if datakit.LogSinkDetail {
			l.Infof("(sink_detail) filtered point: (%s) (%s)", category, pt.String())
		}
	}

	return after, len(conds)
}

func FilterPts(category string, pts []*dkpt.Point) []*dkpt.Point {
	if defaultFilter == nil { // during testing, defaultFilter not initialized
		return pts
	}

	after, _ := defaultFilter.doFilter(category, pts)
	return after
}

func (f *filter) start() {
	defer f.tick.Stop()

	what := "filters=true"

	// Try pull rules ASAP.
	f.pull(what)

	for {
		select {
		case <-f.tick.C:
			l.Debugf("try pull remote filters...")
			f.pull(what)

		case <-datakit.Exit.Wait():
			l.Info("log filter exits")
			return
		}
	}
}

func newFilter(p IPuller) *filter {
	return &filter{
		conditions:    map[string]parser.WhereConditions{},
		rawConditions: map[string]string{},

		puller: p,

		mtx: &sync.RWMutex{},

		tick:         time.NewTicker(pullInterval),
		pullInterval: pullInterval,
	}
}

func StartFilter(p IPuller) {
	l = logger.SLogger(packageName)

	parser.Init()

	var f rtpanic.RecoverCallback

	defaultFilter = newFilter(p)
	defaultFilter.dumpDir = datakit.DataDir

	switch x := p.(type) {
	case *localFilter:
		if len(x.filters) > 0 {
			defaultFilter.source = sourceLocal
		} else {
			defaultFilter.source = sourceRemote
		}
	default:
		defaultFilter.source = sourceRemote
	}

	f = func(trace []byte, err error) {
		defer rtpanic.Recover(f, nil)
		if trace != nil {
			l.Warnf("filter panic: %s: %s", err, string(trace))
		}

		defaultFilter.start()
	}

	f(nil, nil)
}

//nolint:gochecknoinits
func init() {
	setupMetrics()
}
