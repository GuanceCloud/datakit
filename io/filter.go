// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package io

import (

	// nolint:gosec
	"crypto/md5"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"sync"
	"time"

	"gitlab.jiagouyun.com/cloudcare-tools/cliutils/logger"
	"gitlab.jiagouyun.com/cloudcare-tools/cliutils/system/rtpanic"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io/parser"
)

var (
	l         = logger.DefaultSLogger("filter")
	isStarted = false
)

func newFilter(dw IDataway) *filter {
	pullInterval := time.Second * 10

	return &filter{
		conditions: map[string]parser.WhereConditions{},
		dw:         dw,

		RWMutex: sync.RWMutex{},

		metricCh: make(chan *filterMetric, 32),
		qch:      make(chan *qstats),

		tick:         time.NewTicker(pullInterval),
		pullInterval: pullInterval,

		// stats key is category + service/source
		stats: &FilterStats{
			RuleStats: map[string]*ruleStat{},
		},
	}
}

var defaultFilter = newFilter(&datawayImpl{})

type IDataway interface {
	Pull() ([]byte, error)
}

type datawayImpl struct{}

func (dw *datawayImpl) Pull() ([]byte, error) {
	if len(defaultIO.conf.Filters) != 0 {
		// read local filters
		return json.Marshal(&filterPull{Filters: defaultIO.conf.Filters, PullInterval: time.Second * 10})
	} else {
		// pull filters remotely
		return defaultIO.dw.DatakitPull("filters=true")
	}
}

type filter struct {
	conditions map[string]parser.WhereConditions
	dw         IDataway
	md5        string
	sync.RWMutex

	metricCh     chan *filterMetric
	qch          chan *qstats
	pullInterval time.Duration
	tick         *time.Ticker
	stats        *FilterStats
}

type filterPull struct {
	Filters map[string][]string `json:"filters"`
	// other fields ignored
	PullInterval time.Duration `json:"pull_interval"`
}

func dump(rules []byte, dir string) error {
	return ioutil.WriteFile(filepath.Join(dir, ".pull"), rules, os.ModePerm)
}

func (f *filter) pull() {
	start := time.Now()

	f.stats.PullCount++

	if len(defaultIO.conf.Filters) != 0 {
		f.stats.RuleSource = "datakit.conf"
	} else {
		f.stats.RuleSource = "remote"
	}

	body, err := f.dw.Pull()
	if err != nil {
		l.Errorf("dataway Pull: %s", err)
		f.stats.PullFailed++
		f.stats.LastErr = err.Error()
		f.stats.LastErrTime = time.Now()
		return
	}

	l.Debugf("filter condition body: %s", string(body))

	cost := time.Since(start)
	f.stats.PullCost += cost
	f.stats.PullCostAvg = f.stats.PullCost / time.Duration(f.stats.PullCount)
	if cost > f.stats.PullCostMax {
		f.stats.PullCostMax = cost
	}

	bodymd5 := fmt.Sprintf("%x", md5.Sum(body)) //nolint:gosec
	if bodymd5 != f.md5 {                       // try update conditions
		var fp filterPull
		if err := json.Unmarshal(body, &fp); err != nil {
			l.Error("json.Unmarshal: %s", err)
			f.stats.LastErr = err.Error()
			f.stats.LastErrTime = time.Now()
			return
		}

		f.stats.LastUpdate = start
		f.RWMutex.Lock()
		defer f.RWMutex.Unlock()

		if fp.PullInterval > 0 && f.pullInterval != fp.PullInterval {
			f.pullInterval = fp.PullInterval
			f.stats.PullInterval = fp.PullInterval
			f.tick.Reset(f.pullInterval)
		}

		f.md5 = bodymd5
		// clear old conditions: we update all conditions if any changed(new/delete
		// conditons or update old conditions)
		f.conditions = map[string]parser.WhereConditions{}
		for k, v := range fp.Filters {
			for _, condition := range v {
				f.conditions[k] = append(f.conditions[k], parser.GetConds(condition)...)
			}
		}

		if err := dump(body, datakit.DataDir); err != nil {
			l.Warnf("dump: %s, ignored", err)
		}
	}
}

func (f *filter) filterLogging(cond parser.WhereConditions, pts []*Point) []*Point {
	if cond == nil {
		l.Debugf("no condition filter")
		return pts
	}

	var after []*Point
	for _, pt := range pts {
		tags := pt.Point.Tags()
		fields, err := pt.Point.Fields()
		if err != nil {
			continue // filter it!
		}

		tags["source"] = pt.Point.Name() // set measurement name as tag `source'
		if !filtered(cond, tags, fields) {
			after = append(after, pt)
		}
	}

	return after
}

func (f *filter) filterMetric(cond parser.WhereConditions, pts []*Point) []*Point {
	if cond == nil {
		l.Debugf("no condition filter for metric")
		return pts
	}

	var after []*Point

	for _, pt := range pts {
		tags := pt.Point.Tags()
		fields, err := pt.Point.Fields()
		if err != nil {
			l.Errorf("pt.Fields: %s, ignored", err.Error())
			continue // filter it!
		}

		tags["measurement"] = pt.Point.Name() // set measurement name as tag `measurement'

		if !filtered(cond, tags, fields) {
			after = append(after, pt)
		}
	}

	return after
}

func (f *filter) filterObject(cond parser.WhereConditions, pts []*Point) []*Point {
	if cond == nil {
		l.Debugf("no condition filter for object")
		return pts
	}

	var after []*Point

	for _, pt := range pts {
		tags := pt.Point.Tags()
		fields, err := pt.Point.Fields()
		if err != nil {
			l.Errorf("pt.Fields: %s, ignored", err.Error())
			continue // filter it!
		}

		tags["class"] = pt.Point.Name() // set measurement name as tag `class'

		if !filtered(cond, tags, fields) {
			after = append(after, pt)
		}
	}

	return after
}

func (f *filter) filterTracing(cond parser.WhereConditions, pts []*Point) []*Point {
	if cond == nil {
		l.Debugf("no condition filter for tracing")
		return pts
	}

	var after []*Point

	for _, pt := range pts {
		tags := pt.Point.Tags()
		fields, err := pt.Point.Fields()
		if err != nil {
			continue // filter it!
		}

		if !filtered(cond, tags, fields) {
			after = append(after, pt)
		}
	}

	return after
}

func filtered(conds parser.WhereConditions, tags map[string]string, fields map[string]interface{}) bool {
	return conds.Eval(tags, fields)
}

func (f *filter) doFilter(category string, pts []*Point) ([]*Point, int) {
	switch category {
	case datakit.Logging:
		f.RWMutex.RLock()
		defer f.RWMutex.RUnlock()
		return f.filterLogging(f.conditions["logging"], pts), len(f.conditions["logging"])

	case datakit.Tracing:
		f.RWMutex.RLock()
		defer f.RWMutex.RUnlock()
		return f.filterTracing(f.conditions["tracing"], pts), len(f.conditions["tracing"])

	case datakit.Metric:
		f.RWMutex.RLock()
		defer f.RWMutex.RUnlock()
		return f.filterMetric(f.conditions["metric"], pts), len(f.conditions["metric"])

	case datakit.Object:
		f.RWMutex.RLock()
		defer f.RWMutex.RUnlock()
		return f.filterObject(f.conditions["object"], pts), len(f.conditions["object"])

	default: // TODO: not implemented
		l.Warn("unsupport category: %s", category)
		return pts, 0
	}
}

func filterPts(category string, pts []*Point) []*Point {
	start := time.Now()
	after, condCount := defaultFilter.doFilter(category, pts)
	cost := time.Since(start)

	l.Debugf("%s/pts: %d, after: %d", category, len(pts), len(after))

	// report metrics
	fm := &filterMetric{
		key:        category,
		points:     len(pts),
		filtered:   len(pts) - len(after),
		cost:       cost,
		conditions: condCount,
	}
	select {
	case defaultFilter.metricCh <- fm:
	default: // unblocking
		l.Warnf("feed filter metrics failed, ignored: %+#v", fm)
	}

	return after
}

func GetFilterStats() *FilterStats {
	// return nil when not started
	if !isStarted {
		return nil
	}

	q := &qstats{
		ch: make(chan *FilterStats),
	}

	defer close(q.ch)

	tick := time.NewTicker(time.Second * 3)
	defer tick.Stop()

	defaultFilter.qch <- q
	select {
	case s := <-q.ch:
		return s
	case <-tick.C:
		return nil
	}
}

type ruleStat struct {
	Total        int64         `json:"total"`
	Filtered     int64         `json:"filtered"`
	Cost         time.Duration `json:"cost"`
	CostPerPoint time.Duration `json:"cost_per_point"`
	Conditions   int           `json:"conditions"`
}

type FilterStats struct {
	RuleStats map[string]*ruleStat `json:"rule_stats"`

	PullCount    int           `json:"pull_count"`
	PullInterval time.Duration `json:"pull_interval"`
	PullFailed   int           `json:"pull_failed"`

	RuleSource  string        `json:"rule_source"`
	PullCost    time.Duration `json:"pull_cost"`
	PullCostAvg time.Duration `json:"pull_cost_avg"`
	PullCostMax time.Duration `json:"pull_cost_max"`

	LastUpdate  time.Time `json:"last_update"`
	LastErr     string    `json:"last_err"`
	LastErrTime time.Time `json:"last_err_time"`
}

type filterMetric struct {
	key              string
	points, filtered int
	cost             time.Duration
	conditions       int
}

type qstats struct {
	ch chan *FilterStats
}

func copyStats(x *FilterStats) *FilterStats {
	y := &FilterStats{
		RuleStats: map[string]*ruleStat{},

		RuleSource:   x.RuleSource,
		PullInterval: x.PullInterval,
		PullCount:    x.PullCount,
		PullFailed:   x.PullFailed,
		PullCost:     x.PullCost,
		PullCostAvg:  x.PullCostAvg,
		PullCostMax:  x.PullCostMax,
		LastUpdate:   x.LastUpdate,
		LastErr:      x.LastErr,
		LastErrTime:  x.LastErrTime,
	}

	for k, v := range x.RuleStats {
		rs := &ruleStat{
			Total:        v.Total,
			Filtered:     v.Filtered,
			Cost:         v.Cost,
			CostPerPoint: v.CostPerPoint,
			Conditions:   v.Conditions,
		}
		y.RuleStats[k] = rs
	}
	return y
}

func (f *filter) start() {
	ruleStats := defaultFilter.stats.RuleStats
	// first pull: get filter condition ASAP
	defaultFilter.pull()
	defer defaultFilter.tick.Stop()

	for {
		select {
		case <-defaultFilter.tick.C:
			defaultFilter.pull()

		case m := <-defaultFilter.metricCh:
			v, ok := ruleStats[m.key]
			if !ok {
				v = &ruleStat{}
				ruleStats[m.key] = v
			}
			v.Total += int64(m.points)
			v.Filtered += int64(m.filtered)
			v.Cost += m.cost
			v.CostPerPoint = v.Cost / time.Duration(v.Total)
			v.Conditions = m.conditions

		case q := <-defaultFilter.qch:
			select {
			case <-q.ch:
			case q.ch <- copyStats(defaultFilter.stats):
			default: // unblocking
			}

		case <-datakit.Exit.Wait():
			log.Info("log filter exits")
			return
		}
	}
}

func StartFilter() {
	l = logger.SLogger("filter")
	if len(defaultIO.conf.Filters) == 0 && defaultIO.dw == nil {
		l.Warnf("filter not started: neither dataway nor filter conf set!")
		return
	}
	isStarted = true
	parser.Init()

	var f rtpanic.RecoverCallback

	f = func(trace []byte, err error) {
		defer rtpanic.Recover(f, nil)
		if trace != nil {
			l.Warnf("filter panic: %s: %s", err, string(trace))
		}

		defaultFilter.start()
	}

	f(nil, nil)
}
