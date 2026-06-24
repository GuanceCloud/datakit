// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

//nolint:lll,gochecknoinits,goconst,funlen // Kafka metadata descriptions and rule checks are intentionally explicit.
package kafka

import (
	"strings"
	"unicode"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs"
)

type kafkaMetadataRegistration struct {
	measurement string
	fields      map[string]interface{}
	field       string
	info        *inputs.FieldInfo
}

// Source: Kafka monitoring MBean attributes documented at https://kafka.apache.org/42/operations/monitoring/.
// Kafka client and Connect metric names are emitted by the Kafka clients through
// JMX/Jolokia and collected without runtime transformation by this input.
var kafkaMetadataRegistry = []kafkaMetadataRegistration{
	kafkaMeta(
		"kafka_connect", connectFields, "commit_id",
		kafkaStringAttr("Kafka Connect worker commit id string, scoped by client_id when present."),
	),
	kafkaMeta(
		"kafka_connect", connectFields, "start_time_ms",
		kafkaIntGauge(
			inputs.TimestampMS,
			"Kafka Connect worker start time as Unix epoch milliseconds, scoped by client_id when present.",
		),
	),
	kafkaMeta(
		"kafka_connect", connectFields, "version",
		kafkaStringAttr("Kafka Connect worker version string, scoped by client_id when present."),
	),
	kafkaMeta(
		"kafka_connect", connectFields, "running_ratio",
		kafkaFloatGauge(
			inputs.PercentDecimal,
			"Fraction from 0 to 1 of time the Kafka Connect task is running, scoped by connector and task.",
		),
	),
	kafkaMeta(
		"kafka_connect", connectFields, "source_record_poll_rate",
		kafkaPerSecond(
			"Number of source records polled per second by this Kafka Connect source task, scoped by connector and task.",
		),
	),
	kafkaMeta(
		"kafka_connect", connectFields, "sink_record_send_rate",
		kafkaPerSecond(
			"Number of sink records sent per second by this Kafka Connect sink task, scoped by connector and task.",
		),
	),

	kafkaMeta(
		"kafka_consumer", consumerFields, "bytes_consumed_rate",
		kafkaFloatGauge(inputs.BytesPerSec, "Number of bytes consumed per second by this consumer, scoped by client_id."),
	),
	kafkaMeta(
		"kafka_consumer", consumerFields, "fetch_latency_avg",
		kafkaFloatGauge(inputs.DurationMS, "Average consumer fetch request latency in milliseconds, scoped by client_id."),
	),
	kafkaMeta(
		"kafka_consumer", consumerFields, "fetch_rate",
		kafkaFloatGauge(inputs.RequestsPerSec, "Number of consumer fetch requests per second, scoped by client_id."),
	),
	kafkaMeta(
		"kafka_consumer", consumerFields, "records_consumed_rate",
		kafkaPerSecond("Number of records consumed per second by this consumer, scoped by client_id."),
	),

	kafkaMeta(
		"kafka_controller", controllerFields, "AutoLeaderBalanceRateAndTimeMs.EventType",
		kafkaStringAttr(
			"String event type label reported by Kafka for AutoLeaderBalanceRateAndTimeMs, scoped by the controller broker.",
		),
	),
	kafkaMeta(
		"kafka_controller", controllerFields, "AutoLeaderBalanceRateAndTimeMs.LatencyUnit",
		kafkaStringAttr(
			"String unit label reported by Kafka for AutoLeaderBalanceRateAndTimeMs latency values, scoped by the controller broker.",
		),
	),
	kafkaMeta(
		"kafka_controller", controllerFields, "AutoLeaderBalanceRateAndTimeMs.Mean",
		kafkaFloatGauge(inputs.DurationMS, "Mean auto leader balance latency in milliseconds, scoped by the controller broker."),
	),
	kafkaMeta(
		"kafka_controller", controllerFields, "AutoLeaderBalanceRateAndTimeMs.RateUnit",
		kafkaStringAttr(
			"String rate unit label reported by Kafka for AutoLeaderBalanceRateAndTimeMs rate values, scoped by the controller broker.",
		),
	),

	kafkaMeta(
		"kafka_partition", partitionFields, "LogEndOffset",
		kafkaIntGauge(inputs.NCount, "Current log end offset for this topic partition, scoped by topic and partition."),
	),
	kafkaMeta(
		"kafka_partition", partitionFields, "Size",
		kafkaIntGauge(inputs.SizeByte, "Current on-disk size in bytes for this topic partition log, scoped by topic and partition."),
	),

	kafkaMeta(
		"kafka_producer", producerFields, "buffer_total_bytes",
		kafkaIntGauge(
			inputs.SizeByte,
			"Total producer buffer memory in bytes available to this producer, scoped by client_id.",
		),
	),
	kafkaMeta(
		"kafka_producer", producerFields, "incoming_byte_rate",
		kafkaFloatGauge(inputs.BytesPerSec, "Number of bytes received per second by this producer client, scoped by client_id."),
	),
	kafkaMeta(
		"kafka_producer", producerFields, "record_send_rate",
		kafkaPerSecond("Number of records sent per second by this producer, scoped by client_id."),
	),
	kafkaMeta(
		"kafka_producer", producerFields, "record_send_total",
		kafkaIntCount(inputs.NCount, "Total number of records sent by this producer, scoped by client_id."),
	),

	kafkaMeta(
		"kafka_request", requestFields, "LocalTimeMs.Mean",
		kafkaFloatGauge(inputs.DurationMS, "Mean broker request local processing latency in milliseconds, scoped by request."),
	),
	kafkaMeta(
		"kafka_request", requestFields, "RequestBytes.Mean",
		kafkaFloatGauge(inputs.SizeByte, "Mean broker request size in bytes, scoped by request."),
	),

	kafkaMeta(
		"kafka_topic", topicFields, "BytesInPerSec.MeanRate",
		kafkaFloatGauge(inputs.BytesPerSec, "Mean number of bytes received per second for this topic, scoped by topic."),
	),
	kafkaMeta(
		"kafka_topic", topicFields, "BytesInPerSec.RateUnit",
		kafkaStringAttr("String rate unit label reported by Kafka for BytesInPerSec, scoped by topic."),
	),
	kafkaMeta(
		"kafka_topic", topicFields, "MessagesInPerSec.OneMinuteRate",
		kafkaPerSecond("One-minute rate of messages received per second for this topic, scoped by topic."),
	),

	kafkaMeta(
		"kafka_topics", topicsFields, "BytesInPerSec.EventType",
		kafkaStringAttr("String event type label reported by Kafka for cluster-wide BytesInPerSec."),
	),
	kafkaMeta(
		"kafka_topics", topicsFields, "BytesInPerSec.MeanRate",
		kafkaFloatGauge(inputs.BytesPerSec, "Mean number of bytes received per second across broker topics."),
	),
	kafkaMeta(
		"kafka_topics", topicsFields, "MessagesInPerSec.OneMinuteRate",
		kafkaPerSecond("One-minute rate of messages received per second across broker topics."),
	),
}

func init() {
	applyKafkaMetadataToLegacyFields()
	for _, entry := range kafkaMetadataRegistry {
		entry.fields[entry.field] = entry.info
	}
}

func kafkaMeta(measurement string, fields map[string]interface{}, field string, info *inputs.FieldInfo) kafkaMetadataRegistration {
	return kafkaMetadataRegistration{
		measurement: measurement,
		fields:      fields,
		field:       field,
		info:        info,
	}
}

func kafkaStringAttr(desc string) *inputs.FieldInfo {
	return kafkaField(inputs.String, inputs.String, inputs.NoUnit, desc)
}

func kafkaIntGauge(unit, desc string) *inputs.FieldInfo {
	return kafkaField(inputs.Int, inputs.Gauge, unit, desc)
}

func kafkaFloatGauge(unit, desc string) *inputs.FieldInfo {
	return kafkaField(inputs.Float, inputs.Gauge, unit, desc)
}

func kafkaIntCount(unit, desc string) *inputs.FieldInfo {
	return kafkaField(inputs.Int, inputs.Count, unit, desc)
}

func kafkaPerSecond(desc string) *inputs.FieldInfo {
	return kafkaFloatGauge(inputs.NCount, desc)
}

func kafkaField(dataType, metricType, unit, desc string) *inputs.FieldInfo {
	return &inputs.FieldInfo{
		DataType: dataType,
		Type:     metricType,
		Unit:     unit,
		Desc:     desc,
	}
}

func applyKafkaMetadataToLegacyFields() {
	normalizeKafkaFields("kafka_connect", connectFields, "client_id, connector, and task tags when present")
	normalizeKafkaFields("kafka_producer", producerFields, "client_id")
	normalizeKafkaFields("kafka_consumer", consumerFields, "client_id")
	normalizeKafkaFields("kafka_log", logFields, "host and Jolokia agent")
	normalizeKafkaFields("kafka_network", networkFields, "host and Jolokia agent")
	normalizeKafkaFields("kafka_request_handler", requestHandlerFields, "host and Jolokia agent")
	normalizeKafkaFields("kafka_zookeeper", zooKeeperFields, "Jolokia agent")
	normalizeKafkaFields("kafka_controller", controllerFields, "controller broker")
	normalizeKafkaFields("kafka_replication", replicationFields, "host and Jolokia agent")
	normalizeKafkaFields("kafka_purgatory", purgatoryFields, "delayedOperation")
	normalizeKafkaFields("kafka_request", requestFields, "request")
	normalizeKafkaFields("kafka_topics", topicsFields, "host and Jolokia agent")
	normalizeKafkaFields("kafka_topic", topicFields, "topic")
	normalizeKafkaFields("kafka_partition", partitionFields, "topic and partition")
}

func applyKafkaMeasurementMetadata(fields map[string]interface{}) {
	normalizeKafkaFields("kafka", fields, "the Kafka auto-collect measurement tags")
}

func normalizeKafkaFields(measurement string, fields map[string]interface{}, defaultScope string) {
	for name, field := range fields {
		info, ok := field.(*inputs.FieldInfo)
		if !ok {
			continue
		}
		normalizeKafkaField(measurement, name, info, kafkaFieldScope(info, defaultScope))
	}
}

func normalizeKafkaField(_ string, name string, info *inputs.FieldInfo, scope string) {
	lower := strings.ToLower(name)
	attr := kafkaFieldAttr(name)
	attrLower := strings.ToLower(attr)
	base := kafkaFieldBase(name)
	baseText := kafkaHumanize(base)
	attrText := kafkaHumanize(attr)
	label := kafkaMetricLabel(baseText, attrText, attrLower)
	disabled := info.Disabled
	taggedby := info.Taggedby

	dataType := info.DataType
	var metricType string
	var unit string
	desc := ""

	switch {
	case kafkaIsStringAttribute(lower, info):
		dataType = inputs.String
		metricType = inputs.String
		unit = inputs.NoUnit
		desc = kafkaStringAttributeDesc(lower, attrLower, label, scope)
	case kafkaIsTimestamp(lower):
		dataType = kafkaNumericDataType(dataType, inputs.Int)
		metricType = inputs.Gauge
		unit = inputs.TimestampMS
		desc = "Unix epoch timestamp in milliseconds for " + label + ", scoped by " + scope + "."
	case kafkaIsDurationNS(lower):
		dataType = kafkaNumericDataType(dataType, inputs.Float)
		metricType = inputs.Gauge
		unit = inputs.DurationNS
		desc = kafkaDurationDesc(attrLower, label, "nanoseconds", scope)
	case kafkaIsDurationSecond(lower):
		dataType = kafkaNumericDataType(dataType, inputs.Float)
		metricType = inputs.Gauge
		unit = inputs.DurationSecond
		desc = kafkaDurationDesc(attrLower, label, "seconds", scope)
	case kafkaIsTimerDuration(lower, attrLower):
		dataType = kafkaNumericDataType(dataType, inputs.Float)
		metricType = inputs.Gauge
		unit = inputs.DurationMS
		desc = kafkaDurationDesc(attrLower, label, "milliseconds", scope)
	case kafkaIsFraction(lower):
		dataType = kafkaNumericDataType(dataType, inputs.Float)
		metricType = inputs.Gauge
		unit = inputs.PercentDecimal
		desc = kafkaRatioDesc(label, scope)
	case strings.Contains(lower, "percentage") || kafkaContainsPercentUnit(lower):
		dataType = kafkaNumericDataType(dataType, inputs.Float)
		metricType = inputs.Gauge
		unit = inputs.Percent
		desc = kafkaPercentDesc(attrLower, label, scope)
	case kafkaIsByteRate(lower, attrLower):
		dataType = kafkaNumericDataType(dataType, inputs.Float)
		metricType = inputs.Gauge
		unit = inputs.BytesPerSec
		desc = kafkaRateDesc(attrLower, label, "bytes per second", scope)
	case kafkaIsRequestRate(lower, attrLower):
		dataType = kafkaNumericDataType(dataType, inputs.Float)
		metricType = inputs.Gauge
		unit = inputs.RequestsPerSec
		desc = kafkaRateDesc(attrLower, label, "requests per second", scope)
	case kafkaIsByteSize(lower, attrLower):
		dataType = kafkaNumericDataType(dataType, inputs.Float)
		if kafkaIsCountAttr(attrLower) {
			metricType = inputs.Count
			unit = kafkaByteCountUnit(lower)
			desc = kafkaCountDesc(label, scope, true)
		} else {
			metricType = inputs.Gauge
			unit = inputs.SizeByte
			desc = kafkaSizeDesc(attrLower, label, scope)
		}
	case kafkaIsPerSecondRate(lower, attrLower):
		dataType = kafkaNumericDataType(dataType, inputs.Float)
		metricType = inputs.Gauge
		unit = inputs.NCount
		desc = kafkaRateDesc(attrLower, label, "events per second", scope)
	case kafkaIsCountAttr(attrLower) || strings.HasSuffix(lower, "_total"):
		dataType = kafkaNumericDataType(dataType, inputs.Int)
		metricType = inputs.Count
		unit = inputs.NCount
		desc = kafkaCountDesc(label, scope, false)
	case strings.Contains(lower, "offset"):
		dataType = kafkaNumericDataType(dataType, inputs.Int)
		metricType = inputs.Gauge
		unit = inputs.NCount
		desc = "Current offset for " + label + ", scoped by " + scope + "."
	case strings.Contains(lower, "count") || strings.Contains(lower, "queue") ||
		strings.Contains(lower, "partition") || strings.Contains(lower, "replica") ||
		strings.Contains(lower, "segment") || strings.Contains(lower, "attempts") ||
		strings.Contains(lower, "retries") || strings.Contains(lower, "errors") ||
		strings.Contains(lower, "records") || strings.Contains(lower, "messages"):
		dataType = kafkaNumericDataType(dataType, inputs.Int)
		metricType = inputs.Gauge
		unit = inputs.NCount
		desc = kafkaGaugeCountDesc(lower, label, scope)
	default:
		dataType = kafkaNumericDataType(dataType, inputs.Float)
		metricType = inputs.Gauge
		unit = inputs.NCount
		desc = "Current value of " + label + ", scoped by " + scope + "."
	}

	if strings.TrimSpace(desc) == "" {
		desc = "Kafka metric value for " + label + ", scoped by " + scope + "."
	}

	info.DataType = dataType
	info.Type = metricType
	info.Unit = unit
	info.Desc = desc
	info.Taggedby = taggedby
	info.Disabled = disabled
}

func kafkaFieldScope(info *inputs.FieldInfo, defaultScope string) string {
	if len(info.Taggedby) == 0 {
		return defaultScope
	}
	return strings.Join(info.Taggedby, " and ")
}

func kafkaNumericDataType(current, fallback string) string {
	if current == inputs.Int || current == inputs.Float {
		return current
	}
	return fallback
}

func kafkaFieldAttr(name string) string {
	idx := strings.LastIndex(name, ".")
	if idx < 0 {
		return name
	}
	return name[idx+1:]
}

func kafkaFieldBase(name string) string {
	idx := strings.LastIndex(name, ".")
	if idx < 0 {
		return name
	}
	return name[:idx]
}

func kafkaMetricLabel(baseText, attrText, attrLower string) string {
	switch {
	case attrLower == "value" || attrLower == "count" || attrLower == "purgatorysize" || kafkaIsRateAttr(attrLower) || kafkaIsSummaryAttr(attrLower):
		return kafkaLowerFirst(kafkaNormalizeMetricText(baseText))
	case attrLower == "eventtype" || attrLower == "latencyunit" || attrLower == "rateunit":
		return kafkaLowerFirst(kafkaNormalizeMetricText(baseText))
	case strings.EqualFold(baseText, attrText) || baseText == "":
		return kafkaLowerFirst(kafkaNormalizeMetricText(attrText))
	default:
		return kafkaLowerFirst(kafkaNormalizeMetricText(attrText))
	}
}

func kafkaHumanize(name string) string {
	replacer := strings.NewReplacer(".", " ", "_", " ", "-", " ")
	name = replacer.Replace(name)
	var b strings.Builder
	var prev rune
	for _, r := range name {
		if unicode.IsUpper(r) && prev != 0 && (unicode.IsLower(prev) || unicode.IsDigit(prev)) {
			b.WriteByte(' ')
		}
		b.WriteRune(r)
		prev = r
	}
	return strings.Join(strings.Fields(b.String()), " ")
}

func kafkaNormalizeMetricText(text string) string {
	text = strings.TrimSpace(text)
	for _, suffix := range []string{" Ms", " Ns", " Secs", " Seconds Ago"} {
		text = strings.TrimSuffix(text, suffix)
	}
	return strings.TrimSpace(text)
}

func kafkaLowerFirst(text string) string {
	if text == "" {
		return "this Kafka metric"
	}
	rs := []rune(text)
	rs[0] = unicode.ToLower(rs[0])
	return string(rs)
}

func kafkaStringAttributeDesc(lower, attrLower, label, scope string) string {
	switch {
	case attrLower == "eventtype":
		return "Kafka-reported event type label for " + label + ", scoped by " + scope + "."
	case attrLower == "latencyunit":
		return "Kafka-reported latency unit label for " + label + ", scoped by " + scope + "."
	case attrLower == "rateunit":
		return "Kafka-reported rate unit label for " + label + ", scoped by " + scope + "."
	case strings.HasSuffix(lower, "_id"):
		return "Identifier string for " + label + ", scoped by " + scope + "."
	case strings.HasSuffix(lower, "version") || strings.HasSuffix(lower, "_version"):
		return "Version string for " + label + ", scoped by " + scope + "."
	case strings.HasSuffix(lower, "_class"):
		return "Class name for " + label + ", scoped by " + scope + "."
	case strings.HasSuffix(lower, "_type"):
		return "Type label for " + label + ", scoped by " + scope + "."
	case strings.HasSuffix(lower, "status"):
		return "Status label for " + label + ", scoped by " + scope + "."
	default:
		return "String label for " + label + ", scoped by " + scope + "."
	}
}

func kafkaDurationDesc(attrLower, label, unitText, scope string) string {
	switch kafkaSummaryPrefix(attrLower) {
	case "Average":
		return "Average time spent in " + label + ", measured in " + unitText + ", scoped by " + scope + "."
	case "Maximum":
		return "Maximum time spent in " + label + ", measured in " + unitText + ", scoped by " + scope + "."
	case "Minimum":
		return "Minimum time spent in " + label + ", measured in " + unitText + ", scoped by " + scope + "."
	case "Standard deviation":
		return "Standard deviation of time spent in " + label + ", measured in " + unitText + ", scoped by " + scope + "."
	case "50th percentile", "75th percentile", "95th percentile", "98th percentile", "99th percentile", "99.9th percentile":
		return kafkaSummaryPrefix(attrLower) + " of time spent in " + label + ", measured in " + unitText + ", scoped by " + scope + "."
	default:
		return "Time spent in " + label + ", measured in " + unitText + ", scoped by " + scope + "."
	}
}

func kafkaRatioDesc(label, scope string) string {
	return "Fraction from 0 to 1 for " + label + ", scoped by " + scope + "."
}

func kafkaPercentDesc(attrLower, label, scope string) string {
	subject := label
	subject = strings.TrimSuffix(subject, " percentage")
	switch kafkaSummaryPrefix(attrLower) {
	case "Average":
		return "Average percentage for " + subject + " as a value from 0 to 100, scoped by " + scope + "."
	case "Maximum":
		return "Maximum percentage for " + subject + " as a value from 0 to 100, scoped by " + scope + "."
	default:
		return "Percentage for " + subject + " as a value from 0 to 100, scoped by " + scope + "."
	}
}

func kafkaRateDesc(attrLower, label, unitText, scope string) string {
	switch attrLower {
	case "meanrate":
		return "Mean rate for " + label + " in " + unitText + ", scoped by " + scope + "."
	case "oneminuterate":
		return "One-minute moving average rate for " + label + " in " + unitText + ", scoped by " + scope + "."
	case "fiveminuterate":
		return "Five-minute moving average rate for " + label + " in " + unitText + ", scoped by " + scope + "."
	case "fifteenminuterate":
		return "Fifteen-minute moving average rate for " + label + " in " + unitText + ", scoped by " + scope + "."
	default:
		return "Rate for " + label + " in " + unitText + ", scoped by " + scope + "."
	}
}

func kafkaSizeDesc(attrLower, label, scope string) string {
	switch kafkaSummaryPrefix(attrLower) {
	case "Average":
		return "Average size of " + label + " in bytes, scoped by " + scope + "."
	case "Maximum":
		return "Maximum size of " + label + " in bytes, scoped by " + scope + "."
	case "Minimum":
		return "Minimum size of " + label + " in bytes, scoped by " + scope + "."
	case "50th percentile", "75th percentile", "95th percentile", "98th percentile", "99th percentile", "99.9th percentile":
		return kafkaSummaryPrefix(attrLower) + " size of " + label + " in bytes, scoped by " + scope + "."
	default:
		return "Current size of " + label + " in bytes, scoped by " + scope + "."
	}
}

func kafkaCountDesc(label, scope string, bytes bool) string {
	switch {
	case bytes:
		return "Total bytes recorded for " + label + ", scoped by " + scope + "."
	case label == "count":
		return "Total event count, scoped by " + scope + "."
	case strings.HasSuffix(label, " total"):
		return "Total number of " + strings.TrimSuffix(label, " total") + ", scoped by " + scope + "."
	case strings.Contains(label, "count") || strings.Contains(label, "number") || strings.Contains(label, "total"):
		return "Cumulative value of " + label + ", scoped by " + scope + "."
	default:
		return "Total number of " + label + ", scoped by " + scope + "."
	}
}

func kafkaGaugeCountDesc(lower, label, scope string) string {
	switch {
	case strings.Contains(lower, "purgatorysize"):
		return "Current number of requests waiting in " + label + ", scoped by " + scope + "."
	case strings.Contains(lower, "queue"):
		return "Current queue depth for " + label + ", scoped by " + scope + "."
	case label == "count":
		return "Current event count, scoped by " + scope + "."
	case strings.HasSuffix(label, " count"):
		return "Current number of " + strings.TrimSuffix(label, " count") + ", scoped by " + scope + "."
	case strings.Contains(label, "count") || strings.Contains(label, "number"):
		return "Current value of " + label + ", scoped by " + scope + "."
	default:
		return "Current number of " + label + ", scoped by " + scope + "."
	}
}

func kafkaSummaryPrefix(attrLower string) string {
	switch attrLower {
	case "mean", "avg":
		return "Average"
	case "max":
		return "Maximum"
	case "min":
		return "Minimum"
	case "stddev":
		return "Standard deviation"
	case "50thpercentile":
		return "50th percentile"
	case "75thpercentile":
		return "75th percentile"
	case "95thpercentile":
		return "95th percentile"
	case "98thpercentile":
		return "98th percentile"
	case "99thpercentile":
		return "99th percentile"
	case "999thpercentile":
		return "99.9th percentile"
	default:
		return ""
	}
}

func kafkaIsSummaryAttr(attrLower string) bool {
	return kafkaSummaryPrefix(attrLower) != ""
}

func kafkaIsStringAttribute(lower string, info *inputs.FieldInfo) bool {
	return info.DataType == inputs.String || info.Type == inputs.String ||
		strings.HasSuffix(lower, "rateunit") ||
		strings.HasSuffix(lower, "latencyunit") ||
		strings.HasSuffix(lower, "eventtype") ||
		strings.HasSuffix(lower, "_id") ||
		strings.HasSuffix(lower, "version") ||
		strings.HasSuffix(lower, "_version") ||
		strings.HasSuffix(lower, "_type") ||
		strings.HasSuffix(lower, "_class") ||
		strings.HasSuffix(lower, "status")
}

func kafkaIsTimestamp(lower string) bool {
	return lower == "start_time_ms" ||
		strings.Contains(lower, "timestamp") ||
		strings.HasSuffix(lower, "_timestamp")
}

func kafkaIsDurationNS(lower string) bool {
	return strings.Contains(lower, "timens") ||
		strings.Contains(lower, "time_ns") ||
		strings.Contains(lower, "waittime") ||
		strings.Contains(lower, "iotime")
}

func kafkaIsDurationSecond(lower string) bool {
	return strings.Contains(lower, "_secs") ||
		strings.Contains(lower, "seconds_ago") ||
		strings.Contains(lower, "secondsago") ||
		strings.Contains(lower, "metadataage") ||
		strings.Contains(lower, "timesecs") ||
		strings.Contains(lower, "time_secs")
}

func kafkaIsTimerDuration(lower, attrLower string) bool {
	if kafkaIsRateAttr(attrLower) || attrLower == "count" {
		return false
	}
	return strings.Contains(lower, "timems") ||
		strings.Contains(lower, "time_ms") ||
		strings.Contains(lower, "response_time") ||
		strings.Contains(lower, "queue_time") ||
		strings.Contains(lower, "latencyms") ||
		strings.Contains(lower, "latency_") ||
		strings.Contains(lower, "_latency") ||
		strings.Contains(lower, "throttletimems") ||
		strings.Contains(lower, "throttle_time")
}

func kafkaIsByteRate(lower, attrLower string) bool {
	if kafkaIsCountAttr(attrLower) {
		return false
	}
	return (strings.Contains(lower, "bytes") || strings.Contains(lower, "byte")) &&
		(strings.Contains(lower, "persec") || strings.Contains(lower, "_rate") || kafkaIsRateAttr(attrLower))
}

func kafkaIsRequestRate(lower, attrLower string) bool {
	if kafkaIsCountAttr(attrLower) {
		return false
	}
	return strings.Contains(lower, "requestspersec") ||
		(strings.Contains(lower, "request") && (strings.Contains(lower, "_rate") || kafkaIsRateAttr(attrLower)))
}

func kafkaIsByteSize(lower, attrLower string) bool {
	if strings.Contains(lower, "requestbytes") {
		return !kafkaIsCountAttr(attrLower)
	}
	return strings.Contains(lower, "bytes") ||
		strings.Contains(lower, "byte") ||
		strings.Contains(lower, "request_size") ||
		strings.Contains(lower, "record_size") ||
		strings.Contains(lower, "memorypool") ||
		strings.Contains(lower, "batch_size") ||
		strings.HasSuffix(lower, ".size.value")
}

func kafkaByteCountUnit(lower string) string {
	if strings.Contains(lower, "byte") || strings.Contains(lower, "bytes") {
		return inputs.SizeByte
	}
	return inputs.NCount
}

func kafkaIsFraction(lower string) bool {
	return strings.Contains(lower, "ratio") ||
		strings.Contains(lower, "avgidlepercent") ||
		strings.Contains(lower, "idle_percent")
}

func kafkaContainsPercentUnit(lower string) bool {
	return strings.Contains(lower, "_percent") ||
		strings.Contains(lower, "percent_") ||
		strings.HasSuffix(lower, "percent")
}

func kafkaIsPerSecondRate(lower, attrLower string) bool {
	if kafkaIsCountAttr(attrLower) {
		return false
	}
	return strings.Contains(lower, "persec") ||
		strings.Contains(lower, "_rate") ||
		kafkaIsRateAttr(attrLower)
}

func kafkaIsRateAttr(attrLower string) bool {
	return attrLower == "meanrate" ||
		attrLower == "oneminuterate" ||
		attrLower == "fiveminuterate" ||
		attrLower == "fifteenminuterate" ||
		strings.HasSuffix(attrLower, "rate")
}

func kafkaIsCountAttr(attrLower string) bool {
	return attrLower == "count" ||
		strings.HasSuffix(attrLower, "_total")
}
