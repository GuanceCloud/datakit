package io

import (
	"crypto/md5"
	"encoding/json"
	"fmt"
	"strings"
	"sync"
	"time"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io/parser"
	"gitlab.jiagouyun.com/cloudcare-tools/kodo/rtpanic"
)

type IDataway interface {
	Pull() ([]byte, error)
}

type datawayImpl struct{}

func (dw *datawayImpl) Pull() ([]byte, error) {
	return defaultIO.dw.DatakitPull("filters=true")
}

type filter struct {
	conditions map[string]parser.WhereConditions
	dw         IDataway
	md5        string
	sync.RWMutex
}

type filterPull struct {
	Filters map[string][]string `json:"filters"`
	// other fields ignored
	PullInterval time.Duration `json:"pull_interval"`
}

func (f *filter) pull() {
	body, err := f.dw.Pull()
	if err != nil {
		l.Error("dataway Pull: %s", err)
		return
	}

	var fp filterPull
	if err := json.Unmarshal(body, &fp); err != nil {
		l.Error("json.Unmarshal: %s", err)
		return
	}

	bodymd5 := fmt.Sprintf("%x", md5.Sum(body))
	if bodymd5 != f.md5 { // try update conditions
		f.RWMutex.Lock()
		defer f.RWMutex.Unlock()

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

func (f *filter) filter(category string, pts []*Point) []*Point {
	// TODO: add filter metrics
	switch category {
	case datakit.Logging:
		f.RWMutex.RLock()
		defer f.RWMutex.RUnlock()
		return f.filterLogging(f.conditions["logging"], pts)
	case datakit.Tracing:
		f.RWMutex.RLock()
		defer f.RWMutex.RUnlock()
		return f.filterTracing(f.conditions["tracing"], pts)
	default: // TODO:not implemented
		l.Warn("unsupport category: %s", category)
		return pts
	}
}

func StartPull() {
	var f rtpanic.RecoverCallback

	tick := time.NewTicker(time.Second * time.Duration(defIntervalDefault))
	defer tick.Stop()

	filter := filter{}

	f = func(trace []byte, err error) {
		defer rtpanic.Recover(f, nil)
		if trace != nil {
			l.Warnf("filter panic: %s: %s", err, string(trace))
		}

		for {
			filter.pull()

			select {
			case <-tick.C:
			case <-datakit.Exit.Wait():
				log.Info("log filter exits")
				return
			}
		}
	}

	f(nil, nil)
}
