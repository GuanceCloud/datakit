//go:generate stringer -type logFilterStatus -output logfilter_stringer.go

package io

import (
	"context"
	"encoding/json"
	"strings"
	"sync"
	"time"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io/parser"
)

type logFilterStatus uint8

const (
	filterReleased logFilterStatus = iota + 1
	filterRefreshed
)

var (
	defIntervalDefault = 15
	defLogfilter       = &logFilter{status: filterReleased}
)

type rules struct {
	Content struct {
		LogFilter []string `json:"log_filter"`
		Interval  int      `json:"interval"`
	} `json:"content"`
}

type logFilter struct {
	status logFilterStatus
	rules  string
	conds  parser.WhereConditions
	sync.Mutex
}

func (ipt *logFilter) filter(pts []*Point) []*Point {
	// mock data injector
	pts = defLogFilterMock.preparePoints(pts)

	if ipt.status == filterReleased {
		return pts
	}

	var after []*Point
	for _, pt := range pts {
		fields, err := pt.Fields()
		if err != nil {
			log.Error(err)
			continue
		}

		if !ipt.conds.Eval(pt.Name(), pt.Tags(), fields) {
			after = append(after, pt)
		}
	}

	return after
}

func (ipt *logFilter) start() {
	log.Infof("log filter engaged, status: %q", ipt.status.String())

	g := datakit.G("logfilter")
	g.Go(func(ctx context.Context) error {
		tick := time.NewTicker(time.Second * time.Duration(defIntervalDefault))
		defer tick.Stop()
	EXIT:
		for {
			select {
			case <-datakit.Exit.Wait():
				log.Info("log filter exits")
				break EXIT
			case <-tick.C:
				log.Debugf("### enter log filter refresh routine, status: %q", ipt.status.String())
				var err error
				var defInterval int
				if defInterval, err = ipt.refreshRules(); err != nil {
					log.Error(err.Error())
					FeedLastError("logfilter", err.Error())
				}
				if defInterval != defIntervalDefault {
					tick.Reset(time.Second * time.Duration(defInterval))
					defIntervalDefault = defInterval
				}
			}
		}
		return nil
	})
}

func (ipt *logFilter) refreshRules() (int, error) {
	defer func() {
		if err := recover(); err != nil {
			log.Error(err)
		}
	}()
	body, err := defLogFilterMock.getLogFilter()
	if err != nil {
		return defIntervalDefault, err
	}

	if len(body) == 0 {
		ipt.status = filterReleased

		return defIntervalDefault, nil
	}

	rules := rules{}
	if err = json.Unmarshal(body, &rules); err != nil {
		return defIntervalDefault, err
	}
	log.Debugf("logfilter result: %v", rules)
	if len(rules.Content.LogFilter) == 0 {
		ipt.status = filterReleased

		return rules.Content.Interval, nil
	}

	// compare and refresh
	if newRules := strings.Join(rules.Content.LogFilter, ";"); newRules != ipt.rules {
		conds := parser.GetConds(newRules)
		if conds != nil {
			ipt.conds = conds
			ipt.rules = newRules
			ipt.status = filterRefreshed
		}
	}

	return rules.Content.Interval, nil
}
