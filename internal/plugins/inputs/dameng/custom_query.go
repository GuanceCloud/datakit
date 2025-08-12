// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package dameng

import (
	"context"
	"strconv"
	"time"

	"github.com/GuanceCloud/cliutils/point"
	"github.com/jmoiron/sqlx"
	"github.com/spf13/cast"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/config"
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
	Tags     []string         `toml:"tags"`
	Fields   []string         `toml:"fields"`
	Interval datakit.Duration `toml:"interval"`
}

func getCleanCustomQueries(r *sqlx.Rows) []map[string]interface{} {
	l.Debugf("getCleanCustomQueries entry")

	if r == nil {
		l.Debug("r == nil")
		return nil
	}

	defer r.Close() //nolint:errcheck

	var list []map[string]interface{}

	columns, err := r.Columns()
	if err != nil {
		l.Errorf("Columns() failed: %v", err)
		return nil
	}
	l.Debugf("columns = %v", columns)
	columnLength := len(columns)
	l.Debugf("columnLength = %d", columnLength)

	cache := make([]interface{}, columnLength)
	for i := range cache {
		var a interface{}
		cache[i] = &a
	}

	for r.Next() {
		l.Debug("Next() entry")

		if err := r.Scan(cache...); err != nil {
			l.Errorf("Scan failed: %v", err)
			continue
		}

		item := make(map[string]interface{})
		for i, data := range cache {
			key := columns[i]
			val := *data.(*interface{})

			if val != nil {
				switch v := val.(type) {
				case int64:
					item[key] = v
				case float64:
					item[key] = v
				case string:
					if f, err := strconv.ParseFloat(v, 64); err == nil {
						item[key] = f
					} else {
						item[key] = v
					}
				case []byte:
					if f, err := strconv.ParseFloat(string(v), 64); err == nil {
						item[key] = f
					} else {
						item[key] = string(v)
					}
				case time.Time:
					item[key] = v
				default:
					l.Warnf("Unsupported data type %T for column %s, ignored", v, key)
				}
			}
		}

		list = append(list, item)
	}

	if err := r.Err(); err != nil {
		l.Errorf("Err() failed: %v", err)
	}

	l.Debugf("len(list) = %d", len(list))
	return list
}

func (ipt *Input) runCustomQueries() {
	if len(ipt.Query) == 0 {
		return
	}

	l.Infof("start to run custom queries, total %d queries", len(ipt.Query))

	g := goroutine.NewGroup(goroutine.Option{
		Name:         "dameng_custom_query",
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
	if query == nil || query.SQL == "" || query.Metric == "" {
		l.Warnf("Invalid custom query: nil, empty SQL, or empty metric name")
		return
	}

	// Use input interval as default
	duration := ipt.Interval.Duration
	// Use custom query interval if set
	if query.Interval.Duration > 0 {
		duration = config.ProtectedInterval(minInterval, maxInterval, query.Interval.Duration)
	}

	tick := time.NewTicker(duration)
	defer tick.Stop()

	ptsTime := ntp.Now()

	for {
		collectStart := time.Now()
		if ipt.pause {
			l.Debugf("not leader, custom query %s skipped", query.Metric)
		} else {
			l.Debugf("start collecting custom query, metric name: %s", query.Metric)

			// Execute custom query
			rows, err := ipt.db.Queryx(query.SQL)
			if err != nil {
				l.Errorf("Custom query %q failed: %v", query.SQL, err)
				ipt.feeder.FeedLastError(err.Error(),
					metrics.WithLastErrorInput(customQueryFeedName),
					metrics.WithLastErrorCategory(point.Metric),
				)
			} else {
				// Process rows using getCleanCustomQueries
				results := getCleanCustomQueries(rows)

				pts := []*point.Point{}
				opts := point.DefaultMetricOptions()
				opts = append(opts, point.WithTime(ptsTime))
				if ipt.Election {
					opts = append(opts, point.WithExtraTags(datakit.GlobalElectionTags()))
				}

				for _, row := range results {
					var kvs point.KVs
					// Add common tags
					for k, v := range ipt.mergedTags {
						kvs = kvs.AddTag(k, v)
					}

					// Add query-specific tags
					for _, tgKey := range query.Tags {
						if value, ok := row[tgKey]; ok {
							kvs = kvs.AddTag(tgKey, cast.ToString(value))
						}
					}

					// Add query-specific fields
					for _, fdKey := range query.Fields {
						if value, ok := row[fdKey]; ok {
							switch v := value.(type) {
							case int64:
								kvs = kvs.Set(fdKey, float64(v))
							case float64:
								kvs = kvs.Set(fdKey, v)
							case string:
								if f, err := strconv.ParseFloat(v, 64); err == nil {
									kvs = kvs.Set(fdKey, f)
								} else {
									l.Warnf("Field %s is string '%s', cannot convert to float64, ignored", fdKey, v)
								}
							default:
								l.Warnf("Field %s has unsupported type %T, ignored", fdKey, v)
							}
						}
					}

					if kvs.FieldCount() > 0 {
						pts = append(pts, point.NewPoint(query.Metric, kvs, opts...))
					}
				}

				if len(pts) > 0 {
					if err := ipt.feeder.Feed(point.Metric, pts,
						dkio.WithCollectCost(time.Since(collectStart)),
						dkio.WithElection(ipt.Election),
						dkio.WithSource(customQueryFeedName),
					); err != nil {
						ipt.feeder.FeedLastError(err.Error(),
							metrics.WithLastErrorInput(customQueryFeedName),
							metrics.WithLastErrorCategory(point.Metric),
						)
						l.Errorf("Feed failed: %s", err)
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
