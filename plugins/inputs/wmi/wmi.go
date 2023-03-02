// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

//go:build windows
// +build windows

package wmi

import (
	"context"
	"os"
	"time"

	"github.com/GuanceCloud/cliutils"
	"github.com/GuanceCloud/cliutils/logger"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/goroutine"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io/point"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"
)

var l = logger.DefaultSLogger(inputName)

func (*Instance) SampleConfig() string {
	return sampleConfig
}

func (*Instance) Catalog() string {
	return `windows`
}

func (ag *Instance) Run() {
	l = logger.SLogger(inputName)
	g := goroutine.NewGroup(goroutine.Option{Name: "inputs_wmi"})
	g.Go(func(ctx context.Context) error {
		select {
		case <-datakit.Exit.Wait():
		case <-ag.semStop.Wait():
		}
		ag.exit()
		ag.cancelFun()
		return nil
	})

	if ag.MetricName == "" {
		ag.MetricName = "WMI"
	}

	if ag.Interval.Duration == 0 {
		ag.Interval.Duration = time.Minute * 5
	}

	_ = ag.run(ag.ctx)
}

func (ag *Instance) exit() {
	ag.cancelFun()
}

func (ag *Instance) Terminate() {
	if ag.semStop != nil {
		ag.semStop.Close()
	}
}

func (ag *Instance) run(ctx context.Context) error {
	for {
		select {
		case <-ctx.Done():
			return context.Canceled
		default:
		}

		for _, query := range ag.Queries {
			select {
			case <-ctx.Done():
				return context.Canceled
			default:
			}

			if query.lastTime.IsZero() {
				query.lastTime = time.Now()
			} else if time.Since(query.lastTime) < query.Interval.Duration {
				continue
			}

			sql, err := query.ToSQL()
			if err != nil {
				l.Warnf("%s", err)
				continue
			}

			props := []string{}

			for _, ms := range query.Metrics {
				props = append(props, ms[0])
			}

			fieldsArr, err := DefaultClient.QueryEx(sql, props)
			if err != nil {
				l.Errorf("query failed, %s", err)
				continue
			}

			hostname, _ := os.Hostname()

			tags := map[string]string{
				"host": hostname,
			}

			for k, v := range ag.Tags {
				tags[k] = v
			}

			pts := []*point.Point{}
			for _, fields := range fieldsArr {
				if ag.isTest() {
					// pass
				} else {
					pt, err := point.NewPoint(inputName, tags, fields, point.MOpt())
					if err != nil {
						l.Warnf("point.NewPoint: %s, ignored", err)
						continue
					}
					pts = append(pts, pt)
				}
			}

			if err := io.Feed(inputName, datakit.Metric, pts, nil); err != nil {
				l.Warnf("point.NewPoint: %s, ignored", err)
			}

			query.lastTime = time.Now()
		}

		if ag.isTest() {
			return nil
		}
		_ = datakit.SleepContext(ctx, time.Second)
	}
}

func NewAgent() *Instance {
	ac := &Instance{}
	ac.ctx, ac.cancelFun = context.WithCancel(context.Background())
	ac.semStop = cliutils.NewSem()
	return ac
}

func init() { //nolint:gochecknoinits
	inputs.Add(inputName, func() inputs.Input {
		return NewAgent()
	})
}
