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
			"cluster_name":                     inputs.NewTagInfo("Name of the cluster, based on the Cluster name setting setting."),
			"node_attribute_ml.enabled":        inputs.NewTagInfo("Set to true (default) to enable machine learning APIs on the node."),
			"node_attribute_ml.machine_memory": inputs.NewTagInfo("The machine’s memory that machine learning may use for running analytics processes."),
			"node_attribute_ml.max_open_jobs":  inputs.NewTagInfo("The maximum number of jobs that can run simultaneously on a node."),
			"node_attribute_xpack.installed":   inputs.NewTagInfo("Show whether xpack is installed."),
			"node_host":                        inputs.NewTagInfo("Network host for the node, based on the network.host setting."),
			"node_id":                          inputs.NewTagInfo("The id for the node."),
			"node_name":                        inputs.NewTagInfo("Human-readable identifier for the node."),
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
			"cluster_name": inputs.NewTagInfo("Name of the cluster, based on the cluster.name setting."),
			"node_name":    inputs.NewTagInfo("Name of the node."),
			"status":       inputs.NewTagInfo("Health status of the cluster, based on the state of its primary and replica shards."),
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
			"name": inputs.NewTagInfo("Name of the cluster."),
		},
	}
}

type clusterHealthIndicesMeasurement struct {
	elasticsearchMeasurement
}

func (m clusterHealthIndicesMeasurement) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name:   "elasticsearch_cluster_health_indices",
		Fields: clusterHealthIndicesFields,
		Tags: map[string]interface{}{
			"name":  inputs.NewTagInfo("Name of the cluster."),
			"index": inputs.NewTagInfo("Name of the index."),
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
			"cluster_name": inputs.NewTagInfo("Name of the cluster, based on the Cluster name setting setting."),
			"index_name":   inputs.NewTagInfo("Name of the index."),
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
			"index_name": inputs.NewTagInfo("Name of the index."),
			"node_name":  inputs.NewTagInfo("Name of the node."),
			"shard_name": inputs.NewTagInfo("Name of the shard."),
			"type":       inputs.NewTagInfo("Type of the shard."),
		},
	}
}
