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
	filter_released logFilterStatus = iota + 1
	filter_refreshed
)

var (
	defInterval  = 10 * time.Second
	defLogfilter = &logFilter{status: filter_released}
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

func (this *logFilter) filter(pts []*Point) []*Point {
	// mock data injector
	pts = defLogFilterMock.preparePoints(pts)

	if this.status == filter_released {
		return pts
	}

	var after []*Point
	for _, pt := range pts {
		fields, err := pt.Fields()
		if err != nil {
			l.Error(err)
			continue
		}

		if !this.conds.Eval(pt.Name(), pt.Tags(), fields) {
			after = append(after, pt)
		}
	}

	return after
}

func (this *logFilter) start() {
	l.Infof("log filter engaged, status: %q refresh_interval: %ds", this.status.String(), int(defInterval.Seconds()))

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
				l.Debugf("### enter log filter refresh routine, status: %q", this.status.String())
				if err := this.refreshRules(); err != nil {
					l.Error(err.Error())
				}
			}
		}
		return nil
	})
}

func (this *logFilter) refreshRules() error {
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
		this.status = filter_released

		return nil
	}

	var rules rules
	if err = json.Unmarshal(body, &rules); err != nil {
		return err
	}

	if len(rules.Content) == 0 {
		this.status = filter_released

		return nil
	}

	this.Lock()
	defer this.Unlock()

	// compare and refresh
	if newRules := strings.Join(rules.Content, ";"); newRules != this.rules {
		conds := parser.GetConds(newRules)
		if conds != nil {
			this.conds = conds
			this.rules = newRules
			this.status = filter_refreshed
		}
	}

	return nil
}
