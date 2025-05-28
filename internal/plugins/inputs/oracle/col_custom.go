// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package oracle

import (
	"context"
	"time"

	"github.com/GuanceCloud/cliutils/point"
	"github.com/spf13/cast"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/config"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/goroutine"
	dkio "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/io"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/metrics"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs"
)

type customQuery struct {
	SQL      string           `toml:"sql"`
	Metric   string           `toml:"metric"`
	Tags     []string         `toml:"tags"`
	Fields   []string         `toml:"fields"`
	Interval datakit.Duration `toml:"interval"`
}

func (ipt *Input) runCustomQueries() {
	if len(ipt.Query) == 0 {
		return
	}

	l.Infof("start to run custom queries, total %d queries", len(ipt.Query))

	g := goroutine.NewGroup(goroutine.Option{
		Name:         "oracle_custom_query",
		PanicTimes:   6,
		PanicTimeout: 10 * time.Second,
	})
	for _, q := range ipt.Query {
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

	start := time.Now()
	for {
		if ipt.pause {
			l.Debugf("not leader, custom query skipped")
		} else {
			// collect custom query
			l.Debugf("start collecting custom query, metric name: %s", query.Metric)
			arr := getCleanCustomQueries(ipt.q(query.SQL, getMetricName(query.Metric, "custom_query")))

			pts := []*point.Point{}
			opts := point.DefaultMetricOptions()
			opts = append(opts, point.WithTimestamp(start.UnixNano()))
			if ipt.Election {
				opts = append(opts, point.WithExtraTags(datakit.GlobalElectionTags()))
			}
			for _, row := range arr {
				var kvs point.KVs
				// add extended tags
				for k, v := range ipt.mergedTags {
					kvs = kvs.AddTag(k, v)
				}

				for _, tgKey := range query.Tags {
					if value, ok := row[tgKey]; ok {
						kvs = kvs.AddTag(tgKey, cast.ToString(value))
						delete(row, tgKey)
					}
				}

				for _, fdKey := range query.Fields {
					if value, ok := row[fdKey]; ok {
						// transform all fields to float64
						kvs = kvs.Add(fdKey, cast.ToFloat64(value), false, true)
					}
				}

				if kvs.FieldCount() > 0 {
					pts = append(pts, point.NewPointV2(query.Metric, kvs, opts...))
				}
			}
			if len(pts) > 0 {
				if err := ipt.feeder.FeedV2(point.Metric, pts,
					dkio.WithCollectCost(time.Since(start)),
					dkio.WithElection(ipt.Election),
					dkio.WithInputName(customQueryFeedName),
				); err != nil {
					ipt.feeder.FeedLastError(err.Error(),
						metrics.WithLastErrorInput(customQueryFeedName),
						metrics.WithLastErrorCategory(point.Metric),
					)
					l.Errorf("feed failed: %s", err.Error())
				}
			}
		}

		select {
		case <-datakit.Exit.Wait():
			l.Info("custom query exit")
			return

		case <-ipt.semStop.Wait():
			l.Info("custom query return")
			return

		case tt := <-tick.C:
			nextts := inputs.AlignTimeMillSec(tt, start.UnixMilli(), duration.Milliseconds())
			start = time.UnixMilli(nextts)
		}
	}
}
