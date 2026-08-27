// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package mongodb

import (
	"github.com/GuanceCloud/cliutils/point"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs"
)

type mongodbDBMMetricMeasurement struct{}

func (*mongodbDBMMetricMeasurement) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name:   mongodbDBMMetricName,
		Cat:    point.Metric,
		Desc:   "MongoDB Query Metrics collected from $queryStats. Original fields are cumulative values and delta_ fields are changes during the collection interval.",
		DescZh: "通过 $queryStats 采集的 MongoDB Query Metrics，原始字段表示累计值，delta_ 字段表示当前采集周期内的增量。",
		Tags: map[string]interface{}{
			"mongod_host":       inputs.NewTagInfo("MongoDB server host."),
			"server":            inputs.NewTagInfo("MongoDB server address."),
			"database_instance": inputs.NewTagInfo("MongoDB instance identifier."),
			"database_type":     inputs.NewTagInfo("Database type. The value is `MongoDB`."),
			"database_name":     inputs.NewTagInfo("Database name from the query namespace."),
			"collection":        inputs.NewTagInfo("Collection name from the query namespace."),
			"command_type":      inputs.NewTagInfo("MongoDB command type, such as find or aggregate."),
			"query_signature":   inputs.NewTagInfo("xxhash signature generated from database_name, collection, and MongoDB query_shape_hash."),
			"query_shape_hash":  inputs.NewTagInfo("MongoDB-native hash of the query shape."),
			"normalized_query_hash": inputs.NewTagInfo(
				"Hash computed from the complete obfuscated MongoDB command for linking DBM data.",
			),
			"query_text":      inputs.NewTagInfo("A configurable UTF-8-safe prefix of the obfuscated MongoDB command."),
			"query_truncated": inputs.NewTagInfo("Whether query_text was truncated to the configured byte limit."),
		},
		Fields: map[string]interface{}{
			"exec_count": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				Desc:     "Cumulative number of query executions returned by MongoDB.",
			},
			"total_exec_micros_sum": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.DurationUS,
				Desc:     "Cumulative query execution time returned by MongoDB.",
			},
			"first_response_exec_micros_sum": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.DurationUS,
				Desc:     "Cumulative time to first response returned by MongoDB.",
			},
			"keys_examined_sum": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				Desc:     "Cumulative number of index keys examined returned by MongoDB.",
			},
			"docs_examined_sum": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				Desc:     "Cumulative number of documents examined returned by MongoDB.",
			},
			"docs_returned_sum": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				Desc:     "Cumulative number of documents returned by MongoDB.",
			},
			"bytes_read_sum": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.SizeByte,
				Desc:     "Cumulative bytes read returned by MongoDB.",
			},
			"cpu_nanos_sum": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.DurationNS,
				Desc:     "Cumulative CPU time returned by MongoDB.",
			},
			"used_disk_count": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				Desc:     "Cumulative number of executions that used disk returned by MongoDB.",
			},
			"has_sort_stage_count": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				Desc:     "Cumulative number of executions that used a sort stage returned by MongoDB.",
			},
			"read_time_micros_sum": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.DurationUS,
				Desc:     "Cumulative time spent reading data returned by MongoDB.",
			},
			"working_time_millis_sum": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.DurationMS,
				Desc:     "Cumulative working time returned by MongoDB.",
			},
			"delta_exec_count": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Count,
				Unit:     inputs.NCount,
				Desc:     "Change in query executions during the collection interval.",
			},
			"delta_total_exec_micros_sum": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Count,
				Unit:     inputs.DurationUS,
				Desc:     "Change in query execution time during the collection interval.",
			},
			"delta_first_response_exec_micros_sum": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Count,
				Unit:     inputs.DurationUS,
				Desc:     "Change in time to first response during the collection interval.",
			},
			"delta_keys_examined_sum": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Count,
				Unit:     inputs.NCount,
				Desc:     "Change in index keys examined during the collection interval.",
			},
			"delta_docs_examined_sum": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Count,
				Unit:     inputs.NCount,
				Desc:     "Change in documents examined during the collection interval.",
			},
			"delta_docs_returned_sum": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Count,
				Unit:     inputs.NCount,
				Desc:     "Change in documents returned during the collection interval.",
			},
			"delta_bytes_read_sum": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Count,
				Unit:     inputs.SizeByte,
				Desc:     "Change in bytes read during the collection interval.",
			},
			"delta_cpu_nanos_sum": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Count,
				Unit:     inputs.DurationNS,
				Desc:     "Change in CPU time during the collection interval.",
			},
			"delta_used_disk_count": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Count,
				Unit:     inputs.NCount,
				Desc:     "Change in executions that used disk during the collection interval.",
			},
			"delta_has_sort_stage_count": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Count,
				Unit:     inputs.NCount,
				Desc:     "Change in executions that used a sort stage during the collection interval.",
			},
			"delta_read_time_micros_sum": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Count,
				Unit:     inputs.DurationUS,
				Desc:     "Change in time spent reading data during the collection interval.",
			},
			"delta_working_time_millis_sum": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Count,
				Unit:     inputs.DurationMS,
				Desc:     "Change in working time during the collection interval.",
			},
		},
	}
}

type mongodbDBMActivityMeasurement struct{}

type mongodbDBMOperationMeasurement struct{}

type mongodbDBMSlowQueryMeasurement struct{}

func (*mongodbDBMSlowQueryMeasurement) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name:   mongodbDBMSlowQueryName,
		Cat:    point.Logging,
		Desc:   "MongoDB completed slow operations collected from each database's system.profile collection.",
		DescZh: "从各数据库 system.profile 集合采集的 MongoDB 已完成慢操作。",
		Tags: map[string]interface{}{
			"mongod_host":       inputs.NewTagInfo("MongoDB server host."),
			"server":            inputs.NewTagInfo("MongoDB server address."),
			"database_instance": inputs.NewTagInfo("MongoDB instance identifier."),
			"database_type":     inputs.NewTagInfo("Database type. The value is `MongoDB`."),
			"database_name":     inputs.NewTagInfo("Database name of the slow operation."),
			"collection":        inputs.NewTagInfo("Collection name of the slow operation."),
			"command_type":      inputs.NewTagInfo("MongoDB command type, such as find, aggregate, update, or getMore."),
			"operation":         inputs.NewTagInfo("MongoDB operation type reported by system.profile."),
			"plan_summary":      inputs.NewTagInfo("Summary of the execution plan, such as COLLSCAN or IXSCAN."),
			"query_shape_hash":  inputs.NewTagInfo("MongoDB queryShapeHash when it is available."),
			"query_hash":        inputs.NewTagInfo("MongoDB queryHash or planCacheShapeHash."),
			"query_signature": inputs.NewTagInfo(
				"xxhash signature generated from database_name, collection, and query_shape_hash when queryShapeHash is available.",
			),
			"normalized_query_hash": inputs.NewTagInfo(
				"Hash of the canonical obfuscated command used to group equivalent slow operations.",
			),
			"user":        inputs.NewTagInfo("MongoDB user that executed the operation."),
			"application": inputs.NewTagInfo("Client application name."),
		},
		Fields: map[string]interface{}{
			"message": &inputs.FieldInfo{
				DataType: inputs.String,
				Type:     inputs.String,
				Unit:     inputs.NoUnit,
				Desc:     "Complete obfuscated MongoDB command.",
			},
			"namespace": &inputs.FieldInfo{
				DataType: inputs.String,
				Type:     inputs.String,
				Unit:     inputs.NoUnit,
				Desc:     "MongoDB namespace of the slow operation.",
			},
			"plan_cache_key": &inputs.FieldInfo{
				DataType: inputs.String,
				Type:     inputs.String,
				Unit:     inputs.NoUnit,
				Desc:     "MongoDB plan cache key.",
			},
			"query_framework": &inputs.FieldInfo{
				DataType: inputs.String,
				Type:     inputs.String,
				Unit:     inputs.NoUnit,
				Desc:     "Query execution framework reported by MongoDB.",
			},
			"client": &inputs.FieldInfo{
				DataType: inputs.String,
				Type:     inputs.String,
				Unit:     inputs.NoUnit,
				Desc:     "Client address that executed the operation.",
			},
			"query_truncated": &inputs.FieldInfo{
				DataType: inputs.String,
				Type:     inputs.String,
				Unit:     inputs.NoUnit,
				Desc:     "Whether MongoDB reported the command as truncated.",
			},
			"duration_ms": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.DurationMS,
				Desc:     "Total duration of the slow operation.",
			},
			"working_millis": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.DurationMS,
				Desc:     "Time spent working on the operation.",
			},
			"num_yields": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				Desc:     "Number of times the operation yielded.",
			},
			"response_length": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.SizeByte,
				Desc:     "Response size in bytes.",
			},
			"docs_returned": &inputs.FieldInfo{
				DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.NCount,
				Desc: "Number of documents returned.",
			},
			"docs_matched": &inputs.FieldInfo{
				DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.NCount,
				Desc: "Number of documents matched.",
			},
			"docs_modified": &inputs.FieldInfo{
				DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.NCount,
				Desc: "Number of documents modified.",
			},
			"docs_inserted": &inputs.FieldInfo{
				DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.NCount,
				Desc: "Number of documents inserted.",
			},
			"docs_deleted": &inputs.FieldInfo{
				DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.NCount,
				Desc: "Number of documents deleted.",
			},
			"keys_examined": &inputs.FieldInfo{
				DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.NCount,
				Desc: "Number of index keys examined.",
			},
			"docs_examined": &inputs.FieldInfo{
				DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.NCount,
				Desc: "Number of documents examined.",
			},
			"keys_inserted": &inputs.FieldInfo{
				DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.NCount,
				Desc: "Number of index keys inserted.",
			},
			"write_conflicts": &inputs.FieldInfo{
				DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.NCount,
				Desc: "Number of write conflicts encountered.",
			},
			"cpu_nanos": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.DurationNS,
				Desc:     "CPU time consumed by the operation.",
			},
			"planning_time_micros": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.DurationUS,
				Desc:     "Time spent planning the operation.",
			},
			"cursor_exhausted": &inputs.FieldInfo{
				DataType: inputs.Bool, Type: inputs.Bool, Unit: inputs.NoUnit,
				Desc: "Whether the operation exhausted its cursor.",
			},
			"has_sort_stage": &inputs.FieldInfo{
				DataType: inputs.Bool, Type: inputs.Bool, Unit: inputs.NoUnit,
				Desc: "Whether the operation used a sort stage.",
			},
			"used_disk": &inputs.FieldInfo{
				DataType: inputs.Bool, Type: inputs.Bool, Unit: inputs.NoUnit,
				Desc: "Whether the operation used disk.",
			},
			"from_multi_planner": &inputs.FieldInfo{
				DataType: inputs.Bool, Type: inputs.Bool, Unit: inputs.NoUnit,
				Desc: "Whether the operation used the multi-planner.",
			},
			"replanned": &inputs.FieldInfo{
				DataType: inputs.Bool, Type: inputs.Bool, Unit: inputs.NoUnit,
				Desc: "Whether MongoDB replanned the operation.",
			},
		},
	}
}

func (*mongodbDBMOperationMeasurement) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name:   mongodbDBMOperationName,
		Cat:    point.Metric,
		Desc:   "MongoDB active operations aggregated from each $currentOp Activity snapshot.",
		DescZh: "根据每次 $currentOp Activity 快照聚合的 MongoDB 活跃操作指标。",
		Tags: map[string]interface{}{
			"mongod_host":       inputs.NewTagInfo("MongoDB server host."),
			"server":            inputs.NewTagInfo("MongoDB server address."),
			"database_instance": inputs.NewTagInfo("MongoDB instance identifier."),
			"database_type":     inputs.NewTagInfo("Database type. The value is `MongoDB`."),
			"namespace":         inputs.NewTagInfo("MongoDB namespace, usually database.collection."),
			"database_name":     inputs.NewTagInfo("Database name of the active operations."),
			"command_type":      inputs.NewTagInfo("MongoDB command type, such as find, aggregate, or getMore."),
			"operation":         inputs.NewTagInfo("MongoDB operation type reported by $currentOp."),
			"user":              inputs.NewTagInfo("First effective MongoDB user of the active operations."),
			"client_address":    inputs.NewTagInfo("Client host address without the ephemeral port."),
			"application":       inputs.NewTagInfo("Client application name."),
			"shard":             inputs.NewTagInfo("Shard executing the active operations."),
		},
		Fields: map[string]interface{}{
			"active_operation_count": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				Desc:     "Number of active operations in the current Activity snapshot for this tag group.",
			},
			"waiting_for_lock_count": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				Desc:     "Number of active operations waiting for a lock in the current Activity snapshot.",
			},
		},
	}
}

func (*mongodbDBMActivityMeasurement) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name:   mongodbDBMActivityName,
		Cat:    point.Logging,
		Desc:   "MongoDB operations that were active when the $currentOp snapshot was collected.",
		DescZh: "通过 $currentOp 快照采集的 MongoDB 当前活动操作。",
		Tags: map[string]interface{}{
			"mongod_host":       inputs.NewTagInfo("MongoDB server host."),
			"server":            inputs.NewTagInfo("MongoDB server address."),
			"database_instance": inputs.NewTagInfo("MongoDB instance identifier."),
			"database_type":     inputs.NewTagInfo("Database type. The value is `MongoDB`."),
			"database_name":     inputs.NewTagInfo("Database name of the operation."),
			"collection":        inputs.NewTagInfo("Collection name of the operation."),
			"command_type":      inputs.NewTagInfo("MongoDB command type, such as find, aggregate, or getMore."),
			"query_signature": inputs.NewTagInfo(
				"xxhash signature generated from database_name, collection, and MongoDB query_shape_hash.",
			),
			"query_shape_hash": inputs.NewTagInfo("MongoDB-native hash of the query shape."),
			"normalized_query_hash": inputs.NewTagInfo(
				"Hash of the canonical obfuscated command used to link activity to MongoDB Query Metrics.",
			),
			"activity_type": inputs.NewTagInfo("MongoDB current operation type."),
			"operation":     inputs.NewTagInfo("MongoDB operation name."),
			"shard":         inputs.NewTagInfo("Shard that is executing the operation."),
			"application":   inputs.NewTagInfo("Client application name."),
			"user":          inputs.NewTagInfo("First effective MongoDB user of the operation."),
		},
		Fields: map[string]interface{}{
			"message": &inputs.FieldInfo{
				DataType: inputs.String,
				Type:     inputs.String,
				Unit:     inputs.NoUnit,
				Desc:     "Complete obfuscated MongoDB command.",
			},
			"active": &inputs.FieldInfo{
				DataType: inputs.Bool,
				Type:     inputs.Bool,
				Unit:     inputs.NoUnit,
				Desc:     "Whether the operation is active.",
			},
			"description": &inputs.FieldInfo{
				DataType: inputs.String,
				Type:     inputs.String,
				Unit:     inputs.NoUnit,
				Desc:     "MongoDB operation description.",
			},
			"operation_id": &inputs.FieldInfo{
				DataType: inputs.String,
				Type:     inputs.String,
				Unit:     inputs.NoUnit,
				Desc:     "MongoDB operation identifier.",
			},
			"namespace": &inputs.FieldInfo{
				DataType: inputs.String,
				Type:     inputs.String,
				Unit:     inputs.NoUnit,
				Desc:     "MongoDB namespace of the operation.",
			},
			"plan_summary": &inputs.FieldInfo{
				DataType: inputs.String,
				Type:     inputs.String,
				Unit:     inputs.NoUnit,
				Desc:     "Summary of the query execution plan reported by $currentOp.",
			},
			"query_framework": &inputs.FieldInfo{
				DataType: inputs.String,
				Type:     inputs.String,
				Unit:     inputs.NoUnit,
				Desc:     "Query execution framework reported by MongoDB.",
			},
			"current_op_time": &inputs.FieldInfo{
				DataType: inputs.String,
				Type:     inputs.String,
				Unit:     inputs.NoUnit,
				Desc:     "Current operation timestamp reported by MongoDB.",
			},
			"microsecs_running": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.DurationUS,
				Desc:     "Elapsed time of the in-flight operation.",
			},
			"prepare_read_conflicts": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				Desc:     "Number of prepare read conflicts encountered by the operation.",
			},
			"write_conflicts": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				Desc:     "Number of write conflicts encountered by the operation.",
			},
			"num_yields": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				Desc:     "Number of times the operation yielded.",
			},
			"waiting_for_lock": &inputs.FieldInfo{
				DataType: inputs.Bool,
				Type:     inputs.Bool,
				Unit:     inputs.NoUnit,
				Desc:     "Whether the operation is waiting for a lock.",
			},
			"locks": &inputs.FieldInfo{
				DataType: inputs.String,
				Type:     inputs.String,
				Unit:     inputs.NoUnit,
				Desc:     "JSON document containing lock modes held by the operation.",
			},
			"lock_stats": &inputs.FieldInfo{
				DataType: inputs.String,
				Type:     inputs.String,
				Unit:     inputs.NoUnit,
				Desc:     "JSON document containing lock acquisition statistics.",
			},
			"waiting_for_flow_control": &inputs.FieldInfo{
				DataType: inputs.Bool,
				Type:     inputs.Bool,
				Unit:     inputs.NoUnit,
				Desc:     "Whether the operation is waiting for flow control.",
			},
			"flow_control_stats": &inputs.FieldInfo{
				DataType: inputs.String,
				Type:     inputs.String,
				Unit:     inputs.NoUnit,
				Desc:     "JSON document containing flow-control statistics.",
			},
			"waiting_for_latch": &inputs.FieldInfo{
				DataType: inputs.String,
				Type:     inputs.String,
				Unit:     inputs.NoUnit,
				Desc:     "JSON document describing the latch wait.",
			},
			"cursor": &inputs.FieldInfo{
				DataType: inputs.String,
				Type:     inputs.String,
				Unit:     inputs.NoUnit,
				Desc:     "JSON document containing cursor details and the obfuscated originating command.",
			},
			"transaction": &inputs.FieldInfo{
				DataType: inputs.String,
				Type:     inputs.String,
				Unit:     inputs.NoUnit,
				Desc:     "JSON document containing transaction timing and identifiers.",
			},
			"logical_session_id": &inputs.FieldInfo{
				DataType: inputs.String,
				Type:     inputs.String,
				Unit:     inputs.NoUnit,
				Desc:     "JSON document containing the hexadecimal logical session identifier.",
			},
			"client": &inputs.FieldInfo{
				DataType: inputs.String,
				Type:     inputs.String,
				Unit:     inputs.NoUnit,
				Desc:     "JSON document containing client host, driver, operating system, and platform.",
			},
			"query_truncated": &inputs.FieldInfo{
				DataType: inputs.String,
				Type:     inputs.String,
				Unit:     inputs.NoUnit,
				Desc:     "Whether MongoDB reported the command as truncated.",
			},
		},
	}
}

type mongodbDBMQueryObjectMeasurement struct{}

func (*mongodbDBMQueryObjectMeasurement) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name:   mongodbDBMQueryObjectName,
		Cat:    point.Object,
		Desc:   "MongoDB DBM query objects containing the obfuscated command linked to Query Metrics by query_signature and reported at most once every 24 hours per signature.",
		DescZh: "MongoDB DBM 查询对象，包含脱敏后的命令，通过 query_signature 与 Query Metrics 关联，同一签名最多每 24 小时上报一次。",
		Tags: map[string]interface{}{
			"name":              inputs.NewTagInfo("Object identity built from server, database_instance, and query_signature."),
			"mongod_host":       inputs.NewTagInfo("MongoDB server host."),
			"server":            inputs.NewTagInfo("MongoDB server address."),
			"database_instance": inputs.NewTagInfo("MongoDB instance identifier."),
			"database_type":     inputs.NewTagInfo("Database type. The value is `MongoDB`."),
			"database_name":     inputs.NewTagInfo("Database name from the query namespace."),
			"collection":        inputs.NewTagInfo("Collection name from the query namespace."),
			"command_type":      inputs.NewTagInfo("MongoDB command type, such as find or aggregate."),
			"query_signature":   inputs.NewTagInfo("xxhash signature generated from database_name, collection, and MongoDB query_shape_hash."),
			"query_shape_hash":  inputs.NewTagInfo("MongoDB-native hash of the query shape."),
			"normalized_query_hash": inputs.NewTagInfo(
				"Hash computed from the complete obfuscated MongoDB command for linking DBM data.",
			),
		},
		Fields: map[string]interface{}{
			"message": &inputs.FieldInfo{
				DataType: inputs.String,
				Type:     inputs.String,
				Unit:     inputs.NoUnit,
				Desc:     "Obfuscated MongoDB command reconstructed from the query shape.",
			},
		},
	}
}
