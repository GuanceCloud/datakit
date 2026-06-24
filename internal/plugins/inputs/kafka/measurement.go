// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

//nolint:lll,goconst,funlen // Metric descriptions are intentionally long metadata literals.
package kafka

import (
	"github.com/GuanceCloud/cliutils/point"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs"
)

const (
	TagGroupCommon    = "common"
	TagGroupPurgatory = "purgatory"
	TagGroupRequest   = "request"
	TagGroupTopic     = "topic"
	TagGroupPartition = "partition"
)

type kafkaMeasurement struct{}

//nolint:lll
func (m *kafkaMeasurement) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name:   "kafka",
		Desc:   "Metric set for auto collect mode only. Contains all Kafka MBean metrics collected automatically when enable_auto_collect=true. Fields are formatted as domain.type.name.attr.",
		DescZh: "指标集仅用于自动采集模式。当 enable_auto_collect=true 时，包含所有自动采集的 Kafka MBean 指标。字段格式为 domain.type.name.attr。",
		Cat:    point.Metric,
		Tags:   m.getTags(),
		Fields: m.getFields(),
	}
}

func (m *kafkaMeasurement) getTags() map[string]interface{} {
	return mergeMaps(
		m.getCommonTags(),
		m.getPurgatoryTags(),
		m.getRequestTags(),
		m.getTopicTags(),
		m.getPartitionTags(),
	)
}

func (m *kafkaMeasurement) getCommonTags() map[string]interface{} {
	tags := make(map[string]interface{})
	tags["host"] = &inputs.TagInfo{Desc: "Hostname."}
	tags["jolokia_agent_url"] = &inputs.TagInfo{Desc: "Full Jolokia agent URL."}
	return tags
}

func (m *kafkaMeasurement) getPurgatoryTags() map[string]interface{} {
	tags := make(map[string]interface{})
	tags["delayedOperation"] = &inputs.TagInfo{Desc: "Delayed operation type."}
	return tags
}

func (m *kafkaMeasurement) getRequestTags() map[string]interface{} {
	tags := make(map[string]interface{})
	tags["request"] = &inputs.TagInfo{Desc: "Request type."}
	return tags
}

func (m *kafkaMeasurement) getTopicTags() map[string]interface{} {
	tags := make(map[string]interface{})
	tags["topic"] = &inputs.TagInfo{Desc: "Topic name."}
	return tags
}

func (m *kafkaMeasurement) getPartitionTags() map[string]interface{} {
	tags := make(map[string]interface{})
	tags["partition"] = &inputs.TagInfo{Desc: "Partition number."}
	tags["topic"] = &inputs.TagInfo{Desc: "Topic name."}
	return tags
}

func mergeMaps(fieldMaps ...map[string]interface{}) map[string]interface{} {
	result := make(map[string]interface{})
	for _, fields := range fieldMaps {
		for k, v := range fields {
			result[k] = v
		}
	}
	return result
}

func (m *kafkaMeasurement) getFields() map[string]interface{} {
	fields := mergeMaps(
		m.getControllerFields(),
		m.getReplicaFields(),
		m.getZookeeperFields(),
		m.getPurgatoryFields(),
		m.getRequestFields(),
		m.getRequestHandlerFields(),
		m.getNetworkFields(),
		m.getTopicsFields(),
		m.getTopicFields(),
		m.getPartitionFields(),
		m.getLogFields(),
	)
	applyKafkaMeasurementMetadata(fields)
	return fields
}

//nolint:lll,funlen
func (m *kafkaMeasurement) getControllerFields() map[string]interface{} {
	fields := map[string]interface{}{
		"kafka.controller.ControllerStats.AutoLeaderBalanceRateAndTimeMs.50thPercentile": &inputs.FieldInfo{
			DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS,
			Desc: "50th percentile of time spent in kafka controller Controller Stats Auto Leader Balance Rate And Time, measured in milliseconds, scoped by the Kafka auto-collect measurement tags.",
		},

		"kafka.controller.ControllerStats.AutoLeaderBalanceRateAndTimeMs.75thPercentile": &inputs.FieldInfo{
			DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS,
			Desc: "75th percentile of time spent in kafka controller Controller Stats Auto Leader Balance Rate And Time, measured in milliseconds, scoped by the Kafka auto-collect measurement tags.",
		},

		"kafka.controller.ControllerStats.AutoLeaderBalanceRateAndTimeMs.95thPercentile": &inputs.FieldInfo{
			DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS,
			Desc: "95th percentile of time spent in kafka controller Controller Stats Auto Leader Balance Rate And Time, measured in milliseconds, scoped by the Kafka auto-collect measurement tags.",
		},

		"kafka.controller.ControllerStats.AutoLeaderBalanceRateAndTimeMs.98thPercentile": &inputs.FieldInfo{
			DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS,
			Desc: "98th percentile of time spent in kafka controller Controller Stats Auto Leader Balance Rate And Time, measured in milliseconds, scoped by the Kafka auto-collect measurement tags.",
		},

		"kafka.controller.ControllerStats.AutoLeaderBalanceRateAndTimeMs.999thPercentile": &inputs.FieldInfo{
			DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS,
			Desc: "99.9th percentile of time spent in kafka controller Controller Stats Auto Leader Balance Rate And Time, measured in milliseconds, scoped by the Kafka auto-collect measurement tags.",
		},

		"kafka.controller.ControllerStats.AutoLeaderBalanceRateAndTimeMs.99thPercentile": &inputs.FieldInfo{
			DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS,
			Desc: "99th percentile of time spent in kafka controller Controller Stats Auto Leader Balance Rate And Time, measured in milliseconds, scoped by the Kafka auto-collect measurement tags.",
		},

		"kafka.controller.ControllerStats.AutoLeaderBalanceRateAndTimeMs.FifteenMinuteRate": &inputs.FieldInfo{
			DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount,
			Desc: "Fifteen-minute moving average rate for kafka controller Controller Stats Auto Leader Balance Rate And Time in events per second, scoped by the Kafka auto-collect measurement tags.",
		},

		"kafka.controller.ControllerStats.AutoLeaderBalanceRateAndTimeMs.FiveMinuteRate": &inputs.FieldInfo{
			DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount,
			Desc: "Five-minute moving average rate for kafka controller Controller Stats Auto Leader Balance Rate And Time in events per second, scoped by the Kafka auto-collect measurement tags.",
		},

		"kafka.controller.ControllerStats.AutoLeaderBalanceRateAndTimeMs.Max": &inputs.FieldInfo{
			DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS,
			Desc: "Maximum time spent in kafka controller Controller Stats Auto Leader Balance Rate And Time, measured in milliseconds, scoped by the Kafka auto-collect measurement tags.",
		},

		"kafka.controller.ControllerStats.AutoLeaderBalanceRateAndTimeMs.Mean": &inputs.FieldInfo{
			DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS,
			Desc: "Average time spent in kafka controller Controller Stats Auto Leader Balance Rate And Time, measured in milliseconds, scoped by the Kafka auto-collect measurement tags.",
		},

		"kafka.controller.ControllerStats.AutoLeaderBalanceRateAndTimeMs.MeanRate": &inputs.FieldInfo{
			DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount,
			Desc: "Mean rate for kafka controller Controller Stats Auto Leader Balance Rate And Time in events per second, scoped by the Kafka auto-collect measurement tags.",
		},

		"kafka.controller.ControllerStats.AutoLeaderBalanceRateAndTimeMs.Min": &inputs.FieldInfo{
			DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS,
			Desc: "Minimum time spent in kafka controller Controller Stats Auto Leader Balance Rate And Time, measured in milliseconds, scoped by the Kafka auto-collect measurement tags.",
		},

		"kafka.controller.ControllerStats.AutoLeaderBalanceRateAndTimeMs.OneMinuteRate": &inputs.FieldInfo{
			DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount,
			Desc: "One-minute moving average rate for kafka controller Controller Stats Auto Leader Balance Rate And Time in events per second, scoped by the Kafka auto-collect measurement tags.",
		},

		"kafka.controller.ControllerStats.AutoLeaderBalanceRateAndTimeMs.StdDev": &inputs.FieldInfo{
			DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS,
			Desc: "Standard deviation of time spent in kafka controller Controller Stats Auto Leader Balance Rate And Time, measured in milliseconds, scoped by the Kafka auto-collect measurement tags.",
		},

		"kafka.controller.ControllerStats.AutoLeaderBalanceRateAndTimeMs.Count": &inputs.FieldInfo{
			DataType: inputs.Int, Type: inputs.Count, Unit: inputs.NCount,
			Desc: "Total number of kafka controller Controller Stats Auto Leader Balance Rate And Time, scoped by the Kafka auto-collect measurement tags.",
		},

		"kafka.controller.ControllerStats.ControlledShutdownRateAndTimeMs.50thPercentile": &inputs.FieldInfo{
			DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS,
			Desc: "50th percentile of time spent in kafka controller Controller Stats Controlled Shutdown Rate And Time, measured in milliseconds, scoped by the Kafka auto-collect measurement tags.",
		},

		"kafka.controller.ControllerStats.ControlledShutdownRateAndTimeMs.75thPercentile": &inputs.FieldInfo{
			DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS,
			Desc: "75th percentile of time spent in kafka controller Controller Stats Controlled Shutdown Rate And Time, measured in milliseconds, scoped by the Kafka auto-collect measurement tags.",
		},

		"kafka.controller.ControllerStats.ControlledShutdownRateAndTimeMs.95thPercentile": &inputs.FieldInfo{
			DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS,
			Desc: "95th percentile of time spent in kafka controller Controller Stats Controlled Shutdown Rate And Time, measured in milliseconds, scoped by the Kafka auto-collect measurement tags.",
		},

		"kafka.controller.ControllerStats.ControlledShutdownRateAndTimeMs.98thPercentile": &inputs.FieldInfo{
			DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS,
			Desc: "98th percentile of time spent in kafka controller Controller Stats Controlled Shutdown Rate And Time, measured in milliseconds, scoped by the Kafka auto-collect measurement tags.",
		},

		"kafka.controller.ControllerStats.ControlledShutdownRateAndTimeMs.999thPercentile": &inputs.FieldInfo{
			DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS,
			Desc: "99.9th percentile of time spent in kafka controller Controller Stats Controlled Shutdown Rate And Time, measured in milliseconds, scoped by the Kafka auto-collect measurement tags.",
		},

		"kafka.controller.ControllerStats.ControlledShutdownRateAndTimeMs.99thPercentile": &inputs.FieldInfo{
			DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS,
			Desc: "99th percentile of time spent in kafka controller Controller Stats Controlled Shutdown Rate And Time, measured in milliseconds, scoped by the Kafka auto-collect measurement tags.",
		},

		"kafka.controller.ControllerStats.ControlledShutdownRateAndTimeMs.FifteenMinuteRate": &inputs.FieldInfo{
			DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount,
			Desc: "Fifteen-minute moving average rate for kafka controller Controller Stats Controlled Shutdown Rate And Time in events per second, scoped by the Kafka auto-collect measurement tags.",
		},

		"kafka.controller.ControllerStats.ControlledShutdownRateAndTimeMs.FiveMinuteRate": &inputs.FieldInfo{
			DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount,
			Desc: "Five-minute moving average rate for kafka controller Controller Stats Controlled Shutdown Rate And Time in events per second, scoped by the Kafka auto-collect measurement tags.",
		},

		"kafka.controller.ControllerStats.ControlledShutdownRateAndTimeMs.Max": &inputs.FieldInfo{
			DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS,
			Desc: "Maximum time spent in kafka controller Controller Stats Controlled Shutdown Rate And Time, measured in milliseconds, scoped by the Kafka auto-collect measurement tags.",
		},

		"kafka.controller.ControllerStats.ControlledShutdownRateAndTimeMs.Mean": &inputs.FieldInfo{
			DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS,
			Desc: "Average time spent in kafka controller Controller Stats Controlled Shutdown Rate And Time, measured in milliseconds, scoped by the Kafka auto-collect measurement tags.",
		},

		"kafka.controller.ControllerStats.ControlledShutdownRateAndTimeMs.MeanRate": &inputs.FieldInfo{
			DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount,
			Desc: "Mean rate for kafka controller Controller Stats Controlled Shutdown Rate And Time in events per second, scoped by the Kafka auto-collect measurement tags.",
		},

		"kafka.controller.ControllerStats.ControlledShutdownRateAndTimeMs.Min": &inputs.FieldInfo{
			DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS,
			Desc: "Minimum time spent in kafka controller Controller Stats Controlled Shutdown Rate And Time, measured in milliseconds, scoped by the Kafka auto-collect measurement tags.",
		},

		"kafka.controller.ControllerStats.ControlledShutdownRateAndTimeMs.OneMinuteRate": &inputs.FieldInfo{
			DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount,
			Desc: "One-minute moving average rate for kafka controller Controller Stats Controlled Shutdown Rate And Time in events per second, scoped by the Kafka auto-collect measurement tags.",
		},

		"kafka.controller.ControllerStats.ControlledShutdownRateAndTimeMs.StdDev": &inputs.FieldInfo{
			DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS,
			Desc: "Standard deviation of time spent in kafka controller Controller Stats Controlled Shutdown Rate And Time, measured in milliseconds, scoped by the Kafka auto-collect measurement tags.",
		},

		"kafka.controller.ControllerStats.ControlledShutdownRateAndTimeMs.Count": &inputs.FieldInfo{
			DataType: inputs.Int, Type: inputs.Count, Unit: inputs.NCount,
			Desc: "Total number of kafka controller Controller Stats Controlled Shutdown Rate And Time, scoped by the Kafka auto-collect measurement tags.",
		},

		"kafka.controller.ControllerStats.ControllerChangeRateAndTimeMs.50thPercentile": &inputs.FieldInfo{
			DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS,
			Desc: "50th percentile of time spent in kafka controller Controller Stats Controller Change Rate And Time, measured in milliseconds, scoped by the Kafka auto-collect measurement tags.",
		},

		"kafka.controller.ControllerStats.ControllerChangeRateAndTimeMs.75thPercentile": &inputs.FieldInfo{
			DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS,
			Desc: "75th percentile of time spent in kafka controller Controller Stats Controller Change Rate And Time, measured in milliseconds, scoped by the Kafka auto-collect measurement tags.",
		},

		"kafka.controller.ControllerStats.ControllerChangeRateAndTimeMs.95thPercentile": &inputs.FieldInfo{
			DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS,
			Desc: "95th percentile of time spent in kafka controller Controller Stats Controller Change Rate And Time, measured in milliseconds, scoped by the Kafka auto-collect measurement tags.",
		},

		"kafka.controller.ControllerStats.ControllerChangeRateAndTimeMs.98thPercentile": &inputs.FieldInfo{
			DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS,
			Desc: "98th percentile of time spent in kafka controller Controller Stats Controller Change Rate And Time, measured in milliseconds, scoped by the Kafka auto-collect measurement tags.",
		},

		"kafka.controller.ControllerStats.ControllerChangeRateAndTimeMs.999thPercentile": &inputs.FieldInfo{
			DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS,
			Desc: "99.9th percentile of time spent in kafka controller Controller Stats Controller Change Rate And Time, measured in milliseconds, scoped by the Kafka auto-collect measurement tags.",
		},

		"kafka.controller.ControllerStats.ControllerChangeRateAndTimeMs.99thPercentile": &inputs.FieldInfo{
			DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS,
			Desc: "99th percentile of time spent in kafka controller Controller Stats Controller Change Rate And Time, measured in milliseconds, scoped by the Kafka auto-collect measurement tags.",
		},

		"kafka.controller.ControllerStats.ControllerChangeRateAndTimeMs.FifteenMinuteRate": &inputs.FieldInfo{
			DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount,
			Desc: "Fifteen-minute moving average rate for kafka controller Controller Stats Controller Change Rate And Time in events per second, scoped by the Kafka auto-collect measurement tags.",
		},

		"kafka.controller.ControllerStats.ControllerChangeRateAndTimeMs.FiveMinuteRate": &inputs.FieldInfo{
			DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount,
			Desc: "Five-minute moving average rate for kafka controller Controller Stats Controller Change Rate And Time in events per second, scoped by the Kafka auto-collect measurement tags.",
		},

		"kafka.controller.ControllerStats.ControllerChangeRateAndTimeMs.Max": &inputs.FieldInfo{
			DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS,
			Desc: "Maximum time spent in kafka controller Controller Stats Controller Change Rate And Time, measured in milliseconds, scoped by the Kafka auto-collect measurement tags.",
		},

		"kafka.controller.ControllerStats.ControllerChangeRateAndTimeMs.Mean": &inputs.FieldInfo{
			DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS,
			Desc: "Average time spent in kafka controller Controller Stats Controller Change Rate And Time, measured in milliseconds, scoped by the Kafka auto-collect measurement tags.",
		},

		"kafka.controller.ControllerStats.ControllerChangeRateAndTimeMs.MeanRate": &inputs.FieldInfo{
			DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount,
			Desc: "Mean rate for kafka controller Controller Stats Controller Change Rate And Time in events per second, scoped by the Kafka auto-collect measurement tags.",
		},

		"kafka.controller.ControllerStats.ControllerChangeRateAndTimeMs.Min": &inputs.FieldInfo{
			DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS,
			Desc: "Minimum time spent in kafka controller Controller Stats Controller Change Rate And Time, measured in milliseconds, scoped by the Kafka auto-collect measurement tags.",
		},

		"kafka.controller.ControllerStats.ControllerChangeRateAndTimeMs.OneMinuteRate": &inputs.FieldInfo{
			DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount,
			Desc: "One-minute moving average rate for kafka controller Controller Stats Controller Change Rate And Time in events per second, scoped by the Kafka auto-collect measurement tags.",
		},

		"kafka.controller.ControllerStats.ControllerChangeRateAndTimeMs.StdDev": &inputs.FieldInfo{
			DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS,
			Desc: "Standard deviation of time spent in kafka controller Controller Stats Controller Change Rate And Time, measured in milliseconds, scoped by the Kafka auto-collect measurement tags.",
		},

		"kafka.controller.ControllerStats.ControllerChangeRateAndTimeMs.Count": &inputs.FieldInfo{
			DataType: inputs.Int, Type: inputs.Count, Unit: inputs.NCount,
			Desc: "Total number of kafka controller Controller Stats Controller Change Rate And Time, scoped by the Kafka auto-collect measurement tags.",
		},

		"kafka.controller.ControllerStats.ControllerShutdownRateAndTimeMs.50thPercentile": &inputs.FieldInfo{
			DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS,
			Desc: "50th percentile of time spent in kafka controller Controller Stats Controller Shutdown Rate And Time, measured in milliseconds, scoped by the Kafka auto-collect measurement tags.",
		},

		"kafka.controller.ControllerStats.ControllerShutdownRateAndTimeMs.75thPercentile": &inputs.FieldInfo{
			DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS,
			Desc: "75th percentile of time spent in kafka controller Controller Stats Controller Shutdown Rate And Time, measured in milliseconds, scoped by the Kafka auto-collect measurement tags.",
		},

		"kafka.controller.ControllerStats.ControllerShutdownRateAndTimeMs.95thPercentile": &inputs.FieldInfo{
			DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS,
			Desc: "95th percentile of time spent in kafka controller Controller Stats Controller Shutdown Rate And Time, measured in milliseconds, scoped by the Kafka auto-collect measurement tags.",
		},

		"kafka.controller.ControllerStats.ControllerShutdownRateAndTimeMs.98thPercentile": &inputs.FieldInfo{
			DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS,
			Desc: "98th percentile of time spent in kafka controller Controller Stats Controller Shutdown Rate And Time, measured in milliseconds, scoped by the Kafka auto-collect measurement tags.",
		},

		"kafka.controller.ControllerStats.ControllerShutdownRateAndTimeMs.999thPercentile": &inputs.FieldInfo{
			DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS,
			Desc: "99.9th percentile of time spent in kafka controller Controller Stats Controller Shutdown Rate And Time, measured in milliseconds, scoped by the Kafka auto-collect measurement tags.",
		},

		"kafka.controller.ControllerStats.ControllerShutdownRateAndTimeMs.99thPercentile": &inputs.FieldInfo{
			DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS,
			Desc: "99th percentile of time spent in kafka controller Controller Stats Controller Shutdown Rate And Time, measured in milliseconds, scoped by the Kafka auto-collect measurement tags.",
		},

		"kafka.controller.ControllerStats.ControllerShutdownRateAndTimeMs.FifteenMinuteRate": &inputs.FieldInfo{
			DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount,
			Desc: "Fifteen-minute moving average rate for kafka controller Controller Stats Controller Shutdown Rate And Time in events per second, scoped by the Kafka auto-collect measurement tags.",
		},

		"kafka.controller.ControllerStats.ControllerShutdownRateAndTimeMs.FiveMinuteRate": &inputs.FieldInfo{
			DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount,
			Desc: "Five-minute moving average rate for kafka controller Controller Stats Controller Shutdown Rate And Time in events per second, scoped by the Kafka auto-collect measurement tags.",
		},

		"kafka.controller.ControllerStats.ControllerShutdownRateAndTimeMs.Max": &inputs.FieldInfo{
			DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS,
			Desc: "Maximum time spent in kafka controller Controller Stats Controller Shutdown Rate And Time, measured in milliseconds, scoped by the Kafka auto-collect measurement tags.",
		},

		"kafka.controller.ControllerStats.ControllerShutdownRateAndTimeMs.Mean": &inputs.FieldInfo{
			DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS,
			Desc: "Average time spent in kafka controller Controller Stats Controller Shutdown Rate And Time, measured in milliseconds, scoped by the Kafka auto-collect measurement tags.",
		},

		"kafka.controller.ControllerStats.ControllerShutdownRateAndTimeMs.MeanRate": &inputs.FieldInfo{
			DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount,
			Desc: "Mean rate for kafka controller Controller Stats Controller Shutdown Rate And Time in events per second, scoped by the Kafka auto-collect measurement tags.",
		},

		"kafka.controller.ControllerStats.ControllerShutdownRateAndTimeMs.Min": &inputs.FieldInfo{
			DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS,
			Desc: "Minimum time spent in kafka controller Controller Stats Controller Shutdown Rate And Time, measured in milliseconds, scoped by the Kafka auto-collect measurement tags.",
		},

		"kafka.controller.ControllerStats.ControllerShutdownRateAndTimeMs.OneMinuteRate": &inputs.FieldInfo{
			DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount,
			Desc: "One-minute moving average rate for kafka controller Controller Stats Controller Shutdown Rate And Time in events per second, scoped by the Kafka auto-collect measurement tags.",
		},

		"kafka.controller.ControllerStats.ControllerShutdownRateAndTimeMs.StdDev": &inputs.FieldInfo{
			DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS,
			Desc: "Standard deviation of time spent in kafka controller Controller Stats Controller Shutdown Rate And Time, measured in milliseconds, scoped by the Kafka auto-collect measurement tags.",
		},

		"kafka.controller.ControllerStats.ControllerShutdownRateAndTimeMs.Count": &inputs.FieldInfo{
			DataType: inputs.Int, Type: inputs.Count, Unit: inputs.NCount,
			Desc: "Total number of kafka controller Controller Stats Controller Shutdown Rate And Time, scoped by the Kafka auto-collect measurement tags.",
		},

		"kafka.controller.ControllerStats.IsrChangeRateAndTimeMs.50thPercentile": &inputs.FieldInfo{
			DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS,
			Desc: "50th percentile of time spent in kafka controller Controller Stats Isr Change Rate And Time, measured in milliseconds, scoped by the Kafka auto-collect measurement tags.",
		},

		"kafka.controller.ControllerStats.IsrChangeRateAndTimeMs.75thPercentile": &inputs.FieldInfo{
			DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS,
			Desc: "75th percentile of time spent in kafka controller Controller Stats Isr Change Rate And Time, measured in milliseconds, scoped by the Kafka auto-collect measurement tags.",
		},

		"kafka.controller.ControllerStats.IsrChangeRateAndTimeMs.95thPercentile": &inputs.FieldInfo{
			DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS,
			Desc: "95th percentile of time spent in kafka controller Controller Stats Isr Change Rate And Time, measured in milliseconds, scoped by the Kafka auto-collect measurement tags.",
		},

		"kafka.controller.ControllerStats.IsrChangeRateAndTimeMs.98thPercentile": &inputs.FieldInfo{
			DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS,
			Desc: "98th percentile of time spent in kafka controller Controller Stats Isr Change Rate And Time, measured in milliseconds, scoped by the Kafka auto-collect measurement tags.",
		},

		"kafka.controller.ControllerStats.IsrChangeRateAndTimeMs.999thPercentile": &inputs.FieldInfo{
			DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS,
			Desc: "99.9th percentile of time spent in kafka controller Controller Stats Isr Change Rate And Time, measured in milliseconds, scoped by the Kafka auto-collect measurement tags.",
		},

		"kafka.controller.ControllerStats.IsrChangeRateAndTimeMs.99thPercentile": &inputs.FieldInfo{
			DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS,
			Desc: "99th percentile of time spent in kafka controller Controller Stats Isr Change Rate And Time, measured in milliseconds, scoped by the Kafka auto-collect measurement tags.",
		},

		"kafka.controller.ControllerStats.IsrChangeRateAndTimeMs.FifteenMinuteRate": &inputs.FieldInfo{
			DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount,
			Desc: "Fifteen-minute moving average rate for kafka controller Controller Stats Isr Change Rate And Time in events per second, scoped by the Kafka auto-collect measurement tags.",
		},

		"kafka.controller.ControllerStats.IsrChangeRateAndTimeMs.FiveMinuteRate": &inputs.FieldInfo{
			DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount,
			Desc: "Five-minute moving average rate for kafka controller Controller Stats Isr Change Rate And Time in events per second, scoped by the Kafka auto-collect measurement tags.",
		},

		"kafka.controller.ControllerStats.IsrChangeRateAndTimeMs.Max": &inputs.FieldInfo{
			DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS,
			Desc: "Maximum time spent in kafka controller Controller Stats Isr Change Rate And Time, measured in milliseconds, scoped by the Kafka auto-collect measurement tags.",
		},

		"kafka.controller.ControllerStats.IsrChangeRateAndTimeMs.Mean": &inputs.FieldInfo{
			DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS,
			Desc: "Average time spent in kafka controller Controller Stats Isr Change Rate And Time, measured in milliseconds, scoped by the Kafka auto-collect measurement tags.",
		},

		"kafka.controller.ControllerStats.IsrChangeRateAndTimeMs.MeanRate": &inputs.FieldInfo{
			DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount,
			Desc: "Mean rate for kafka controller Controller Stats Isr Change Rate And Time in events per second, scoped by the Kafka auto-collect measurement tags.",
		},

		"kafka.controller.ControllerStats.IsrChangeRateAndTimeMs.Min": &inputs.FieldInfo{
			DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS,
			Desc: "Minimum time spent in kafka controller Controller Stats Isr Change Rate And Time, measured in milliseconds, scoped by the Kafka auto-collect measurement tags.",
		},

		"kafka.controller.ControllerStats.IsrChangeRateAndTimeMs.OneMinuteRate": &inputs.FieldInfo{
			DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount,
			Desc: "One-minute moving average rate for kafka controller Controller Stats Isr Change Rate And Time in events per second, scoped by the Kafka auto-collect measurement tags.",
		},

		"kafka.controller.ControllerStats.IsrChangeRateAndTimeMs.StdDev": &inputs.FieldInfo{
			DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS,
			Desc: "Standard deviation of time spent in kafka controller Controller Stats Isr Change Rate And Time, measured in milliseconds, scoped by the Kafka auto-collect measurement tags.",
		},

		"kafka.controller.ControllerStats.IsrChangeRateAndTimeMs.Count": &inputs.FieldInfo{
			DataType: inputs.Int, Type: inputs.Count, Unit: inputs.NCount,
			Desc: "Total number of kafka controller Controller Stats Isr Change Rate And Time, scoped by the Kafka auto-collect measurement tags.",
		},

		"kafka.controller.ControllerStats.LeaderAndIsrResponseReceivedRateAndTimeMs.50thPercentile": &inputs.FieldInfo{
			DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS,
			Desc: "50th percentile of time spent in kafka controller Controller Stats Leader And Isr Response Received Rate And Time, measured in milliseconds, scoped by the Kafka auto-collect measurement tags.",
		},

		"kafka.controller.ControllerStats.LeaderAndIsrResponseReceivedRateAndTimeMs.75thPercentile": &inputs.FieldInfo{
			DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS,
			Desc: "75th percentile of time spent in kafka controller Controller Stats Leader And Isr Response Received Rate And Time, measured in milliseconds, scoped by the Kafka auto-collect measurement tags.",
		},

		"kafka.controller.ControllerStats.LeaderAndIsrResponseReceivedRateAndTimeMs.95thPercentile": &inputs.FieldInfo{
			DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS,
			Desc: "95th percentile of time spent in kafka controller Controller Stats Leader And Isr Response Received Rate And Time, measured in milliseconds, scoped by the Kafka auto-collect measurement tags.",
		},

		"kafka.controller.ControllerStats.LeaderAndIsrResponseReceivedRateAndTimeMs.98thPercentile": &inputs.FieldInfo{
			DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS,
			Desc: "98th percentile of time spent in kafka controller Controller Stats Leader And Isr Response Received Rate And Time, measured in milliseconds, scoped by the Kafka auto-collect measurement tags.",
		},

		"kafka.controller.ControllerStats.LeaderAndIsrResponseReceivedRateAndTimeMs.999thPercentile": &inputs.FieldInfo{
			DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS,
			Desc: "99.9th percentile of time spent in kafka controller Controller Stats Leader And Isr Response Received Rate And Time, measured in milliseconds, scoped by the Kafka auto-collect measurement tags.",
		},

		"kafka.controller.ControllerStats.LeaderAndIsrResponseReceivedRateAndTimeMs.99thPercentile": &inputs.FieldInfo{
			DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS,
			Desc: "99th percentile of time spent in kafka controller Controller Stats Leader And Isr Response Received Rate And Time, measured in milliseconds, scoped by the Kafka auto-collect measurement tags.",
		},

		"kafka.controller.ControllerStats.LeaderAndIsrResponseReceivedRateAndTimeMs.FifteenMinuteRate": &inputs.FieldInfo{
			DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount,
			Desc: "Fifteen-minute moving average rate for kafka controller Controller Stats Leader And Isr Response Received Rate And Time in events per second, scoped by the Kafka auto-collect measurement tags.",
		},

		"kafka.controller.ControllerStats.LeaderAndIsrResponseReceivedRateAndTimeMs.FiveMinuteRate": &inputs.FieldInfo{
			DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount,
			Desc: "Five-minute moving average rate for kafka controller Controller Stats Leader And Isr Response Received Rate And Time in events per second, scoped by the Kafka auto-collect measurement tags.",
		},

		"kafka.controller.ControllerStats.LeaderAndIsrResponseReceivedRateAndTimeMs.Max": &inputs.FieldInfo{
			DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS,
			Desc: "Maximum time spent in kafka controller Controller Stats Leader And Isr Response Received Rate And Time, measured in milliseconds, scoped by the Kafka auto-collect measurement tags.",
		},

		"kafka.controller.ControllerStats.LeaderAndIsrResponseReceivedRateAndTimeMs.Mean": &inputs.FieldInfo{
			DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS,
			Desc: "Average time spent in kafka controller Controller Stats Leader And Isr Response Received Rate And Time, measured in milliseconds, scoped by the Kafka auto-collect measurement tags.",
		},

		"kafka.controller.ControllerStats.LeaderAndIsrResponseReceivedRateAndTimeMs.MeanRate": &inputs.FieldInfo{
			DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount,
			Desc: "Mean rate for kafka controller Controller Stats Leader And Isr Response Received Rate And Time in events per second, scoped by the Kafka auto-collect measurement tags.",
		},

		"kafka.controller.ControllerStats.LeaderAndIsrResponseReceivedRateAndTimeMs.Min": &inputs.FieldInfo{
			DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS,
			Desc: "Minimum time spent in kafka controller Controller Stats Leader And Isr Response Received Rate And Time, measured in milliseconds, scoped by the Kafka auto-collect measurement tags.",
		},

		"kafka.controller.ControllerStats.LeaderAndIsrResponseReceivedRateAndTimeMs.OneMinuteRate": &inputs.FieldInfo{
			DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount,
			Desc: "One-minute moving average rate for kafka controller Controller Stats Leader And Isr Response Received Rate And Time in events per second, scoped by the Kafka auto-collect measurement tags.",
		},

		"kafka.controller.ControllerStats.LeaderAndIsrResponseReceivedRateAndTimeMs.StdDev": &inputs.FieldInfo{
			DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS,
			Desc: "Standard deviation of time spent in kafka controller Controller Stats Leader And Isr Response Received Rate And Time, measured in milliseconds, scoped by the Kafka auto-collect measurement tags.",
		},

		"kafka.controller.ControllerStats.LeaderAndIsrResponseReceivedRateAndTimeMs.Count": &inputs.FieldInfo{
			DataType: inputs.Int, Type: inputs.Count, Unit: inputs.NCount,
			Desc: "Total number of kafka controller Controller Stats Leader And Isr Response Received Rate And Time, scoped by the Kafka auto-collect measurement tags.",
		},

		"kafka.controller.ControllerStats.LeaderElectionRateAndTimeMs.50thPercentile": &inputs.FieldInfo{
			DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS,
			Desc: "50th percentile of time spent in kafka controller Controller Stats Leader Election Rate And Time, measured in milliseconds, scoped by the Kafka auto-collect measurement tags.",
		},

		"kafka.controller.ControllerStats.LeaderElectionRateAndTimeMs.75thPercentile": &inputs.FieldInfo{
			DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS,
			Desc: "75th percentile of time spent in kafka controller Controller Stats Leader Election Rate And Time, measured in milliseconds, scoped by the Kafka auto-collect measurement tags.",
		},

		"kafka.controller.ControllerStats.LeaderElectionRateAndTimeMs.95thPercentile": &inputs.FieldInfo{
			DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS,
			Desc: "95th percentile of time spent in kafka controller Controller Stats Leader Election Rate And Time, measured in milliseconds, scoped by the Kafka auto-collect measurement tags.",
		},

		"kafka.controller.ControllerStats.LeaderElectionRateAndTimeMs.98thPercentile": &inputs.FieldInfo{
			DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS,
			Desc: "98th percentile of time spent in kafka controller Controller Stats Leader Election Rate And Time, measured in milliseconds, scoped by the Kafka auto-collect measurement tags.",
		},

		"kafka.controller.ControllerStats.LeaderElectionRateAndTimeMs.999thPercentile": &inputs.FieldInfo{
			DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS,
			Desc: "99.9th percentile of time spent in kafka controller Controller Stats Leader Election Rate And Time, measured in milliseconds, scoped by the Kafka auto-collect measurement tags.",
		},

		"kafka.controller.ControllerStats.LeaderElectionRateAndTimeMs.99thPercentile": &inputs.FieldInfo{
			DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS,
			Desc: "99th percentile of time spent in kafka controller Controller Stats Leader Election Rate And Time, measured in milliseconds, scoped by the Kafka auto-collect measurement tags.",
		},

		"kafka.controller.ControllerStats.LeaderElectionRateAndTimeMs.FifteenMinuteRate": &inputs.FieldInfo{
			DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount,
			Desc: "Fifteen-minute moving average rate for kafka controller Controller Stats Leader Election Rate And Time in events per second, scoped by the Kafka auto-collect measurement tags.",
		},

		"kafka.controller.ControllerStats.LeaderElectionRateAndTimeMs.FiveMinuteRate": &inputs.FieldInfo{
			DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount,
			Desc: "Five-minute moving average rate for kafka controller Controller Stats Leader Election Rate And Time in events per second, scoped by the Kafka auto-collect measurement tags.",
		},

		"kafka.controller.ControllerStats.LeaderElectionRateAndTimeMs.Max": &inputs.FieldInfo{
			DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS,
			Desc: "Maximum time spent in kafka controller Controller Stats Leader Election Rate And Time, measured in milliseconds, scoped by the Kafka auto-collect measurement tags.",
		},

		"kafka.controller.ControllerStats.LeaderElectionRateAndTimeMs.Mean": &inputs.FieldInfo{
			DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS,
			Desc: "Average time spent in kafka controller Controller Stats Leader Election Rate And Time, measured in milliseconds, scoped by the Kafka auto-collect measurement tags.",
		},

		"kafka.controller.ControllerStats.LeaderElectionRateAndTimeMs.MeanRate": &inputs.FieldInfo{
			DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount,
			Desc: "Mean rate for kafka controller Controller Stats Leader Election Rate And Time in events per second, scoped by the Kafka auto-collect measurement tags.",
		},

		"kafka.controller.ControllerStats.LeaderElectionRateAndTimeMs.Min": &inputs.FieldInfo{
			DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS,
			Desc: "Minimum time spent in kafka controller Controller Stats Leader Election Rate And Time, measured in milliseconds, scoped by the Kafka auto-collect measurement tags.",
		},

		"kafka.controller.ControllerStats.LeaderElectionRateAndTimeMs.OneMinuteRate": &inputs.FieldInfo{
			DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount,
			Desc: "One-minute moving average rate for kafka controller Controller Stats Leader Election Rate And Time in events per second, scoped by the Kafka auto-collect measurement tags.",
		},

		"kafka.controller.ControllerStats.LeaderElectionRateAndTimeMs.StdDev": &inputs.FieldInfo{
			DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS,
			Desc: "Standard deviation of time spent in kafka controller Controller Stats Leader Election Rate And Time, measured in milliseconds, scoped by the Kafka auto-collect measurement tags.",
		},

		"kafka.controller.ControllerStats.LeaderElectionRateAndTimeMs.Count": &inputs.FieldInfo{
			DataType: inputs.Int, Type: inputs.Count, Unit: inputs.NCount,
			Desc: "Total number of kafka controller Controller Stats Leader Election Rate And Time, scoped by the Kafka auto-collect measurement tags.",
		},

		"kafka.controller.ControllerStats.ListPartitionReassignmentRateAndTimeMs.50thPercentile": &inputs.FieldInfo{
			DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS,
			Desc: "50th percentile of time spent in kafka controller Controller Stats List Partition Reassignment Rate And Time, measured in milliseconds, scoped by the Kafka auto-collect measurement tags.",
		},

		"kafka.controller.ControllerStats.ListPartitionReassignmentRateAndTimeMs.75thPercentile": &inputs.FieldInfo{
			DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS,
			Desc: "75th percentile of time spent in kafka controller Controller Stats List Partition Reassignment Rate And Time, measured in milliseconds, scoped by the Kafka auto-collect measurement tags.",
		},

		"kafka.controller.ControllerStats.ListPartitionReassignmentRateAndTimeMs.95thPercentile": &inputs.FieldInfo{
			DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS,
			Desc: "95th percentile of time spent in kafka controller Controller Stats List Partition Reassignment Rate And Time, measured in milliseconds, scoped by the Kafka auto-collect measurement tags.",
		},

		"kafka.controller.ControllerStats.ListPartitionReassignmentRateAndTimeMs.98thPercentile": &inputs.FieldInfo{
			DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS,
			Desc: "98th percentile of time spent in kafka controller Controller Stats List Partition Reassignment Rate And Time, measured in milliseconds, scoped by the Kafka auto-collect measurement tags.",
		},

		"kafka.controller.ControllerStats.ListPartitionReassignmentRateAndTimeMs.999thPercentile": &inputs.FieldInfo{
			DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS,
			Desc: "99.9th percentile of time spent in kafka controller Controller Stats List Partition Reassignment Rate And Time, measured in milliseconds, scoped by the Kafka auto-collect measurement tags.",
		},

		"kafka.controller.ControllerStats.ListPartitionReassignmentRateAndTimeMs.99thPercentile": &inputs.FieldInfo{
			DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS,
			Desc: "99th percentile of time spent in kafka controller Controller Stats List Partition Reassignment Rate And Time, measured in milliseconds, scoped by the Kafka auto-collect measurement tags.",
		},

		"kafka.controller.ControllerStats.ListPartitionReassignmentRateAndTimeMs.FifteenMinuteRate": &inputs.FieldInfo{
			DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount,
			Desc: "Fifteen-minute moving average rate for kafka controller Controller Stats List Partition Reassignment Rate And Time in events per second, scoped by the Kafka auto-collect measurement tags.",
		},

		"kafka.controller.ControllerStats.ListPartitionReassignmentRateAndTimeMs.FiveMinuteRate": &inputs.FieldInfo{
			DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount,
			Desc: "Five-minute moving average rate for kafka controller Controller Stats List Partition Reassignment Rate And Time in events per second, scoped by the Kafka auto-collect measurement tags.",
		},

		"kafka.controller.ControllerStats.ListPartitionReassignmentRateAndTimeMs.Max": &inputs.FieldInfo{
			DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS,
			Desc: "Maximum time spent in kafka controller Controller Stats List Partition Reassignment Rate And Time, measured in milliseconds, scoped by the Kafka auto-collect measurement tags.",
		},

		"kafka.controller.ControllerStats.ListPartitionReassignmentRateAndTimeMs.Mean": &inputs.FieldInfo{
			DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS,
			Desc: "Average time spent in kafka controller Controller Stats List Partition Reassignment Rate And Time, measured in milliseconds, scoped by the Kafka auto-collect measurement tags.",
		},

		"kafka.controller.ControllerStats.ListPartitionReassignmentRateAndTimeMs.MeanRate": &inputs.FieldInfo{
			DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount,
			Desc: "Mean rate for kafka controller Controller Stats List Partition Reassignment Rate And Time in events per second, scoped by the Kafka auto-collect measurement tags.",
		},

		"kafka.controller.ControllerStats.ListPartitionReassignmentRateAndTimeMs.Min": &inputs.FieldInfo{
			DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS,
			Desc: "Minimum time spent in kafka controller Controller Stats List Partition Reassignment Rate And Time, measured in milliseconds, scoped by the Kafka auto-collect measurement tags.",
		},

		"kafka.controller.ControllerStats.ListPartitionReassignmentRateAndTimeMs.OneMinuteRate": &inputs.FieldInfo{
			DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount,
			Desc: "One-minute moving average rate for kafka controller Controller Stats List Partition Reassignment Rate And Time in events per second, scoped by the Kafka auto-collect measurement tags.",
		},

		"kafka.controller.ControllerStats.ListPartitionReassignmentRateAndTimeMs.StdDev": &inputs.FieldInfo{
			DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS,
			Desc: "Standard deviation of time spent in kafka controller Controller Stats List Partition Reassignment Rate And Time, measured in milliseconds, scoped by the Kafka auto-collect measurement tags.",
		},

		"kafka.controller.ControllerStats.ListPartitionReassignmentRateAndTimeMs.Count": &inputs.FieldInfo{
			DataType: inputs.Int, Type: inputs.Count, Unit: inputs.NCount,
			Desc: "Total number of kafka controller Controller Stats List Partition Reassignment Rate And Time, scoped by the Kafka auto-collect measurement tags.",
		},

		"kafka.controller.ControllerStats.LogDirChangeRateAndTimeMs.50thPercentile": &inputs.FieldInfo{
			DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS,
			Desc: "50th percentile of time spent in kafka controller Controller Stats Log Dir Change Rate And Time, measured in milliseconds, scoped by the Kafka auto-collect measurement tags.",
		},

		"kafka.controller.ControllerStats.LogDirChangeRateAndTimeMs.75thPercentile": &inputs.FieldInfo{
			DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS,
			Desc: "75th percentile of time spent in kafka controller Controller Stats Log Dir Change Rate And Time, measured in milliseconds, scoped by the Kafka auto-collect measurement tags.",
		},

		"kafka.controller.ControllerStats.LogDirChangeRateAndTimeMs.95thPercentile": &inputs.FieldInfo{
			DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS,
			Desc: "95th percentile of time spent in kafka controller Controller Stats Log Dir Change Rate And Time, measured in milliseconds, scoped by the Kafka auto-collect measurement tags.",
		},

		"kafka.controller.ControllerStats.LogDirChangeRateAndTimeMs.98thPercentile": &inputs.FieldInfo{
			DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS,
			Desc: "98th percentile of time spent in kafka controller Controller Stats Log Dir Change Rate And Time, measured in milliseconds, scoped by the Kafka auto-collect measurement tags.",
		},

		"kafka.controller.ControllerStats.LogDirChangeRateAndTimeMs.999thPercentile": &inputs.FieldInfo{
			DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS,
			Desc: "99.9th percentile of time spent in kafka controller Controller Stats Log Dir Change Rate And Time, measured in milliseconds, scoped by the Kafka auto-collect measurement tags.",
		},

		"kafka.controller.ControllerStats.LogDirChangeRateAndTimeMs.99thPercentile": &inputs.FieldInfo{
			DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS,
			Desc: "99th percentile of time spent in kafka controller Controller Stats Log Dir Change Rate And Time, measured in milliseconds, scoped by the Kafka auto-collect measurement tags.",
		},

		"kafka.controller.ControllerStats.LogDirChangeRateAndTimeMs.FifteenMinuteRate": &inputs.FieldInfo{
			DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount,
			Desc: "Fifteen-minute moving average rate for kafka controller Controller Stats Log Dir Change Rate And Time in events per second, scoped by the Kafka auto-collect measurement tags.",
		},

		"kafka.controller.ControllerStats.LogDirChangeRateAndTimeMs.FiveMinuteRate": &inputs.FieldInfo{
			DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount,
			Desc: "Five-minute moving average rate for kafka controller Controller Stats Log Dir Change Rate And Time in events per second, scoped by the Kafka auto-collect measurement tags.",
		},

		"kafka.controller.ControllerStats.LogDirChangeRateAndTimeMs.Max": &inputs.FieldInfo{
			DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS,
			Desc: "Maximum time spent in kafka controller Controller Stats Log Dir Change Rate And Time, measured in milliseconds, scoped by the Kafka auto-collect measurement tags.",
		},

		"kafka.controller.ControllerStats.LogDirChangeRateAndTimeMs.Mean": &inputs.FieldInfo{
			DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS,
			Desc: "Average time spent in kafka controller Controller Stats Log Dir Change Rate And Time, measured in milliseconds, scoped by the Kafka auto-collect measurement tags.",
		},

		"kafka.controller.ControllerStats.LogDirChangeRateAndTimeMs.MeanRate": &inputs.FieldInfo{
			DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount,
			Desc: "Mean rate for kafka controller Controller Stats Log Dir Change Rate And Time in events per second, scoped by the Kafka auto-collect measurement tags.",
		},

		"kafka.controller.ControllerStats.LogDirChangeRateAndTimeMs.Min": &inputs.FieldInfo{
			DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS,
			Desc: "Minimum time spent in kafka controller Controller Stats Log Dir Change Rate And Time, measured in milliseconds, scoped by the Kafka auto-collect measurement tags.",
		},

		"kafka.controller.ControllerStats.LogDirChangeRateAndTimeMs.OneMinuteRate": &inputs.FieldInfo{
			DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount,
			Desc: "One-minute moving average rate for kafka controller Controller Stats Log Dir Change Rate And Time in events per second, scoped by the Kafka auto-collect measurement tags.",
		},

		"kafka.controller.ControllerStats.LogDirChangeRateAndTimeMs.StdDev": &inputs.FieldInfo{
			DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS,
			Desc: "Standard deviation of time spent in kafka controller Controller Stats Log Dir Change Rate And Time, measured in milliseconds, scoped by the Kafka auto-collect measurement tags.",
		},

		"kafka.controller.ControllerStats.LogDirChangeRateAndTimeMs.Count": &inputs.FieldInfo{
			DataType: inputs.Int, Type: inputs.Count, Unit: inputs.NCount,
			Desc: "Total number of kafka controller Controller Stats Log Dir Change Rate And Time, scoped by the Kafka auto-collect measurement tags.",
		},

		"kafka.controller.ControllerStats.ManualLeaderBalanceRateAndTimeMs.50thPercentile": &inputs.FieldInfo{
			DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS,
			Desc: "50th percentile of time spent in kafka controller Controller Stats Manual Leader Balance Rate And Time, measured in milliseconds, scoped by the Kafka auto-collect measurement tags.",
		},

		"kafka.controller.ControllerStats.ManualLeaderBalanceRateAndTimeMs.75thPercentile": &inputs.FieldInfo{
			DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS,
			Desc: "75th percentile of time spent in kafka controller Controller Stats Manual Leader Balance Rate And Time, measured in milliseconds, scoped by the Kafka auto-collect measurement tags.",
		},

		"kafka.controller.ControllerStats.ManualLeaderBalanceRateAndTimeMs.95thPercentile": &inputs.FieldInfo{
			DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS,
			Desc: "95th percentile of time spent in kafka controller Controller Stats Manual Leader Balance Rate And Time, measured in milliseconds, scoped by the Kafka auto-collect measurement tags.",
		},

		"kafka.controller.ControllerStats.ManualLeaderBalanceRateAndTimeMs.98thPercentile": &inputs.FieldInfo{
			DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS,
			Desc: "98th percentile of time spent in kafka controller Controller Stats Manual Leader Balance Rate And Time, measured in milliseconds, scoped by the Kafka auto-collect measurement tags.",
		},

		"kafka.controller.ControllerStats.ManualLeaderBalanceRateAndTimeMs.999thPercentile": &inputs.FieldInfo{
			DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS,
			Desc: "99.9th percentile of time spent in kafka controller Controller Stats Manual Leader Balance Rate And Time, measured in milliseconds, scoped by the Kafka auto-collect measurement tags.",
		},

		"kafka.controller.ControllerStats.ManualLeaderBalanceRateAndTimeMs.99thPercentile": &inputs.FieldInfo{
			DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS,
			Desc: "99th percentile of time spent in kafka controller Controller Stats Manual Leader Balance Rate And Time, measured in milliseconds, scoped by the Kafka auto-collect measurement tags.",
		},

		"kafka.controller.ControllerStats.ManualLeaderBalanceRateAndTimeMs.FifteenMinuteRate": &inputs.FieldInfo{
			DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount,
			Desc: "Fifteen-minute moving average rate for kafka controller Controller Stats Manual Leader Balance Rate And Time in events per second, scoped by the Kafka auto-collect measurement tags.",
		},

		"kafka.controller.ControllerStats.ManualLeaderBalanceRateAndTimeMs.FiveMinuteRate": &inputs.FieldInfo{
			DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount,
			Desc: "Five-minute moving average rate for kafka controller Controller Stats Manual Leader Balance Rate And Time in events per second, scoped by the Kafka auto-collect measurement tags.",
		},

		"kafka.controller.ControllerStats.ManualLeaderBalanceRateAndTimeMs.Max": &inputs.FieldInfo{
			DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS,
			Desc: "Maximum time spent in kafka controller Controller Stats Manual Leader Balance Rate And Time, measured in milliseconds, scoped by the Kafka auto-collect measurement tags.",
		},

		"kafka.controller.ControllerStats.ManualLeaderBalanceRateAndTimeMs.Mean": &inputs.FieldInfo{
			DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS,
			Desc: "Average time spent in kafka controller Controller Stats Manual Leader Balance Rate And Time, measured in milliseconds, scoped by the Kafka auto-collect measurement tags.",
		},

		"kafka.controller.ControllerStats.ManualLeaderBalanceRateAndTimeMs.MeanRate": &inputs.FieldInfo{
			DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount,
			Desc: "Mean rate for kafka controller Controller Stats Manual Leader Balance Rate And Time in events per second, scoped by the Kafka auto-collect measurement tags.",
		},

		"kafka.controller.ControllerStats.ManualLeaderBalanceRateAndTimeMs.Min": &inputs.FieldInfo{
			DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS,
			Desc: "Minimum time spent in kafka controller Controller Stats Manual Leader Balance Rate And Time, measured in milliseconds, scoped by the Kafka auto-collect measurement tags.",
		},

		"kafka.controller.ControllerStats.ManualLeaderBalanceRateAndTimeMs.OneMinuteRate": &inputs.FieldInfo{
			DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount,
			Desc: "One-minute moving average rate for kafka controller Controller Stats Manual Leader Balance Rate And Time in events per second, scoped by the Kafka auto-collect measurement tags.",
		},

		"kafka.controller.ControllerStats.ManualLeaderBalanceRateAndTimeMs.StdDev": &inputs.FieldInfo{
			DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS,
			Desc: "Standard deviation of time spent in kafka controller Controller Stats Manual Leader Balance Rate And Time, measured in milliseconds, scoped by the Kafka auto-collect measurement tags.",
		},

		"kafka.controller.ControllerStats.ManualLeaderBalanceRateAndTimeMs.Count": &inputs.FieldInfo{
			DataType: inputs.Int, Type: inputs.Count, Unit: inputs.NCount,
			Desc: "Total number of kafka controller Controller Stats Manual Leader Balance Rate And Time, scoped by the Kafka auto-collect measurement tags.",
		},

		"kafka.controller.ControllerStats.PartitionReassignmentRateAndTimeMs.50thPercentile": &inputs.FieldInfo{
			DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS,
			Desc: "50th percentile of time spent in kafka controller Controller Stats Partition Reassignment Rate And Time, measured in milliseconds, scoped by the Kafka auto-collect measurement tags.",
		},

		"kafka.controller.ControllerStats.PartitionReassignmentRateAndTimeMs.75thPercentile": &inputs.FieldInfo{
			DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS,
			Desc: "75th percentile of time spent in kafka controller Controller Stats Partition Reassignment Rate And Time, measured in milliseconds, scoped by the Kafka auto-collect measurement tags.",
		},

		"kafka.controller.ControllerStats.PartitionReassignmentRateAndTimeMs.95thPercentile": &inputs.FieldInfo{
			DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS,
			Desc: "95th percentile of time spent in kafka controller Controller Stats Partition Reassignment Rate And Time, measured in milliseconds, scoped by the Kafka auto-collect measurement tags.",
		},

		"kafka.controller.ControllerStats.PartitionReassignmentRateAndTimeMs.98thPercentile": &inputs.FieldInfo{
			DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS,
			Desc: "98th percentile of time spent in kafka controller Controller Stats Partition Reassignment Rate And Time, measured in milliseconds, scoped by the Kafka auto-collect measurement tags.",
		},

		"kafka.controller.ControllerStats.PartitionReassignmentRateAndTimeMs.999thPercentile": &inputs.FieldInfo{
			DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS,
			Desc: "99.9th percentile of time spent in kafka controller Controller Stats Partition Reassignment Rate And Time, measured in milliseconds, scoped by the Kafka auto-collect measurement tags.",
		},

		"kafka.controller.ControllerStats.PartitionReassignmentRateAndTimeMs.99thPercentile": &inputs.FieldInfo{
			DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS,
			Desc: "99th percentile of time spent in kafka controller Controller Stats Partition Reassignment Rate And Time, measured in milliseconds, scoped by the Kafka auto-collect measurement tags.",
		},

		"kafka.controller.ControllerStats.PartitionReassignmentRateAndTimeMs.FifteenMinuteRate": &inputs.FieldInfo{
			DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount,
			Desc: "Fifteen-minute moving average rate for kafka controller Controller Stats Partition Reassignment Rate And Time in events per second, scoped by the Kafka auto-collect measurement tags.",
		},

		"kafka.controller.ControllerStats.PartitionReassignmentRateAndTimeMs.FiveMinuteRate": &inputs.FieldInfo{
			DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount,
			Desc: "Five-minute moving average rate for kafka controller Controller Stats Partition Reassignment Rate And Time in events per second, scoped by the Kafka auto-collect measurement tags.",
		},

		"kafka.controller.ControllerStats.PartitionReassignmentRateAndTimeMs.Max": &inputs.FieldInfo{
			DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS,
			Desc: "Maximum time spent in kafka controller Controller Stats Partition Reassignment Rate And Time, measured in milliseconds, scoped by the Kafka auto-collect measurement tags.",
		},

		"kafka.controller.ControllerStats.PartitionReassignmentRateAndTimeMs.Mean": &inputs.FieldInfo{
			DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS,
			Desc: "Average time spent in kafka controller Controller Stats Partition Reassignment Rate And Time, measured in milliseconds, scoped by the Kafka auto-collect measurement tags.",
		},

		"kafka.controller.ControllerStats.PartitionReassignmentRateAndTimeMs.MeanRate": &inputs.FieldInfo{
			DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount,
			Desc: "Mean rate for kafka controller Controller Stats Partition Reassignment Rate And Time in events per second, scoped by the Kafka auto-collect measurement tags.",
		},

		"kafka.controller.ControllerStats.PartitionReassignmentRateAndTimeMs.Min": &inputs.FieldInfo{
			DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS,
			Desc: "Minimum time spent in kafka controller Controller Stats Partition Reassignment Rate And Time, measured in milliseconds, scoped by the Kafka auto-collect measurement tags.",
		},

		"kafka.controller.ControllerStats.PartitionReassignmentRateAndTimeMs.OneMinuteRate": &inputs.FieldInfo{
			DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount,
			Desc: "One-minute moving average rate for kafka controller Controller Stats Partition Reassignment Rate And Time in events per second, scoped by the Kafka auto-collect measurement tags.",
		},

		"kafka.controller.ControllerStats.PartitionReassignmentRateAndTimeMs.StdDev": &inputs.FieldInfo{
			DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS,
			Desc: "Standard deviation of time spent in kafka controller Controller Stats Partition Reassignment Rate And Time, measured in milliseconds, scoped by the Kafka auto-collect measurement tags.",
		},

		"kafka.controller.ControllerStats.PartitionReassignmentRateAndTimeMs.Count": &inputs.FieldInfo{
			DataType: inputs.Int, Type: inputs.Count, Unit: inputs.NCount,
			Desc: "Total number of kafka controller Controller Stats Partition Reassignment Rate And Time, scoped by the Kafka auto-collect measurement tags.",
		},

		"kafka.controller.ControllerStats.TopicChangeRateAndTimeMs.50thPercentile": &inputs.FieldInfo{
			DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS,
			Desc: "50th percentile of time spent in kafka controller Controller Stats Topic Change Rate And Time, measured in milliseconds, scoped by the Kafka auto-collect measurement tags.",
		},

		"kafka.controller.ControllerStats.TopicChangeRateAndTimeMs.75thPercentile": &inputs.FieldInfo{
			DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS,
			Desc: "75th percentile of time spent in kafka controller Controller Stats Topic Change Rate And Time, measured in milliseconds, scoped by the Kafka auto-collect measurement tags.",
		},

		"kafka.controller.ControllerStats.TopicChangeRateAndTimeMs.95thPercentile": &inputs.FieldInfo{
			DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS,
			Desc: "95th percentile of time spent in kafka controller Controller Stats Topic Change Rate And Time, measured in milliseconds, scoped by the Kafka auto-collect measurement tags.",
		},

		"kafka.controller.ControllerStats.TopicChangeRateAndTimeMs.98thPercentile": &inputs.FieldInfo{
			DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS,
			Desc: "98th percentile of time spent in kafka controller Controller Stats Topic Change Rate And Time, measured in milliseconds, scoped by the Kafka auto-collect measurement tags.",
		},

		"kafka.controller.ControllerStats.TopicChangeRateAndTimeMs.999thPercentile": &inputs.FieldInfo{
			DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS,
			Desc: "99.9th percentile of time spent in kafka controller Controller Stats Topic Change Rate And Time, measured in milliseconds, scoped by the Kafka auto-collect measurement tags.",
		},

		"kafka.controller.ControllerStats.TopicChangeRateAndTimeMs.99thPercentile": &inputs.FieldInfo{
			DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS,
			Desc: "99th percentile of time spent in kafka controller Controller Stats Topic Change Rate And Time, measured in milliseconds, scoped by the Kafka auto-collect measurement tags.",
		},

		"kafka.controller.ControllerStats.TopicChangeRateAndTimeMs.FifteenMinuteRate": &inputs.FieldInfo{
			DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount,
			Desc: "Fifteen-minute moving average rate for kafka controller Controller Stats Topic Change Rate And Time in events per second, scoped by the Kafka auto-collect measurement tags.",
		},

		"kafka.controller.ControllerStats.TopicChangeRateAndTimeMs.FiveMinuteRate": &inputs.FieldInfo{
			DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount,
			Desc: "Five-minute moving average rate for kafka controller Controller Stats Topic Change Rate And Time in events per second, scoped by the Kafka auto-collect measurement tags.",
		},

		"kafka.controller.ControllerStats.TopicChangeRateAndTimeMs.Max": &inputs.FieldInfo{
			DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS,
			Desc: "Maximum time spent in kafka controller Controller Stats Topic Change Rate And Time, measured in milliseconds, scoped by the Kafka auto-collect measurement tags.",
		},

		"kafka.controller.ControllerStats.TopicChangeRateAndTimeMs.Mean": &inputs.FieldInfo{
			DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS,
			Desc: "Average time spent in kafka controller Controller Stats Topic Change Rate And Time, measured in milliseconds, scoped by the Kafka auto-collect measurement tags.",
		},

		"kafka.controller.ControllerStats.TopicChangeRateAndTimeMs.MeanRate": &inputs.FieldInfo{
			DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount,
			Desc: "Mean rate for kafka controller Controller Stats Topic Change Rate And Time in events per second, scoped by the Kafka auto-collect measurement tags.",
		},

		"kafka.controller.ControllerStats.TopicChangeRateAndTimeMs.Min": &inputs.FieldInfo{
			DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS,
			Desc: "Minimum time spent in kafka controller Controller Stats Topic Change Rate And Time, measured in milliseconds, scoped by the Kafka auto-collect measurement tags.",
		},

		"kafka.controller.ControllerStats.TopicChangeRateAndTimeMs.OneMinuteRate": &inputs.FieldInfo{
			DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount,
			Desc: "One-minute moving average rate for kafka controller Controller Stats Topic Change Rate And Time in events per second, scoped by the Kafka auto-collect measurement tags.",
		},

		"kafka.controller.ControllerStats.TopicChangeRateAndTimeMs.StdDev": &inputs.FieldInfo{
			DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS,
			Desc: "Standard deviation of time spent in kafka controller Controller Stats Topic Change Rate And Time, measured in milliseconds, scoped by the Kafka auto-collect measurement tags.",
		},

		"kafka.controller.ControllerStats.TopicChangeRateAndTimeMs.Count": &inputs.FieldInfo{
			DataType: inputs.Int, Type: inputs.Count, Unit: inputs.NCount,
			Desc: "Total number of kafka controller Controller Stats Topic Change Rate And Time, scoped by the Kafka auto-collect measurement tags.",
		},

		"kafka.controller.ControllerStats.TopicDeletionRateAndTimeMs.50thPercentile": &inputs.FieldInfo{
			DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS,
			Desc: "50th percentile of time spent in kafka controller Controller Stats Topic Deletion Rate And Time, measured in milliseconds, scoped by the Kafka auto-collect measurement tags.",
		},

		"kafka.controller.ControllerStats.TopicDeletionRateAndTimeMs.75thPercentile": &inputs.FieldInfo{
			DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS,
			Desc: "75th percentile of time spent in kafka controller Controller Stats Topic Deletion Rate And Time, measured in milliseconds, scoped by the Kafka auto-collect measurement tags.",
		},

		"kafka.controller.ControllerStats.TopicDeletionRateAndTimeMs.95thPercentile": &inputs.FieldInfo{
			DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS,
			Desc: "95th percentile of time spent in kafka controller Controller Stats Topic Deletion Rate And Time, measured in milliseconds, scoped by the Kafka auto-collect measurement tags.",
		},

		"kafka.controller.ControllerStats.TopicDeletionRateAndTimeMs.98thPercentile": &inputs.FieldInfo{
			DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS,
			Desc: "98th percentile of time spent in kafka controller Controller Stats Topic Deletion Rate And Time, measured in milliseconds, scoped by the Kafka auto-collect measurement tags.",
		},

		"kafka.controller.ControllerStats.TopicDeletionRateAndTimeMs.999thPercentile": &inputs.FieldInfo{
			DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS,
			Desc: "99.9th percentile of time spent in kafka controller Controller Stats Topic Deletion Rate And Time, measured in milliseconds, scoped by the Kafka auto-collect measurement tags.",
		},

		"kafka.controller.ControllerStats.TopicDeletionRateAndTimeMs.99thPercentile": &inputs.FieldInfo{
			DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS,
			Desc: "99th percentile of time spent in kafka controller Controller Stats Topic Deletion Rate And Time, measured in milliseconds, scoped by the Kafka auto-collect measurement tags.",
		},

		"kafka.controller.ControllerStats.TopicDeletionRateAndTimeMs.FifteenMinuteRate": &inputs.FieldInfo{
			DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount,
			Desc: "Fifteen-minute moving average rate for kafka controller Controller Stats Topic Deletion Rate And Time in events per second, scoped by the Kafka auto-collect measurement tags.",
		},

		"kafka.controller.ControllerStats.TopicDeletionRateAndTimeMs.FiveMinuteRate": &inputs.FieldInfo{
			DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount,
			Desc: "Five-minute moving average rate for kafka controller Controller Stats Topic Deletion Rate And Time in events per second, scoped by the Kafka auto-collect measurement tags.",
		},

		"kafka.controller.ControllerStats.TopicDeletionRateAndTimeMs.Max": &inputs.FieldInfo{
			DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS,
			Desc: "Maximum time spent in kafka controller Controller Stats Topic Deletion Rate And Time, measured in milliseconds, scoped by the Kafka auto-collect measurement tags.",
		},

		"kafka.controller.ControllerStats.TopicDeletionRateAndTimeMs.Mean": &inputs.FieldInfo{
			DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS,
			Desc: "Average time spent in kafka controller Controller Stats Topic Deletion Rate And Time, measured in milliseconds, scoped by the Kafka auto-collect measurement tags.",
		},

		"kafka.controller.ControllerStats.TopicDeletionRateAndTimeMs.MeanRate": &inputs.FieldInfo{
			DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount,
			Desc: "Mean rate for kafka controller Controller Stats Topic Deletion Rate And Time in events per second, scoped by the Kafka auto-collect measurement tags.",
		},

		"kafka.controller.ControllerStats.TopicDeletionRateAndTimeMs.Min": &inputs.FieldInfo{
			DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS,
			Desc: "Minimum time spent in kafka controller Controller Stats Topic Deletion Rate And Time, measured in milliseconds, scoped by the Kafka auto-collect measurement tags.",
		},

		"kafka.controller.ControllerStats.TopicDeletionRateAndTimeMs.OneMinuteRate": &inputs.FieldInfo{
			DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount,
			Desc: "One-minute moving average rate for kafka controller Controller Stats Topic Deletion Rate And Time in events per second, scoped by the Kafka auto-collect measurement tags.",
		},

		"kafka.controller.ControllerStats.TopicDeletionRateAndTimeMs.StdDev": &inputs.FieldInfo{
			DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS,
			Desc: "Standard deviation of time spent in kafka controller Controller Stats Topic Deletion Rate And Time, measured in milliseconds, scoped by the Kafka auto-collect measurement tags.",
		},

		"kafka.controller.ControllerStats.TopicDeletionRateAndTimeMs.Count": &inputs.FieldInfo{
			DataType: inputs.Int, Type: inputs.Count, Unit: inputs.NCount,
			Desc: "Total number of kafka controller Controller Stats Topic Deletion Rate And Time, scoped by the Kafka auto-collect measurement tags.",
		},

		"kafka.controller.ControllerStats.TopicUncleanLeaderElectionEnableRateAndTimeMs.50thPercentile": &inputs.FieldInfo{
			DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS,
			Desc: "50th percentile of time spent in kafka controller Controller Stats Topic Unclean Leader Election Enable Rate And Time, measured in milliseconds, scoped by the Kafka auto-collect measurement tags.",
		},

		"kafka.controller.ControllerStats.TopicUncleanLeaderElectionEnableRateAndTimeMs.75thPercentile": &inputs.FieldInfo{
			DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS,
			Desc: "75th percentile of time spent in kafka controller Controller Stats Topic Unclean Leader Election Enable Rate And Time, measured in milliseconds, scoped by the Kafka auto-collect measurement tags.",
		},

		"kafka.controller.ControllerStats.TopicUncleanLeaderElectionEnableRateAndTimeMs.95thPercentile": &inputs.FieldInfo{
			DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS,
			Desc: "95th percentile of time spent in kafka controller Controller Stats Topic Unclean Leader Election Enable Rate And Time, measured in milliseconds, scoped by the Kafka auto-collect measurement tags.",
		},

		"kafka.controller.ControllerStats.TopicUncleanLeaderElectionEnableRateAndTimeMs.98thPercentile": &inputs.FieldInfo{
			DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS,
			Desc: "98th percentile of time spent in kafka controller Controller Stats Topic Unclean Leader Election Enable Rate And Time, measured in milliseconds, scoped by the Kafka auto-collect measurement tags.",
		},

		"kafka.controller.ControllerStats.TopicUncleanLeaderElectionEnableRateAndTimeMs.999thPercentile": &inputs.FieldInfo{
			DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS,
			Desc: "99.9th percentile of time spent in kafka controller Controller Stats Topic Unclean Leader Election Enable Rate And Time, measured in milliseconds, scoped by the Kafka auto-collect measurement tags.",
		},

		"kafka.controller.ControllerStats.TopicUncleanLeaderElectionEnableRateAndTimeMs.99thPercentile": &inputs.FieldInfo{
			DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS,
			Desc: "99th percentile of time spent in kafka controller Controller Stats Topic Unclean Leader Election Enable Rate And Time, measured in milliseconds, scoped by the Kafka auto-collect measurement tags.",
		},

		"kafka.controller.ControllerStats.TopicUncleanLeaderElectionEnableRateAndTimeMs.FifteenMinuteRate": &inputs.FieldInfo{
			DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount,
			Desc: "Fifteen-minute moving average rate for kafka controller Controller Stats Topic Unclean Leader Election Enable Rate And Time in events per second, scoped by the Kafka auto-collect measurement tags.",
		},

		"kafka.controller.ControllerStats.TopicUncleanLeaderElectionEnableRateAndTimeMs.FiveMinuteRate": &inputs.FieldInfo{
			DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount,
			Desc: "Five-minute moving average rate for kafka controller Controller Stats Topic Unclean Leader Election Enable Rate And Time in events per second, scoped by the Kafka auto-collect measurement tags.",
		},

		"kafka.controller.ControllerStats.TopicUncleanLeaderElectionEnableRateAndTimeMs.Max": &inputs.FieldInfo{
			DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS,
			Desc: "Maximum time spent in kafka controller Controller Stats Topic Unclean Leader Election Enable Rate And Time, measured in milliseconds, scoped by the Kafka auto-collect measurement tags.",
		},

		"kafka.controller.ControllerStats.TopicUncleanLeaderElectionEnableRateAndTimeMs.Mean": &inputs.FieldInfo{
			DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS,
			Desc: "Average time spent in kafka controller Controller Stats Topic Unclean Leader Election Enable Rate And Time, measured in milliseconds, scoped by the Kafka auto-collect measurement tags.",
		},

		"kafka.controller.ControllerStats.TopicUncleanLeaderElectionEnableRateAndTimeMs.MeanRate": &inputs.FieldInfo{
			DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount,
			Desc: "Mean rate for kafka controller Controller Stats Topic Unclean Leader Election Enable Rate And Time in events per second, scoped by the Kafka auto-collect measurement tags.",
		},

		"kafka.controller.ControllerStats.TopicUncleanLeaderElectionEnableRateAndTimeMs.Min": &inputs.FieldInfo{
			DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS,
			Desc: "Minimum time spent in kafka controller Controller Stats Topic Unclean Leader Election Enable Rate And Time, measured in milliseconds, scoped by the Kafka auto-collect measurement tags.",
		},

		"kafka.controller.ControllerStats.TopicUncleanLeaderElectionEnableRateAndTimeMs.OneMinuteRate": &inputs.FieldInfo{
			DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount,
			Desc: "One-minute moving average rate for kafka controller Controller Stats Topic Unclean Leader Election Enable Rate And Time in events per second, scoped by the Kafka auto-collect measurement tags.",
		},

		"kafka.controller.ControllerStats.TopicUncleanLeaderElectionEnableRateAndTimeMs.StdDev": &inputs.FieldInfo{
			DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS,
			Desc: "Standard deviation of time spent in kafka controller Controller Stats Topic Unclean Leader Election Enable Rate And Time, measured in milliseconds, scoped by the Kafka auto-collect measurement tags.",
		},

		"kafka.controller.ControllerStats.TopicUncleanLeaderElectionEnableRateAndTimeMs.Count": &inputs.FieldInfo{
			DataType: inputs.Int, Type: inputs.Count, Unit: inputs.NCount,
			Desc: "Total number of kafka controller Controller Stats Topic Unclean Leader Election Enable Rate And Time, scoped by the Kafka auto-collect measurement tags.",
		},

		"kafka.controller.ControllerStats.UncleanLeaderElectionEnableRateAndTimeMs.50thPercentile": &inputs.FieldInfo{
			DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS,
			Desc: "50th percentile of time spent in kafka controller Controller Stats Unclean Leader Election Enable Rate And Time, measured in milliseconds, scoped by the Kafka auto-collect measurement tags.",
		},

		"kafka.controller.ControllerStats.UncleanLeaderElectionEnableRateAndTimeMs.75thPercentile": &inputs.FieldInfo{
			DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS,
			Desc: "75th percentile of time spent in kafka controller Controller Stats Unclean Leader Election Enable Rate And Time, measured in milliseconds, scoped by the Kafka auto-collect measurement tags.",
		},

		"kafka.controller.ControllerStats.UncleanLeaderElectionEnableRateAndTimeMs.95thPercentile": &inputs.FieldInfo{
			DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS,
			Desc: "95th percentile of time spent in kafka controller Controller Stats Unclean Leader Election Enable Rate And Time, measured in milliseconds, scoped by the Kafka auto-collect measurement tags.",
		},

		"kafka.controller.ControllerStats.UncleanLeaderElectionEnableRateAndTimeMs.98thPercentile": &inputs.FieldInfo{
			DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS,
			Desc: "98th percentile of time spent in kafka controller Controller Stats Unclean Leader Election Enable Rate And Time, measured in milliseconds, scoped by the Kafka auto-collect measurement tags.",
		},

		"kafka.controller.ControllerStats.UncleanLeaderElectionEnableRateAndTimeMs.999thPercentile": &inputs.FieldInfo{
			DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS,
			Desc: "99.9th percentile of time spent in kafka controller Controller Stats Unclean Leader Election Enable Rate And Time, measured in milliseconds, scoped by the Kafka auto-collect measurement tags.",
		},

		"kafka.controller.ControllerStats.UncleanLeaderElectionEnableRateAndTimeMs.99thPercentile": &inputs.FieldInfo{
			DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS,
			Desc: "99th percentile of time spent in kafka controller Controller Stats Unclean Leader Election Enable Rate And Time, measured in milliseconds, scoped by the Kafka auto-collect measurement tags.",
		},

		"kafka.controller.ControllerStats.UncleanLeaderElectionEnableRateAndTimeMs.FifteenMinuteRate": &inputs.FieldInfo{
			DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount,
			Desc: "Fifteen-minute moving average rate for kafka controller Controller Stats Unclean Leader Election Enable Rate And Time in events per second, scoped by the Kafka auto-collect measurement tags.",
		},

		"kafka.controller.ControllerStats.UncleanLeaderElectionEnableRateAndTimeMs.FiveMinuteRate": &inputs.FieldInfo{
			DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount,
			Desc: "Five-minute moving average rate for kafka controller Controller Stats Unclean Leader Election Enable Rate And Time in events per second, scoped by the Kafka auto-collect measurement tags.",
		},

		"kafka.controller.ControllerStats.UncleanLeaderElectionEnableRateAndTimeMs.Max": &inputs.FieldInfo{
			DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS,
			Desc: "Maximum time spent in kafka controller Controller Stats Unclean Leader Election Enable Rate And Time, measured in milliseconds, scoped by the Kafka auto-collect measurement tags.",
		},

		"kafka.controller.ControllerStats.UncleanLeaderElectionEnableRateAndTimeMs.Mean": &inputs.FieldInfo{
			DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS,
			Desc: "Average time spent in kafka controller Controller Stats Unclean Leader Election Enable Rate And Time, measured in milliseconds, scoped by the Kafka auto-collect measurement tags.",
		},

		"kafka.controller.ControllerStats.UncleanLeaderElectionEnableRateAndTimeMs.MeanRate": &inputs.FieldInfo{
			DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount,
			Desc: "Mean rate for kafka controller Controller Stats Unclean Leader Election Enable Rate And Time in events per second, scoped by the Kafka auto-collect measurement tags.",
		},

		"kafka.controller.ControllerStats.UncleanLeaderElectionEnableRateAndTimeMs.Min": &inputs.FieldInfo{
			DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS,
			Desc: "Minimum time spent in kafka controller Controller Stats Unclean Leader Election Enable Rate And Time, measured in milliseconds, scoped by the Kafka auto-collect measurement tags.",
		},

		"kafka.controller.ControllerStats.UncleanLeaderElectionEnableRateAndTimeMs.OneMinuteRate": &inputs.FieldInfo{
			DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount,
			Desc: "One-minute moving average rate for kafka controller Controller Stats Unclean Leader Election Enable Rate And Time in events per second, scoped by the Kafka auto-collect measurement tags.",
		},

		"kafka.controller.ControllerStats.UncleanLeaderElectionEnableRateAndTimeMs.StdDev": &inputs.FieldInfo{
			DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS,
			Desc: "Standard deviation of time spent in kafka controller Controller Stats Unclean Leader Election Enable Rate And Time, measured in milliseconds, scoped by the Kafka auto-collect measurement tags.",
		},

		"kafka.controller.ControllerStats.UncleanLeaderElectionEnableRateAndTimeMs.Count": &inputs.FieldInfo{
			DataType: inputs.Int, Type: inputs.Count, Unit: inputs.NCount,
			Desc: "Total number of kafka controller Controller Stats Unclean Leader Election Enable Rate And Time, scoped by the Kafka auto-collect measurement tags.",
		},

		"kafka.controller.ControllerStats.UpdateFeaturesRateAndTimeMs.50thPercentile": &inputs.FieldInfo{
			DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS,
			Desc: "50th percentile of time spent in kafka controller Controller Stats Update Features Rate And Time, measured in milliseconds, scoped by the Kafka auto-collect measurement tags.",
		},

		"kafka.controller.ControllerStats.UpdateFeaturesRateAndTimeMs.75thPercentile": &inputs.FieldInfo{
			DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS,
			Desc: "75th percentile of time spent in kafka controller Controller Stats Update Features Rate And Time, measured in milliseconds, scoped by the Kafka auto-collect measurement tags.",
		},

		"kafka.controller.ControllerStats.UpdateFeaturesRateAndTimeMs.95thPercentile": &inputs.FieldInfo{
			DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS,
			Desc: "95th percentile of time spent in kafka controller Controller Stats Update Features Rate And Time, measured in milliseconds, scoped by the Kafka auto-collect measurement tags.",
		},

		"kafka.controller.ControllerStats.UpdateFeaturesRateAndTimeMs.98thPercentile": &inputs.FieldInfo{
			DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS,
			Desc: "98th percentile of time spent in kafka controller Controller Stats Update Features Rate And Time, measured in milliseconds, scoped by the Kafka auto-collect measurement tags.",
		},

		"kafka.controller.ControllerStats.UpdateFeaturesRateAndTimeMs.999thPercentile": &inputs.FieldInfo{
			DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS,
			Desc: "99.9th percentile of time spent in kafka controller Controller Stats Update Features Rate And Time, measured in milliseconds, scoped by the Kafka auto-collect measurement tags.",
		},

		"kafka.controller.ControllerStats.UpdateFeaturesRateAndTimeMs.99thPercentile": &inputs.FieldInfo{
			DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS,
			Desc: "99th percentile of time spent in kafka controller Controller Stats Update Features Rate And Time, measured in milliseconds, scoped by the Kafka auto-collect measurement tags.",
		},

		"kafka.controller.ControllerStats.UpdateFeaturesRateAndTimeMs.FifteenMinuteRate": &inputs.FieldInfo{
			DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount,
			Desc: "Fifteen-minute moving average rate for kafka controller Controller Stats Update Features Rate And Time in events per second, scoped by the Kafka auto-collect measurement tags.",
		},

		"kafka.controller.ControllerStats.UpdateFeaturesRateAndTimeMs.FiveMinuteRate": &inputs.FieldInfo{
			DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount,
			Desc: "Five-minute moving average rate for kafka controller Controller Stats Update Features Rate And Time in events per second, scoped by the Kafka auto-collect measurement tags.",
		},

		"kafka.controller.ControllerStats.UpdateFeaturesRateAndTimeMs.Max": &inputs.FieldInfo{
			DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS,
			Desc: "Maximum time spent in kafka controller Controller Stats Update Features Rate And Time, measured in milliseconds, scoped by the Kafka auto-collect measurement tags.",
		},

		"kafka.controller.ControllerStats.UpdateFeaturesRateAndTimeMs.Mean": &inputs.FieldInfo{
			DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS,
			Desc: "Average time spent in kafka controller Controller Stats Update Features Rate And Time, measured in milliseconds, scoped by the Kafka auto-collect measurement tags.",
		},

		"kafka.controller.ControllerStats.UpdateFeaturesRateAndTimeMs.MeanRate": &inputs.FieldInfo{
			DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount,
			Desc: "Mean rate for kafka controller Controller Stats Update Features Rate And Time in events per second, scoped by the Kafka auto-collect measurement tags.",
		},

		"kafka.controller.ControllerStats.UpdateFeaturesRateAndTimeMs.Min": &inputs.FieldInfo{
			DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS,
			Desc: "Minimum time spent in kafka controller Controller Stats Update Features Rate And Time, measured in milliseconds, scoped by the Kafka auto-collect measurement tags.",
		},

		"kafka.controller.ControllerStats.UpdateFeaturesRateAndTimeMs.OneMinuteRate": &inputs.FieldInfo{
			DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount,
			Desc: "One-minute moving average rate for kafka controller Controller Stats Update Features Rate And Time in events per second, scoped by the Kafka auto-collect measurement tags.",
		},

		"kafka.controller.ControllerStats.UpdateFeaturesRateAndTimeMs.StdDev": &inputs.FieldInfo{
			DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS,
			Desc: "Standard deviation of time spent in kafka controller Controller Stats Update Features Rate And Time, measured in milliseconds, scoped by the Kafka auto-collect measurement tags.",
		},

		"kafka.controller.ControllerStats.UpdateFeaturesRateAndTimeMs.Count": &inputs.FieldInfo{
			DataType: inputs.Int, Type: inputs.Count, Unit: inputs.NCount,
			Desc: "Total number of kafka controller Controller Stats Update Features Rate And Time, scoped by the Kafka auto-collect measurement tags.",
		},

		"kafka.controller.ControllerStats.UncleanLeaderElectionsPerSec.FifteenMinuteRate": &inputs.FieldInfo{
			DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount,
			Desc: "Fifteen-minute moving average rate for kafka controller Controller Stats Unclean Leader Elections Per Sec in events per second, scoped by the Kafka auto-collect measurement tags.",
		},

		"kafka.controller.ControllerStats.UncleanLeaderElectionsPerSec.FiveMinuteRate": &inputs.FieldInfo{
			DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount,
			Desc: "Five-minute moving average rate for kafka controller Controller Stats Unclean Leader Elections Per Sec in events per second, scoped by the Kafka auto-collect measurement tags.",
		},

		"kafka.controller.ControllerStats.UncleanLeaderElectionsPerSec.MeanRate": &inputs.FieldInfo{
			DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount,
			Desc: "Mean rate for kafka controller Controller Stats Unclean Leader Elections Per Sec in events per second, scoped by the Kafka auto-collect measurement tags.",
		},

		"kafka.controller.ControllerStats.UncleanLeaderElectionsPerSec.OneMinuteRate": &inputs.FieldInfo{
			DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount,
			Desc: "One-minute moving average rate for kafka controller Controller Stats Unclean Leader Elections Per Sec in events per second, scoped by the Kafka auto-collect measurement tags.",
		},

		"kafka.controller.ControllerStats.UncleanLeaderElectionsPerSec.Count": &inputs.FieldInfo{
			DataType: inputs.Int, Type: inputs.Count, Unit: inputs.NCount,
			Desc: "Total number of kafka controller Controller Stats Unclean Leader Elections Per Sec, scoped by the Kafka auto-collect measurement tags.",
		},

		"kafka.controller.ControllerEventManager.EventQueueTimeMs.50thPercentile": &inputs.FieldInfo{
			DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS,
			Desc: "50th percentile of time spent in kafka controller Controller Event Manager Event Queue Time, measured in milliseconds, scoped by the Kafka auto-collect measurement tags.",
		},

		"kafka.controller.ControllerEventManager.EventQueueTimeMs.75thPercentile": &inputs.FieldInfo{
			DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS,
			Desc: "75th percentile of time spent in kafka controller Controller Event Manager Event Queue Time, measured in milliseconds, scoped by the Kafka auto-collect measurement tags.",
		},

		"kafka.controller.ControllerEventManager.EventQueueTimeMs.95thPercentile": &inputs.FieldInfo{
			DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS,
			Desc: "95th percentile of time spent in kafka controller Controller Event Manager Event Queue Time, measured in milliseconds, scoped by the Kafka auto-collect measurement tags.",
		},

		"kafka.controller.ControllerEventManager.EventQueueTimeMs.98thPercentile": &inputs.FieldInfo{
			DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS,
			Desc: "98th percentile of time spent in kafka controller Controller Event Manager Event Queue Time, measured in milliseconds, scoped by the Kafka auto-collect measurement tags.",
		},

		"kafka.controller.ControllerEventManager.EventQueueTimeMs.999thPercentile": &inputs.FieldInfo{
			DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS,
			Desc: "99.9th percentile of time spent in kafka controller Controller Event Manager Event Queue Time, measured in milliseconds, scoped by the Kafka auto-collect measurement tags.",
		},

		"kafka.controller.ControllerEventManager.EventQueueTimeMs.99thPercentile": &inputs.FieldInfo{
			DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS,
			Desc: "99th percentile of time spent in kafka controller Controller Event Manager Event Queue Time, measured in milliseconds, scoped by the Kafka auto-collect measurement tags.",
		},

		"kafka.controller.ControllerEventManager.EventQueueTimeMs.Max": &inputs.FieldInfo{
			DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS,
			Desc: "Maximum time spent in kafka controller Controller Event Manager Event Queue Time, measured in milliseconds, scoped by the Kafka auto-collect measurement tags.",
		},

		"kafka.controller.ControllerEventManager.EventQueueTimeMs.Mean": &inputs.FieldInfo{
			DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS,
			Desc: "Average time spent in kafka controller Controller Event Manager Event Queue Time, measured in milliseconds, scoped by the Kafka auto-collect measurement tags.",
		},

		"kafka.controller.ControllerEventManager.EventQueueTimeMs.Min": &inputs.FieldInfo{
			DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS,
			Desc: "Minimum time spent in kafka controller Controller Event Manager Event Queue Time, measured in milliseconds, scoped by the Kafka auto-collect measurement tags.",
		},

		"kafka.controller.ControllerEventManager.EventQueueTimeMs.StdDev": &inputs.FieldInfo{
			DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS,
			Desc: "Standard deviation of time spent in kafka controller Controller Event Manager Event Queue Time, measured in milliseconds, scoped by the Kafka auto-collect measurement tags.",
		},

		"kafka.controller.ControllerEventManager.EventQueueTimeMs.Count": &inputs.FieldInfo{
			DataType: inputs.Int, Type: inputs.Count, Unit: inputs.NCount,
			Desc: "Total number of kafka controller Controller Event Manager Event Queue Time, scoped by the Kafka auto-collect measurement tags.",
		},

		"kafka.controller.KafkaController.GlobalPartitionCount.Value": &inputs.FieldInfo{
			DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.NCount,
			Desc: "Current number of kafka controller Kafka Controller Global Partition Count, scoped by the Kafka auto-collect measurement tags.",
		},

		"kafka.controller.KafkaController.GlobalTopicCount.Value": &inputs.FieldInfo{
			DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.NCount,
			Desc: "Current number of kafka controller Kafka Controller Global Topic Count, scoped by the Kafka auto-collect measurement tags.",
		},

		"kafka.controller.KafkaController.OfflinePartitionsCount.Value": &inputs.FieldInfo{
			DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.NCount,
			Desc: "Current number of kafka controller Kafka Controller Offline Partitions Count, scoped by the Kafka auto-collect measurement tags.",
		},

		"kafka.controller.KafkaController.PreferredReplicaImbalanceCount.Value": &inputs.FieldInfo{
			DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.NCount,
			Desc: "Current number of kafka controller Kafka Controller Preferred Replica Imbalance Count, scoped by the Kafka auto-collect measurement tags.",
		},

		"kafka.controller.KafkaController.ReplicasIneligibleToDeleteCount.Value": &inputs.FieldInfo{
			DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.NCount,
			Desc: "Current number of kafka controller Kafka Controller Replicas Ineligible To Delete Count, scoped by the Kafka auto-collect measurement tags.",
		},

		"kafka.controller.KafkaController.ReplicasToDeleteCount.Value": &inputs.FieldInfo{
			DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.NCount,
			Desc: "Current number of kafka controller Kafka Controller Replicas To Delete Count, scoped by the Kafka auto-collect measurement tags.",
		},

		"kafka.controller.KafkaController.TopicsIneligibleToDeleteCount.Value": &inputs.FieldInfo{
			DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.NCount,
			Desc: "Current number of kafka controller Kafka Controller Topics Ineligible To Delete Count, scoped by the Kafka auto-collect measurement tags.",
		},

		"kafka.controller.KafkaController.TopicsToDeleteCount.Value": &inputs.FieldInfo{
			DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.NCount,
			Desc: "Current number of kafka controller Kafka Controller Topics To Delete Count, scoped by the Kafka auto-collect measurement tags.",
		},

		"kafka.controller.KafkaController.ActiveControllerCount.Value": &inputs.FieldInfo{
			DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.NCount,
			Desc: "Current number of kafka controller Kafka Controller Active Controller Count, scoped by the Kafka auto-collect measurement tags.",
		},

		"kafka.controller.KafkaController.ControllerState.Value": &inputs.FieldInfo{
			DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.NCount,
			Desc: "Current value of kafka controller Kafka Controller Controller State, scoped by the Kafka auto-collect measurement tags.",
		},

		"kafka.controller.ControllerEventManager.EventQueueSize.Value": &inputs.FieldInfo{
			DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.NCount,
			Desc: "Current queue depth for kafka controller Controller Event Manager Event Queue Size, scoped by the Kafka auto-collect measurement tags.",
		},

		"kafka.controller.ControllerChannelManager.TotalQueueSize.Value": &inputs.FieldInfo{
			DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.NCount,
			Desc: "Current queue depth for kafka controller Controller Channel Manager Total Queue Size, scoped by the Kafka auto-collect measurement tags.",
		},
	}

	return fields
}

//nolint:lll
func (m *kafkaMeasurement) getReplicaFields() map[string]interface{} {
	fields := map[string]interface{}{
		"kafka.server.ReplicaManager.FailedIsrUpdatesPerSec.Count": &inputs.FieldInfo{
			DataType: inputs.Int, Type: inputs.Count, Unit: inputs.NCount,
			Desc: "Total number of kafka server Replica Manager Failed Isr Updates Per Sec, scoped by the Kafka auto-collect measurement tags.",
		},

		"kafka.server.ReplicaManager.FailedIsrUpdatesPerSec.FifteenMinuteRate": &inputs.FieldInfo{
			DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount,
			Desc: "Fifteen-minute moving average rate for kafka server Replica Manager Failed Isr Updates Per Sec in events per second, scoped by the Kafka auto-collect measurement tags.",
		},

		"kafka.server.ReplicaManager.FailedIsrUpdatesPerSec.FiveMinuteRate": &inputs.FieldInfo{
			DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount,
			Desc: "Five-minute moving average rate for kafka server Replica Manager Failed Isr Updates Per Sec in events per second, scoped by the Kafka auto-collect measurement tags.",
		},

		"kafka.server.ReplicaManager.FailedIsrUpdatesPerSec.MeanRate": &inputs.FieldInfo{
			DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount,
			Desc: "Mean rate for kafka server Replica Manager Failed Isr Updates Per Sec in events per second, scoped by the Kafka auto-collect measurement tags.",
		},

		"kafka.server.ReplicaManager.FailedIsrUpdatesPerSec.OneMinuteRate": &inputs.FieldInfo{
			DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount,
			Desc: "One-minute moving average rate for kafka server Replica Manager Failed Isr Updates Per Sec in events per second, scoped by the Kafka auto-collect measurement tags.",
		},

		"kafka.server.ReplicaManager.IsrExpandsPerSec.Count": &inputs.FieldInfo{
			DataType: inputs.Int, Type: inputs.Count, Unit: inputs.NCount,
			Desc: "Total number of kafka server Replica Manager Isr Expands Per Sec, scoped by the Kafka auto-collect measurement tags.",
		},

		"kafka.server.ReplicaManager.IsrExpandsPerSec.FifteenMinuteRate": &inputs.FieldInfo{
			DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount,
			Desc: "Fifteen-minute moving average rate for kafka server Replica Manager Isr Expands Per Sec in events per second, scoped by the Kafka auto-collect measurement tags.",
		},

		"kafka.server.ReplicaManager.IsrExpandsPerSec.FiveMinuteRate": &inputs.FieldInfo{
			DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount,
			Desc: "Five-minute moving average rate for kafka server Replica Manager Isr Expands Per Sec in events per second, scoped by the Kafka auto-collect measurement tags.",
		},

		"kafka.server.ReplicaManager.IsrExpandsPerSec.MeanRate": &inputs.FieldInfo{
			DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount,
			Desc: "Mean rate for kafka server Replica Manager Isr Expands Per Sec in events per second, scoped by the Kafka auto-collect measurement tags.",
		},

		"kafka.server.ReplicaManager.IsrExpandsPerSec.OneMinuteRate": &inputs.FieldInfo{
			DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount,
			Desc: "One-minute moving average rate for kafka server Replica Manager Isr Expands Per Sec in events per second, scoped by the Kafka auto-collect measurement tags.",
		},

		"kafka.server.ReplicaManager.IsrShrinksPerSec.Count": &inputs.FieldInfo{
			DataType: inputs.Int, Type: inputs.Count, Unit: inputs.NCount,
			Desc: "Total number of kafka server Replica Manager Isr Shrinks Per Sec, scoped by the Kafka auto-collect measurement tags.",
		},

		"kafka.server.ReplicaManager.IsrShrinksPerSec.FifteenMinuteRate": &inputs.FieldInfo{
			DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount,
			Desc: "Fifteen-minute moving average rate for kafka server Replica Manager Isr Shrinks Per Sec in events per second, scoped by the Kafka auto-collect measurement tags.",
		},

		"kafka.server.ReplicaManager.IsrShrinksPerSec.FiveMinuteRate": &inputs.FieldInfo{
			DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount,
			Desc: "Five-minute moving average rate for kafka server Replica Manager Isr Shrinks Per Sec in events per second, scoped by the Kafka auto-collect measurement tags.",
		},

		"kafka.server.ReplicaManager.IsrShrinksPerSec.MeanRate": &inputs.FieldInfo{
			DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount,
			Desc: "Mean rate for kafka server Replica Manager Isr Shrinks Per Sec in events per second, scoped by the Kafka auto-collect measurement tags.",
		},

		"kafka.server.ReplicaManager.IsrShrinksPerSec.OneMinuteRate": &inputs.FieldInfo{
			DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount,
			Desc: "One-minute moving average rate for kafka server Replica Manager Isr Shrinks Per Sec in events per second, scoped by the Kafka auto-collect measurement tags.",
		},

		"kafka.server.ReplicaManager.AtMinIsrPartitionCount.Value": &inputs.FieldInfo{
			DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.NCount,
			Desc: "Current number of kafka server Replica Manager At Min Isr Partition Count, scoped by the Kafka auto-collect measurement tags.",
		},

		"kafka.server.ReplicaManager.LeaderCount.Value": &inputs.FieldInfo{
			DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.NCount,
			Desc: "Current number of kafka server Replica Manager Leader Count, scoped by the Kafka auto-collect measurement tags.",
		},

		"kafka.server.ReplicaManager.OfflineReplicaCount.Value": &inputs.FieldInfo{
			DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.NCount,
			Desc: "Current number of kafka server Replica Manager Offline Replica Count, scoped by the Kafka auto-collect measurement tags.",
		},

		"kafka.server.ReplicaManager.PartitionCount.Value": &inputs.FieldInfo{
			DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.NCount,
			Desc: "Current number of kafka server Replica Manager Partition Count, scoped by the Kafka auto-collect measurement tags.",
		},

		"kafka.server.ReplicaManager.ReassigningPartitions.Value": &inputs.FieldInfo{
			DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.NCount,
			Desc: "Current number of kafka server Replica Manager Reassigning Partitions, scoped by the Kafka auto-collect measurement tags.",
		},

		"kafka.server.ReplicaManager.UnderMinIsrPartitionCount.Value": &inputs.FieldInfo{
			DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.NCount,
			Desc: "Current number of kafka server Replica Manager Under Min Isr Partition Count, scoped by the Kafka auto-collect measurement tags.",
		},

		"kafka.server.ReplicaManager.UnderReplicatedPartitions.Value": &inputs.FieldInfo{
			DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.NCount,
			Desc: "Current number of kafka server Replica Manager Under Replicated Partitions, scoped by the Kafka auto-collect measurement tags.",
		},
	}
	return fields
}

//nolint:lll
func (m *kafkaMeasurement) getZookeeperFields() map[string]interface{} {
	fields := map[string]interface{}{
		"kafka.server.ZooKeeperClientMetrics.ZooKeeperRequestLatencyMs.Mean": &inputs.FieldInfo{
			DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS,
			Desc: "Average time spent in kafka server Zoo Keeper Client Metrics Zoo Keeper Request Latency, measured in milliseconds, scoped by the Kafka auto-collect measurement tags.",
		},

		"kafka.server.ZooKeeperClientMetrics.ZooKeeperRequestLatencyMs.98thPercentile": &inputs.FieldInfo{
			DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS,
			Desc: "98th percentile of time spent in kafka server Zoo Keeper Client Metrics Zoo Keeper Request Latency, measured in milliseconds, scoped by the Kafka auto-collect measurement tags.",
		},

		"kafka.server.ZooKeeperClientMetrics.ZooKeeperRequestLatencyMs.75thPercentile": &inputs.FieldInfo{
			DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS,
			Desc: "75th percentile of time spent in kafka server Zoo Keeper Client Metrics Zoo Keeper Request Latency, measured in milliseconds, scoped by the Kafka auto-collect measurement tags.",
		},

		"kafka.server.ZooKeeperClientMetrics.ZooKeeperRequestLatencyMs.95thPercentile": &inputs.FieldInfo{
			DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS,
			Desc: "95th percentile of time spent in kafka server Zoo Keeper Client Metrics Zoo Keeper Request Latency, measured in milliseconds, scoped by the Kafka auto-collect measurement tags.",
		},

		"kafka.server.ZooKeeperClientMetrics.ZooKeeperRequestLatencyMs.Min": &inputs.FieldInfo{
			DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS,
			Desc: "Minimum time spent in kafka server Zoo Keeper Client Metrics Zoo Keeper Request Latency, measured in milliseconds, scoped by the Kafka auto-collect measurement tags.",
		},

		"kafka.server.ZooKeeperClientMetrics.ZooKeeperRequestLatencyMs.999thPercentile": &inputs.FieldInfo{
			DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS,
			Desc: "99.9th percentile of time spent in kafka server Zoo Keeper Client Metrics Zoo Keeper Request Latency, measured in milliseconds, scoped by the Kafka auto-collect measurement tags.",
		},

		"kafka.server.ZooKeeperClientMetrics.ZooKeeperRequestLatencyMs.Max": &inputs.FieldInfo{
			DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS,
			Desc: "Maximum time spent in kafka server Zoo Keeper Client Metrics Zoo Keeper Request Latency, measured in milliseconds, scoped by the Kafka auto-collect measurement tags.",
		},

		"kafka.server.ZooKeeperClientMetrics.ZooKeeperRequestLatencyMs.StdDev": &inputs.FieldInfo{
			DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS,
			Desc: "Standard deviation of time spent in kafka server Zoo Keeper Client Metrics Zoo Keeper Request Latency, measured in milliseconds, scoped by the Kafka auto-collect measurement tags.",
		},

		"kafka.server.ZooKeeperClientMetrics.ZooKeeperRequestLatencyMs.99thPercentile": &inputs.FieldInfo{
			DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS,
			Desc: "99th percentile of time spent in kafka server Zoo Keeper Client Metrics Zoo Keeper Request Latency, measured in milliseconds, scoped by the Kafka auto-collect measurement tags.",
		},

		"kafka.server.ZooKeeperClientMetrics.ZooKeeperRequestLatencyMs.Count": &inputs.FieldInfo{
			DataType: inputs.Int, Type: inputs.Count, Unit: inputs.NCount,
			Desc: "Total number of kafka server Zoo Keeper Client Metrics Zoo Keeper Request Latency, scoped by the Kafka auto-collect measurement tags.",
		},

		"kafka.server.ZooKeeperClientMetrics.ZooKeeperRequestLatencyMs.50thPercentile": &inputs.FieldInfo{
			DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS,
			Desc: "50th percentile of time spent in kafka server Zoo Keeper Client Metrics Zoo Keeper Request Latency, measured in milliseconds, scoped by the Kafka auto-collect measurement tags.",
		},
	}
	return fields
}

//nolint:lll
func (m *kafkaMeasurement) getPurgatoryFields() map[string]interface{} {
	fields := map[string]interface{}{
		"kafka.server.DelayedOperationPurgatory.NumDelayedOperations.Value": &inputs.FieldInfo{
			DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.PercentDecimal,
			Desc: "Fraction from 0 to 1 for kafka server Delayed Operation Purgatory Num Delayed Operations, scoped by delayedOperation.",
		},

		"kafka.server.DelayedOperationPurgatory.PurgatorySize.Value": &inputs.FieldInfo{
			DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.PercentDecimal,
			Desc: "Fraction from 0 to 1 for kafka server Delayed Operation Purgatory Purgatory Size, scoped by delayedOperation.",
		},
	}
	m.addTaggedbyToFields(fields, TagGroupPurgatory)
	return fields
}

//nolint:lll
func (m *kafkaMeasurement) getRequestFields() map[string]interface{} {
	fields := map[string]interface{}{
		"kafka.network.RequestMetrics.LocalTimeMs.50thPercentile": &inputs.FieldInfo{
			DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS,
			Desc: "50th percentile of time spent in kafka network Request Metrics Local Time, measured in milliseconds, scoped by request.",
		},

		"kafka.network.RequestMetrics.LocalTimeMs.75thPercentile": &inputs.FieldInfo{
			DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS,
			Desc: "75th percentile of time spent in kafka network Request Metrics Local Time, measured in milliseconds, scoped by request.",
		},

		"kafka.network.RequestMetrics.LocalTimeMs.95thPercentile": &inputs.FieldInfo{
			DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS,
			Desc: "95th percentile of time spent in kafka network Request Metrics Local Time, measured in milliseconds, scoped by request.",
		},

		"kafka.network.RequestMetrics.LocalTimeMs.98thPercentile": &inputs.FieldInfo{
			DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS,
			Desc: "98th percentile of time spent in kafka network Request Metrics Local Time, measured in milliseconds, scoped by request.",
		},

		"kafka.network.RequestMetrics.LocalTimeMs.999thPercentile": &inputs.FieldInfo{
			DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS,
			Desc: "99.9th percentile of time spent in kafka network Request Metrics Local Time, measured in milliseconds, scoped by request.",
		},

		"kafka.network.RequestMetrics.LocalTimeMs.99thPercentile": &inputs.FieldInfo{
			DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS,
			Desc: "99th percentile of time spent in kafka network Request Metrics Local Time, measured in milliseconds, scoped by request.",
		},

		"kafka.network.RequestMetrics.LocalTimeMs.Max": &inputs.FieldInfo{
			DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS,
			Desc: "Maximum time spent in kafka network Request Metrics Local Time, measured in milliseconds, scoped by request.",
		},

		"kafka.network.RequestMetrics.LocalTimeMs.Mean": &inputs.FieldInfo{
			DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS,
			Desc: "Average time spent in kafka network Request Metrics Local Time, measured in milliseconds, scoped by request.",
		},

		"kafka.network.RequestMetrics.LocalTimeMs.Min": &inputs.FieldInfo{
			DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS,
			Desc: "Minimum time spent in kafka network Request Metrics Local Time, measured in milliseconds, scoped by request.",
		},

		"kafka.network.RequestMetrics.LocalTimeMs.StdDev": &inputs.FieldInfo{
			DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS,
			Desc: "Standard deviation of time spent in kafka network Request Metrics Local Time, measured in milliseconds, scoped by request.",
		},

		"kafka.network.RequestMetrics.LocalTimeMs.Count": &inputs.FieldInfo{
			DataType: inputs.Int, Type: inputs.Count, Unit: inputs.NCount,
			Desc: "Total number of kafka network Request Metrics Local Time, scoped by request.",
		},

		"kafka.network.RequestMetrics.RemoteTimeMs.50thPercentile": &inputs.FieldInfo{
			DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS,
			Desc: "50th percentile of time spent in kafka network Request Metrics Remote Time, measured in milliseconds, scoped by request.",
		},

		"kafka.network.RequestMetrics.RemoteTimeMs.75thPercentile": &inputs.FieldInfo{
			DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS,
			Desc: "75th percentile of time spent in kafka network Request Metrics Remote Time, measured in milliseconds, scoped by request.",
		},

		"kafka.network.RequestMetrics.RemoteTimeMs.95thPercentile": &inputs.FieldInfo{
			DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS,
			Desc: "95th percentile of time spent in kafka network Request Metrics Remote Time, measured in milliseconds, scoped by request.",
		},

		"kafka.network.RequestMetrics.RemoteTimeMs.98thPercentile": &inputs.FieldInfo{
			DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS,
			Desc: "98th percentile of time spent in kafka network Request Metrics Remote Time, measured in milliseconds, scoped by request.",
		},

		"kafka.network.RequestMetrics.RemoteTimeMs.999thPercentile": &inputs.FieldInfo{
			DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS,
			Desc: "99.9th percentile of time spent in kafka network Request Metrics Remote Time, measured in milliseconds, scoped by request.",
		},

		"kafka.network.RequestMetrics.RemoteTimeMs.99thPercentile": &inputs.FieldInfo{
			DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS,
			Desc: "99th percentile of time spent in kafka network Request Metrics Remote Time, measured in milliseconds, scoped by request.",
		},

		"kafka.network.RequestMetrics.RemoteTimeMs.Max": &inputs.FieldInfo{
			DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS,
			Desc: "Maximum time spent in kafka network Request Metrics Remote Time, measured in milliseconds, scoped by request.",
		},

		"kafka.network.RequestMetrics.RemoteTimeMs.Mean": &inputs.FieldInfo{
			DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS,
			Desc: "Average time spent in kafka network Request Metrics Remote Time, measured in milliseconds, scoped by request.",
		},

		"kafka.network.RequestMetrics.RemoteTimeMs.Min": &inputs.FieldInfo{
			DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS,
			Desc: "Minimum time spent in kafka network Request Metrics Remote Time, measured in milliseconds, scoped by request.",
		},

		"kafka.network.RequestMetrics.RemoteTimeMs.StdDev": &inputs.FieldInfo{
			DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS,
			Desc: "Standard deviation of time spent in kafka network Request Metrics Remote Time, measured in milliseconds, scoped by request.",
		},

		"kafka.network.RequestMetrics.RemoteTimeMs.Count": &inputs.FieldInfo{
			DataType: inputs.Int, Type: inputs.Count, Unit: inputs.NCount,
			Desc: "Total number of kafka network Request Metrics Remote Time, scoped by request.",
		},

		"kafka.network.RequestMetrics.RequestBytes.50thPercentile": &inputs.FieldInfo{
			DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.SizeByte,
			Desc: "50th percentile size of kafka network Request Metrics Request Bytes in bytes, scoped by request.",
		},

		"kafka.network.RequestMetrics.RequestBytes.75thPercentile": &inputs.FieldInfo{
			DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.SizeByte,
			Desc: "75th percentile size of kafka network Request Metrics Request Bytes in bytes, scoped by request.",
		},

		"kafka.network.RequestMetrics.RequestBytes.95thPercentile": &inputs.FieldInfo{
			DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.SizeByte,
			Desc: "95th percentile size of kafka network Request Metrics Request Bytes in bytes, scoped by request.",
		},

		"kafka.network.RequestMetrics.RequestBytes.98thPercentile": &inputs.FieldInfo{
			DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.SizeByte,
			Desc: "98th percentile size of kafka network Request Metrics Request Bytes in bytes, scoped by request.",
		},

		"kafka.network.RequestMetrics.RequestBytes.999thPercentile": &inputs.FieldInfo{
			DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.SizeByte,
			Desc: "99.9th percentile size of kafka network Request Metrics Request Bytes in bytes, scoped by request.",
		},

		"kafka.network.RequestMetrics.RequestBytes.99thPercentile": &inputs.FieldInfo{
			DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.SizeByte,
			Desc: "99th percentile size of kafka network Request Metrics Request Bytes in bytes, scoped by request.",
		},

		"kafka.network.RequestMetrics.RequestBytes.Max": &inputs.FieldInfo{
			DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.SizeByte,
			Desc: "Maximum size of kafka network Request Metrics Request Bytes in bytes, scoped by request.",
		},

		"kafka.network.RequestMetrics.RequestBytes.Mean": &inputs.FieldInfo{
			DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.SizeByte,
			Desc: "Average size of kafka network Request Metrics Request Bytes in bytes, scoped by request.",
		},

		"kafka.network.RequestMetrics.RequestBytes.Min": &inputs.FieldInfo{
			DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.SizeByte,
			Desc: "Minimum size of kafka network Request Metrics Request Bytes in bytes, scoped by request.",
		},

		"kafka.network.RequestMetrics.RequestBytes.StdDev": &inputs.FieldInfo{
			DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.SizeByte,
			Desc: "Current size of kafka network Request Metrics Request Bytes in bytes, scoped by request.",
		},

		"kafka.network.RequestMetrics.RequestBytes.Count": &inputs.FieldInfo{
			DataType: inputs.Int, Type: inputs.Count, Unit: inputs.NCount,
			Desc: "Total number of kafka network Request Metrics Request Bytes, scoped by request.",
		},

		"kafka.network.RequestMetrics.RequestQueueTimeMs.50thPercentile": &inputs.FieldInfo{
			DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS,
			Desc: "50th percentile of time spent in kafka network Request Metrics Request Queue Time, measured in milliseconds, scoped by request.",
		},

		"kafka.network.RequestMetrics.RequestQueueTimeMs.75thPercentile": &inputs.FieldInfo{
			DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS,
			Desc: "75th percentile of time spent in kafka network Request Metrics Request Queue Time, measured in milliseconds, scoped by request.",
		},

		"kafka.network.RequestMetrics.RequestQueueTimeMs.95thPercentile": &inputs.FieldInfo{
			DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS,
			Desc: "95th percentile of time spent in kafka network Request Metrics Request Queue Time, measured in milliseconds, scoped by request.",
		},

		"kafka.network.RequestMetrics.RequestQueueTimeMs.98thPercentile": &inputs.FieldInfo{
			DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS,
			Desc: "98th percentile of time spent in kafka network Request Metrics Request Queue Time, measured in milliseconds, scoped by request.",
		},

		"kafka.network.RequestMetrics.RequestQueueTimeMs.999thPercentile": &inputs.FieldInfo{
			DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS,
			Desc: "99.9th percentile of time spent in kafka network Request Metrics Request Queue Time, measured in milliseconds, scoped by request.",
		},

		"kafka.network.RequestMetrics.RequestQueueTimeMs.99thPercentile": &inputs.FieldInfo{
			DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS,
			Desc: "99th percentile of time spent in kafka network Request Metrics Request Queue Time, measured in milliseconds, scoped by request.",
		},

		"kafka.network.RequestMetrics.RequestQueueTimeMs.Max": &inputs.FieldInfo{
			DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS,
			Desc: "Maximum time spent in kafka network Request Metrics Request Queue Time, measured in milliseconds, scoped by request.",
		},

		"kafka.network.RequestMetrics.RequestQueueTimeMs.Mean": &inputs.FieldInfo{
			DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS,
			Desc: "Average time spent in kafka network Request Metrics Request Queue Time, measured in milliseconds, scoped by request.",
		},

		"kafka.network.RequestMetrics.RequestQueueTimeMs.Min": &inputs.FieldInfo{
			DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS,
			Desc: "Minimum time spent in kafka network Request Metrics Request Queue Time, measured in milliseconds, scoped by request.",
		},

		"kafka.network.RequestMetrics.RequestQueueTimeMs.StdDev": &inputs.FieldInfo{
			DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS,
			Desc: "Standard deviation of time spent in kafka network Request Metrics Request Queue Time, measured in milliseconds, scoped by request.",
		},

		"kafka.network.RequestMetrics.RequestQueueTimeMs.Count": &inputs.FieldInfo{
			DataType: inputs.Int, Type: inputs.Count, Unit: inputs.NCount,
			Desc: "Total number of kafka network Request Metrics Request Queue Time, scoped by request.",
		},

		"kafka.network.RequestMetrics.ResponseQueueTimeMs.50thPercentile": &inputs.FieldInfo{
			DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS,
			Desc: "50th percentile of time spent in kafka network Request Metrics Response Queue Time, measured in milliseconds, scoped by request.",
		},

		"kafka.network.RequestMetrics.ResponseQueueTimeMs.75thPercentile": &inputs.FieldInfo{
			DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS,
			Desc: "75th percentile of time spent in kafka network Request Metrics Response Queue Time, measured in milliseconds, scoped by request.",
		},

		"kafka.network.RequestMetrics.ResponseQueueTimeMs.95thPercentile": &inputs.FieldInfo{
			DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS,
			Desc: "95th percentile of time spent in kafka network Request Metrics Response Queue Time, measured in milliseconds, scoped by request.",
		},

		"kafka.network.RequestMetrics.ResponseQueueTimeMs.98thPercentile": &inputs.FieldInfo{
			DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS,
			Desc: "98th percentile of time spent in kafka network Request Metrics Response Queue Time, measured in milliseconds, scoped by request.",
		},

		"kafka.network.RequestMetrics.ResponseQueueTimeMs.999thPercentile": &inputs.FieldInfo{
			DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS,
			Desc: "99.9th percentile of time spent in kafka network Request Metrics Response Queue Time, measured in milliseconds, scoped by request.",
		},

		"kafka.network.RequestMetrics.ResponseQueueTimeMs.99thPercentile": &inputs.FieldInfo{
			DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS,
			Desc: "99th percentile of time spent in kafka network Request Metrics Response Queue Time, measured in milliseconds, scoped by request.",
		},

		"kafka.network.RequestMetrics.ResponseQueueTimeMs.Max": &inputs.FieldInfo{
			DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS,
			Desc: "Maximum time spent in kafka network Request Metrics Response Queue Time, measured in milliseconds, scoped by request.",
		},

		"kafka.network.RequestMetrics.ResponseQueueTimeMs.Mean": &inputs.FieldInfo{
			DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS,
			Desc: "Average time spent in kafka network Request Metrics Response Queue Time, measured in milliseconds, scoped by request.",
		},

		"kafka.network.RequestMetrics.ResponseQueueTimeMs.Min": &inputs.FieldInfo{
			DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS,
			Desc: "Minimum time spent in kafka network Request Metrics Response Queue Time, measured in milliseconds, scoped by request.",
		},

		"kafka.network.RequestMetrics.ResponseQueueTimeMs.StdDev": &inputs.FieldInfo{
			DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS,
			Desc: "Standard deviation of time spent in kafka network Request Metrics Response Queue Time, measured in milliseconds, scoped by request.",
		},

		"kafka.network.RequestMetrics.ResponseQueueTimeMs.Count": &inputs.FieldInfo{
			DataType: inputs.Int, Type: inputs.Count, Unit: inputs.NCount,
			Desc: "Total number of kafka network Request Metrics Response Queue Time, scoped by request.",
		},

		"kafka.network.RequestMetrics.ResponseSendTimeMs.50thPercentile": &inputs.FieldInfo{
			DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS,
			Desc: "50th percentile of time spent in kafka network Request Metrics Response Send Time, measured in milliseconds, scoped by request.",
		},

		"kafka.network.RequestMetrics.ResponseSendTimeMs.75thPercentile": &inputs.FieldInfo{
			DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS,
			Desc: "75th percentile of time spent in kafka network Request Metrics Response Send Time, measured in milliseconds, scoped by request.",
		},

		"kafka.network.RequestMetrics.ResponseSendTimeMs.95thPercentile": &inputs.FieldInfo{
			DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS,
			Desc: "95th percentile of time spent in kafka network Request Metrics Response Send Time, measured in milliseconds, scoped by request.",
		},

		"kafka.network.RequestMetrics.ResponseSendTimeMs.98thPercentile": &inputs.FieldInfo{
			DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS,
			Desc: "98th percentile of time spent in kafka network Request Metrics Response Send Time, measured in milliseconds, scoped by request.",
		},

		"kafka.network.RequestMetrics.ResponseSendTimeMs.999thPercentile": &inputs.FieldInfo{
			DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS,
			Desc: "99.9th percentile of time spent in kafka network Request Metrics Response Send Time, measured in milliseconds, scoped by request.",
		},

		"kafka.network.RequestMetrics.ResponseSendTimeMs.99thPercentile": &inputs.FieldInfo{
			DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS,
			Desc: "99th percentile of time spent in kafka network Request Metrics Response Send Time, measured in milliseconds, scoped by request.",
		},

		"kafka.network.RequestMetrics.ResponseSendTimeMs.Max": &inputs.FieldInfo{
			DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS,
			Desc: "Maximum time spent in kafka network Request Metrics Response Send Time, measured in milliseconds, scoped by request.",
		},

		"kafka.network.RequestMetrics.ResponseSendTimeMs.Mean": &inputs.FieldInfo{
			DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS,
			Desc: "Average time spent in kafka network Request Metrics Response Send Time, measured in milliseconds, scoped by request.",
		},

		"kafka.network.RequestMetrics.ResponseSendTimeMs.Min": &inputs.FieldInfo{
			DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS,
			Desc: "Minimum time spent in kafka network Request Metrics Response Send Time, measured in milliseconds, scoped by request.",
		},

		"kafka.network.RequestMetrics.ResponseSendTimeMs.StdDev": &inputs.FieldInfo{
			DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS,
			Desc: "Standard deviation of time spent in kafka network Request Metrics Response Send Time, measured in milliseconds, scoped by request.",
		},

		"kafka.network.RequestMetrics.ResponseSendTimeMs.Count": &inputs.FieldInfo{
			DataType: inputs.Int, Type: inputs.Count, Unit: inputs.NCount,
			Desc: "Total number of kafka network Request Metrics Response Send Time, scoped by request.",
		},

		"kafka.network.RequestMetrics.ThrottleTimeMs.50thPercentile": &inputs.FieldInfo{
			DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS,
			Desc: "50th percentile of time spent in kafka network Request Metrics Throttle Time, measured in milliseconds, scoped by request.",
		},

		"kafka.network.RequestMetrics.ThrottleTimeMs.75thPercentile": &inputs.FieldInfo{
			DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS,
			Desc: "75th percentile of time spent in kafka network Request Metrics Throttle Time, measured in milliseconds, scoped by request.",
		},

		"kafka.network.RequestMetrics.ThrottleTimeMs.95thPercentile": &inputs.FieldInfo{
			DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS,
			Desc: "95th percentile of time spent in kafka network Request Metrics Throttle Time, measured in milliseconds, scoped by request.",
		},

		"kafka.network.RequestMetrics.ThrottleTimeMs.98thPercentile": &inputs.FieldInfo{
			DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS,
			Desc: "98th percentile of time spent in kafka network Request Metrics Throttle Time, measured in milliseconds, scoped by request.",
		},

		"kafka.network.RequestMetrics.ThrottleTimeMs.999thPercentile": &inputs.FieldInfo{
			DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS,
			Desc: "99.9th percentile of time spent in kafka network Request Metrics Throttle Time, measured in milliseconds, scoped by request.",
		},

		"kafka.network.RequestMetrics.ThrottleTimeMs.99thPercentile": &inputs.FieldInfo{
			DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS,
			Desc: "99th percentile of time spent in kafka network Request Metrics Throttle Time, measured in milliseconds, scoped by request.",
		},

		"kafka.network.RequestMetrics.ThrottleTimeMs.Max": &inputs.FieldInfo{
			DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS,
			Desc: "Maximum time spent in kafka network Request Metrics Throttle Time, measured in milliseconds, scoped by request.",
		},

		"kafka.network.RequestMetrics.ThrottleTimeMs.Mean": &inputs.FieldInfo{
			DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS,
			Desc: "Average time spent in kafka network Request Metrics Throttle Time, measured in milliseconds, scoped by request.",
		},

		"kafka.network.RequestMetrics.ThrottleTimeMs.Min": &inputs.FieldInfo{
			DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS,
			Desc: "Minimum time spent in kafka network Request Metrics Throttle Time, measured in milliseconds, scoped by request.",
		},

		"kafka.network.RequestMetrics.ThrottleTimeMs.StdDev": &inputs.FieldInfo{
			DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS,
			Desc: "Standard deviation of time spent in kafka network Request Metrics Throttle Time, measured in milliseconds, scoped by request.",
		},

		"kafka.network.RequestMetrics.ThrottleTimeMs.Count": &inputs.FieldInfo{
			DataType: inputs.Int, Type: inputs.Count, Unit: inputs.NCount,
			Desc: "Total number of kafka network Request Metrics Throttle Time, scoped by request.",
		},

		"kafka.network.RequestMetrics.TotalTimeMs.50thPercentile": &inputs.FieldInfo{
			DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS,
			Desc: "50th percentile of time spent in kafka network Request Metrics Total Time, measured in milliseconds, scoped by request.",
		},

		"kafka.network.RequestMetrics.TotalTimeMs.75thPercentile": &inputs.FieldInfo{
			DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS,
			Desc: "75th percentile of time spent in kafka network Request Metrics Total Time, measured in milliseconds, scoped by request.",
		},

		"kafka.network.RequestMetrics.TotalTimeMs.95thPercentile": &inputs.FieldInfo{
			DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS,
			Desc: "95th percentile of time spent in kafka network Request Metrics Total Time, measured in milliseconds, scoped by request.",
		},

		"kafka.network.RequestMetrics.TotalTimeMs.98thPercentile": &inputs.FieldInfo{
			DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS,
			Desc: "98th percentile of time spent in kafka network Request Metrics Total Time, measured in milliseconds, scoped by request.",
		},

		"kafka.network.RequestMetrics.TotalTimeMs.999thPercentile": &inputs.FieldInfo{
			DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS,
			Desc: "99.9th percentile of time spent in kafka network Request Metrics Total Time, measured in milliseconds, scoped by request.",
		},

		"kafka.network.RequestMetrics.TotalTimeMs.99thPercentile": &inputs.FieldInfo{
			DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS,
			Desc: "99th percentile of time spent in kafka network Request Metrics Total Time, measured in milliseconds, scoped by request.",
		},

		"kafka.network.RequestMetrics.TotalTimeMs.Max": &inputs.FieldInfo{
			DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS,
			Desc: "Maximum time spent in kafka network Request Metrics Total Time, measured in milliseconds, scoped by request.",
		},

		"kafka.network.RequestMetrics.TotalTimeMs.Mean": &inputs.FieldInfo{
			DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS,
			Desc: "Average time spent in kafka network Request Metrics Total Time, measured in milliseconds, scoped by request.",
		},

		"kafka.network.RequestMetrics.TotalTimeMs.Min": &inputs.FieldInfo{
			DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS,
			Desc: "Minimum time spent in kafka network Request Metrics Total Time, measured in milliseconds, scoped by request.",
		},

		"kafka.network.RequestMetrics.TotalTimeMs.StdDev": &inputs.FieldInfo{
			DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS,
			Desc: "Standard deviation of time spent in kafka network Request Metrics Total Time, measured in milliseconds, scoped by request.",
		},

		"kafka.network.RequestMetrics.TotalTimeMs.Count": &inputs.FieldInfo{
			DataType: inputs.Int, Type: inputs.Count, Unit: inputs.NCount,
			Desc: "Total number of kafka network Request Metrics Total Time, scoped by request.",
		},
	}
	m.addTaggedbyToFields(fields, TagGroupRequest)
	return fields
}

//nolint:lll
func (m *kafkaMeasurement) getRequestHandlerFields() map[string]interface{} {
	fields := map[string]interface{}{
		"kafka.server.KafkaRequestHandlerPool.RequestHandlerAvgIdlePercent.FifteenMinuteRate": &inputs.FieldInfo{
			DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.PercentDecimal,
			Desc: "Fraction from 0 to 1 for kafka server Kafka Request Handler Pool Request Handler Avg Idle Percent, scoped by the Kafka auto-collect measurement tags.",
		},

		"kafka.server.KafkaRequestHandlerPool.RequestHandlerAvgIdlePercent.OneMinuteRate": &inputs.FieldInfo{
			DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.PercentDecimal,
			Desc: "Fraction from 0 to 1 for kafka server Kafka Request Handler Pool Request Handler Avg Idle Percent, scoped by the Kafka auto-collect measurement tags.",
		},

		"kafka.server.KafkaRequestHandlerPool.RequestHandlerAvgIdlePercent.Count": &inputs.FieldInfo{
			DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.PercentDecimal,
			Desc: "Fraction from 0 to 1 for kafka server Kafka Request Handler Pool Request Handler Avg Idle Percent, scoped by the Kafka auto-collect measurement tags.",
		},

		"kafka.server.KafkaRequestHandlerPool.RequestHandlerAvgIdlePercent.FiveMinuteRate": &inputs.FieldInfo{
			DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.PercentDecimal,
			Desc: "Fraction from 0 to 1 for kafka server Kafka Request Handler Pool Request Handler Avg Idle Percent, scoped by the Kafka auto-collect measurement tags.",
		},

		"kafka.server.KafkaRequestHandlerPool.RequestHandlerAvgIdlePercent.MeanRate": &inputs.FieldInfo{
			DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.PercentDecimal,
			Desc: "Fraction from 0 to 1 for kafka server Kafka Request Handler Pool Request Handler Avg Idle Percent, scoped by the Kafka auto-collect measurement tags.",
		},
	}
	return fields
}

//nolint:lll
func (m *kafkaMeasurement) getNetworkFields() map[string]interface{} {
	fields := map[string]interface{}{
		"kafka.network.SocketServer.NetworkProcessorAvgIdlePercent.Value": &inputs.FieldInfo{
			DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.PercentDecimal,
			Desc: "Fraction from 0 to 1 for kafka network Socket Server Network Processor Avg Idle Percent, scoped by the Kafka auto-collect measurement tags.",
		},

		"kafka.network.SocketServer.ExpiredConnectionsKilledCount.Value": &inputs.FieldInfo{
			DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.NCount,
			Desc: "Current number of kafka network Socket Server Expired Connections Killed Count, scoped by the Kafka auto-collect measurement tags.",
		},

		"kafka.network.SocketServer.kafka.network.SocketServer.ControlPlaneExpiredConnectionsKilledCount.Value": &inputs.FieldInfo{
			DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.NCount,
			Desc: "Current number of kafka network Socket Server kafka network Socket Server Control Plane Expired Connections Killed Count, scoped by the Kafka auto-collect measurement tags.",
		},

		"kafka.network.SocketServer.MemoryPoolAvailable.Value": &inputs.FieldInfo{
			DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.SizeByte,
			Desc: "Current size of kafka network Socket Server Memory Pool Available in bytes, scoped by the Kafka auto-collect measurement tags.",
		},

		"kafka.network.SocketServer.MemoryPoolUsed.Value": &inputs.FieldInfo{
			DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.SizeByte,
			Desc: "Current size of kafka network Socket Server Memory Pool Used in bytes, scoped by the Kafka auto-collect measurement tags.",
		},
	}
	return fields
}

//nolint:lll
func (m *kafkaMeasurement) getTopicsFields() map[string]interface{} {
	fields := map[string]interface{}{
		"kafka.server.BrokerTopicMetrics.BytesInPerSec.Count": &inputs.FieldInfo{
			DataType: inputs.Int, Type: inputs.Count, Unit: inputs.SizeByte,
			Desc: "Total bytes recorded for kafka server Broker Topic Metrics Bytes In Per Sec, scoped by topic.",
		},

		"kafka.server.BrokerTopicMetrics.BytesInPerSec.FifteenMinuteRate": &inputs.FieldInfo{
			DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.BytesPerSec,
			Desc: "Fifteen-minute moving average rate for kafka server Broker Topic Metrics Bytes In Per Sec in bytes per second, scoped by topic.",
		},

		"kafka.server.BrokerTopicMetrics.BytesInPerSec.FiveMinuteRate": &inputs.FieldInfo{
			DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.BytesPerSec,
			Desc: "Five-minute moving average rate for kafka server Broker Topic Metrics Bytes In Per Sec in bytes per second, scoped by topic.",
		},

		"kafka.server.BrokerTopicMetrics.BytesInPerSec.MeanRate": &inputs.FieldInfo{
			DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.BytesPerSec,
			Desc: "Mean rate for kafka server Broker Topic Metrics Bytes In Per Sec in bytes per second, scoped by topic.",
		},

		"kafka.server.BrokerTopicMetrics.BytesInPerSec.OneMinuteRate": &inputs.FieldInfo{
			DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.BytesPerSec,
			Desc: "One-minute moving average rate for kafka server Broker Topic Metrics Bytes In Per Sec in bytes per second, scoped by topic.",
		},

		"kafka.server.BrokerTopicMetrics.BytesOutPerSec.Count": &inputs.FieldInfo{
			DataType: inputs.Int, Type: inputs.Count, Unit: inputs.SizeByte,
			Desc: "Total bytes recorded for kafka server Broker Topic Metrics Bytes Out Per Sec, scoped by topic.",
		},

		"kafka.server.BrokerTopicMetrics.BytesOutPerSec.FifteenMinuteRate": &inputs.FieldInfo{
			DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.BytesPerSec,
			Desc: "Fifteen-minute moving average rate for kafka server Broker Topic Metrics Bytes Out Per Sec in bytes per second, scoped by topic.",
		},

		"kafka.server.BrokerTopicMetrics.BytesOutPerSec.FiveMinuteRate": &inputs.FieldInfo{
			DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.BytesPerSec,
			Desc: "Five-minute moving average rate for kafka server Broker Topic Metrics Bytes Out Per Sec in bytes per second, scoped by topic.",
		},

		"kafka.server.BrokerTopicMetrics.BytesOutPerSec.MeanRate": &inputs.FieldInfo{
			DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.BytesPerSec,
			Desc: "Mean rate for kafka server Broker Topic Metrics Bytes Out Per Sec in bytes per second, scoped by topic.",
		},

		"kafka.server.BrokerTopicMetrics.BytesOutPerSec.OneMinuteRate": &inputs.FieldInfo{
			DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.BytesPerSec,
			Desc: "One-minute moving average rate for kafka server Broker Topic Metrics Bytes Out Per Sec in bytes per second, scoped by topic.",
		},

		"kafka.server.BrokerTopicMetrics.BytesRejectedPerSec.Count": &inputs.FieldInfo{
			DataType: inputs.Int, Type: inputs.Count, Unit: inputs.SizeByte,
			Desc: "Total bytes recorded for kafka server Broker Topic Metrics Bytes Rejected Per Sec, scoped by the Kafka auto-collect measurement tags.",
		},

		"kafka.server.BrokerTopicMetrics.BytesRejectedPerSec.FifteenMinuteRate": &inputs.FieldInfo{
			DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.BytesPerSec,
			Desc: "Fifteen-minute moving average rate for kafka server Broker Topic Metrics Bytes Rejected Per Sec in bytes per second, scoped by the Kafka auto-collect measurement tags.",
		},

		"kafka.server.BrokerTopicMetrics.BytesRejectedPerSec.FiveMinuteRate": &inputs.FieldInfo{
			DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.BytesPerSec,
			Desc: "Five-minute moving average rate for kafka server Broker Topic Metrics Bytes Rejected Per Sec in bytes per second, scoped by the Kafka auto-collect measurement tags.",
		},

		"kafka.server.BrokerTopicMetrics.BytesRejectedPerSec.MeanRate": &inputs.FieldInfo{
			DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.BytesPerSec,
			Desc: "Mean rate for kafka server Broker Topic Metrics Bytes Rejected Per Sec in bytes per second, scoped by the Kafka auto-collect measurement tags.",
		},

		"kafka.server.BrokerTopicMetrics.BytesRejectedPerSec.OneMinuteRate": &inputs.FieldInfo{
			DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.BytesPerSec,
			Desc: "One-minute moving average rate for kafka server Broker Topic Metrics Bytes Rejected Per Sec in bytes per second, scoped by the Kafka auto-collect measurement tags.",
		},

		"kafka.server.BrokerTopicMetrics.FailedFetchRequestsPerSec.Count": &inputs.FieldInfo{
			DataType: inputs.Int, Type: inputs.Count, Unit: inputs.NCount,
			Desc: "Total number of kafka server Broker Topic Metrics Failed Fetch Requests Per Sec, scoped by the Kafka auto-collect measurement tags.",
		},

		"kafka.server.BrokerTopicMetrics.FailedFetchRequestsPerSec.FifteenMinuteRate": &inputs.FieldInfo{
			DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.RequestsPerSec,
			Desc: "Fifteen-minute moving average rate for kafka server Broker Topic Metrics Failed Fetch Requests Per Sec in requests per second, scoped by the Kafka auto-collect measurement tags.",
		},

		"kafka.server.BrokerTopicMetrics.FailedFetchRequestsPerSec.FiveMinuteRate": &inputs.FieldInfo{
			DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.RequestsPerSec,
			Desc: "Five-minute moving average rate for kafka server Broker Topic Metrics Failed Fetch Requests Per Sec in requests per second, scoped by the Kafka auto-collect measurement tags.",
		},

		"kafka.server.BrokerTopicMetrics.FailedFetchRequestsPerSec.MeanRate": &inputs.FieldInfo{
			DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.RequestsPerSec,
			Desc: "Mean rate for kafka server Broker Topic Metrics Failed Fetch Requests Per Sec in requests per second, scoped by the Kafka auto-collect measurement tags.",
		},

		"kafka.server.BrokerTopicMetrics.FailedFetchRequestsPerSec.OneMinuteRate": &inputs.FieldInfo{
			DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.RequestsPerSec,
			Desc: "One-minute moving average rate for kafka server Broker Topic Metrics Failed Fetch Requests Per Sec in requests per second, scoped by the Kafka auto-collect measurement tags.",
		},

		"kafka.server.BrokerTopicMetrics.FailedProduceRequestsPerSec.Count": &inputs.FieldInfo{
			DataType: inputs.Int, Type: inputs.Count, Unit: inputs.NCount,
			Desc: "Total number of kafka server Broker Topic Metrics Failed Produce Requests Per Sec, scoped by the Kafka auto-collect measurement tags.",
		},

		"kafka.server.BrokerTopicMetrics.FailedProduceRequestsPerSec.FifteenMinuteRate": &inputs.FieldInfo{
			DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.RequestsPerSec,
			Desc: "Fifteen-minute moving average rate for kafka server Broker Topic Metrics Failed Produce Requests Per Sec in requests per second, scoped by the Kafka auto-collect measurement tags.",
		},

		"kafka.server.BrokerTopicMetrics.FailedProduceRequestsPerSec.FiveMinuteRate": &inputs.FieldInfo{
			DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.RequestsPerSec,
			Desc: "Five-minute moving average rate for kafka server Broker Topic Metrics Failed Produce Requests Per Sec in requests per second, scoped by the Kafka auto-collect measurement tags.",
		},

		"kafka.server.BrokerTopicMetrics.FailedProduceRequestsPerSec.MeanRate": &inputs.FieldInfo{
			DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.RequestsPerSec,
			Desc: "Mean rate for kafka server Broker Topic Metrics Failed Produce Requests Per Sec in requests per second, scoped by the Kafka auto-collect measurement tags.",
		},

		"kafka.server.BrokerTopicMetrics.FailedProduceRequestsPerSec.OneMinuteRate": &inputs.FieldInfo{
			DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.RequestsPerSec,
			Desc: "One-minute moving average rate for kafka server Broker Topic Metrics Failed Produce Requests Per Sec in requests per second, scoped by the Kafka auto-collect measurement tags.",
		},

		"kafka.server.BrokerTopicMetrics.FetchMessageConversionsPerSec.Count": &inputs.FieldInfo{
			DataType: inputs.Int, Type: inputs.Count, Unit: inputs.NCount,
			Desc: "Total number of kafka server Broker Topic Metrics Fetch Message Conversions Per Sec, scoped by the Kafka auto-collect measurement tags.",
		},

		"kafka.server.BrokerTopicMetrics.FetchMessageConversionsPerSec.FifteenMinuteRate": &inputs.FieldInfo{
			DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount,
			Desc: "Fifteen-minute moving average rate for kafka server Broker Topic Metrics Fetch Message Conversions Per Sec in events per second, scoped by the Kafka auto-collect measurement tags.",
		},

		"kafka.server.BrokerTopicMetrics.FetchMessageConversionsPerSec.FiveMinuteRate": &inputs.FieldInfo{
			DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount,
			Desc: "Five-minute moving average rate for kafka server Broker Topic Metrics Fetch Message Conversions Per Sec in events per second, scoped by the Kafka auto-collect measurement tags.",
		},

		"kafka.server.BrokerTopicMetrics.FetchMessageConversionsPerSec.MeanRate": &inputs.FieldInfo{
			DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount,
			Desc: "Mean rate for kafka server Broker Topic Metrics Fetch Message Conversions Per Sec in events per second, scoped by the Kafka auto-collect measurement tags.",
		},

		"kafka.server.BrokerTopicMetrics.FetchMessageConversionsPerSec.OneMinuteRate": &inputs.FieldInfo{
			DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount,
			Desc: "One-minute moving average rate for kafka server Broker Topic Metrics Fetch Message Conversions Per Sec in events per second, scoped by the Kafka auto-collect measurement tags.",
		},

		"kafka.server.BrokerTopicMetrics.InvalidMagicNumberRecordsPerSec.Count": &inputs.FieldInfo{
			DataType: inputs.Int, Type: inputs.Count, Unit: inputs.NCount,
			Desc: "Total number of kafka server Broker Topic Metrics Invalid Magic Number Records Per Sec, scoped by the Kafka auto-collect measurement tags.",
		},

		"kafka.server.BrokerTopicMetrics.InvalidMagicNumberRecordsPerSec.FifteenMinuteRate": &inputs.FieldInfo{
			DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount,
			Desc: "Fifteen-minute moving average rate for kafka server Broker Topic Metrics Invalid Magic Number Records Per Sec in events per second, scoped by the Kafka auto-collect measurement tags.",
		},

		"kafka.server.BrokerTopicMetrics.InvalidMagicNumberRecordsPerSec.FiveMinuteRate": &inputs.FieldInfo{
			DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount,
			Desc: "Five-minute moving average rate for kafka server Broker Topic Metrics Invalid Magic Number Records Per Sec in events per second, scoped by the Kafka auto-collect measurement tags.",
		},

		"kafka.server.BrokerTopicMetrics.InvalidMagicNumberRecordsPerSec.MeanRate": &inputs.FieldInfo{
			DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount,
			Desc: "Mean rate for kafka server Broker Topic Metrics Invalid Magic Number Records Per Sec in events per second, scoped by the Kafka auto-collect measurement tags.",
		},

		"kafka.server.BrokerTopicMetrics.InvalidMagicNumberRecordsPerSec.OneMinuteRate": &inputs.FieldInfo{
			DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount,
			Desc: "One-minute moving average rate for kafka server Broker Topic Metrics Invalid Magic Number Records Per Sec in events per second, scoped by the Kafka auto-collect measurement tags.",
		},

		"kafka.server.BrokerTopicMetrics.InvalidMessageCrcRecordsPerSec.Count": &inputs.FieldInfo{
			DataType: inputs.Int, Type: inputs.Count, Unit: inputs.NCount,
			Desc: "Total number of kafka server Broker Topic Metrics Invalid Message Crc Records Per Sec, scoped by the Kafka auto-collect measurement tags.",
		},

		"kafka.server.BrokerTopicMetrics.InvalidMessageCrcRecordsPerSec.FifteenMinuteRate": &inputs.FieldInfo{
			DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount,
			Desc: "Fifteen-minute moving average rate for kafka server Broker Topic Metrics Invalid Message Crc Records Per Sec in events per second, scoped by the Kafka auto-collect measurement tags.",
		},

		"kafka.server.BrokerTopicMetrics.InvalidMessageCrcRecordsPerSec.FiveMinuteRate": &inputs.FieldInfo{
			DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount,
			Desc: "Five-minute moving average rate for kafka server Broker Topic Metrics Invalid Message Crc Records Per Sec in events per second, scoped by the Kafka auto-collect measurement tags.",
		},

		"kafka.server.BrokerTopicMetrics.InvalidMessageCrcRecordsPerSec.MeanRate": &inputs.FieldInfo{
			DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount,
			Desc: "Mean rate for kafka server Broker Topic Metrics Invalid Message Crc Records Per Sec in events per second, scoped by the Kafka auto-collect measurement tags.",
		},

		"kafka.server.BrokerTopicMetrics.InvalidMessageCrcRecordsPerSec.OneMinuteRate": &inputs.FieldInfo{
			DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount,
			Desc: "One-minute moving average rate for kafka server Broker Topic Metrics Invalid Message Crc Records Per Sec in events per second, scoped by the Kafka auto-collect measurement tags.",
		},

		"kafka.server.BrokerTopicMetrics.InvalidOffsetOrSequenceRecordsPerSec.Count": &inputs.FieldInfo{
			DataType: inputs.Int, Type: inputs.Count, Unit: inputs.NCount,
			Desc: "Total number of kafka server Broker Topic Metrics Invalid Offset Or Sequence Records Per Sec, scoped by the Kafka auto-collect measurement tags.",
		},

		"kafka.server.BrokerTopicMetrics.InvalidOffsetOrSequenceRecordsPerSec.FifteenMinuteRate": &inputs.FieldInfo{
			DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount,
			Desc: "Fifteen-minute moving average rate for kafka server Broker Topic Metrics Invalid Offset Or Sequence Records Per Sec in events per second, scoped by the Kafka auto-collect measurement tags.",
		},

		"kafka.server.BrokerTopicMetrics.InvalidOffsetOrSequenceRecordsPerSec.FiveMinuteRate": &inputs.FieldInfo{
			DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount,
			Desc: "Five-minute moving average rate for kafka server Broker Topic Metrics Invalid Offset Or Sequence Records Per Sec in events per second, scoped by the Kafka auto-collect measurement tags.",
		},

		"kafka.server.BrokerTopicMetrics.InvalidOffsetOrSequenceRecordsPerSec.MeanRate": &inputs.FieldInfo{
			DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount,
			Desc: "Mean rate for kafka server Broker Topic Metrics Invalid Offset Or Sequence Records Per Sec in events per second, scoped by the Kafka auto-collect measurement tags.",
		},

		"kafka.server.BrokerTopicMetrics.InvalidOffsetOrSequenceRecordsPerSec.OneMinuteRate": &inputs.FieldInfo{
			DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount,
			Desc: "One-minute moving average rate for kafka server Broker Topic Metrics Invalid Offset Or Sequence Records Per Sec in events per second, scoped by the Kafka auto-collect measurement tags.",
		},

		"kafka.server.BrokerTopicMetrics.MessagesInPerSec.Count": &inputs.FieldInfo{
			DataType: inputs.Int, Type: inputs.Count, Unit: inputs.NCount,
			Desc: "Total number of kafka server Broker Topic Metrics Messages In Per Sec, scoped by topic.",
		},

		"kafka.server.BrokerTopicMetrics.MessagesInPerSec.FifteenMinuteRate": &inputs.FieldInfo{
			DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount,
			Desc: "Fifteen-minute moving average rate for kafka server Broker Topic Metrics Messages In Per Sec in events per second, scoped by topic.",
		},

		"kafka.server.BrokerTopicMetrics.MessagesInPerSec.FiveMinuteRate": &inputs.FieldInfo{
			DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount,
			Desc: "Five-minute moving average rate for kafka server Broker Topic Metrics Messages In Per Sec in events per second, scoped by topic.",
		},

		"kafka.server.BrokerTopicMetrics.MessagesInPerSec.MeanRate": &inputs.FieldInfo{
			DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount,
			Desc: "Mean rate for kafka server Broker Topic Metrics Messages In Per Sec in events per second, scoped by topic.",
		},

		"kafka.server.BrokerTopicMetrics.MessagesInPerSec.OneMinuteRate": &inputs.FieldInfo{
			DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount,
			Desc: "One-minute moving average rate for kafka server Broker Topic Metrics Messages In Per Sec in events per second, scoped by topic.",
		},

		"kafka.server.BrokerTopicMetrics.NoKeyCompactedTopicRecordsPerSec.Count": &inputs.FieldInfo{
			DataType: inputs.Int, Type: inputs.Count, Unit: inputs.NCount,
			Desc: "Total number of kafka server Broker Topic Metrics No Key Compacted Topic Records Per Sec, scoped by the Kafka auto-collect measurement tags.",
		},

		"kafka.server.BrokerTopicMetrics.NoKeyCompactedTopicRecordsPerSec.FifteenMinuteRate": &inputs.FieldInfo{
			DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount,
			Desc: "Fifteen-minute moving average rate for kafka server Broker Topic Metrics No Key Compacted Topic Records Per Sec in events per second, scoped by the Kafka auto-collect measurement tags.",
		},

		"kafka.server.BrokerTopicMetrics.NoKeyCompactedTopicRecordsPerSec.FiveMinuteRate": &inputs.FieldInfo{
			DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount,
			Desc: "Five-minute moving average rate for kafka server Broker Topic Metrics No Key Compacted Topic Records Per Sec in events per second, scoped by the Kafka auto-collect measurement tags.",
		},

		"kafka.server.BrokerTopicMetrics.NoKeyCompactedTopicRecordsPerSec.MeanRate": &inputs.FieldInfo{
			DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount,
			Desc: "Mean rate for kafka server Broker Topic Metrics No Key Compacted Topic Records Per Sec in events per second, scoped by the Kafka auto-collect measurement tags.",
		},

		"kafka.server.BrokerTopicMetrics.NoKeyCompactedTopicRecordsPerSec.OneMinuteRate": &inputs.FieldInfo{
			DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount,
			Desc: "One-minute moving average rate for kafka server Broker Topic Metrics No Key Compacted Topic Records Per Sec in events per second, scoped by the Kafka auto-collect measurement tags.",
		},

		"kafka.server.BrokerTopicMetrics.ProduceMessageConversionsPerSec.Count": &inputs.FieldInfo{
			DataType: inputs.Int, Type: inputs.Count, Unit: inputs.NCount,
			Desc: "Total number of kafka server Broker Topic Metrics Produce Message Conversions Per Sec, scoped by the Kafka auto-collect measurement tags.",
		},

		"kafka.server.BrokerTopicMetrics.ProduceMessageConversionsPerSec.FifteenMinuteRate": &inputs.FieldInfo{
			DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount,
			Desc: "Fifteen-minute moving average rate for kafka server Broker Topic Metrics Produce Message Conversions Per Sec in events per second, scoped by the Kafka auto-collect measurement tags.",
		},

		"kafka.server.BrokerTopicMetrics.ProduceMessageConversionsPerSec.FiveMinuteRate": &inputs.FieldInfo{
			DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount,
			Desc: "Five-minute moving average rate for kafka server Broker Topic Metrics Produce Message Conversions Per Sec in events per second, scoped by the Kafka auto-collect measurement tags.",
		},

		"kafka.server.BrokerTopicMetrics.ProduceMessageConversionsPerSec.MeanRate": &inputs.FieldInfo{
			DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount,
			Desc: "Mean rate for kafka server Broker Topic Metrics Produce Message Conversions Per Sec in events per second, scoped by the Kafka auto-collect measurement tags.",
		},

		"kafka.server.BrokerTopicMetrics.ProduceMessageConversionsPerSec.OneMinuteRate": &inputs.FieldInfo{
			DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount,
			Desc: "One-minute moving average rate for kafka server Broker Topic Metrics Produce Message Conversions Per Sec in events per second, scoped by the Kafka auto-collect measurement tags.",
		},

		"kafka.server.BrokerTopicMetrics.ReassignmentBytesInPerSec.Count": &inputs.FieldInfo{
			DataType: inputs.Int, Type: inputs.Count, Unit: inputs.SizeByte,
			Desc: "Total bytes recorded for kafka server Broker Topic Metrics Reassignment Bytes In Per Sec, scoped by the Kafka auto-collect measurement tags.",
		},

		"kafka.server.BrokerTopicMetrics.ReassignmentBytesInPerSec.FifteenMinuteRate": &inputs.FieldInfo{
			DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.BytesPerSec,
			Desc: "Fifteen-minute moving average rate for kafka server Broker Topic Metrics Reassignment Bytes In Per Sec in bytes per second, scoped by the Kafka auto-collect measurement tags.",
		},

		"kafka.server.BrokerTopicMetrics.ReassignmentBytesInPerSec.FiveMinuteRate": &inputs.FieldInfo{
			DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.BytesPerSec,
			Desc: "Five-minute moving average rate for kafka server Broker Topic Metrics Reassignment Bytes In Per Sec in bytes per second, scoped by the Kafka auto-collect measurement tags.",
		},

		"kafka.server.BrokerTopicMetrics.ReassignmentBytesInPerSec.MeanRate": &inputs.FieldInfo{
			DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.BytesPerSec,
			Desc: "Mean rate for kafka server Broker Topic Metrics Reassignment Bytes In Per Sec in bytes per second, scoped by the Kafka auto-collect measurement tags.",
		},

		"kafka.server.BrokerTopicMetrics.ReassignmentBytesInPerSec.OneMinuteRate": &inputs.FieldInfo{
			DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.BytesPerSec,
			Desc: "One-minute moving average rate for kafka server Broker Topic Metrics Reassignment Bytes In Per Sec in bytes per second, scoped by the Kafka auto-collect measurement tags.",
		},

		"kafka.server.BrokerTopicMetrics.ReassignmentBytesOutPerSec.Count": &inputs.FieldInfo{
			DataType: inputs.Int, Type: inputs.Count, Unit: inputs.SizeByte,
			Desc: "Total bytes recorded for kafka server Broker Topic Metrics Reassignment Bytes Out Per Sec, scoped by the Kafka auto-collect measurement tags.",
		},

		"kafka.server.BrokerTopicMetrics.ReassignmentBytesOutPerSec.FifteenMinuteRate": &inputs.FieldInfo{
			DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.BytesPerSec,
			Desc: "Fifteen-minute moving average rate for kafka server Broker Topic Metrics Reassignment Bytes Out Per Sec in bytes per second, scoped by the Kafka auto-collect measurement tags.",
		},

		"kafka.server.BrokerTopicMetrics.ReassignmentBytesOutPerSec.FiveMinuteRate": &inputs.FieldInfo{
			DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.BytesPerSec,
			Desc: "Five-minute moving average rate for kafka server Broker Topic Metrics Reassignment Bytes Out Per Sec in bytes per second, scoped by the Kafka auto-collect measurement tags.",
		},

		"kafka.server.BrokerTopicMetrics.ReassignmentBytesOutPerSec.MeanRate": &inputs.FieldInfo{
			DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.BytesPerSec,
			Desc: "Mean rate for kafka server Broker Topic Metrics Reassignment Bytes Out Per Sec in bytes per second, scoped by the Kafka auto-collect measurement tags.",
		},

		"kafka.server.BrokerTopicMetrics.ReassignmentBytesOutPerSec.OneMinuteRate": &inputs.FieldInfo{
			DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.BytesPerSec,
			Desc: "One-minute moving average rate for kafka server Broker Topic Metrics Reassignment Bytes Out Per Sec in bytes per second, scoped by the Kafka auto-collect measurement tags.",
		},

		"kafka.server.BrokerTopicMetrics.ReplicationBytesInPerSec.Count": &inputs.FieldInfo{
			DataType: inputs.Int, Type: inputs.Count, Unit: inputs.SizeByte,
			Desc: "Total bytes recorded for kafka server Broker Topic Metrics Replication Bytes In Per Sec, scoped by the Kafka auto-collect measurement tags.",
		},

		"kafka.server.BrokerTopicMetrics.ReplicationBytesInPerSec.FifteenMinuteRate": &inputs.FieldInfo{
			DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.BytesPerSec,
			Desc: "Fifteen-minute moving average rate for kafka server Broker Topic Metrics Replication Bytes In Per Sec in bytes per second, scoped by the Kafka auto-collect measurement tags.",
		},

		"kafka.server.BrokerTopicMetrics.ReplicationBytesInPerSec.FiveMinuteRate": &inputs.FieldInfo{
			DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.BytesPerSec,
			Desc: "Five-minute moving average rate for kafka server Broker Topic Metrics Replication Bytes In Per Sec in bytes per second, scoped by the Kafka auto-collect measurement tags.",
		},

		"kafka.server.BrokerTopicMetrics.ReplicationBytesInPerSec.MeanRate": &inputs.FieldInfo{
			DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.BytesPerSec,
			Desc: "Mean rate for kafka server Broker Topic Metrics Replication Bytes In Per Sec in bytes per second, scoped by the Kafka auto-collect measurement tags.",
		},

		"kafka.server.BrokerTopicMetrics.ReplicationBytesInPerSec.OneMinuteRate": &inputs.FieldInfo{
			DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.BytesPerSec,
			Desc: "One-minute moving average rate for kafka server Broker Topic Metrics Replication Bytes In Per Sec in bytes per second, scoped by the Kafka auto-collect measurement tags.",
		},

		"kafka.server.BrokerTopicMetrics.ReplicationBytesOutPerSec.Count": &inputs.FieldInfo{
			DataType: inputs.Int, Type: inputs.Count, Unit: inputs.SizeByte,
			Desc: "Total bytes recorded for kafka server Broker Topic Metrics Replication Bytes Out Per Sec, scoped by the Kafka auto-collect measurement tags.",
		},

		"kafka.server.BrokerTopicMetrics.ReplicationBytesOutPerSec.FifteenMinuteRate": &inputs.FieldInfo{
			DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.BytesPerSec,
			Desc: "Fifteen-minute moving average rate for kafka server Broker Topic Metrics Replication Bytes Out Per Sec in bytes per second, scoped by the Kafka auto-collect measurement tags.",
		},

		"kafka.server.BrokerTopicMetrics.ReplicationBytesOutPerSec.FiveMinuteRate": &inputs.FieldInfo{
			DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.BytesPerSec,
			Desc: "Five-minute moving average rate for kafka server Broker Topic Metrics Replication Bytes Out Per Sec in bytes per second, scoped by the Kafka auto-collect measurement tags.",
		},

		"kafka.server.BrokerTopicMetrics.ReplicationBytesOutPerSec.MeanRate": &inputs.FieldInfo{
			DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.BytesPerSec,
			Desc: "Mean rate for kafka server Broker Topic Metrics Replication Bytes Out Per Sec in bytes per second, scoped by the Kafka auto-collect measurement tags.",
		},

		"kafka.server.BrokerTopicMetrics.ReplicationBytesOutPerSec.OneMinuteRate": &inputs.FieldInfo{
			DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.BytesPerSec,
			Desc: "One-minute moving average rate for kafka server Broker Topic Metrics Replication Bytes Out Per Sec in bytes per second, scoped by the Kafka auto-collect measurement tags.",
		},

		"kafka.server.BrokerTopicMetrics.TotalFetchRequestsPerSec.Count": &inputs.FieldInfo{
			DataType: inputs.Int, Type: inputs.Count, Unit: inputs.NCount,
			Desc: "Total number of kafka server Broker Topic Metrics Total Fetch Requests Per Sec, scoped by topic.",
		},

		"kafka.server.BrokerTopicMetrics.TotalFetchRequestsPerSec.FifteenMinuteRate": &inputs.FieldInfo{
			DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.RequestsPerSec,
			Desc: "Fifteen-minute moving average rate for kafka server Broker Topic Metrics Total Fetch Requests Per Sec in requests per second, scoped by topic.",
		},

		"kafka.server.BrokerTopicMetrics.TotalFetchRequestsPerSec.FiveMinuteRate": &inputs.FieldInfo{
			DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.RequestsPerSec,
			Desc: "Five-minute moving average rate for kafka server Broker Topic Metrics Total Fetch Requests Per Sec in requests per second, scoped by topic.",
		},

		"kafka.server.BrokerTopicMetrics.TotalFetchRequestsPerSec.MeanRate": &inputs.FieldInfo{
			DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.RequestsPerSec,
			Desc: "Mean rate for kafka server Broker Topic Metrics Total Fetch Requests Per Sec in requests per second, scoped by topic.",
		},

		"kafka.server.BrokerTopicMetrics.TotalFetchRequestsPerSec.OneMinuteRate": &inputs.FieldInfo{
			DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.RequestsPerSec,
			Desc: "One-minute moving average rate for kafka server Broker Topic Metrics Total Fetch Requests Per Sec in requests per second, scoped by topic.",
		},

		"kafka.server.BrokerTopicMetrics.TotalProduceRequestsPerSec.Count": &inputs.FieldInfo{
			DataType: inputs.Int, Type: inputs.Count, Unit: inputs.NCount,
			Desc: "Total number of kafka server Broker Topic Metrics Total Produce Requests Per Sec, scoped by topic.",
		},

		"kafka.server.BrokerTopicMetrics.TotalProduceRequestsPerSec.FifteenMinuteRate": &inputs.FieldInfo{
			DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.RequestsPerSec,
			Desc: "Fifteen-minute moving average rate for kafka server Broker Topic Metrics Total Produce Requests Per Sec in requests per second, scoped by topic.",
		},

		"kafka.server.BrokerTopicMetrics.TotalProduceRequestsPerSec.FiveMinuteRate": &inputs.FieldInfo{
			DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.RequestsPerSec,
			Desc: "Five-minute moving average rate for kafka server Broker Topic Metrics Total Produce Requests Per Sec in requests per second, scoped by topic.",
		},

		"kafka.server.BrokerTopicMetrics.TotalProduceRequestsPerSec.MeanRate": &inputs.FieldInfo{
			DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.RequestsPerSec,
			Desc: "Mean rate for kafka server Broker Topic Metrics Total Produce Requests Per Sec in requests per second, scoped by topic.",
		},

		"kafka.server.BrokerTopicMetrics.TotalProduceRequestsPerSec.OneMinuteRate": &inputs.FieldInfo{
			DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.RequestsPerSec,
			Desc: "One-minute moving average rate for kafka server Broker Topic Metrics Total Produce Requests Per Sec in requests per second, scoped by topic.",
		},
	}
	return fields
}

//nolint:lll
func (m *kafkaMeasurement) getTopicFields() map[string]interface{} {
	fields := map[string]interface{}{
		"kafka.server.BrokerTopicMetrics.BytesInPerSec.Count": &inputs.FieldInfo{
			DataType: inputs.Int, Type: inputs.Count, Unit: inputs.SizeByte,
			Desc: "Total bytes recorded for kafka server Broker Topic Metrics Bytes In Per Sec, scoped by topic.",
		},

		"kafka.server.BrokerTopicMetrics.BytesInPerSec.FifteenMinuteRate": &inputs.FieldInfo{
			DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.BytesPerSec,
			Desc: "Fifteen-minute moving average rate for kafka server Broker Topic Metrics Bytes In Per Sec in bytes per second, scoped by topic.",
		},

		"kafka.server.BrokerTopicMetrics.BytesInPerSec.FiveMinuteRate": &inputs.FieldInfo{
			DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.BytesPerSec,
			Desc: "Five-minute moving average rate for kafka server Broker Topic Metrics Bytes In Per Sec in bytes per second, scoped by topic.",
		},

		"kafka.server.BrokerTopicMetrics.BytesInPerSec.MeanRate": &inputs.FieldInfo{
			DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.BytesPerSec,
			Desc: "Mean rate for kafka server Broker Topic Metrics Bytes In Per Sec in bytes per second, scoped by topic.",
		},

		"kafka.server.BrokerTopicMetrics.BytesInPerSec.OneMinuteRate": &inputs.FieldInfo{
			DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.BytesPerSec,
			Desc: "One-minute moving average rate for kafka server Broker Topic Metrics Bytes In Per Sec in bytes per second, scoped by topic.",
		},

		"kafka.server.BrokerTopicMetrics.BytesOutPerSec.Count": &inputs.FieldInfo{
			DataType: inputs.Int, Type: inputs.Count, Unit: inputs.SizeByte,
			Desc: "Total bytes recorded for kafka server Broker Topic Metrics Bytes Out Per Sec, scoped by topic.",
		},

		"kafka.server.BrokerTopicMetrics.BytesOutPerSec.FifteenMinuteRate": &inputs.FieldInfo{
			DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.BytesPerSec,
			Desc: "Fifteen-minute moving average rate for kafka server Broker Topic Metrics Bytes Out Per Sec in bytes per second, scoped by topic.",
		},

		"kafka.server.BrokerTopicMetrics.BytesOutPerSec.FiveMinuteRate": &inputs.FieldInfo{
			DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.BytesPerSec,
			Desc: "Five-minute moving average rate for kafka server Broker Topic Metrics Bytes Out Per Sec in bytes per second, scoped by topic.",
		},

		"kafka.server.BrokerTopicMetrics.BytesOutPerSec.MeanRate": &inputs.FieldInfo{
			DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.BytesPerSec,
			Desc: "Mean rate for kafka server Broker Topic Metrics Bytes Out Per Sec in bytes per second, scoped by topic.",
		},

		"kafka.server.BrokerTopicMetrics.BytesOutPerSec.OneMinuteRate": &inputs.FieldInfo{
			DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.BytesPerSec,
			Desc: "One-minute moving average rate for kafka server Broker Topic Metrics Bytes Out Per Sec in bytes per second, scoped by topic.",
		},

		"kafka.server.BrokerTopicMetrics.MessagesInPerSec.Count": &inputs.FieldInfo{
			DataType: inputs.Int, Type: inputs.Count, Unit: inputs.NCount,
			Desc: "Total number of kafka server Broker Topic Metrics Messages In Per Sec, scoped by topic.",
		},

		"kafka.server.BrokerTopicMetrics.MessagesInPerSec.FifteenMinuteRate": &inputs.FieldInfo{
			DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount,
			Desc: "Fifteen-minute moving average rate for kafka server Broker Topic Metrics Messages In Per Sec in events per second, scoped by topic.",
		},

		"kafka.server.BrokerTopicMetrics.MessagesInPerSec.FiveMinuteRate": &inputs.FieldInfo{
			DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount,
			Desc: "Five-minute moving average rate for kafka server Broker Topic Metrics Messages In Per Sec in events per second, scoped by topic.",
		},

		"kafka.server.BrokerTopicMetrics.MessagesInPerSec.MeanRate": &inputs.FieldInfo{
			DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount,
			Desc: "Mean rate for kafka server Broker Topic Metrics Messages In Per Sec in events per second, scoped by topic.",
		},

		"kafka.server.BrokerTopicMetrics.MessagesInPerSec.OneMinuteRate": &inputs.FieldInfo{
			DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount,
			Desc: "One-minute moving average rate for kafka server Broker Topic Metrics Messages In Per Sec in events per second, scoped by topic.",
		},

		"kafka.server.BrokerTopicMetrics.TotalFetchRequestsPerSec.Count": &inputs.FieldInfo{
			DataType: inputs.Int, Type: inputs.Count, Unit: inputs.NCount,
			Desc: "Total number of kafka server Broker Topic Metrics Total Fetch Requests Per Sec, scoped by topic.",
		},

		"kafka.server.BrokerTopicMetrics.TotalFetchRequestsPerSec.FifteenMinuteRate": &inputs.FieldInfo{
			DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.RequestsPerSec,
			Desc: "Fifteen-minute moving average rate for kafka server Broker Topic Metrics Total Fetch Requests Per Sec in requests per second, scoped by topic.",
		},

		"kafka.server.BrokerTopicMetrics.TotalFetchRequestsPerSec.FiveMinuteRate": &inputs.FieldInfo{
			DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.RequestsPerSec,
			Desc: "Five-minute moving average rate for kafka server Broker Topic Metrics Total Fetch Requests Per Sec in requests per second, scoped by topic.",
		},

		"kafka.server.BrokerTopicMetrics.TotalFetchRequestsPerSec.MeanRate": &inputs.FieldInfo{
			DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.RequestsPerSec,
			Desc: "Mean rate for kafka server Broker Topic Metrics Total Fetch Requests Per Sec in requests per second, scoped by topic.",
		},

		"kafka.server.BrokerTopicMetrics.TotalFetchRequestsPerSec.OneMinuteRate": &inputs.FieldInfo{
			DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.RequestsPerSec,
			Desc: "One-minute moving average rate for kafka server Broker Topic Metrics Total Fetch Requests Per Sec in requests per second, scoped by topic.",
		},

		"kafka.server.BrokerTopicMetrics.TotalProduceRequestsPerSec.Count": &inputs.FieldInfo{
			DataType: inputs.Int, Type: inputs.Count, Unit: inputs.NCount,
			Desc: "Total number of kafka server Broker Topic Metrics Total Produce Requests Per Sec, scoped by topic.",
		},

		"kafka.server.BrokerTopicMetrics.TotalProduceRequestsPerSec.FifteenMinuteRate": &inputs.FieldInfo{
			DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.RequestsPerSec,
			Desc: "Fifteen-minute moving average rate for kafka server Broker Topic Metrics Total Produce Requests Per Sec in requests per second, scoped by topic.",
		},

		"kafka.server.BrokerTopicMetrics.TotalProduceRequestsPerSec.FiveMinuteRate": &inputs.FieldInfo{
			DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.RequestsPerSec,
			Desc: "Five-minute moving average rate for kafka server Broker Topic Metrics Total Produce Requests Per Sec in requests per second, scoped by topic.",
		},

		"kafka.server.BrokerTopicMetrics.TotalProduceRequestsPerSec.MeanRate": &inputs.FieldInfo{
			DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.RequestsPerSec,
			Desc: "Mean rate for kafka server Broker Topic Metrics Total Produce Requests Per Sec in requests per second, scoped by topic.",
		},

		"kafka.server.BrokerTopicMetrics.TotalProduceRequestsPerSec.OneMinuteRate": &inputs.FieldInfo{
			DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.RequestsPerSec,
			Desc: "One-minute moving average rate for kafka server Broker Topic Metrics Total Produce Requests Per Sec in requests per second, scoped by topic.",
		},
	}
	m.addTaggedbyToFields(fields, TagGroupTopic)
	return fields
}

//nolint:lll
func (m *kafkaMeasurement) getPartitionFields() map[string]interface{} {
	fields := map[string]interface{}{
		"kafka.log.Log.LogEndOffset.Value": &inputs.FieldInfo{
			DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.NCount,
			Desc: "Current offset for kafka log Log Log End Offset, scoped by partition and topic.",
		},

		"kafka.log.Log.LogStartOffset.Value": &inputs.FieldInfo{
			DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.NCount,
			Desc: "Current offset for kafka log Log Log Start Offset, scoped by partition and topic.",
		},

		"kafka.log.Log.NumLogSegments.Value": &inputs.FieldInfo{
			DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.NCount,
			Desc: "Current number of kafka log Log Num Log Segments, scoped by partition and topic.",
		},

		"kafka.log.Log.Size.Value": &inputs.FieldInfo{
			DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.SizeByte,
			Desc: "Current size of kafka log Log Size in bytes, scoped by partition and topic.",
		},
	}
	m.addTaggedbyToFields(fields, TagGroupPartition)
	return fields
}

//nolint:lll
func (m *kafkaMeasurement) getLogFields() map[string]interface{} {
	fields := map[string]interface{}{
		"kafka.log.LogManager.OfflineLogDirectoryCount.Value": &inputs.FieldInfo{
			DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.NCount,
			Desc: "Current number of kafka log Log Manager Offline Log Directory Count, scoped by the Kafka auto-collect measurement tags.",
		},

		"kafka.log.LogCleaner.cleaner_recopy_percent.Value": &inputs.FieldInfo{
			DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.Percent,
			Desc: "Percentage for kafka log Log Cleaner cleaner recopy percent as a value from 0 to 100, scoped by the Kafka auto-collect measurement tags.",
		},

		"kafka.log.LogCleaner.max_compaction_delay_secs.Value": &inputs.FieldInfo{
			DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.DurationSecond,
			Desc: "Time spent in kafka log Log Cleaner max compaction delay secs, measured in seconds, scoped by the Kafka auto-collect measurement tags.",
		},

		"kafka.log.LogCleaner.max_clean_time_secs.Value": &inputs.FieldInfo{
			DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.DurationSecond,
			Desc: "Time spent in kafka log Log Cleaner max clean time secs, measured in seconds, scoped by the Kafka auto-collect measurement tags.",
		},

		"kafka.log.LogCleaner.DeadThreadCount.Value": &inputs.FieldInfo{
			DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.NCount,
			Desc: "Current number of kafka log Log Cleaner Dead Thread Count, scoped by the Kafka auto-collect measurement tags.",
		},

		"kafka.log.LogCleaner.max_buffer_utilization_percent.Value": &inputs.FieldInfo{
			DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.Percent,
			Desc: "Percentage for kafka log Log Cleaner max buffer utilization percent as a value from 0 to 100, scoped by the Kafka auto-collect measurement tags.",
		},
	}
	return fields
}

func (m *kafkaMeasurement) addTaggedbyToFields(fields map[string]interface{}, tagGroup string) {
	var tags map[string]interface{}

	// Select corresponding getTags function based on tagGroup
	switch tagGroup {
	case TagGroupPurgatory:
		tags = m.getPurgatoryTags()
	case TagGroupRequest:
		tags = m.getRequestTags()
	case TagGroupTopic:
		tags = m.getTopicTags()
	case TagGroupPartition:
		tags = m.getPartitionTags()
	default:
		return
	}

	// Extract tag keys
	taggedBy := make([]string, 0, len(tags))
	for tag := range tags {
		taggedBy = append(taggedBy, tag)
	}

	// Add Taggedby to each field
	for _, field := range fields {
		if fieldInfo, ok := field.(*inputs.FieldInfo); ok {
			fieldInfo.Taggedby = taggedBy
		}
	}
}
