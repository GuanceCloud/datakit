// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package postgresql

import (
	"context"
	"time"

	"github.com/GuanceCloud/cliutils/point"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/config"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/goroutine"
	dkio "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/io"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/metrics"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/ntp"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs"
)

func (ipt *Input) runCustomQueries() {
	if len(ipt.CustomQuery) == 0 {
		return
	}

	l.Infof("start to run custom queries, total %d queries", len(ipt.CustomQuery))

	g := goroutine.NewGroup(goroutine.Option{
		Name:         "postgresql_custom_query",
		PanicTimes:   6,
		PanicTimeout: 10 * time.Second,
	})
	for _, q := range ipt.CustomQuery {
		func(q *customQuery) {
			g.Go(func(ctx context.Context) error {
				ipt.runCustomQuery(q)
				return nil
			})
		}(q)
	}
}

func (ipt *Input) runCustomQuery(query *customQuery) {
	if query == nil {
		return
	}

	// use input interval as default
	duration := ipt.Interval.Duration
	// use custom query interval if set
	if query.Interval.Duration > 0 {
		duration = config.ProtectedInterval(minInterval, maxInterval, query.Interval.Duration)
	}

	tick := time.NewTicker(duration)
	defer tick.Stop()

	ptsTime := ntp.Now()

	for {
		if ipt.pause {
			l.Debugf("not leader, custom query skipped")
		} else {
			collectStart := time.Now()
			// collect custom query
			l.Debugf("start collecting custom query, metric name: %s", query.Metric)

			tags := map[string]interface{}{}
			fields := map[string]interface{}{}
			for _, tag := range query.Tags {
				tags[tag] = tag
			}
			for _, field := range query.Fields {
				fields[field] = field
			}

			queryItem := &queryCacheItem{
				q:       query.SQL,
				ptsTime: ptsTime, // set point's time on custome query aligned-time
				measurementInfo: &inputs.MeasurementInfo{
					Name:   query.Metric,
					Tags:   tags,
					Fields: fields,
				},
			}

			if err := ipt.startService(); err != nil {
				l.Warnf("start service failed: %s", err.Error())
			} else {
				if points, err := ipt.getQueryPoints(queryItem); err != nil {
					l.Errorf("collect custom query [%s] failed: %s", query.SQL, err.Error())
				} else if len(points) > 0 {
					if err := ipt.feeder.Feed(point.Metric, points,
						dkio.WithCollectCost(time.Since(collectStart)),
						dkio.WithElection(ipt.Election),
						dkio.WithSource(customQueryFeedName),
					); err != nil {
						ipt.feeder.FeedLastError(err.Error(),
							metrics.WithLastErrorInput(customQueryFeedName),
						)
						l.Errorf("feed failed: %s", err.Error())
					}
				}
			}
		}

		select {
		case <-datakit.Exit.Wait():
			ipt.exit()
			l.Info("custom query exit")
			return

		case <-ipt.semStop.Wait():
			ipt.exit()
			l.Info("custom query return")
			return

		case tt := <-tick.C:
			ptsTime = inputs.AlignTime(tt, ptsTime, duration)
		}
	}
}
