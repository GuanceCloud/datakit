// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

// Package filter contains filter logic.
package filter

import (
	"crypto/md5" // nolint:gosec
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/GuanceCloud/cliutils/logger"
	"github.com/GuanceCloud/cliutils/system/rtpanic"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io/dataway"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io/parser"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io/point"
)

const packageName = "filter"

var (
	l            = logger.DefaultSLogger(packageName)
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
	Pull(map[string][]string, dataway.DataWay) ([]byte, error)
}

type datawayImpl struct{}

func (dw *datawayImpl) Pull(filters map[string][]string, inDW dataway.DataWay) ([]byte, error) {
	if len(filters) != 0 {
		// read local filters
		return json.Marshal(&filterPull{Filters: filters, PullInterval: pullInterval})
	} else {
		// pull filters remotely
		return inDW.DatakitPull("filters=true")
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
		l.Errorf("json.Unmarshal: %v", err)
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
		conds, err := GetConds(v)
		if err != nil {
			l.Errorf("GetConds failed: %v", err)
			return err
		}
		f.conditions[k] = conds
	}

	if err := dump(body, dumpdir); err != nil {
		l.Warnf("dump: %s, ignored", err)
	}
	return nil
}

func (f *filter) pull(filters map[string][]string, dw dataway.DataWay) {
	start := time.Now()

	body, err := f.dw.Pull(filters, dw)
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
func CheckPointFiltered(conds parser.WhereConditions, category string, pt *point.Point) (bool, error) {
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

func (f *filter) doFilter(category string, pts []*point.Point) ([]*point.Point, int) {
	f.RWMutex.RLock()
	defer f.RWMutex.RUnlock()

	// "/v1/write/metric" => "metric"
	categoryPureStr, ok := datakit.CategoryPureMap[category]
	if !ok {
		l.Warnf("unsupport category: %s", category)
		return pts, 0
	}

	conds, ok := f.conditions[categoryPureStr]
	if !ok || conds == nil {
		l.Debugf("no condition filter for %s", categoryPureStr)
		return pts, 0
	}

	var after []*point.Point

	for _, pt := range pts {
		isFiltered, err := CheckPointFiltered(conds, category, pt)
		if err != nil {
			l.Errorf("pt.Fields: %s, ignored", err.Error())
			continue // filter it!
		}
		if !isFiltered { // Pick those points that not matched filter rules.
			after = append(after, pt)
		} else if datakit.LogSinkDetail {
			l.Infof("(sink_detail) defaultFilter filtered point: (%s) (%s)", category, pt.String())
		}
	}

	return after, len(conds)
}

func FilterPts(category string, pts []*point.Point) []*point.Point {
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
	if v.Total > 0 {
		v.CostPerPoint = v.Cost / time.Duration(v.Total)
	}
	v.Conditions = m.conditions
}

func (f *filter) start(filters map[string][]string, dw dataway.DataWay) {
	defer defaultFilter.tick.Stop()

	if len(filters) != 0 {
		f.stats.RuleSource = datakit.StrDefaultConfFile
	} else {
		f.stats.RuleSource = "remote"
	}

	// Try pull rules ASAP.
	defaultFilter.pull(filters, dw)

	for {
		select {
		case <-defaultFilter.tick.C:
			l.Debugf("try pull remote filters...")
			defaultFilter.pull(filters, dw)

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
			l.Info("log filter exits")
			return
		}
	}
}

func StartFilter(filters map[string][]string, dw dataway.DataWay) {
	l = logger.SLogger(packageName)
	if len(filters) == 0 && dw == nil {
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

		defaultFilter.start(filters, dw)
	}

	f(nil, nil)
}
