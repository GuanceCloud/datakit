package elasticsearch

import (
	"time"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"
)

type elasticsearchMeasurement struct {
	name   string
	tags   map[string]string
	fields map[string]interface{}
	ts     time.Time
}

func (m elasticsearchMeasurement) LineProto() (*io.Point, error) {
	return io.MakePoint(m.name, m.tags, m.fields, m.ts)
}

func (m elasticsearchMeasurement) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name: "elasticsearch",
		Fields: map[string]interface{}{
			"active_shards_percent_as_number": &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.Percent, Desc: "active shards percent"},
			"active_primary_shards":           &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: "active primary shards"},
			"status":                          &inputs.FieldInfo{DataType: inputs.String, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: "status"},
			"timed_out":                       &inputs.FieldInfo{DataType: inputs.Bool, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: "timed_out"},
		},
	}
}

type nodeStatsMeasurement struct {
	elasticsearchMeasurement
}

func (m nodeStatsMeasurement) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name:   "elasticsearch_node_stats",
		Fields: nodeStatsFields,
		Tags: map[string]interface{}{
			"cluster_name":                     inputs.NewTagInfo("cluster name"),
			"node_attribute_ml.enabled":        inputs.NewTagInfo("machine learning enabled"),
			"node_attribute_ml.machine_memory": inputs.NewTagInfo("machine learning memory"),
			"node_attribute_ml.max_open_jobs":  inputs.NewTagInfo("The maximum number of jobs that can run simultaneously on a node"),
			"node_attribute_xpack.installed":   inputs.NewTagInfo("xpack installed"),
			"node_host":                        inputs.NewTagInfo("node host"),
			"node_id":                          inputs.NewTagInfo("node id"),
			"node_name":                        inputs.NewTagInfo("node name"),
		},
	}
}

type clusterStatsMeasurement struct {
	elasticsearchMeasurement
}

func (m clusterStatsMeasurement) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name:   "elasticsearch_cluster_stats",
		Fields: clusterStatsFields,
		Tags: map[string]interface{}{
			"cluster_name": inputs.NewTagInfo("cluster name"),
			"node_name":    inputs.NewTagInfo("node name"),
			"status":       inputs.NewTagInfo("status"),
		},
	}
}

type clusterHealthMeasurement struct {
	elasticsearchMeasurement
}

func (m clusterHealthMeasurement) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name:   "elasticsearch_cluster_health",
		Fields: clusterHealthFields,
		Tags: map[string]interface{}{
			"name": inputs.NewTagInfo("cluster name"),
		},
	}
}

type clusterHealthIndicesMeasurement struct {
	elasticsearchMeasurement
}

func (m clusterHealthIndicesMeasurement) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name:   "elasticsearch_cluster_health",
		Fields: clusterHealthIndicesFields,
		Tags: map[string]interface{}{
			"name":  inputs.NewTagInfo("cluster name"),
			"index": inputs.NewTagInfo("index"),
		},
	}
}

type indicesStatsShardsTotalMeasurement struct {
	elasticsearchMeasurement
}

func (m indicesStatsShardsTotalMeasurement) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name:   "elasticsearch_indices_stats_shards_total",
		Fields: indicesStatsShardsTotalFields,
	}
}

type indicesStatsMeasurement struct {
	elasticsearchMeasurement
}

func (m indicesStatsMeasurement) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name:   "elasticsearch_indices_stats",
		Fields: indicesStatsFields,
		Tags: map[string]interface{}{
			"index_name": inputs.NewTagInfo("index name"),
		},
	}
}

type indicesStatsShardsMeasurement struct {
	elasticsearchMeasurement
}

func (m indicesStatsShardsMeasurement) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name:   "elasticsearch_indices_stats_shards",
		Fields: indicesStatsShardsFields,
		Tags: map[string]interface{}{
			"index_name": inputs.NewTagInfo("index name"),
			"node_name":  inputs.NewTagInfo("node name"),
			"shard_name": inputs.NewTagInfo("shard name"),
			"type":       inputs.NewTagInfo("type"),
		},
	}
}
