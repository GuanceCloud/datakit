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
	defInterval  = 10 * time.Second
	defLogfilter = &logFilter{status: filterReleased}
)

type rules struct {
	Content []string `json:"content"`
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
			l.Error(err)
			continue
		}

		if !ipt.conds.Eval(pt.Name(), pt.Tags(), fields) {
			after = append(after, pt)
		}
	}

	return after
}

func (ipt *logFilter) start() {
	l.Infof("log filter engaged, status: %q refresh_interval: %ds", ipt.status.String(), int(defInterval.Seconds()))

	g := datakit.G("logfilter")
	g.Go(func(ctx context.Context) error {
		tick := time.NewTicker(defInterval)
	EXIT:
		for {
			select {
			case <-datakit.Exit.Wait():
				l.Info("log filter exits")
				break EXIT
			case <-tick.C:
				l.Debugf("### enter log filter refresh routine, status: %q", ipt.status.String())
				if err := ipt.refreshRules(); err != nil {
					l.Error(err.Error())
					FeedLastError("logfilter", err.Error())
				}
			}
		}
		return nil
	})
}

func (ipt *logFilter) refreshRules() error {
	defer func() {
		if err := recover(); err != nil {
			l.Error(err)
		}
	}()
	body, err := defLogFilterMock.getLogFilter()
	if err != nil {
		return err
	}

	if len(body) == 0 {
		ipt.status = filterReleased

		return nil
	}

	var rules rules
	if err = json.Unmarshal(body, &rules); err != nil {
		return err
	}

	if len(rules.Content) == 0 {
		ipt.status = filterReleased

		return nil
	}

	ipt.Lock()
	defer ipt.Unlock()

	// compare and refresh
	if newRules := strings.Join(rules.Content, ";"); newRules != ipt.rules {
		conds := parser.GetConds(newRules)
		if conds != nil {
			ipt.conds = conds
			ipt.rules = newRules
			ipt.status = filterRefreshed
		}
	}

	return nil
}
