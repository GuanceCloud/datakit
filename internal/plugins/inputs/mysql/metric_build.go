// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package mysql

import (
	"fmt"
	"time"

	gcPoint "github.com/GuanceCloud/cliutils/point"
	"github.com/spf13/cast"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/ntp"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs"
)

// metric build line proto

func (ipt *Input) buildMysql() ([]*gcPoint.Point, error) {
	var pts []*gcPoint.Point
	opts := ipt.getKVsOpts()
	kvs := ipt.getKVs()

	m := &baseMeasurement{
		i:       ipt,
		resData: make(map[string]interface{}),
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

	if ipt.objectMetric != nil {
		if v, ok := m.resData["Slow_queries"]; ok {
			ipt.objectMetric.SlowQueries = cast.ToInt64(v)
		}
		var costTimeSeconds float64

		if !ipt.objectMetric.Time.IsZero() {
			costTimeSeconds = time.Since(ipt.objectMetric.Time).Seconds()
		}

		if costTimeSeconds > 0 {
			// calculate QPS
			if v, ok := m.resData["Queries"]; ok {
				val := cast.ToInt64(v)
				if val >= ipt.objectMetric.Queries {
					ipt.objectMetric.QPS = float64(val-ipt.objectMetric.Queries) / costTimeSeconds
				}

				ipt.objectMetric.Queries = val
			}

			// calculate TPS
			var trans int64
			if v, ok := m.resData["Com_commit"]; ok {
				trans += cast.ToInt64(v)
			}

			if v, ok := m.resData["Com_rollback"]; ok {
				trans += cast.ToInt64(v)
			}

			if trans >= ipt.objectMetric.Trans {
				ipt.objectMetric.TPS = float64(trans-ipt.objectMetric.Trans) / costTimeSeconds
			}

			ipt.objectMetric.Trans = trans
		}

		ipt.objectMetric.Time = ntp.Now()
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
				return pts, err
			} else {
				kvs = kvs.Add(key, val, false, true)
			}
		} else {
			l.Warnf("field %q:%v not in list", key, value)
		}
	}

	if kvs.FieldCount() > 0 {
		pts = append(pts, gcPoint.NewPointV2(metricNameMySQL, kvs, opts...))
	}

	return pts, nil
}

func (ipt *Input) buildMysqlReplication() ([]*gcPoint.Point, error) {
	var pts []*gcPoint.Point
	opts := ipt.getKVsOpts()

	m := &replicationMeasurement{}

	kvs := ipt.getKVs()
	// Replication
	for k, v := range getMetricFields(ipt.mReplication, m.Info()) {
		kvs = kvs.Add(k, v, false, true)
	}

	// Group Replication
	for k, v := range getMetricFields(ipt.mGroupReplication, m.Info()) {
		kvs = kvs.Add(k, v, false, true)
	}

	if kvs.FieldCount() > 0 {
		pts = append(pts, gcPoint.NewPointV2(metricNameMySQLReplication, kvs, opts...))
	}

	return pts, nil
}

func (ipt *Input) buildMysqlReplicationLog() ([]*gcPoint.Point, error) {
	var pts []*gcPoint.Point
	opts := ipt.getKVsOpts(gcPoint.Logging)

	m := &replicationLogMeasurement{}

	kvs := ipt.getKVs()
	for k, v := range getMetricFields(ipt.mReplication, m.Info()) {
		kvs = kvs.Add(k, v, false, true)
	}

	if kvs.FieldCount() > 0 {
		pts = append(pts, gcPoint.NewPointV2(metricNameMySQLReplicationLog, kvs, opts...))
	}

	return pts, nil
}

func (ipt *Input) buildMysqlSchema() ([]*gcPoint.Point, error) {
	var pts []*gcPoint.Point
	opts := ipt.getKVsOpts()

	// SchemaSize
	for k, v := range ipt.mSchemaSize {
		kvs := ipt.getKVs()

		kvs = kvs.AddTag("schema_name", k)

		size := cast.ToFloat64(v)
		kvs = kvs.Add("schema_size", size, false, true)

		pts = append(pts, gcPoint.NewPointV2(metricNameMySQLSchema, kvs, opts...))
	}

	for k, v := range ipt.mSchemaQueryExecTime {
		kvs := ipt.getKVs()

		kvs = kvs.AddTag("schema_name", k)

		size := cast.ToInt64(v)
		kvs = kvs.Add("query_run_time_avg", size, false, true)

		pts = append(pts, gcPoint.NewPointV2(metricNameMySQLSchema, kvs, opts...))
	}

	return pts, nil
}

func (ipt *Input) buildMysqlInnodb() ([]*gcPoint.Point, error) {
	var pts []*gcPoint.Point
	opts := ipt.getKVsOpts()
	kvs := ipt.getKVs()

	m := &innodbMeasurement{}
	for k, v := range getMetricFields(ipt.mInnodb, m.Info()) {
		kvs = kvs.Add(k, v, false, true)
	}

	if kvs.FieldCount() > 0 {
		pts = append(pts, gcPoint.NewPointV2(metricNameMySQLInnodb, kvs, opts...))
	}
	return pts, nil
}

func (ipt *Input) buildMysqlTableSchema() ([]*gcPoint.Point, error) {
	var pts []*gcPoint.Point
	opts := ipt.getKVsOpts()

	for _, v := range ipt.mTableSchema {
		kvs := ipt.getKVs()

		for kk, vv := range v {
			switch kk {
			case "table_schema", "table_name", "table_type", "engine", "version":
				if vvv, ok := vv.(string); ok {
					kvs = kvs.AddTag(kk, vvv)
				}
			case "table_rows", "data_length", "index_length", "data_free":
				kvs = kvs.Add(kk, vv, false, true)
			}
		}

		if kvs.FieldCount() > 0 {
			pts = append(pts, gcPoint.NewPointV2(metricNameMySQLTableSchema, kvs, opts...))
		}
	}

	return pts, nil
}

func (ipt *Input) buildMysqlUserStatus() ([]*gcPoint.Point, error) {
	var pts []*gcPoint.Point
	opts := ipt.getKVsOpts()

	for user := range ipt.mUserStatusName {
		kvs := ipt.getKVs()

		// tags
		kvs = kvs.AddTag("user", user)

		// fields
		for k, v := range ipt.mUserStatusVariable[user] {
			kvs = kvs.Add(k, v, false, true)
		}

		if _, ok := ipt.mUserStatusConnection[user]["current_connect"]; ok {
			kvs = kvs.Add("current_connect", ipt.mUserStatusConnection[user]["current_connect"], false, true)
		}

		if _, ok := ipt.mUserStatusConnection[user]["total_connect"]; ok {
			kvs = kvs.Add("total_connect", ipt.mUserStatusConnection[user]["total_connect"], false, true)
		}

		if kvs.FieldCount() > 0 {
			pts = append(pts, gcPoint.NewPointV2(metricNameMySQLUserStatus, kvs, opts...))
		}
	}

	return pts, nil
}

func (ipt *Input) buildMysqlDbmMetric() ([]*gcPoint.Point, error) {
	var pts []*gcPoint.Point
	opts := ipt.getKVsOpts(gcPoint.Logging)

	for _, row := range ipt.dbmMetricRows {
		kvs := ipt.getKVs()

		// tags
		kvs = kvs.AddTag("service", "mysql")
		kvs = kvs.AddTag("status", "info")
		if len(row.schemaName) > 0 {
			kvs = kvs.AddTag("schema_name", row.schemaName)
		} else {
			kvs = kvs.AddTag("schema_name", "-")
		}

		// fields
		if len(row.digestText) > 0 {
			kvs = kvs.Add("message", row.digestText, false, true)
		}

		if len(row.digest) > 0 {
			kvs = kvs.Add("digest", row.digest, false, true)
		}

		if len(row.querySignature) > 0 {
			kvs = kvs.Add("query_signature", row.querySignature, false, true)
		}

		kvs = kvs.Add("count_star", row.countStar, false, true)
		kvs = kvs.Add("sum_timer_wait", row.sumTimerWait, false, true)
		kvs = kvs.Add("sum_lock_time", row.sumLockTime, false, true)
		kvs = kvs.Add("sum_errors", row.sumErrors, false, true)
		kvs = kvs.Add("sum_rows_affected", row.sumRowsAffected, false, true)
		kvs = kvs.Add("sum_rows_sent", row.sumRowsSent, false, true)
		kvs = kvs.Add("sum_rows_examined", row.sumRowsExamined, false, true)
		kvs = kvs.Add("sum_select_scan", row.sumSelectScan, false, true)
		kvs = kvs.Add("sum_select_full_join", row.sumSelectFullJoin, false, true)
		kvs = kvs.Add("sum_no_index_used", row.sumNoIndexUsed, false, true)
		kvs = kvs.Add("sum_no_good_index_used", row.sumNoGoodIndexUsed, false, true)

		pts = append(pts, gcPoint.NewPointV2(metricNameMySQLDbmMetric, kvs, opts...))
	}

	return pts, nil
}

func (ipt *Input) buildMysqlDbmSample() ([]*gcPoint.Point, error) {
	var pts []*gcPoint.Point
	opts := ipt.getKVsOpts(gcPoint.Logging)

	for _, plan := range ipt.dbmSamplePlans {
		kvs := ipt.getKVs()

		// tags
		kvs = kvs.AddTag("service", "mysql")
		kvs = kvs.AddTag("current_schema", plan.currentSchema)
		kvs = kvs.AddTag("plan_definition", plan.planDefinition)
		kvs = kvs.AddTag("plan_signature", plan.planSignature)
		kvs = kvs.AddTag("query_signature", plan.querySignature)
		kvs = kvs.AddTag("resource_hash", plan.resourceHash)
		kvs = kvs.AddTag("digest_text", plan.digestText)
		kvs = kvs.AddTag("query_truncated", plan.queryTruncated)
		kvs = kvs.AddTag("network_client_ip", plan.networkClientIP)
		kvs = kvs.AddTag("digest", plan.digest)
		kvs = kvs.AddTag("processlist_db", plan.processlistDB)
		kvs = kvs.AddTag("processlist_user", plan.processlistUser)
		kvs = kvs.AddTag("status", "info")
		// fields
		kvs = kvs.Add("timestamp", plan.timestamp, false, true)
		kvs = kvs.Add("duration", plan.duration, false, true)
		kvs = kvs.Add("lock_time_ns", plan.lockTimeNs, false, true)
		kvs = kvs.Add("no_good_index_used", plan.noGoodIndexUsed, false, true)
		kvs = kvs.Add("no_index_used", plan.noIndexUsed, false, true)
		kvs = kvs.Add("rows_affected", plan.rowsAffected, false, true)
		kvs = kvs.Add("rows_examined", plan.rowsExamined, false, true)
		kvs = kvs.Add("rows_sent", plan.rowsSent, false, true)
		kvs = kvs.Add("select_full_join", plan.selectFullJoin, false, true)
		kvs = kvs.Add("select_full_range_join", plan.selectFullRangeJoin, false, true)
		kvs = kvs.Add("select_range", plan.selectRange, false, true)
		kvs = kvs.Add("select_range_check", plan.selectRangeCheck, false, true)
		kvs = kvs.Add("select_scan", plan.selectScan, false, true)
		kvs = kvs.Add("sort_merge_passes", plan.sortMergePasses, false, true)
		kvs = kvs.Add("sort_range", plan.sortRange, false, true)
		kvs = kvs.Add("sort_rows", plan.sortRows, false, true)
		kvs = kvs.Add("sort_scan", plan.sortScan, false, true)
		kvs = kvs.Add("timer_wait_ns", plan.duration, false, true)
		kvs = kvs.Add("message", plan.digestText, false, true)

		pts = append(pts, gcPoint.NewPointV2(metricNameMySQLDbmSample, kvs, opts...))
	}

	return pts, nil
}

func (ipt *Input) getCustomQueryPoints(query *customQuery, arr []map[string]interface{}) []*gcPoint.Point {
	var pts []*gcPoint.Point

	if query == nil {
		return pts
	}

	opts := append(ipt.getKVsOpts(), gcPoint.WithTime(query.ptsTime)) // use custom query's aligned time

	for _, item := range arr {
		kvs := ipt.getKVs()

		for _, tgKey := range query.Tags {
			if value, ok := item[tgKey]; ok {
				kvs = kvs.AddTag(tgKey, cast.ToString(value))
				delete(item, tgKey)
			}
		}

		for _, fdKey := range query.Fields {
			if value, ok := item[fdKey]; ok {
				// transform all fields to float64
				kvs = kvs.Add(fdKey, cast.ToFloat64(value), false, true)
			}
		}

		if kvs.FieldCount() > 0 {
			pts = append(pts, gcPoint.NewPointV2(query.Metric, kvs, opts...))
		}
	}

	return pts
}

func (ipt *Input) buildMysqlCustomerObject() ([]*gcPoint.Point, error) {
	ms := []inputs.MeasurementV2{}
	version := ""
	if ipt.Version != nil {
		version = ipt.Version.version
	}
	fields := map[string]interface{}{
		"display_name": fmt.Sprintf("%s:%d", ipt.Host, ipt.Port),
		"uptime":       ipt.Uptime,
		"version":      version,
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

func (ipt *Input) getKVsOpts(categorys ...gcPoint.Category) []gcPoint.Option {
	var opts []gcPoint.Option

	category := gcPoint.Metric
	if len(categorys) > 0 {
		category = categorys[0]
	}

	switch category { //nolint:exhaustive
	case gcPoint.Logging:
		opts = gcPoint.DefaultLoggingOptions()
	case gcPoint.Metric:
		opts = gcPoint.DefaultMetricOptions()
	case gcPoint.Object:
		opts = gcPoint.DefaultObjectOptions()
	default:
		opts = gcPoint.DefaultMetricOptions()
	}

	if ipt.Election {
		opts = append(opts, gcPoint.WithExtraTags(datakit.GlobalElectionTags()))
	}

	opts = append(opts, gcPoint.WithTimestamp(ipt.ptsTime.UnixNano()))

	return opts
}

func (ipt *Input) getKVs() gcPoint.KVs {
	var kvs gcPoint.KVs

	// add extended tags
	for k, v := range ipt.mergedTags {
		kvs = kvs.AddTag(k, v)
	}

	return kvs
}
