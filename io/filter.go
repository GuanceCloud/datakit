package io

import (

	// nolint:gosec
	"crypto/md5"
	"encoding/json"
	"fmt"
	"strings"
	"sync"
	"time"

	"gitlab.jiagouyun.com/cloudcare-tools/cliutils/logger"
	"gitlab.jiagouyun.com/cloudcare-tools/cliutils/system/rtpanic"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io/parser"
)

var l = logger.DefaultSLogger("filter")

var defaultFilter = &filter{
	conditions:   map[string]parser.WhereConditions{},
	dw:           &datawayImpl{},
	RWMutex:      sync.RWMutex{},
	metricCh:     make(chan *filterMetric, 32),
	qch:          make(chan *qstats),
	pullInterval: time.Second * 10,

	// stats key is category + service/source
	stats: &FilterStats{
		RuleStats: map[string]*ruleStat{},
	},
}

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
	stats        *FilterStats
}

type filterPull struct {
	Filters map[string][]string `json:"filters"`
	// other fields ignored
	PullInterval time.Duration `json:"pull_interval"`
}

func (f *filter) pull() {
	start := time.Now()

	f.stats.PullCount++

	body, err := f.dw.Pull()
	if err != nil {
		l.Error("dataway Pull: %s", err)
		f.stats.PullFailed++
		f.stats.LastErr = err.Error()
		f.stats.LastErrTime = time.Now()
		return
	}

	if len(defaultIO.conf.Filters) != 0 {
		f.stats.RuleSource = "datakit.conf"
	} else {
		f.stats.RuleSource = "remote"
	}

	cost := time.Since(start)
	f.stats.PullCost += cost
	f.stats.PullCostAvg = f.stats.PullCost / time.Duration(f.stats.PullCount)
	if cost > f.stats.PullCostMax {
		f.stats.PullCostMax = cost
	}

	var fp filterPull
	if err := json.Unmarshal(body, &fp); err != nil {
		l.Error("json.Unmarshal: %s", err)
		f.stats.LastErr = err.Error()
		f.stats.LastErrTime = time.Now()
		return
	}

	bodymd5 := fmt.Sprintf("%x", md5.Sum(body)) //nolint:gosec
	if bodymd5 != f.md5 {                       // try update conditions
		f.stats.LastUpdate = start
		f.RWMutex.Lock()
		defer f.RWMutex.Unlock()

		if fp.PullInterval > 0 {
			f.pullInterval = fp.PullInterval
			f.stats.PullInterval = fp.PullInterval
		}

		f.md5 = bodymd5
		for k, v := range fp.Filters {
			f.conditions[k] = parser.GetConds(strings.Join(v, ";"))
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
		tags := pt.Tags()
		fields, err := pt.Fields()
		if err != nil {
			continue // filter it!
		}

		tags["source"] = pt.Name() // set measurement name as tag `source'
		if !filtered(cond, tags, fields) {
			after = append(after, pt)
		}
	}

	return after
}

func (f *filter) filterMetric(cond parser.WhereConditions, pts []*Point) []*Point {
	if cond == nil {
		l.Debugf("no condition filter")
		return pts
	}

	var after []*Point

	for _, pt := range pts {
		tags := pt.Tags()
		fields, err := pt.Fields()
		if err != nil {
			continue // filter it!
		}

		tags["measurement"] = pt.Name() // set measurement name as tag `measurement'

		if !filtered(cond, tags, fields) {
			after = append(after, pt)
		}
	}

	return after
}

func (f *filter) filterTracing(cond parser.WhereConditions, pts []*Point) []*Point {
	if cond == nil {
		l.Debugf("no condition filter")
		return pts
	}

	var after []*Point

	for _, pt := range pts {
		tags := pt.Tags()
		fields, err := pt.Fields()
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

func (f *filter) filter(category string, pts []*Point) ([]*Point, int) {
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
	default: // TODO: not implemented
		l.Warn("unsupport category: %s", category)
		return pts, 0
	}
}

func filterPts(category string, pts []*Point) []*Point {
	start := time.Now()
	after, condCount := defaultFilter.filter(category, pts)
	cost := time.Since(start)

	l.Debugf("%s/pts: %d, after: %d", category, len(pts), len(after))

	// report metrics
	select {
	case defaultFilter.metricCh <- &filterMetric{
		key:        category,
		points:     len(pts),
		filtered:   len(pts) - len(after),
		cost:       cost,
		conditions: condCount,
	}:
	default: // unblocking
	}

	return after
}

func GetFilterStats() *FilterStats {
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

func StartFilter() {
	var f rtpanic.RecoverCallback

	l = logger.SLogger("filter")
	parser.Init()

	ruleStats := defaultFilter.stats.RuleStats

	// first pull: get filter condition ASAP
	defaultFilter.pull()

	tick := time.NewTicker(defaultFilter.pullInterval)
	defer tick.Stop()

	f = func(trace []byte, err error) {
		defer rtpanic.Recover(f, nil)
		if trace != nil {
			l.Warnf("filter panic: %s: %s", err, string(trace))
		}

		for {
			select {
			case <-tick.C:
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

	f(nil, nil)
}
