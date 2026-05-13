// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

//nolint:lll // Metric descriptions are intentionally long for clarity
package postgresql

import (
	"github.com/GuanceCloud/cliutils/point"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs"
)

const (
	measurementPostgreSQL = "postgresql_metric"

	TagGroupDatabase        = "database"
	TagGroupFunction        = "function"
	TagGroupLock            = "lock"
	TagGroupStat            = "stat"
	TagGroupIndex           = "index"
	TagGroupSize            = "size"
	TagGroupStatIO          = "statio"
	TagGroupReplication     = "replication"
	TagGroupReplicationSlot = "replication_slot"
	TagGroupSlru            = "slru"
	TagGroupIO              = "io"
	TagGroupBgwriter        = "bgwriter"
	TagGroupConnection      = "connection"
	TagGroupConflict        = "conflict"
	TagGroupArchiver        = "archiver"
	TagGroupDbmMetric       = "dbm_metric"
	TagGroupDbmSession      = "dbm_session"
	TagGroupDbmConnection   = "dbm_connection"
)

var postgresqlMetricFieldDescOverrides = map[string]string{
	"blks_hit":      "Number of block cache hits. Emitted either as database-wide statistics (tagged by `db`) or SLRU cache statistics (tagged by `name`).",
	"blks_read":     "Number of disk blocks read. Emitted either as database-wide statistics (tagged by `db`) or SLRU cache statistics (tagged by `name`).",
	"idx_scan":      "Number of index scans. When `pg_index` is present the metric comes from per-index statistics; otherwise it comes from table-level statistics.",
	"idx_tup_fetch": "Number of live rows fetched by index scans. When `pg_index` is present the metric comes from per-index statistics; otherwise it comes from table-level statistics.",
}

type postgresqlMeasurement struct{}

func (m *postgresqlMeasurement) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name: measurementPostgreSQL,
		//nolint:lll
		Desc:   "Metric set including PostgreSQL database, function, lock, statistics, index, size, relation I/O, replication, replication slot, simple LRU cache, I/O, background writer, connection, conflict, archiver, and DBM (metric/session/connection) statistics, unified in v2",
		DescZh: "指标集包含 PostgreSQL database、function、lock、stat、index、size、relation I/O、replication、replication slot、simple LRU cache、I/O、background writer、connection、conflict、archiver 以及 DBM（metric/session/connection）相关指标，v2 版本统一为 postgresql_metric",
		Cat:    point.Metric,
		Tags:   m.getTags(),
		Fields: m.getFields(),
	}
}

func (m *postgresqlMeasurement) getTags() map[string]interface{} {
	return mergePostgreSQLMetricMaps(
		m.getCommonTags(),
		m.getDatabaseTags(),
		m.getFunctionTags(),
		m.getLockTags(),
		m.getStatTags(),
		m.getIndexTags(),
		m.getSizeTags(),
		m.getStatIOTags(),
		m.getReplicationTags(),
		m.getReplicationSlotTags(),
		m.getSlruTags(),
		m.getIOTags(),
		m.getBgwriterTags(),
		m.getConnectionTags(),
		m.getConflictTags(),
		m.getArchiverTags(),
		m.getDbmMetricTags(),
		m.getDbmSessionTags(),
		m.getDbmConnectionTags(),
	)
}

func (m *postgresqlMeasurement) getFields() map[string]interface{} {
	return mergePostgreSQLMetricMaps(
		m.getDatabaseFields(),
		m.getFunctionFields(),
		m.getLockFields(),
		m.getStatFields(),
		m.getIndexFields(),
		m.getSizeFields(),
		m.getStatIOFields(),
		m.getReplicationFields(),
		m.getReplicationSlotFields(),
		m.getSlruFields(),
		m.getIOFields(),
		m.getBgwriterFields(),
		m.getConnectionFields(),
		m.getConflictFields(),
		m.getArchiverFields(),
		m.getDbmMetricFields(),
		m.getDbmSessionFields(),
		m.getDbmConnectionFields(),
	)
}

func (m *postgresqlMeasurement) getCommonTags() map[string]interface{} {
	return map[string]interface{}{
		"server":            &inputs.TagInfo{Desc: "The address of the server. The value is `host:port`. Common tag."},
		"host":              &inputs.TagInfo{Desc: "The server host address. Common tag."},
		"database_instance": &inputs.TagInfo{Desc: "PostgreSQL instance identifier from configured tag `database_instance` or system_identifier. Common tag."},
	}
}

func (m *postgresqlMeasurement) filterCommonMetricTags(tags map[string]interface{}) map[string]interface{} {
	out := map[string]interface{}{}
	commonTags := m.getCommonTags()
	for k, v := range tags {
		if _, ok := commonTags[k]; ok {
			continue
		}
		out[k] = v
	}
	return out
}

func (m *postgresqlMeasurement) getDatabaseTags() map[string]interface{} {
	return m.filterCommonMetricTags((&inputMeasurement{}).Info().Tags)
}

func (m *postgresqlMeasurement) getFunctionTags() map[string]interface{} {
	return m.filterCommonMetricTags((&functionMeasurement{}).Info().Tags)
}

func (m *postgresqlMeasurement) getLockTags() map[string]interface{} {
	return m.filterCommonMetricTags((&lockMeasurement{}).Info().Tags)
}

func (m *postgresqlMeasurement) getStatTags() map[string]interface{} {
	return m.filterCommonMetricTags((&statMeasurement{}).Info().Tags)
}

func (m *postgresqlMeasurement) getIndexTags() map[string]interface{} {
	return m.filterCommonMetricTags((&indexMeasurement{}).Info().Tags)
}

func (m *postgresqlMeasurement) getSizeTags() map[string]interface{} {
	return m.filterCommonMetricTags((&sizeMeasurement{}).Info().Tags)
}

func (m *postgresqlMeasurement) getStatIOTags() map[string]interface{} {
	return m.filterCommonMetricTags((&statIOMeasurement{}).Info().Tags)
}

func (m *postgresqlMeasurement) getReplicationTags() map[string]interface{} {
	return m.filterCommonMetricTags((&replicationMeasurement{}).Info().Tags)
}

func (m *postgresqlMeasurement) getReplicationSlotTags() map[string]interface{} {
	return m.filterCommonMetricTags((&replicationSlotMeasurement{}).Info().Tags)
}

func (m *postgresqlMeasurement) getSlruTags() map[string]interface{} {
	return m.filterCommonMetricTags((&slruMeasurement{}).Info().Tags)
}

func (m *postgresqlMeasurement) getIOTags() map[string]interface{} {
	return m.filterCommonMetricTags((&ioMeasurement{}).Info().Tags)
}

func (m *postgresqlMeasurement) getBgwriterTags() map[string]interface{} {
	return m.filterCommonMetricTags((&bgwriterMeasurement{}).Info().Tags)
}

func (m *postgresqlMeasurement) getConnectionTags() map[string]interface{} {
	return m.filterCommonMetricTags((&connectionMeasurement{}).Info().Tags)
}

func (m *postgresqlMeasurement) getConflictTags() map[string]interface{} {
	return m.filterCommonMetricTags((&conflictMeasurement{}).Info().Tags)
}

func (m *postgresqlMeasurement) getArchiverTags() map[string]interface{} {
	return m.filterCommonMetricTags((&archiverMeasurement{}).Info().Tags)
}

func (m *postgresqlMeasurement) getDbmMetricTags() map[string]interface{} {
	return m.filterCommonMetricTags((&dbmMetricMeasurement{}).Info().Tags)
}

func (m *postgresqlMeasurement) getDbmSessionTags() map[string]interface{} {
	return m.filterCommonMetricTags((&dbmSessionMeasurement{}).Info().Tags)
}

func (m *postgresqlMeasurement) getDbmConnectionTags() map[string]interface{} {
	return m.filterCommonMetricTags((&dbmConnectionMeasurement{}).Info().Tags)
}

func mergePostgreSQLMetricMaps(maps ...map[string]interface{}) map[string]interface{} {
	result := make(map[string]interface{})
	for _, m := range maps {
		if m == nil {
			continue
		}
		for k, v := range m {
			result[k] = v
		}
	}
	return result
}

func (m *postgresqlMeasurement) addTaggedbyToFields(fields map[string]interface{}, tagGroup string) {
	var tags map[string]interface{}

	switch tagGroup {
	case TagGroupDatabase:
		tags = m.getDatabaseTags()
	case TagGroupFunction:
		tags = m.getFunctionTags()
	case TagGroupLock:
		tags = m.getLockTags()
	case TagGroupStat:
		tags = m.getStatTags()
	case TagGroupIndex:
		tags = m.getIndexTags()
	case TagGroupSize:
		tags = m.getSizeTags()
	case TagGroupStatIO:
		tags = m.getStatIOTags()
	case TagGroupReplication:
		tags = m.getReplicationTags()
	case TagGroupReplicationSlot:
		tags = m.getReplicationSlotTags()
	case TagGroupSlru:
		tags = m.getSlruTags()
	case TagGroupIO:
		tags = m.getIOTags()
	case TagGroupBgwriter:
		tags = m.getBgwriterTags()
	case TagGroupConnection:
		tags = m.getConnectionTags()
	case TagGroupConflict:
		tags = m.getConflictTags()
	case TagGroupArchiver:
		tags = m.getArchiverTags()
	case TagGroupDbmMetric:
		tags = m.getDbmMetricTags()
	case TagGroupDbmSession:
		tags = m.getDbmSessionTags()
	case TagGroupDbmConnection:
		tags = m.getDbmConnectionTags()
	default:
		return
	}

	taggedBy := make([]string, 0, len(tags))
	for tag := range tags {
		taggedBy = append(taggedBy, tag)
	}

	for _, field := range fields {
		if fieldInfo, ok := field.(*inputs.FieldInfo); ok {
			fieldInfo.Taggedby = taggedBy
		}
	}
}

func (m *postgresqlMeasurement) applyFieldDescOverrides(fields map[string]interface{}) map[string]interface{} {
	for fieldName, raw := range fields {
		fieldInfo, ok := raw.(*inputs.FieldInfo)
		if !ok {
			continue
		}
		if overrideDesc, ok := postgresqlMetricFieldDescOverrides[fieldName]; ok {
			fieldInfo.Desc = overrideDesc
		}
	}
	return fields
}

func (m *postgresqlMeasurement) getDatabaseFields() map[string]interface{} {
	fields := (&inputMeasurement{}).Info().Fields
	fields = m.applyFieldDescOverrides(fields)
	m.addTaggedbyToFields(fields, TagGroupDatabase)
	return fields
}

func (m *postgresqlMeasurement) getFunctionFields() map[string]interface{} {
	fields := (&functionMeasurement{}).Info().Fields
	fields = m.applyFieldDescOverrides(fields)
	m.addTaggedbyToFields(fields, TagGroupFunction)
	return fields
}

func (m *postgresqlMeasurement) getLockFields() map[string]interface{} {
	fields := (&lockMeasurement{}).Info().Fields
	fields = m.applyFieldDescOverrides(fields)
	m.addTaggedbyToFields(fields, TagGroupLock)
	return fields
}

func (m *postgresqlMeasurement) getStatFields() map[string]interface{} {
	fields := (&statMeasurement{}).Info().Fields
	fields = m.applyFieldDescOverrides(fields)
	m.addTaggedbyToFields(fields, TagGroupStat)
	return fields
}

func (m *postgresqlMeasurement) getIndexFields() map[string]interface{} {
	fields := (&indexMeasurement{}).Info().Fields
	fields = m.applyFieldDescOverrides(fields)
	m.addTaggedbyToFields(fields, TagGroupIndex)
	return fields
}

func (m *postgresqlMeasurement) getSizeFields() map[string]interface{} {
	fields := (&sizeMeasurement{}).Info().Fields
	fields = m.applyFieldDescOverrides(fields)
	m.addTaggedbyToFields(fields, TagGroupSize)
	return fields
}

func (m *postgresqlMeasurement) getStatIOFields() map[string]interface{} {
	fields := (&statIOMeasurement{}).Info().Fields
	fields = m.applyFieldDescOverrides(fields)
	m.addTaggedbyToFields(fields, TagGroupStatIO)
	return fields
}

func (m *postgresqlMeasurement) getReplicationFields() map[string]interface{} {
	fields := (&replicationMeasurement{}).Info().Fields
	fields = m.applyFieldDescOverrides(fields)
	m.addTaggedbyToFields(fields, TagGroupReplication)
	return fields
}

func (m *postgresqlMeasurement) getReplicationSlotFields() map[string]interface{} {
	fields := (&replicationSlotMeasurement{}).Info().Fields
	fields = m.applyFieldDescOverrides(fields)
	m.addTaggedbyToFields(fields, TagGroupReplicationSlot)
	return fields
}

func (m *postgresqlMeasurement) getSlruFields() map[string]interface{} {
	fields := (&slruMeasurement{}).Info().Fields
	fields = m.applyFieldDescOverrides(fields)
	m.addTaggedbyToFields(fields, TagGroupSlru)
	return fields
}

func (m *postgresqlMeasurement) getIOFields() map[string]interface{} {
	fields := (&ioMeasurement{}).Info().Fields
	fields = m.applyFieldDescOverrides(fields)
	m.addTaggedbyToFields(fields, TagGroupIO)
	return fields
}

func (m *postgresqlMeasurement) getBgwriterFields() map[string]interface{} {
	fields := (&bgwriterMeasurement{}).Info().Fields
	fields = m.applyFieldDescOverrides(fields)
	m.addTaggedbyToFields(fields, TagGroupBgwriter)
	return fields
}

func (m *postgresqlMeasurement) getConnectionFields() map[string]interface{} {
	fields := (&connectionMeasurement{}).Info().Fields
	fields = m.applyFieldDescOverrides(fields)
	m.addTaggedbyToFields(fields, TagGroupConnection)
	return fields
}

func (m *postgresqlMeasurement) getConflictFields() map[string]interface{} {
	fields := (&conflictMeasurement{}).Info().Fields
	fields = m.applyFieldDescOverrides(fields)
	m.addTaggedbyToFields(fields, TagGroupConflict)
	return fields
}

func (m *postgresqlMeasurement) getArchiverFields() map[string]interface{} {
	fields := (&archiverMeasurement{}).Info().Fields
	fields = m.applyFieldDescOverrides(fields)
	m.addTaggedbyToFields(fields, TagGroupArchiver)
	return fields
}

func (m *postgresqlMeasurement) getDbmMetricFields() map[string]interface{} {
	fields := (&dbmMetricMeasurement{}).Info().Fields
	fields = m.applyFieldDescOverrides(fields)
	m.addTaggedbyToFields(fields, TagGroupDbmMetric)
	m.clearTaggedbyForFields(fields, "total_calls", "delta_total_calls", "dbm_qps")
	return fields
}

func (m *postgresqlMeasurement) clearTaggedbyForFields(fields map[string]interface{}, fieldNames ...string) {
	for _, fieldName := range fieldNames {
		fieldInfo, ok := fields[fieldName].(*inputs.FieldInfo)
		if !ok {
			continue
		}
		fieldInfo.Taggedby = nil
	}
}

func (m *postgresqlMeasurement) getDbmSessionFields() map[string]interface{} {
	fields := (&dbmSessionMeasurement{}).Info().Fields
	fields = m.applyFieldDescOverrides(fields)
	m.addTaggedbyToFields(fields, TagGroupDbmSession)
	return fields
}

func (m *postgresqlMeasurement) getDbmConnectionFields() map[string]interface{} {
	fields := (&dbmConnectionMeasurement{}).Info().Fields
	fields = m.applyFieldDescOverrides(fields)
	m.addTaggedbyToFields(fields, TagGroupDbmConnection)
	return fields
}
