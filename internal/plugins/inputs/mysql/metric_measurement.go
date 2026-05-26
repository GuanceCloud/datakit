// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

//nolint:lll // Metric descriptions are intentionally long for clarity
package mysql

import (
	"github.com/GuanceCloud/cliutils/point"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs"
)

const (
	measurementMySQL = "mysql_metric"

	TagGroupBase          = "base"
	TagGroupReplication   = "replication"
	TagGroupSchema        = "schema"
	TagGroupTableSchema   = "table_schema"
	TagGroupUserStatus    = "user_status"
	TagGroupInnodb        = "innodb"
	TagGroupDbmMetric     = "dbm_metric"
	TagGroupDbmSession    = "dbm_session"
	TagGroupDbmConnection = "dbm_connection"
)

type mysqlMeasurement struct{}

//nolint:funlen
func (m *mysqlMeasurement) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name:   measurementMySQL,
		Desc:   "Metric set including MySQL server, replication, schema, table schema, user status, InnoDB, and DBM (metric/session/connection) statistics, unified in v2",
		DescZh: "指标集包含 MySQL 基础、复制、schema、表级、用户状态、InnoDB 以及 DBM（mysql_dbm_metric/session/connection）相关指标，v2 版本统一为 mysql_metric",
		Cat:    point.Metric,
		Tags:   m.getTags(),
		Fields: m.getFields(),
	}
}

func (m *mysqlMeasurement) getTags() map[string]interface{} {
	return mergeMetricMaps(
		m.getCommonTags(),
		m.getBaseTags(),
		m.getReplicationTags(),
		m.getSchemaTags(),
		m.getTableSchemaTags(),
		m.getUserStatusTags(),
		m.getInnodbTags(),
		m.getDbmMetricTags(),
		m.getDbmSessionTags(),
		m.getDbmConnectionTags(),
	)
}

func (m *mysqlMeasurement) getCommonTags() map[string]interface{} {
	tags := make(map[string]interface{})
	tags["server"] = &inputs.TagInfo{Desc: "The address of the server. The value is `host:port`. Common tag."}
	tags["host"] = &inputs.TagInfo{Desc: "The server host address. Common tag."}
	tags["database_instance"] = &inputs.TagInfo{Desc: "MySQL instance identifier from configured tag or @@server_uuid. Common tag."}
	return tags
}

func (m *mysqlMeasurement) getBaseTags() map[string]interface{} {
	return m.filterCommonMetricTags(baseMeasurementInfo.Tags)
}

func (m *mysqlMeasurement) getReplicationTags() map[string]interface{} {
	return m.filterCommonMetricTags(replicationMeasurementInfo.Tags)
}

func (m *mysqlMeasurement) getInnodbTags() map[string]interface{} {
	return m.filterCommonMetricTags(innoDBMeasurementInfo.Tags)
}

func (m *mysqlMeasurement) getSchemaTags() map[string]interface{} {
	return m.filterCommonMetricTags((&schemaMeasurement{}).Info().Tags)
}

func (m *mysqlMeasurement) getTableSchemaTags() map[string]interface{} {
	return m.filterCommonMetricTags((&tbMeasurement{}).Info().Tags)
}

func (m *mysqlMeasurement) getUserStatusTags() map[string]interface{} {
	return m.filterCommonMetricTags((&userMeasurement{}).Info().Tags)
}

func (m *mysqlMeasurement) getDbmMetricTags() map[string]interface{} {
	return m.filterCommonMetricTags((&dbmStateMeasurement{}).Info().Tags)
}

func (m *mysqlMeasurement) getDbmSessionTags() map[string]interface{} {
	return m.filterCommonMetricTags((&dbmSessionMeasurement{}).Info().Tags)
}

func (m *mysqlMeasurement) getDbmConnectionTags() map[string]interface{} {
	return m.filterCommonMetricTags((&dbmConnectionMeasurement{}).Info().Tags)
}

func (m *mysqlMeasurement) filterCommonMetricTags(tags map[string]interface{}) map[string]interface{} {
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

func (m *mysqlMeasurement) getFields() map[string]interface{} {
	return mergeMetricMaps(
		m.getBaseFields(),
		m.getReplicationFields(),
		m.getSchemaFields(),
		m.getTableSchemaFields(),
		m.getUserStatusFields(),
		m.getInnodbFields(),
		m.getDbmMetricFields(),
		m.getDbmSessionFields(),
		m.getDbmConnectionFields(),
	)
}

func mergeMetricMaps(maps ...map[string]interface{}) map[string]interface{} {
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

func (m *mysqlMeasurement) addTaggedbyToFields(fields map[string]interface{}, tagGroup string) {
	var tags map[string]interface{}

	switch tagGroup {
	case TagGroupBase:
		tags = m.getBaseTags()
	case TagGroupReplication:
		tags = m.getReplicationTags()
	case TagGroupInnodb:
		tags = m.getInnodbTags()
	case TagGroupSchema:
		tags = m.getSchemaTags()
	case TagGroupTableSchema:
		tags = m.getTableSchemaTags()
	case TagGroupUserStatus:
		tags = m.getUserStatusTags()
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

func (m *mysqlMeasurement) getBaseFields() map[string]interface{} {
	fields := baseMeasurementInfo.Fields
	m.addTaggedbyToFields(fields, TagGroupBase)
	return fields
}

func (m *mysqlMeasurement) getReplicationFields() map[string]interface{} {
	fields := replicationMeasurementInfo.Fields
	m.addTaggedbyToFields(fields, TagGroupReplication)
	return fields
}

func (m *mysqlMeasurement) getSchemaFields() map[string]interface{} {
	fields := (&schemaMeasurement{}).Info().Fields
	m.addTaggedbyToFields(fields, TagGroupSchema)
	return fields
}

func (m *mysqlMeasurement) getTableSchemaFields() map[string]interface{} {
	fields := (&tbMeasurement{}).Info().Fields
	m.addTaggedbyToFields(fields, TagGroupTableSchema)
	return fields
}

func (m *mysqlMeasurement) getUserStatusFields() map[string]interface{} {
	fields := (&userMeasurement{}).Info().Fields
	m.addTaggedbyToFields(fields, TagGroupUserStatus)
	return fields
}

func (m *mysqlMeasurement) getInnodbFields() map[string]interface{} {
	fields := innoDBMeasurementInfo.Fields
	m.addTaggedbyToFields(fields, TagGroupInnodb)
	return fields
}

func (m *mysqlMeasurement) getDbmMetricFields() map[string]interface{} {
	fields := (&dbmStateMeasurement{}).Info().Fields
	m.addTaggedbyToFields(fields, TagGroupDbmMetric)
	return fields
}

func (m *mysqlMeasurement) getDbmSessionFields() map[string]interface{} {
	fields := (&dbmSessionMeasurement{}).Info().Fields
	m.addTaggedbyToFields(fields, TagGroupDbmSession)
	return fields
}

func (m *mysqlMeasurement) getDbmConnectionFields() map[string]interface{} {
	fields := (&dbmConnectionMeasurement{}).Info().Fields
	m.addTaggedbyToFields(fields, TagGroupDbmConnection)
	return fields
}
