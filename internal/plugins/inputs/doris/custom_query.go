// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package doris

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/GuanceCloud/cliutils/point"
	"github.com/spf13/cast"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/goroutine"
	dkio "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/io"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/metrics"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/ntp"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs"
)

type customQuery struct {
	SQL      string           `toml:"sql"`
	Metric   string           `toml:"metric"`
	Interval datakit.Duration `toml:"interval"`
	Tags     []string         `toml:"tags"`
	Fields   []string         `toml:"fields"`
}

func (ipt *Input) runCustomQueries() {
	if len(ipt.Query) == 0 {
		return
	}

	l.Infof("start to run custom queries, total %d queries", len(ipt.Query))

	g := goroutine.NewGroup(goroutine.Option{
		Name:         "doris_custom_query",
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
	if err := validateCustomQuery(query); err != nil {
		l.Warnf("invalid doris custom query: %s", err.Error())
		return
	}

	duration := ipt.Interval.Duration
	if query.Interval.Duration > 0 {
		duration = query.Interval.Duration
	}

	tick := time.NewTicker(duration)
	defer tick.Stop()

	ptsTime := ntp.Now()
	for {
		collectStart := time.Now()
		if ipt.pause.Load() {
			l.Debugf("not leader, custom query skipped")
		} else {
			l.Debugf("start collecting custom query, metric name: %s", query.Metric)

			ctx, cancel := context.WithTimeout(context.Background(), ipt.ConnectTimeout.Duration)
			rows, err := ipt.queryCustomRows(ctx, query.SQL)
			cancel()
			if err != nil {
				l.Errorf("collect custom query [%s] failed: %s", query.SQL, err.Error())
			} else if points := ipt.getCustomQueryPoints(query, rows, ptsTime); len(points) > 0 {
				if err := ipt.feeder.Feed(point.Metric, points,
					dkio.WithCollectCost(time.Since(collectStart)),
					dkio.WithElection(ipt.Election),
					dkio.WithSource(customQueryFeedName),
				); err != nil {
					ipt.feeder.FeedLastError(err.Error(),
						metrics.WithLastErrorInput(customQueryFeedName),
						metrics.WithLastErrorCategory(point.Metric),
					)
					l.Errorf("feed custom query failed: %s", err.Error())
				}
			}
		}

		select {
		case <-datakit.Exit.Wait():
			l.Info("doris custom query exit")
			return

		case <-ipt.semStop.Wait():
			l.Info("doris custom query return")
			return

		case tt := <-tick.C:
			ptsTime = inputs.AlignTime(tt, ptsTime, duration)
		}
	}
}

func validateCustomQuery(query *customQuery) error {
	if query == nil {
		return fmt.Errorf("nil query")
	}
	if strings.TrimSpace(query.SQL) == "" {
		return fmt.Errorf("empty sql")
	}
	if strings.TrimSpace(query.Metric) == "" {
		return fmt.Errorf("empty metric")
	}
	if len(query.Fields) == 0 {
		return fmt.Errorf("empty fields")
	}

	return nil
}

func (ipt *Input) queryCustomRows(ctx context.Context, query string) ([]map[string]interface{}, error) {
	rows, err := ipt.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("query custom sql: %w", err)
	}
	defer func() {
		if err := rows.Close(); err != nil {
			l.Warnf("close custom query rows failed: %s", err.Error())
		}
	}()

	return customRowsToMaps(rows)
}

func customRowsToMaps(rows *sql.Rows) ([]map[string]interface{}, error) {
	columns, err := rows.Columns()
	if err != nil {
		return nil, fmt.Errorf("get custom query columns: %w", err)
	}

	values := make([]interface{}, len(columns))
	scanArgs := make([]interface{}, len(columns))
	for i := range values {
		scanArgs[i] = &values[i]
	}

	var result []map[string]interface{}
	for rows.Next() {
		if err := rows.Scan(scanArgs...); err != nil {
			return nil, fmt.Errorf("scan custom query row: %w", err)
		}

		row := make(map[string]interface{}, len(columns))
		for i, col := range columns {
			if values[i] != nil {
				row[col] = normalizeCustomValue(values[i])
			}
		}
		result = append(result, row)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate custom query rows: %w", err)
	}

	return result, nil
}

func normalizeCustomValue(v interface{}) interface{} {
	switch value := v.(type) {
	case []byte:
		return string(value)
	case string:
		return value
	default:
		return value
	}
}

func (ipt *Input) getCustomQueryPoints(query *customQuery, rows []map[string]interface{}, ptsTime time.Time) []*point.Point {
	if query == nil {
		return nil
	}

	var pts []*point.Point
	opts := point.DefaultMetricOptions()
	opts = append(opts, point.WithTime(ptsTime))

	for _, row := range rows {
		kvs := point.KVs{}
		for k, v := range ipt.mergedTags {
			kvs = kvs.AddTag(k, v)
		}

		for _, tgKey := range query.Tags {
			if value, ok := row[tgKey]; ok {
				kvs = kvs.AddTag(tgKey, cast.ToString(value))
			}
		}

		// Custom query fields are metric values, so only numeric columns are kept.
		for _, fdKey := range query.Fields {
			if value, ok := row[fdKey]; ok {
				if f, ok := customFieldFloat(value); ok {
					kvs = kvs.Set(fdKey, f)
				} else {
					l.Warnf("custom query field %s has non-numeric value %v, ignored", fdKey, value)
				}
			}
		}

		if kvs.FieldCount() > 0 {
			pts = append(pts, point.NewPoint(query.Metric, kvs, opts...))
		}
	}

	return pts
}

func customFieldFloat(v interface{}) (float64, bool) {
	f, err := cast.ToFloat64E(v)
	if err != nil {
		return 0, false
	}
	return f, true
}
