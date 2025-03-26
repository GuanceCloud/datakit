// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package mysql

import (
	"fmt"

	gcPoint "github.com/GuanceCloud/cliutils/point"
	"github.com/spf13/cast"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs"
)

// metric build line proto

func (ipt *Input) buildMysql() ([]*gcPoint.Point, error) {
	m := &baseMeasurement{
		i:        ipt,
		resData:  make(map[string]interface{}),
		tags:     map[string]string{},
		fields:   make(map[string]interface{}),
		election: ipt.Election,
	}
	setHostTagIfNotLoopback(m.tags, ipt.Host)

	m.name = metricNameMySQL
	for key, value := range ipt.Tags {
		m.tags[key] = value
	}

	for k, v := range ipt.globalStatus {
		m.resData[k] = v
	}
	for k, v := range ipt.globalVariables {
		m.resData[k] = v
	}
	if ipt.binLogOn {
		m.resData["Binlog_space_usage_bytes"] = ipt.binlog["Binlog_space_usage_bytes"]
	}

	if hasKey(m.resData, "Key_blocks_unused") &&
		hasKey(m.resData, "key_cache_block_size") &&
		hasKey(m.resData, "key_buffer_size") {
		m.resData["Key_buffer_size"] = m.resData["key_buffer_size"]
		if keyBufferSize, ok := m.resData["key_buffer_size"].(int64); ok {
			if keyBufferSize != 0 {
				keyBlocksUnused, ok1 := m.resData["Key_blocks_unused"].(int64)
				keyCacheBlockSize, ok2 := m.resData["key_cache_block_size"].(int64)
				if ok1 && ok2 {
					m.resData["Key_cache_utilization"] = 1.0 - float64(keyBlocksUnused*keyCacheBlockSize)/float64(keyBufferSize)
					if hasKey(m.resData, "Key_blocks_used") {
						keyBlocksUsed, ok := m.resData["Key_blocks_used"].(int64)
						if ok {
							m.resData["Key_buffer_bytes_used"] = keyBlocksUsed * keyCacheBlockSize
						}
					}

					if hasKey(m.resData, "Key_blocks_not_flushed") {
						keyBufferBytesUnflushed, ok := m.resData["Key_blocks_not_flushed"].(int64)
						if ok {
							m.resData["Key_buffer_bytes_unflushed"] = keyBufferBytesUnflushed * keyCacheBlockSize
						}
					}
				}
			}
		}
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
		} else {
			l.Warnf("field %q:%v not in list", key, value)
		}
	}

	if len(m.fields) > 0 {
		pts := getPointsFromMeasurement([]inputs.MeasurementV2{m})
		return pts, nil
	}

	return []*gcPoint.Point{}, nil
}

func (ipt *Input) buildMysqlReplication() ([]*gcPoint.Point, error) {
	ms := []inputs.MeasurementV2{}

	m := &replicationMeasurement{
		tags:     map[string]string{},
		resData:  make(map[string]interface{}),
		fields:   make(map[string]interface{}),
		election: ipt.Election,
	}
	setHostTagIfNotLoopback(m.tags, ipt.Host)

	m.name = metricNameMySQLReplication

	for key, value := range ipt.Tags {
		m.tags[key] = value
	}

	// Replication
	m.fields = getMetricFields(ipt.mReplication, m.Info())

	// Group Replication
	m.resData = getMetricFields(ipt.mGroupReplication, m.Info())

	for k, v := range m.resData {
		m.fields[k] = v
	}

	if len(m.fields) > 0 {
		ms = append(ms, m)
	}

	if len(ms) > 0 {
		pts := getPointsFromMeasurement(ms)
		return pts, nil
	}
	return []*gcPoint.Point{}, nil
}

func (ipt *Input) buildMysqlReplicationLog() ([]*gcPoint.Point, error) {
	ms := []inputs.MeasurementV2{}

	m := &replicationLogMeasurement{
		tags:     map[string]string{},
		fields:   make(map[string]interface{}),
		election: ipt.Election,
	}
	setHostTagIfNotLoopback(m.tags, ipt.Host)

	m.name = metricNameMySQLReplicationLog

	for key, value := range ipt.Tags {
		m.tags[key] = value
	}

	m.fields = getMetricFields(ipt.mReplication, m.Info())

	if len(m.fields) > 0 {
		ms = append(ms, m)
	}

	if len(ms) > 0 {
		pts := getPointsFromMeasurement(ms)
		return pts, nil
	}
	return []*gcPoint.Point{}, nil
}

func (ipt *Input) buildMysqlSchema() ([]*gcPoint.Point, error) {
	ms := []inputs.MeasurementV2{}

	// SchemaSize
	for k, v := range ipt.mSchemaSize {
		m := &schemaMeasurement{
			name:     metricNameMySQLSchema,
			tags:     map[string]string{},
			fields:   make(map[string]interface{}),
			election: ipt.Election,
		}
		setHostTagIfNotLoopback(m.tags, ipt.Host)

		for key, value := range ipt.Tags {
			m.tags[key] = value
		}

		size := cast.ToFloat64(v)

		m.fields["schema_size"] = size
		m.tags["schema_name"] = k
		m.ts = ipt.start.UnixNano()

		if len(m.fields) > 0 {
			ms = append(ms, m)
		}
	}

	for k, v := range ipt.mSchemaQueryExecTime {
		m := &schemaMeasurement{
			name:     metricNameMySQLSchema,
			tags:     make(map[string]string),
			fields:   make(map[string]interface{}),
			election: ipt.Election,
		}
		setHostTagIfNotLoopback(m.tags, ipt.Host)

		for key, value := range ipt.Tags {
			m.tags[key] = value
		}

		size := cast.ToInt64(v)

		m.fields["query_run_time_avg"] = size
		m.tags["schema_name"] = k
		m.ts = ipt.start.UnixNano()

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

func (ipt *Input) buildMysqlInnodb() ([]*gcPoint.Point, error) {
	ms := []inputs.MeasurementV2{}

	m := &innodbMeasurement{
		tags:     map[string]string{},
		fields:   make(map[string]interface{}),
		election: ipt.Election,
	}
	setHostTagIfNotLoopback(m.tags, ipt.Host)

	m.name = metricNameMySQLInnodb

	for key, value := range ipt.Tags {
		m.tags[key] = value
	}

	m.fields = getMetricFields(ipt.mInnodb, m.Info())

	ms = append(ms, m)

	if len(ms) > 0 {
		pts := getPointsFromMeasurement(ms)
		return pts, nil
	}
	return []*gcPoint.Point{}, nil
}

func (ipt *Input) buildMysqlTableSchema() ([]*gcPoint.Point, error) {
	ms := []inputs.MeasurementV2{}

	for _, v := range ipt.mTableSchema {
		m := &tbMeasurement{
			tags:     map[string]string{},
			fields:   make(map[string]interface{}),
			election: ipt.Election,
		}
		setHostTagIfNotLoopback(m.tags, ipt.Host)

		m.name = metricNameMySQLTableSchema

		for key, value := range ipt.Tags {
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

func (ipt *Input) buildMysqlUserStatus() ([]*gcPoint.Point, error) {
	ms := []inputs.MeasurementV2{}

	for user := range ipt.mUserStatusName {
		m := &userMeasurement{
			tags:     map[string]string{},
			fields:   make(map[string]interface{}),
			election: ipt.Election,
		}
		setHostTagIfNotLoopback(m.tags, ipt.Host)

		m.name = metricNameMySQLUserStatus

		for key, value := range ipt.Tags {
			m.tags[key] = value
		}

		m.tags["user"] = user

		for k, v := range ipt.mUserStatusVariable[user] {
			m.fields[k] = v
		}

		if _, ok := ipt.mUserStatusConnection[user]["current_connect"]; ok {
			m.fields["current_connect"] = ipt.mUserStatusConnection[user]["current_connect"]
		}

		if _, ok := ipt.mUserStatusConnection[user]["total_connect"]; ok {
			m.fields["total_connect"] = ipt.mUserStatusConnection[user]["total_connect"]
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

func (ipt *Input) buildMysqlDbmMetric() ([]*gcPoint.Point, error) {
	ms := []inputs.MeasurementV2{}

	for _, row := range ipt.dbmMetricRows {
		m := &dbmStateMeasurement{
			name: metricNameMySQLDbmMetric,
			tags: map[string]string{
				"service": "mysql",
				"status":  "info",
			},
			fields:   make(map[string]interface{}),
			election: ipt.Election,
		}
		setHostTagIfNotLoopback(m.tags, ipt.Host)

		if len(row.digestText) > 0 {
			m.fields["message"] = row.digestText
		}

		if len(row.digest) > 0 {
			m.tags["digest"] = row.digest
		}

		if len(row.schemaName) > 0 {
			m.tags["schema_name"] = row.schemaName
		} else {
			m.tags["schema_name"] = "-"
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

		for key, value := range ipt.Tags {
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

func (ipt *Input) buildMysqlDbmSample() ([]*gcPoint.Point, error) {
	ms := []inputs.MeasurementV2{}

	for _, plan := range ipt.dbmSamplePlans {
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
		setHostTagIfNotLoopback(tags, ipt.Host)

		// append extra tags
		for key, value := range ipt.Tags {
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
			name:     metricNameMySQLDbmSample,
			tags:     tags,
			fields:   fields,
			election: ipt.Election,
		}
		ms = append(ms, m)
	}

	if len(ms) > 0 {
		pts := getPointsFromMeasurement(ms)
		return pts, nil
	}

	return []*gcPoint.Point{}, nil
}

func (ipt *Input) buildMysqlCustomQueries() ([]*gcPoint.Point, error) {
	ms := []inputs.MeasurementV2{}

	for hs, items := range ipt.mCustomQueries {
		var qy *customQuery
		for _, v := range ipt.Query {
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
				name:     qy.Metric,
				tags:     map[string]string{},
				fields:   make(map[string]interface{}),
				election: ipt.Election,
			}
			setHostTagIfNotLoopback(m.tags, ipt.Host)

			for key, value := range ipt.Tags {
				m.tags[key] = value
			}

			for _, tgKey := range qy.Tags {
				if value, ok := item[tgKey]; ok {
					m.tags[tgKey] = cast.ToString(value)
					delete(item, tgKey)
				}
			}

			for _, fdKey := range qy.Fields {
				if value, ok := item[fdKey]; ok {
					// transform all fields to float64
					m.fields[fdKey] = cast.ToFloat64(value)
				}
			}

			m.ts = ipt.start.UnixNano()

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

func (ipt *Input) buildMysqlCustomerObject() ([]*gcPoint.Point, error) {
	ms := []inputs.MeasurementV2{}
	fields := map[string]interface{}{
		"display_name": fmt.Sprintf("%s:%d", ipt.Host, ipt.Port),
		"uptime":       ipt.Uptime,
		"version":      ipt.Version,
	}
	tags := map[string]string{
		"name":          fmt.Sprintf("mysql-%s:%d", ipt.Host, ipt.Port),
		"host":          ipt.Host,
		"ip":            fmt.Sprintf("%s:%d", ipt.Host, ipt.Port),
		"col_co_status": ipt.CollectCoStatus,
	}
	m := &customerObjectMeasurement{
		name:     "database",
		tags:     tags,
		fields:   fields,
		election: ipt.Election,
	}
	ipt.LastCustomerObject = m
	ms = append(ms, m)
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
