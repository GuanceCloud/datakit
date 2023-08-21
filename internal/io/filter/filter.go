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

	fp "github.com/GuanceCloud/cliutils/filter"
	"github.com/GuanceCloud/cliutils/logger"
	"github.com/GuanceCloud/cliutils/point"
	"github.com/GuanceCloud/cliutils/system/rtpanic"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/datakit"
	dkpt "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/io/point"
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

	conditions    map[string]fp.WhereConditions
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
func GetConds(filterArr []string) (fp.WhereConditions, error) {
	var conds fp.WhereConditions
	for _, v := range filterArr {
		cond, err := fp.GetConds(v)
		if err != nil {
			filterParseErrorVec.WithLabelValues(err.Error(), v).Set(float64(time.Now().Unix()))
			return nil, err
		}

		if cond == nil {
			return nil, fmt.Errorf("condition empty")
		}
		conds = append(conds, cond...)
	}
	return conds, nil
}

// CheckPointFiltered returns whether the point matches the fitler rule.
// If returns true means they are matched.
func CheckPointFiltered(conds fp.WhereConditions, category point.Category, pt *dkpt.Point) (bool, error) {
	return filtered(conds, NewTFData(category, pt)), nil
}

func filtered(conds fp.WhereConditions, data fp.KVs) bool {
	return conds.Eval(data)
}

func (f *filter) doFilter(category point.Category, pts []*dkpt.Point) ([]*dkpt.Point, int) {
	l.Debugf("doFilter: %+#v", f)

	start := time.Now()

	f.mtx.RLock()
	defer f.mtx.RUnlock()

	// "/v1/write/metric" => "metric"
	catStr := category.String()

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

func FilterPts(category point.Category, pts []*dkpt.Point) []*dkpt.Point {
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
		conditions:    map[string]fp.WhereConditions{},
		rawConditions: map[string]string{},

		puller: p,

		mtx: &sync.RWMutex{},

		tick:         time.NewTicker(pullInterval),
		pullInterval: pullInterval,
	}
}

func StartFilter(p IPuller) {
	l = logger.SLogger(packageName)

	fp.Init()

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
