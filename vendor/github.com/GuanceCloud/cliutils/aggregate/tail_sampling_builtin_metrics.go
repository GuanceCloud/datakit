package aggregate

import (
	"time"

	"github.com/GuanceCloud/cliutils/point"
)

func traceBuiltinMetricNames() []string {
	return []string{
		"trace_total_count",
		"trace_kept_count",
		"trace_dropped_count",
		"trace_error_count",
		"span_total_count",
		"trace_duration",
	}
}

func loggingBuiltinMetricNames() []string {
	return []string{
		"logging_total_count",
		"logging_error_count",
		"logging_kept_count",
		"logging_dropped_count",
	}
}

func rumBuiltinMetricNames() []string {
	return []string{
		"rum_total_count",
		"rum_kept_count",
		"rum_dropped_count",
	}
}

type countBuiltinDerivedMetric struct {
	name       string
	onIngest   func(packet *DataPacket) (float64, bool)
	onDecision func(packet *DataPacket, decision DerivedMetricDecision) (float64, bool)
}

func (m *countBuiltinDerivedMetric) Name() string {
	return m.name
}

func (m *countBuiltinDerivedMetric) OnIngest(packet *DataPacket) []DerivedMetricRecord {
	if m == nil || m.onIngest == nil || packet == nil {
		return nil
	}

	value, ok := m.onIngest(packet)
	if !ok {
		return nil
	}

	return []DerivedMetricRecord{newDerivedMetricRecord(packet, m.name, DerivedMetricStageIngest, DerivedMetricDecisionUnknown, value)}
}

func (m *countBuiltinDerivedMetric) OnDecision(packet *DataPacket, decision DerivedMetricDecision) []DerivedMetricRecord {
	if m == nil || m.onDecision == nil || packet == nil {
		return nil
	}

	value, ok := m.onDecision(packet, decision)
	if !ok {
		return nil
	}

	return []DerivedMetricRecord{newDerivedMetricRecord(packet, m.name, DerivedMetricStageDecision, decision, value)}
}

func (m *countBuiltinDerivedMetric) OnPreDecision(packet *DataPacket) []DerivedMetricRecord {
	return nil
}

type histogramBuiltinDerivedMetric struct {
	name       string
	buckets    []float64
	onObserve  func(packet *DataPacket) (float64, bool)
	recordTags func(packet *DataPacket) map[string]string
}

func (m *histogramBuiltinDerivedMetric) Name() string {
	return m.name
}

func (m *histogramBuiltinDerivedMetric) OnIngest(packet *DataPacket) []DerivedMetricRecord {
	return nil
}

func (m *histogramBuiltinDerivedMetric) OnDecision(packet *DataPacket, decision DerivedMetricDecision) []DerivedMetricRecord {
	return nil
}

func (m *histogramBuiltinDerivedMetric) OnPreDecision(packet *DataPacket) []DerivedMetricRecord {
	if m == nil || packet == nil || m.onObserve == nil {
		return nil
	}

	val, ok := m.onObserve(packet)
	if !ok {
		return nil
	}

	tags := builtinRecordTags(packet)
	if m.recordTags != nil {
		tags = m.recordTags(packet)
	}

	return []DerivedMetricRecord{newHistogramDerivedMetricRecord(
		packet,
		m.name,
		DerivedMetricStagePreDecision,
		DerivedMetricDecisionUnknown,
		val,
		tags,
		m.buckets,
	)}
}

var defaultTraceDurationBucketsMS = []float64{1, 10, 20, 50, 100, 200, 500, 1000, 2000, 5000}

func DefaultTailSamplingBuiltinMetrics() TailSamplingBuiltinMetrics {
	return TailSamplingBuiltinMetrics{
		&countBuiltinDerivedMetric{
			name: "trace_total_count",
			onIngest: func(packet *DataPacket) (float64, bool) {
				return builtinDataTypeCount(packet, point.STracing)
			},
		},
		&countBuiltinDerivedMetric{
			name: "trace_kept_count",
			onDecision: func(packet *DataPacket, decision DerivedMetricDecision) (float64, bool) {
				if packet != nil && packet.DataType == point.STracing && decision == DerivedMetricDecisionKept {
					return 1, true
				}
				return 0, false
			},
		},
		&countBuiltinDerivedMetric{
			name: "trace_dropped_count",
			onDecision: func(packet *DataPacket, decision DerivedMetricDecision) (float64, bool) {
				if packet != nil && packet.DataType == point.STracing && decision == DerivedMetricDecisionDropped {
					return 1, true
				}
				return 0, false
			},
		},
		&countBuiltinDerivedMetric{
			name: "trace_error_count",
			onIngest: func(packet *DataPacket) (float64, bool) {
				if packet != nil && packet.DataType == point.STracing && packet.HasError {
					return 1, true
				}
				return 0, false
			},
		},
		&countBuiltinDerivedMetric{
			name: "span_total_count",
			onIngest: func(packet *DataPacket) (float64, bool) {
				if packet == nil || packet.DataType != point.STracing {
					return 0, false
				}
				if packet.PointCount <= 0 {
					return 0, false
				}
				return float64(packet.PointCount), true
			},
		},
		&histogramBuiltinDerivedMetric{
			name:    "trace_duration",
			buckets: defaultTraceDurationBucketsMS,
			onObserve: func(packet *DataPacket) (float64, bool) {
				if packet == nil || packet.DataType != point.STracing {
					return 0, false
				}
				if packet.TraceEndTimeUnixNano <= packet.TraceStartTimeUnixNano {
					return 0, false
				}
				durationMS := float64(packet.TraceEndTimeUnixNano-packet.TraceStartTimeUnixNano) / float64(time.Millisecond)
				return durationMS, true
			},
		},
		&countBuiltinDerivedMetric{
			name: "logging_total_count",
			onIngest: func(packet *DataPacket) (float64, bool) {
				return builtinDataTypeCount(packet, point.SLogging)
			},
		},
		&countBuiltinDerivedMetric{
			name: "logging_error_count",
			onIngest: func(packet *DataPacket) (float64, bool) {
				if packet != nil && packet.DataType == point.SLogging && packet.HasError {
					return 1, true
				}
				return 0, false
			},
		},
		&countBuiltinDerivedMetric{
			name: "logging_kept_count",
			onDecision: func(packet *DataPacket, decision DerivedMetricDecision) (float64, bool) {
				if packet != nil && packet.DataType == point.SLogging && decision == DerivedMetricDecisionKept {
					return 1, true
				}
				return 0, false
			},
		},
		&countBuiltinDerivedMetric{
			name: "logging_dropped_count",
			onDecision: func(packet *DataPacket, decision DerivedMetricDecision) (float64, bool) {
				if packet != nil && packet.DataType == point.SLogging && decision == DerivedMetricDecisionDropped {
					return 1, true
				}
				return 0, false
			},
		},
		&countBuiltinDerivedMetric{
			name: "rum_total_count",
			onIngest: func(packet *DataPacket) (float64, bool) {
				return builtinDataTypeCount(packet, point.SRUM)
			},
		},
		&countBuiltinDerivedMetric{
			name: "rum_kept_count",
			onDecision: func(packet *DataPacket, decision DerivedMetricDecision) (float64, bool) {
				if packet != nil && packet.DataType == point.SRUM && decision == DerivedMetricDecisionKept {
					return 1, true
				}
				return 0, false
			},
		},
		&countBuiltinDerivedMetric{
			name: "rum_dropped_count",
			onDecision: func(packet *DataPacket, decision DerivedMetricDecision) (float64, bool) {
				if packet != nil && packet.DataType == point.SRUM && decision == DerivedMetricDecisionDropped {
					return 1, true
				}
				return 0, false
			},
		},
	}
}

func builtinDataTypeCount(packet *DataPacket, dataType string) (float64, bool) {
	if packet != nil && packet.DataType == dataType {
		return 1, true
	}
	return 0, false
}

func newDerivedMetricRecord(
	packet *DataPacket,
	metricName string,
	stage DerivedMetricStage,
	decision DerivedMetricDecision,
	value float64,
) DerivedMetricRecord {
	return DerivedMetricRecord{
		Token:       packet.Token,
		DataType:    point.SMetric,
		MetricName:  metricName,
		Kind:        DerivedMetricKindSum,
		Stage:       stage,
		Decision:    decision,
		Measurement: TailSamplingDerivedMeasurement,
		Tags:        builtinRecordTags(packet),
		Value:       value,
		Time:        packetTime(packet),
	}
}

func newHistogramDerivedMetricRecord(
	packet *DataPacket,
	metricName string,
	stage DerivedMetricStage,
	decision DerivedMetricDecision,
	value float64,
	tags map[string]string,
	buckets []float64,
) DerivedMetricRecord {
	return DerivedMetricRecord{
		Token:       packet.Token,
		DataType:    point.SMetric,
		MetricName:  metricName,
		Kind:        DerivedMetricKindHistogram,
		Stage:       stage,
		Decision:    decision,
		Measurement: TailSamplingDerivedMeasurement,
		Tags:        tags,
		Value:       value,
		Buckets:     append([]float64(nil), buckets...),
		Time:        packetTime(packet),
	}
}

func packetTime(packet *DataPacket) time.Time {
	if packet == nil {
		return time.Now()
	}

	maxPointTS := packet.MaxPointTimeUnixNano
	if maxPointTS > 0 {
		return time.Unix(0, maxPointTS)
	}

	switch {
	case packet.TraceEndTimeUnixNano > 0:
		return time.Unix(0, packet.TraceEndTimeUnixNano)
	default:
		return time.Now()
	}
}

func builtinRecordTags(packet *DataPacket) map[string]string {
	tags := map[string]string{
		"data_type": packet.DataType,
	}

	if packet.Source != "" {
		tags["source"] = packet.Source
	}

	if packet.GroupKey != "" {
		tags["group_key"] = packet.GroupKey
	}

	return tags
}
