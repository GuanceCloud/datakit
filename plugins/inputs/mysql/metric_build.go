package mysql

import (
	"time"

	"github.com/spf13/cast"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"
)

// metric build line proto

func (i *Input) buildMysql() ([]*io.Point, error) {
	m := &baseMeasurement{
		i:       i,
		resData: make(map[string]interface{}),
		tags:    make(map[string]string),
		fields:  make(map[string]interface{}),
	}

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
				return []*io.Point{}, err
			} else {
				m.fields[key] = val
			}
		}
	}

	if len(m.fields) > 0 {
		pts, err := inputs.GetPointsFromMeasurement([]inputs.Measurement{m})
		if err != nil {
			return []*io.Point{}, err
		}
		return pts, nil
	}

	return []*io.Point{}, nil
}

func (i *Input) buildMysqlSchema() ([]*io.Point, error) {
	ms := []inputs.Measurement{}

	// SchemaSize
	for k, v := range i.mSchemaSize {
		m := &schemaMeasurement{
			name:   "mysql_schema",
			tags:   make(map[string]string),
			fields: make(map[string]interface{}),
		}

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
			name:   "mysql_schema",
			tags:   make(map[string]string),
			fields: make(map[string]interface{}),
		}

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
		pts, err := inputs.GetPointsFromMeasurement(ms)
		if err != nil {
			return []*io.Point{}, err
		}
		return pts, nil
	}

	return []*io.Point{}, nil
}

func (i *Input) buildMysqlInnodb() ([]*io.Point, error) {
	ms := []inputs.Measurement{}

	m := &innodbMeasurement{
		tags:   make(map[string]string),
		fields: make(map[string]interface{}),
	}

	m.name = "mysql_innodb"

	for key, value := range i.Tags {
		m.tags[key] = value
	}

	for k, v := range i.mInnodb {
		m.fields[k] = v
	}

	ms = append(ms, m)

	if len(ms) > 0 {
		pts, err := inputs.GetPointsFromMeasurement(ms)
		if err != nil {
			return []*io.Point{}, err
		}
		return pts, nil
	}
	return []*io.Point{}, nil
}

func (i *Input) buildMysqlTableSchema() ([]*io.Point, error) {
	ms := []inputs.Measurement{}

	for _, v := range i.mTableSchema {
		m := &tbMeasurement{
			tags:   make(map[string]string),
			fields: make(map[string]interface{}),
		}

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
		pts, err := inputs.GetPointsFromMeasurement(ms)
		if err != nil {
			return []*io.Point{}, err
		}
		return pts, nil
	}

	return []*io.Point{}, nil
}

func (i *Input) buildMysqlUserStatus() ([]*io.Point, error) {
	ms := []inputs.Measurement{}

	for user := range i.mUserStatusName {
		m := &userMeasurement{
			tags:   make(map[string]string),
			fields: make(map[string]interface{}),
		}

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
		pts, err := inputs.GetPointsFromMeasurement(ms)
		if err != nil {
			return []*io.Point{}, err
		}
		return pts, nil
	}
	return []*io.Point{}, nil
}

func (i *Input) buildMysqlDbmMetric() ([]*io.Point, error) {
	ms := []inputs.Measurement{}

	now := time.Now()
	for _, row := range i.dbmMetricRows {
		m := &dbmStateMeasurement{
			name: dbmMetricName,
			tags: map[string]string{
				"service": "mysql_dbm_metric",
			},
			fields: make(map[string]interface{}),
			ts:     now,
		}

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
		pts, err := inputs.GetPointsFromMeasurement(ms)
		if err != nil {
			return []*io.Point{}, err
		}
		return pts, nil
	}
	return []*io.Point{}, nil
}

func (i *Input) buildMysqlDbmSample() ([]*io.Point, error) {
	ms := []inputs.Measurement{}

	now := time.Now()
	for _, plan := range i.dbmSamplePlans {
		tags := map[string]string{"service": "mysql_dbm_sample"}
		fields := make(map[string]interface{})
		tags["current_schema"] = plan.currentSchema
		tags["plan_definition"] = plan.planDefinition
		tags["plan_signature"] = plan.planSignature
		tags["query_signature"] = plan.querySignature
		tags["resource_hash"] = plan.resourceHash
		tags["digest_text"] = plan.digestText
		tags["query_truncated"] = plan.queryTruncated
		tags["network_client_ip"] = plan.networkClientIP
		tags["digest"] = plan.digest
		tags["processlist_db"] = plan.processlistDB
		tags["processlist_user"] = plan.processlistUser

		fields["timestamp"] = plan.timestamp
		fields["duration"] = plan.duration
		fields["lock_time_ns"] = plan.lockTimeNs
		fields["no_good_index_used"] = plan.noGoodIndexUsed
		fields["no_index_used"] = plan.noIndexUsed
		fields["rows_affected"] = plan.rowsAffected
		fields["rows_examined"] = plan.rowsExamined
		fields["rows_sent"] = plan.rowsSent
		fields["select_full_join"] = plan.selectFullJoin
		fields["select_full_range_join"] = plan.selectFullRangeJoin
		fields["select_range"] = plan.selectRange
		fields["select_range_check"] = plan.selectRangeCheck
		fields["select_scan"] = plan.selectScan
		fields["sort_merge_passes"] = plan.sortMergePasses
		fields["sort_range"] = plan.sortRange
		fields["sort_rows"] = plan.sortRows
		fields["sort_scan"] = plan.sortScan
		fields["timer_wait_ns"] = plan.duration
		fields["message"] = plan.digestText

		for key, value := range i.Tags {
			tags[key] = value
		}

		m := &dbmSampleMeasurement{
			name:   dbmMetricName,
			tags:   tags,
			fields: fields,
			ts:     now,
		}
		ms = append(ms, m)
	}

	if len(ms) > 0 {
		pts, err := inputs.GetPointsFromMeasurement(ms)
		if err != nil {
			return []*io.Point{}, err
		}
		return pts, nil
	}
	return []*io.Point{}, nil
}

func (i *Input) buildMysqlCustomQueries() ([]*io.Point, error) {
	ms := []inputs.Measurement{}

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
				name:   qy.metric,
				tags:   make(map[string]string),
				fields: make(map[string]interface{}),
			}

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
		pts, err := inputs.GetPointsFromMeasurement(ms)
		if err != nil {
			return []*io.Point{}, err
		}
		return pts, nil
	}
	return []*io.Point{}, nil
}
