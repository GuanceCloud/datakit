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
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io/point"
)

var (
	l            = logger.DefaultSLogger("filter")
	isStarted    = false
	pullInterval = time.Second * 30
)

func newFilter(dw IDataway) *filter {
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
		return json.Marshal(&filterPull{Filters: defaultIO.conf.Filters, PullInterval: pullInterval})
	} else {
		// pull filters remotely
		return defaultIO.dw.DatakitPull("filters=true")
	}
}

type filter struct {
	conditions map[string]parser.WhereConditions
	dw         IDataway
	md5        string

	// Mutex to R/W on rules: rules are updated(Write) from remote center, or
	// applied to(Read) filter points
	sync.RWMutex

	metricCh chan *filterMetric
	qch      chan *qstats

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

func (f *filter) update(body []byte, dumpdir string) error {
	bodymd5 := fmt.Sprintf("%x", md5.Sum(body)) //nolint:gosec
	if bodymd5 == f.md5 {
		return nil
	}

	// try update conditions
	var fp filterPull
	if err := json.Unmarshal(body, &fp); err != nil {
		l.Error("json.Unmarshal: %s", err)
		f.stats.LastErr = err.Error()
		f.stats.LastErrTime = time.Now()
		return err
	}

	if fp.PullInterval > 0 && f.pullInterval != fp.PullInterval {
		l.Infof("set pull interval from %s to %s", f.pullInterval, fp.PullInterval)
		f.pullInterval = fp.PullInterval
		f.stats.PullInterval = fp.PullInterval
		f.tick.Reset(f.pullInterval)
	}

	f.md5 = bodymd5
	// Clear old conditions: we update all conditions if any changed(new/delete
	// conditons or update old conditions)
	f.conditions = map[string]parser.WhereConditions{}
	for k, v := range fp.Filters {
		for _, condition := range v {
			f.conditions[k] = append(f.conditions[k], parser.GetConds(condition)...)
		}
	}

	if err := dump(body, dumpdir); err != nil {
		l.Warnf("dump: %s, ignored", err)
	}
	return nil
}

func (f *filter) pull() {
	start := time.Now()

	body, err := f.dw.Pull()
	if err != nil {
		l.Errorf("dataway Pull: %s", err)

		// keep mutex away from HTTP request
		f.RWMutex.Lock()
		defer f.RWMutex.Unlock()
		f.stats.PullFailed++
		f.stats.LastUpdate = start
		f.stats.LastErr = err.Error()
		f.stats.LastErrTime = time.Now()
		return
	}

	l.Debugf("filter condition body: %s", string(body))
	cost := time.Since(start)

	f.RWMutex.Lock()
	defer f.RWMutex.Unlock()

	f.stats.PullCount++
	f.stats.LastUpdate = start
	f.stats.PullCost += cost
	f.stats.PullCostAvg = f.stats.PullCost / time.Duration(f.stats.PullCount)
	if cost > f.stats.PullCostMax {
		f.stats.PullCostMax = cost
	}

	if err := f.update(body, datakit.DataDir); err != nil {
		l.Warnf("update filters failed: %s, ignored", err)
	}
}

func (f *filter) filterLogging(cond parser.WhereConditions, pts []*point.Point) []*point.Point {
	if cond == nil {
		l.Debugf("no condition filter")
		return pts
	}

	var after []*point.Point
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

func (f *filter) filterMetric(cond parser.WhereConditions, pts []*point.Point) []*point.Point {
	if cond == nil {
		l.Debugf("no condition filter for metric")
		return pts
	}

	var after []*point.Point

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

func (f *filter) filterObject(cond parser.WhereConditions, pts []*point.Point) []*point.Point {
	if cond == nil {
		l.Debugf("no condition filter for object")
		return pts
	}

	var after []*point.Point

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

// using measurement name as tag `service'.
func (f *filter) filterTracing(cond parser.WhereConditions, pts []*point.Point) []*point.Point {
	if cond == nil {
		l.Debugf("no condition filter for tracing")
		return pts
	}

	var after []*point.Point

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

func (f *filter) filterNetwork(cond parser.WhereConditions, pts []*point.Point) []*point.Point {
	if cond == nil {
		l.Debugf("no condition filter for network")
		return pts
	}

	var after []*point.Point

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

func (f *filter) filterKeyEvent(cond parser.WhereConditions, pts []*point.Point) []*point.Point {
	if cond == nil {
		l.Debugf("no condition filter for key event")
		return pts
	}

	var after []*point.Point

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

func (f *filter) filterCustomObject(cond parser.WhereConditions, pts []*point.Point) []*point.Point {
	if cond == nil {
		l.Debugf("no condition filter for custom object")
		return pts
	}

	var after []*point.Point

	for _, pt := range pts {
		tags := pt.Point.Tags()
		fields, err := pt.Point.Fields()
		if err != nil {
			continue // filter it!
		}

		tags["class"] = pt.Point.Name() // set measurement name as tag `class'

		if !filtered(cond, tags, fields) {
			after = append(after, pt)
		}
	}

	return after
}

func (f *filter) filterRUM(cond parser.WhereConditions, pts []*point.Point) []*point.Point {
	if cond == nil {
		l.Debugf("no condition filter for rum")
		return pts
	}

	var after []*point.Point

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

// using measurement name as tag `service'.
func (f *filter) filterSecurity(cond parser.WhereConditions, pts []*point.Point) []*point.Point {
	if cond == nil {
		l.Debugf("no condition filter for security")
		return pts
	}

	var after []*point.Point

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

// using measurement name as tag `service'.
func (f *filter) filterProfiling(cond parser.WhereConditions, pts []*point.Point) []*point.Point {
	if cond == nil {
		l.Debugf("no condition filter for profiling")
		return pts
	}

	var after []*point.Point

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

func (f *filter) doFilter(category string, pts []*point.Point) ([]*point.Point, int) {
	f.RWMutex.RLock()
	defer f.RWMutex.RUnlock()

	switch category {
	case datakit.Logging:
		return f.filterLogging(f.conditions["logging"], pts), len(f.conditions["logging"])

	case datakit.Tracing:
		return f.filterTracing(f.conditions["tracing"], pts), len(f.conditions["tracing"])

	case datakit.Metric:
		return f.filterMetric(f.conditions["metric"], pts), len(f.conditions["metric"])

	case datakit.Object:
		return f.filterObject(f.conditions["object"], pts), len(f.conditions["object"])

	case datakit.Network:
		return f.filterNetwork(f.conditions["network"], pts), len(f.conditions["network"])

	case datakit.KeyEvent:
		return f.filterKeyEvent(f.conditions["keyevent"], pts), len(f.conditions["keyevent"])

	case datakit.CustomObject:
		return f.filterCustomObject(f.conditions["custom_object"], pts), len(f.conditions["custom_object"])

	case datakit.RUM:
		return f.filterRUM(f.conditions["rum"], pts), len(f.conditions["rum"])

	case datakit.Security:
		return f.filterSecurity(f.conditions["security"], pts), len(f.conditions["security"])

	case datakit.Profiling:
		return f.filterProfiling(f.conditions["profiling"], pts), len(f.conditions["profiling"])

	default: // TODO: not implemented
		l.Warn("unsupport category: %s", category)
		return pts, 0
	}
}

func filterPts(category string, pts []*point.Point) []*point.Point {
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
		l.Debug("feed filter metrics failed, ignored: %+#v", fm)
	}

	return after
}

type qstats struct {
	ch chan *FilterStats
}

func GetFilterStats() *FilterStats {
	// return nil when not started
	if !isStarted {
		return nil
	}

	q := &qstats{
		ch: make(chan *FilterStats),
	}

	tick := time.NewTicker(time.Second)
	defer tick.Stop()

	select {
	case defaultFilter.qch <- q:
	case <-tick.C:
		l.Warnf("query failed, filter busy")
		return nil
	}

	select {
	case s := <-q.ch:
		return s
	case <-tick.C:
		l.Warnf("filter timeout")
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

func (f *filter) updateMetric(m *filterMetric) {
	ruleStats := defaultFilter.stats.RuleStats

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
}

func (f *filter) start() {
	defer defaultFilter.tick.Stop()

	if len(defaultIO.conf.Filters) != 0 {
		f.stats.RuleSource = "datakit.conf"
	} else {
		f.stats.RuleSource = "remote"
	}

	for {
		select {
		case <-defaultFilter.tick.C:
			l.Debugf("try pull remote filters...")
			defaultFilter.pull()

		case m := <-defaultFilter.metricCh:
			l.Debugf("update metrics...")
			f.updateMetric(m)

		case q := <-defaultFilter.qch:
			l.Debugf("accept stats query...")

			select {
			case <-q.ch:
			case q.ch <- copyStats(defaultFilter.stats):
			default: // pass
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
