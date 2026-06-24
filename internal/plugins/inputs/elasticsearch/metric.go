// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package elasticsearch

import (
	"github.com/GuanceCloud/cliutils/point"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs"
)

type elasticsearchMeasurement struct {
	name     string
	tags     map[string]string
	fields   map[string]interface{}
	ts       int64
	election bool
}

// Point implement MeasurementV2.
func (m *elasticsearchMeasurement) Point() *point.Point {
	opts := append(point.DefaultMetricOptions(), point.WithTimestamp(m.ts))

	if m.election {
		opts = append(opts, point.WithExtraTags(datakit.GlobalElectionTags()))
	}

	return point.NewPoint(m.name,
		append(point.NewTags(m.tags), point.NewKVs(m.fields)...),
		opts...)
}

func (m elasticsearchMeasurement) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name:   "elasticsearch",
		Cat:    point.Metric,
		Desc:   "Elasticsearch cluster-level metric set collected from Elasticsearch APIs, covering general cluster and node runtime fields.",
		DescZh: "从 Elasticsearch API 采集的集群级指标集，包含集群和节点运行状态相关字段。",
		Fields: elasticsearchMeasurementFields,
	}
}

type nodeStatsMeasurement struct {
	elasticsearchMeasurement
}

func (m nodeStatsMeasurement) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name:   "elasticsearch_node_stats",
		Desc:   "Elasticsearch node statistics collected from the node stats API, including JVM, filesystem, thread pool, transport, indexing, search, merge, and cache metrics.",
		DescZh: "从 Elasticsearch node stats API 采集的节点统计指标，包含 JVM、文件系统、线程池、传输、索引、搜索、合并和缓存指标。",
		Fields: nodeStatsFields,
		Tags:   nodeStatsTags,
		Cat:    point.Metric,
	}
}

type clusterStatsMeasurement struct {
	elasticsearchMeasurement
}

func (m clusterStatsMeasurement) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name:   "elasticsearch_cluster_stats",
		Desc:   "Elasticsearch cluster statistics collected from the cluster stats API, including node, index, shard, document, store, and cluster health summary metrics.",
		DescZh: "从 Elasticsearch cluster stats API 采集的集群统计指标，包含节点、索引、分片、文档、存储和集群健康概览。",
		Fields: clusterStatsFields,
		Tags:   clusterStatsTags,
		Cat:    point.Metric,
	}
}

type clusterHealthMeasurement struct {
	elasticsearchMeasurement
}

func (m clusterHealthMeasurement) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name:   "elasticsearch_cluster_health",
		Desc:   "Elasticsearch cluster health metrics collected from the cluster health API, including shard state, node counts, pending tasks, and health status.",
		DescZh: "从 Elasticsearch cluster health API 采集的集群健康指标，包含分片状态、节点数量、等待任务和健康状态。",
		Fields: clusterHealthFields,
		Tags:   clusterHealthTags,
		Cat:    point.Metric,
	}
}

type clusterHealthIndicesMeasurement struct {
	elasticsearchMeasurement
}

func (m clusterHealthIndicesMeasurement) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name:   "elasticsearch_cluster_health_indices",
		Desc:   "Elasticsearch per-index cluster health metrics, including active, relocating, initializing, unassigned shard counts, and index health status.",
		DescZh: "Elasticsearch 索引维度的集群健康指标，包含活跃、迁移中、初始化、未分配分片数量和索引健康状态。",
		Fields: clusterHealthIndicesFields,
		Tags:   clusterHealthIndicesTags,
		Cat:    point.Metric,
	}
}

type indicesStatsShardsTotalMeasurement struct {
	elasticsearchMeasurement
}

func (m indicesStatsShardsTotalMeasurement) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name:   "elasticsearch_indices_stats_shards_total",
		Desc:   "Elasticsearch aggregate shard statistics across indices, summarizing shard-level index, search, merge, refresh, flush, and storage activity.",
		DescZh: "Elasticsearch 跨索引聚合的分片统计指标，汇总分片级索引、搜索、合并、刷新、flush 和存储活动。",
		Fields: indicesStatsShardsTotalFields,
		Cat:    point.Metric,
		// No tags.
	}
}

type indicesStatsMeasurement struct {
	elasticsearchMeasurement
}

func (m indicesStatsMeasurement) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name:   "elasticsearch_indices_stats",
		Desc:   "Elasticsearch index statistics collected from the indices stats API, including document, store, indexing, search, merge, refresh, flush, and cache metrics.",
		DescZh: "从 Elasticsearch indices stats API 采集的索引统计指标，包含文档、存储、索引、搜索、合并、刷新、flush 和缓存指标。",
		Fields: indicesStatsFields,
		Tags:   indicesStatsTags,
		Cat:    point.Metric,
	}
}

type indicesStatsShardsMeasurement struct {
	elasticsearchMeasurement
}

func (m indicesStatsShardsMeasurement) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name:   "elasticsearch_indices_stats_shards",
		Desc:   "Elasticsearch per-shard index statistics, including shard-level document, store, indexing, search, merge, refresh, flush, and cache metrics.",
		DescZh: "Elasticsearch 分片维度的索引统计指标，包含分片级文档、存储、索引、搜索、合并、刷新、flush 和缓存指标。",
		Fields: indicesStatsShardsFields,
		Tags:   indicesStatsShardsTags,
		Cat:    point.Metric,
	}
}
