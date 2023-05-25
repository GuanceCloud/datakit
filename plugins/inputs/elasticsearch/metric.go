// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package elasticsearch

import (
	"time"

	"github.com/GuanceCloud/cliutils/point"
	dkpt "gitlab.jiagouyun.com/cloudcare-tools/datakit/io/point"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"
)

type elasticsearchMeasurement struct {
	name     string
	tags     map[string]string
	fields   map[string]interface{}
	ts       time.Time
	election bool
}

func (m elasticsearchMeasurement) LineProto() (*dkpt.Point, error) {
	return dkpt.NewPoint(m.name, m.tags, m.fields, dkpt.MOptElectionV2(m.election))
}

// Point implement MeasurementV2.
func (m *elasticsearchMeasurement) Point() *point.Point {
	opts := point.DefaultMetricOptions()

	if m.election {
		opts = append(opts, point.WithExtraTags(dkpt.GlobalElectionTags()))
	}

	return point.NewPointV2([]byte(m.name),
		append(point.NewTags(m.tags), point.NewKVs(m.fields)...),
		opts...)
}

func (m elasticsearchMeasurement) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name:   "elasticsearch",
		Fields: elasticsearchMeasurementFields,
	}
}

type nodeStatsMeasurement struct {
	elasticsearchMeasurement
}

func (m nodeStatsMeasurement) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name:   "elasticsearch_node_stats",
		Fields: nodeStatsFields,
		Tags:   nodeStatsTags,
	}
}

type clusterStatsMeasurement struct {
	elasticsearchMeasurement
}

func (m clusterStatsMeasurement) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name:   "elasticsearch_cluster_stats",
		Fields: clusterStatsFields,
		Tags:   clusterStatsTags,
	}
}

type clusterHealthMeasurement struct {
	elasticsearchMeasurement
}

func (m clusterHealthMeasurement) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name:   "elasticsearch_cluster_health",
		Fields: clusterHealthFields,
		Tags:   clusterHealthTags,
	}
}

type clusterHealthIndicesMeasurement struct {
	elasticsearchMeasurement
}

func (m clusterHealthIndicesMeasurement) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name:   "elasticsearch_cluster_health_indices",
		Fields: clusterHealthIndicesFields,
		Tags:   clusterHealthIndicesTags,
	}
}

type indicesStatsShardsTotalMeasurement struct {
	elasticsearchMeasurement
}

func (m indicesStatsShardsTotalMeasurement) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name:   "elasticsearch_indices_stats_shards_total",
		Fields: indicesStatsShardsTotalFields,
		// No tags.
	}
}

type indicesStatsMeasurement struct {
	elasticsearchMeasurement
}

func (m indicesStatsMeasurement) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name:   "elasticsearch_indices_stats",
		Fields: indicesStatsFields,
		Tags:   indicesStatsTags,
	}
}

type indicesStatsShardsMeasurement struct {
	elasticsearchMeasurement
}

func (m indicesStatsShardsMeasurement) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name:   "elasticsearch_indices_stats_shards",
		Fields: indicesStatsShardsFields,
		Tags:   indicesStatsShardsTags,
	}
}
