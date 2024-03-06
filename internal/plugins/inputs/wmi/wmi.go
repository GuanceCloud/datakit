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
	"github.com/GuanceCloud/cliutils/point"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/goroutine"
	dkio "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/io"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs"
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
			now := time.Now()
			for _, fields := range fieldsArr {
				if ag.isTest() {
					// pass
				} else {
					opts := point.DefaultMetricOptions()
					opts = append(opts, point.WithTime(now))

					pt := point.NewPointV2(inputName,
						append(point.NewTags(tags), point.NewKVs(fields)...),
						opts...)

					pts = append(pts, pt)
				}
			}
			if len(pts) > 0 {
				if err := ag.feeder.FeedV2(point.Metric, pts,
					dkio.WithCollectCost(time.Since(start)),
					dkio.WithInputName(inputName),
				); err != nil {
					l.Warnf("feeder.Feed failed: %s, ignored", err)
				}
			}
			query.lastTime = time.Now()
		}

		if ag.isTest() {
			return nil
		}
		_ = datakit.SleepContext(ctx, time.Second)
	}
}

func defaultInput() *Instance {
	ac := &Instance{}
	ac.ctx, ac.cancelFun = context.WithCancel(context.Background())
	ac.semStop = cliutils.NewSem()
	ac.feeder = dkio.DefaultFeeder()
	ac.Tagger = datakit.DefaultGlobalTagger()
	return ac
}

func init() { //nolint:gochecknoinits
	inputs.Add(inputName, func() inputs.Input {
		return defaultInput()
	})
}
