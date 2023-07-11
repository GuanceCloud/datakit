// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package mysql

import (
	"time"

	gcPoint "github.com/GuanceCloud/cliutils/point"
	"github.com/spf13/cast"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs"
)

// metric build line proto

func (i *Input) buildMysql() ([]*gcPoint.Point, error) {
	m := &baseMeasurement{
		i:        i,
		resData:  make(map[string]interface{}),
		tags:     map[string]string{},
		fields:   make(map[string]interface{}),
		election: i.Election,
	}
	setHostTagIfNotLoopback(m.tags, i.Host)

	m.name = "mysql"
	for key, value := range i.Tags {
		m.tags[key] = value
	}

	for k, v := range i.globalStatus {
		m.resData[k] = v
	}
	for k, v := range i.globalVariables {
		m.resData[k] = v
	}
	if i.binLogOn {
		m.resData["Binlog_space_usage_bytes"] = i.binlog["Binlog_space_usage_bytes"]
	}

	metricInfo := m.Info()
	for key, item := range metricInfo.Fields {
		if value, ok := m.resData[key]; ok {
			val, err := Conv(value, item.(*inputs.FieldInfo).DataType)
			if err != nil {
				l.Errorf("buildMysql metric %v value %v parse error %v", key, value, err)
				return []*gcPoint.Point{}, err
			} else {
				m.fields[key] = val
			}
		}
	}

	if len(m.fields) > 0 {
		pts := getPointsFromMeasurement([]inputs.MeasurementV2{m})
		return pts, nil
	}

	return []*gcPoint.Point{}, nil
}

func (i *Input) buildMysqlSchema() ([]*gcPoint.Point, error) {
	ms := []inputs.MeasurementV2{}

	// SchemaSize
	for k, v := range i.mSchemaSize {
		m := &schemaMeasurement{
			name:     "mysql_schema",
			tags:     map[string]string{},
			fields:   make(map[string]interface{}),
			election: i.Election,
		}
		setHostTagIfNotLoopback(m.tags, i.Host)

		for key, value := range i.Tags {
			m.tags[key] = value
		}

		size := cast.ToFloat64(v)

		m.fields["schema_size"] = size
		m.tags["schema_name"] = k
		m.ts = time.Now()

		if len(m.fields) > 0 {
			ms = append(ms, m)
		}
	}

	for k, v := range i.mSchemaQueryExecTime {
		m := &schemaMeasurement{
			name:     "mysql_schema",
			tags:     make(map[string]string),
			fields:   make(map[string]interface{}),
			election: i.Election,
		}
		setHostTagIfNotLoopback(m.tags, i.Host)

		for key, value := range i.Tags {
			m.tags[key] = value
		}

		size := cast.ToInt64(v)

		m.fields["query_run_time_avg"] = size
		m.tags["schema_name"] = k
		m.ts = time.Now()

		if len(m.fields) > 0 {
			ms = append(ms, m)
		}
	}

	if len(ms) > 0 {
		pts := getPointsFromMeasurement(ms)
		return pts, nil
	}

	return []*gcPoint.Point{}, nil
}

func (i *Input) buildMysqlInnodb() ([]*gcPoint.Point, error) {
	ms := []inputs.MeasurementV2{}

	m := &innodbMeasurement{
		tags:     map[string]string{},
		fields:   make(map[string]interface{}),
		election: i.Election,
	}
	setHostTagIfNotLoopback(m.tags, i.Host)

	m.name = "mysql_innodb"

	for key, value := range i.Tags {
		m.tags[key] = value
	}

	m.fields = getMetricFields(i.mInnodb, m.Info())

	ms = append(ms, m)

	if len(ms) > 0 {
		pts := getPointsFromMeasurement(ms)
		return pts, nil
	}
	return []*gcPoint.Point{}, nil
}

func (i *Input) buildMysqlTableSchema() ([]*gcPoint.Point, error) {
	ms := []inputs.MeasurementV2{}

	for _, v := range i.mTableSchema {
		m := &tbMeasurement{
			tags:     map[string]string{},
			fields:   make(map[string]interface{}),
			election: i.Election,
		}
		setHostTagIfNotLoopback(m.tags, i.Host)

		m.name = "mysql_table_schema"

		for key, value := range i.Tags {
			m.tags[key] = value
		}

		for kk, vv := range v {
			switch kk {
			case "table_schema", "table_name", "table_type", "engine", "version":
				if vvv, ok := vv.(string); ok {
					m.tags[kk] = vvv
				}
			case "table_rows", "data_length", "index_length", "data_free":
				m.fields[kk] = vv
			}
		}

		ms = append(ms, m)
	}

	if len(ms) > 0 {
		pts := getPointsFromMeasurement(ms)
		return pts, nil
	}

	return []*gcPoint.Point{}, nil
}

func (i *Input) buildMysqlUserStatus() ([]*gcPoint.Point, error) {
	ms := []inputs.MeasurementV2{}

	for user := range i.mUserStatusName {
		m := &userMeasurement{
			tags:     map[string]string{},
			fields:   make(map[string]interface{}),
			election: i.Election,
		}
		setHostTagIfNotLoopback(m.tags, i.Host)

		m.name = "mysql_user_status"

		for key, value := range i.Tags {
			m.tags[key] = value
		}

		m.tags["user"] = user

		for k, v := range i.mUserStatusVariable[user] {
			m.fields[k] = v
		}

		if _, ok := i.mUserStatusConnection[user]["current_connect"]; ok {
			m.fields["current_connect"] = i.mUserStatusConnection[user]["current_connect"]
		}

		if _, ok := i.mUserStatusConnection[user]["total_connect"]; ok {
			m.fields["total_connect"] = i.mUserStatusConnection[user]["total_connect"]
		}

		if len(m.fields) == 0 {
			continue
		}

		ms = append(ms, m)
	}

	if len(ms) > 0 {
		pts := getPointsFromMeasurement(ms)
		return pts, nil
	}
	return []*gcPoint.Point{}, nil
}

func (i *Input) buildMysqlDbmMetric() ([]*gcPoint.Point, error) {
	ms := []inputs.MeasurementV2{}

	for _, row := range i.dbmMetricRows {
		m := &dbmStateMeasurement{
			name: "mysql_dbm_metric",
			tags: map[string]string{
				"service": "mysql",
				"status":  "info",
			},
			fields:   make(map[string]interface{}),
			election: i.Election,
		}
		setHostTagIfNotLoopback(m.tags, i.Host)

		if len(row.digestText) > 0 {
			m.fields["message"] = row.digestText
		}

		if len(row.digest) > 0 {
			m.tags["digest"] = row.digest
		}

		if len(row.schemaName) > 0 {
			m.tags["schema_name"] = row.schemaName
		}

		if len(row.querySignature) > 0 {
			m.tags["query_signature"] = row.querySignature
		}

		m.fields["count_star"] = row.countStar
		m.fields["sum_timer_wait"] = row.sumTimerWait
		m.fields["sum_lock_time"] = row.sumLockTime
		m.fields["sum_errors"] = row.sumErrors
		m.fields["sum_rows_affected"] = row.sumRowsAffected
		m.fields["sum_rows_sent"] = row.sumRowsSent
		m.fields["sum_rows_examined"] = row.sumRowsExamined
		m.fields["sum_select_scan"] = row.sumSelectScan
		m.fields["sum_select_full_join"] = row.sumSelectFullJoin
		m.fields["sum_no_index_used"] = row.sumNoIndexUsed
		m.fields["sum_no_good_index_used"] = row.sumNoGoodIndexUsed

		for key, value := range i.Tags {
			m.tags[key] = value
		}

		ms = append(ms, m)
	}

	if len(ms) > 0 {
		pts := getPointsFromMeasurement(ms)
		return pts, nil
	}
	return []*gcPoint.Point{}, nil
}

func (i *Input) buildMysqlDbmSample() ([]*gcPoint.Point, error) {
	ms := []inputs.MeasurementV2{}

	for _, plan := range i.dbmSamplePlans {
		tags := map[string]string{
			"service":           "mysql",
			"current_schema":    plan.currentSchema,
			"plan_definition":   plan.planDefinition,
			"plan_signature":    plan.planSignature,
			"query_signature":   plan.querySignature,
			"resource_hash":     plan.resourceHash,
			"digest_text":       plan.digestText,
			"query_truncated":   plan.queryTruncated,
			"network_client_ip": plan.networkClientIP,
			"digest":            plan.digest,
			"processlist_db":    plan.processlistDB,
			"processlist_user":  plan.processlistUser,
			"status":            "info",
		}
		setHostTagIfNotLoopback(tags, i.Host)

		// append extra tags
		for key, value := range i.Tags {
			tags[key] = value
		}

		fields := map[string]interface{}{
			"timestamp":              plan.timestamp,
			"duration":               plan.duration,
			"lock_time_ns":           plan.lockTimeNs,
			"no_good_index_used":     plan.noGoodIndexUsed,
			"no_index_used":          plan.noIndexUsed,
			"rows_affected":          plan.rowsAffected,
			"rows_examined":          plan.rowsExamined,
			"rows_sent":              plan.rowsSent,
			"select_full_join":       plan.selectFullJoin,
			"select_full_range_join": plan.selectFullRangeJoin,
			"select_range":           plan.selectRange,
			"select_range_check":     plan.selectRangeCheck,
			"select_scan":            plan.selectScan,
			"sort_merge_passes":      plan.sortMergePasses,
			"sort_range":             plan.sortRange,
			"sort_rows":              plan.sortRows,
			"sort_scan":              plan.sortScan,
			"timer_wait_ns":          plan.duration,
			"message":                plan.digestText,
		}

		m := &dbmSampleMeasurement{
			name:     "mysql_dbm_sample",
			tags:     tags,
			fields:   fields,
			election: i.Election,
		}
		ms = append(ms, m)
	}

	if len(ms) > 0 {
		pts := getPointsFromMeasurement(ms)
		return pts, nil
	}

	return []*gcPoint.Point{}, nil
}

func (i *Input) buildMysqlCustomQueries() ([]*gcPoint.Point, error) {
	ms := []inputs.MeasurementV2{}

	for hs, items := range i.mCustomQueries {
		var qy *customQuery
		for _, v := range i.Query {
			if hs == v.md5Hash {
				qy = v
				break
			}
		}
		if qy == nil {
			continue
		}

		for _, item := range items {
			m := &customerMeasurement{
				name:     qy.metric,
				tags:     map[string]string{},
				fields:   make(map[string]interface{}),
				election: i.Election,
			}
			setHostTagIfNotLoopback(m.tags, i.Host)

			for key, value := range i.Tags {
				m.tags[key] = value
			}

			if len(qy.tags) > 0 && len(qy.fields) == 0 {
				for _, tgKey := range qy.tags {
					if value, ok := item[tgKey]; ok {
						m.tags[tgKey] = cast.ToString(value)
						delete(item, tgKey)
					}
				}
				m.fields = item
			}

			if len(qy.tags) > 0 && len(qy.fields) > 0 {
				for _, tgKey := range qy.tags {
					if value, ok := item[tgKey]; ok {
						m.tags[tgKey] = cast.ToString(value)
						delete(item, tgKey)
					}
				}

				for _, fdKey := range qy.fields {
					if value, ok := item[fdKey]; ok {
						m.fields[fdKey] = value
					}
				}
			}

			if len(qy.tags) == 0 && len(qy.fields) == 0 {
				m.fields = item
			}
			m.ts = time.Now()

			if len(m.fields) > 0 {
				ms = append(ms, m)
			}
		}
	}

	if len(ms) > 0 {
		pts := getPointsFromMeasurement(ms)
		return pts, nil
	}
	return []*gcPoint.Point{}, nil
}

func getPointsFromMeasurement(ms []inputs.MeasurementV2) []*gcPoint.Point {
	pts := []*gcPoint.Point{}
	for _, m := range ms {
		pts = append(pts, m.Point())
	}

	return pts
}

func getMetricFields(fields map[string]interface{}, info *inputs.MeasurementInfo) map[string]interface{} {
	if info == nil {
		return fields
	}
	newFields := map[string]interface{}{}

	for k, v := range fields {
		if _, ok := info.Fields[k]; ok {
			newFields[k] = v
		}
	}

	return newFields
}
