// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

//go:build linux
// +build linux

// Package coceanbase contains collect OceanBase code.
package coceanbase

import (
	"crypto/md5" //nolint:gosec
	"fmt"
	"reflect"
	"strconv"
	"time"

	"github.com/GuanceCloud/cliutils/point"
	"github.com/spf13/cast"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/externals/oceanbase/collect/ccommon"
)

type customQueryCollector struct {
	x collectParameters
}

var _ ccommon.DBMetricsCollector = (*customQueryCollector)(nil)

func newCustomQueryCollector(opts ...collectOption) *customQueryCollector {
	m := &customQueryCollector{}

	for _, opt := range opts {
		if opt != nil {
			opt(&m.x)
		}
	}

	return m
}

func (m *customQueryCollector) Collect() ([]*point.Point, error) {
	l.Debug("Collect entry")
	return m.metricCollectMysqlCustomQueries()
}

func (m *customQueryCollector) metricCollectMysqlCustomQueries() ([]*point.Point, error) {
	if err := m.collectMysqlCustomQueries(); err != nil {
		return nil, err
	}

	return m.buildMysqlCustomQueries()
}

func (m *customQueryCollector) collectMysqlCustomQueries() error {
	m.x.Ipt.mCustomQueries = map[string][]map[string]interface{}{}

	l.Debugf("m.x.Ipt.Query = %v", m.x.Ipt.Query)

	for _, item := range m.x.Ipt.Query {
		l.Debugf("item = %s", item.SQL)

		arr := getCleanMysqlCustomQueries(m.x.Ipt.q(item.SQL))
		if arr == nil {
			l.Debug("arr == nil")
			continue
		}
		if item.MD5Hash == "" {
			hs := GetMD5String32([]byte(item.SQL))
			item.MD5Hash = hs
		}
		m.x.Ipt.mCustomQueries[item.MD5Hash] = make([]map[string]interface{}, 0)
		m.x.Ipt.mCustomQueries[item.MD5Hash] = arr
	}

	return nil
}

func (m *customQueryCollector) buildMysqlCustomQueries() ([]*point.Point, error) {
	var pts []*point.Point

	hostTag := ccommon.GetHostTag(l, m.x.Ipt.host)

	for hs, items := range m.x.Ipt.mCustomQueries {
		var qy *customQuery
		for _, v := range m.x.Ipt.Query {
			if hs == v.MD5Hash {
				qy = v
				break
			}
		}
		if qy == nil {
			continue
		}

		for _, item := range items {
			var kvs point.KVs

			for _, tgKey := range qy.Tags {
				if value, ok := item[tgKey]; ok {
					kvs = kvs.AddTag(tgKey, cast.ToString(value))
					delete(item, tgKey)
				}
			}

			for _, fdKey := range qy.Fields {
				if value, ok := item[fdKey]; ok {
					// transform all fields to float64
					kvs = kvs.Add(fdKey, cast.ToFloat64(value), false, false)
				}
			}

			if kvs.FieldCount() > 0 {
				pts = append(pts, ccommon.BuildPointMetric(
					kvs, qy.Metric,
					m.x.Ipt.tags, hostTag,
				))
			}
		}
	}

	return pts, nil
}

func getCleanMysqlCustomQueries(r rows) []map[string]interface{} {
	l.Debugf("getCleanMysqlCustomQueries entry")

	if r == nil {
		l.Debug("r == nil")
		return nil
	}

	defer closeRows(r)

	var list []map[string]interface{}

	columns, err := r.Columns()
	if err != nil {
		l.Errorf("Columns() failed: %v", err)
	}
	l.Debugf("columns = %v", columns)
	columnLength := len(columns)
	l.Debugf("columnLength = %d", columnLength)

	cache := make([]interface{}, columnLength)
	for idx := range cache {
		var a interface{}
		cache[idx] = &a
	}

	for r.Next() {
		l.Debug("Next() entry")

		if err := r.Scan(cache...); err != nil {
			l.Errorf("Scan failed: %v", err)
		}

		l.Debugf("len(cache) = %d", len(cache))

		item := make(map[string]interface{})
		for i, data := range cache {
			key := columns[i]
			val := *data.(*interface{})

			if val != nil {
				vType := reflect.TypeOf(val)

				l.Debugf("key = %s, vType = %s, %s", key, vType.String(), vType.Name())

				switch vType.String() {
				case "int64":
					if v, ok := val.(int64); ok {
						item[key] = v
					} else {
						l.Warn("expect int64, ignored")
					}
				case "string":
					var data interface{}
					data, err := strconv.ParseFloat(val.(string), 64)
					if err != nil {
						data = val
					}
					item[key] = data
				case "time.Time":
					if v, ok := val.(time.Time); ok {
						item[key] = v
					} else {
						l.Warn("expect time.Time, ignored")
					}
				case "[]uint8":
					item[key] = string(val.([]uint8))
				default:
					l.Warn("unsupport data type '%s', ignored", vType)
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

func GetMD5String32(bt []byte) string {
	return fmt.Sprintf("%X", md5.Sum(bt)) // nolint:gosec
}
