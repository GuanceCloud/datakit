// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

//nolint:lll,goconst,funlen // Kafka measurement metadata contains intentionally long repeated literals.
package kafka

import (
	"github.com/GuanceCloud/cliutils/point"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs"
)

type KafkaMeasurement struct{}

type KafkaControllerMment struct {
	KafkaMeasurement
}

type KafkaReplicaMment struct {
	KafkaMeasurement
}

type KafkaPurgatoryMment struct {
	KafkaMeasurement
}

type KafkaClientMment struct {
	KafkaMeasurement
}

type KafkaRequestMment struct {
	KafkaMeasurement
}

type KafkaTopicsMment struct {
	KafkaMeasurement
}

type KafkaTopicMment struct {
	KafkaMeasurement
}

type KafkaPartitionMment struct {
	KafkaMeasurement
}

type KafkaZooKeeperMment struct {
	KafkaMeasurement
}

type KafkaRequestHandlerMment struct {
	KafkaMeasurement
}

type KafkaNetworkMment struct {
	KafkaMeasurement
}

type KafkaLogMment struct {
	KafkaMeasurement
}

type KafkaConsumerMment struct {
	KafkaMeasurement
}

type KafkaProducerMment struct {
	KafkaMeasurement
}

type KafkaConnectMment struct {
	KafkaMeasurement
}

// TODO: add more desc & units
//
//	refer to https://github.com/DataDog/integrations-core/blob/master/confluent_platform/metadata.csv
//
//nolint:lll
var connectFields = map[string]interface{}{
	"commit_id": &inputs.FieldInfo{
		DataType: inputs.String, Type: inputs.String, Unit: inputs.NoUnit,
		Desc: "Kafka Connect worker commit id string, scoped by client_id when present.",
	},

	"start_time_ms": &inputs.FieldInfo{
		DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.TimestampMS,
		Desc: "Kafka Connect worker start time as Unix epoch milliseconds, scoped by client_id when present.",
	},

	"version": &inputs.FieldInfo{
		DataType: inputs.String, Type: inputs.String, Unit: inputs.NoUnit,
		Desc: "Kafka Connect worker version string, scoped by client_id when present.",
	},

	"count": &inputs.FieldInfo{
		DataType: inputs.Int, Type: inputs.Count, Unit: inputs.NCount,
		Desc: "Total event count, scoped by client_id, connector, and task tags when present.",
	},

	"connector_count": &inputs.FieldInfo{
		DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.NCount,
		Desc: "Current number of connector, scoped by client_id, connector, and task tags when present.",
	},

	"connector_startup_success_percentage": &inputs.FieldInfo{
		DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.Percent,
		Desc: "Percentage for connector startup success as a value from 0 to 100, scoped by client_id, connector, and task tags when present.",
	},

	"connector_startup_success_total": &inputs.FieldInfo{
		DataType: inputs.Int, Type: inputs.Count, Unit: inputs.NCount,
		Desc: "Total number of connector startup success, scoped by client_id, connector, and task tags when present.",
	},

	"task_startup_failure_percentage": &inputs.FieldInfo{
		DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.Percent,
		Desc: "Percentage for task startup failure as a value from 0 to 100, scoped by client_id, connector, and task tags when present.",
	},

	"task_startup_success_percentage": &inputs.FieldInfo{
		DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.Percent,
		Desc: "Percentage for task startup success as a value from 0 to 100, scoped by client_id, connector, and task tags when present.",
	},

	"task_startup_success_total": &inputs.FieldInfo{
		DataType: inputs.Int, Type: inputs.Count, Unit: inputs.NCount,
		Desc: "Total number of task startup success, scoped by client_id, connector, and task tags when present.",
	},

	"connector_startup_attempts_total": &inputs.FieldInfo{
		DataType: inputs.Int, Type: inputs.Count, Unit: inputs.NCount,
		Desc: "Total number of connector startup attempts, scoped by client_id, connector, and task tags when present.",
	},

	"connector_startup_failure_percentage": &inputs.FieldInfo{
		DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.Percent,
		Desc: "Percentage for connector startup failure as a value from 0 to 100, scoped by client_id, connector, and task tags when present.",
	},

	"connector_startup_failure_total": &inputs.FieldInfo{
		DataType: inputs.Int, Type: inputs.Count, Unit: inputs.NCount,
		Desc: "Total number of connector startup failure, scoped by client_id, connector, and task tags when present.",
	},

	"task_count": &inputs.FieldInfo{
		DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.NCount,
		Desc: "Current number of task, scoped by client_id, connector, and task tags when present.",
	},

	"task_startup_attempts_total": &inputs.FieldInfo{
		DataType: inputs.Int, Type: inputs.Count, Unit: inputs.NCount,
		Desc: "Total number of task startup attempts, scoped by client_id, connector, and task tags when present.",
	},

	"task_startup_failure_total": &inputs.FieldInfo{
		DataType: inputs.Int, Type: inputs.Count, Unit: inputs.NCount,
		Desc: "Total number of task startup failure, scoped by client_id, connector, and task tags when present.",
	},

	"connector_type": &inputs.FieldInfo{
		DataType: inputs.String, Type: inputs.String, Unit: inputs.NoUnit,
		Desc: "Type label for connector type, scoped by client_id, connector, and task tags when present.",
	},

	"connector_unassigned_task_count": &inputs.FieldInfo{
		DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.NCount,
		Desc: "Current number of connector unassigned task, scoped by client_id, connector, and task tags when present.",
	},

	"connector_version": &inputs.FieldInfo{
		DataType: inputs.String, Type: inputs.String, Unit: inputs.NoUnit,
		Desc: "Version string for connector version, scoped by client_id, connector, and task tags when present.",
	},

	"status": &inputs.FieldInfo{
		DataType: inputs.String, Type: inputs.String, Unit: inputs.NoUnit,
		Desc: "Status label for status, scoped by client_id, connector, and task tags when present.",
	},

	"connector_class": &inputs.FieldInfo{
		DataType: inputs.String, Type: inputs.String, Unit: inputs.NoUnit,
		Desc: "Class name for connector class, scoped by client_id, connector, and task tags when present.",
	},

	"connector_failed_task_count": &inputs.FieldInfo{
		DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.NCount,
		Desc: "Current number of connector failed task, scoped by client_id, connector, and task tags when present.",
	},

	"connector_paused_task_count": &inputs.FieldInfo{
		DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.NCount,
		Desc: "Current number of connector paused task, scoped by client_id, connector, and task tags when present.",
	},

	"connector_total_task_count": &inputs.FieldInfo{
		DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.NCount,
		Desc: "Current number of connector total task, scoped by client_id, connector, and task tags when present.",
	},

	"connector_destroyed_task_count": &inputs.FieldInfo{
		DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.NCount,
		Desc: "Current number of connector destroyed task, scoped by client_id, connector, and task tags when present.",
	},

	"connector_restarting_task_count": &inputs.FieldInfo{
		DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.NCount,
		Desc: "Current number of connector restarting task, scoped by client_id, connector, and task tags when present.",
	},

	"connector_running_task_count": &inputs.FieldInfo{
		DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.NCount,
		Desc: "Current number of connector running task, scoped by client_id, connector, and task tags when present.",
	},

	"total_records_skipped": &inputs.FieldInfo{
		DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.NCount,
		Desc: "Current number of total records skipped, scoped by client_id, connector, and task tags when present.",
	},

	"total_retries": &inputs.FieldInfo{
		DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.NCount,
		Desc: "Current number of total retries, scoped by client_id, connector, and task tags when present.",
	},

	"offset_commit_failure_percentage": &inputs.FieldInfo{
		DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.Percent,
		Desc: "Percentage for offset commit failure as a value from 0 to 100, scoped by client_id, connector, and task tags when present.",
	},

	"running_ratio": &inputs.FieldInfo{
		DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.PercentDecimal,
		Desc: "Fraction from 0 to 1 of time the Kafka Connect task is running, scoped by connector and task.",
	},

	"source_record_poll_total": &inputs.FieldInfo{
		DataType: inputs.Int, Type: inputs.Count, Unit: inputs.NCount,
		Desc: "Total number of source record poll, scoped by client_id, connector, and task tags when present.",
	},

	"total_record_failures": &inputs.FieldInfo{
		DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.NCount,
		Desc: "Current value of total record failures, scoped by client_id, connector, and task tags when present.",
	},

	"last_error_timestamp": &inputs.FieldInfo{
		DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.TimestampMS,
		Desc: "Unix epoch timestamp in milliseconds for last error timestamp, scoped by client_id, connector, and task tags when present.",
	},

	"offset_commit_success_percentage": &inputs.FieldInfo{
		DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.Percent,
		Desc: "Percentage for offset commit success as a value from 0 to 100, scoped by client_id, connector, and task tags when present.",
	},

	"source_record_poll_rate": &inputs.FieldInfo{
		DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount,
		Desc: "Number of source records polled per second by this Kafka Connect source task, scoped by connector and task.",
	},

	"source_record_write_rate": &inputs.FieldInfo{
		DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount,
		Desc: "Rate for source record write rate in events per second, scoped by client_id, connector, and task tags when present.",
	},

	"total_errors_logged": &inputs.FieldInfo{
		DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.NCount,
		Desc: "Current number of total errors logged, scoped by client_id, connector, and task tags when present.",
	},

	"offset_commit_max_time_ms": &inputs.FieldInfo{
		DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS,
		Desc: "Time spent in offset commit max time ms, measured in milliseconds, scoped by client_id, connector, and task tags when present.",
	},

	"source_record_active_count": &inputs.FieldInfo{
		DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.NCount,
		Desc: "Current number of source record active, scoped by client_id, connector, and task tags when present.",
	},

	"source_record_write_total": &inputs.FieldInfo{
		DataType: inputs.Int, Type: inputs.Count, Unit: inputs.NCount,
		Desc: "Total number of source record write, scoped by client_id, connector, and task tags when present.",
	},

	"total_record_errors": &inputs.FieldInfo{
		DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.NCount,
		Desc: "Current number of total record errors, scoped by client_id, connector, and task tags when present.",
	},

	"deadletterqueue_produce_failures": &inputs.FieldInfo{
		DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.NCount,
		Desc: "Current queue depth for deadletterqueue produce failures, scoped by client_id, connector, and task tags when present.",
	},

	"deadletterqueue_produce_requests": &inputs.FieldInfo{
		DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.NCount,
		Desc: "Current queue depth for deadletterqueue produce requests, scoped by client_id, connector, and task tags when present.",
	},

	"offset_commit_avg_time_ms": &inputs.FieldInfo{
		DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS,
		Desc: "Time spent in offset commit avg time ms, measured in milliseconds, scoped by client_id, connector, and task tags when present.",
	},

	"pause_ratio": &inputs.FieldInfo{
		DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.PercentDecimal,
		Desc: "Fraction from 0 to 1 for pause ratio, scoped by client_id, connector, and task tags when present.",
	},

	"offset_commit_skip_rate": &inputs.FieldInfo{
		DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount,
		Desc: "Rate for offset commit skip rate in events per second, scoped by client_id, connector, and task tags when present.",
	},

	"sink_record_send_total": &inputs.FieldInfo{
		DataType: inputs.Int, Type: inputs.Count, Unit: inputs.NCount,
		Desc: "Total number of sink record send, scoped by client_id, connector, and task tags when present.",
	},

	"offset_commit_seq_no": &inputs.FieldInfo{
		DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.NCount,
		Desc: "Current offset for offset commit seq no, scoped by client_id, connector, and task tags when present.",
	},

	"put_batch_avg_time_ms": &inputs.FieldInfo{
		DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS,
		Desc: "Time spent in put batch avg time ms, measured in milliseconds, scoped by client_id, connector, and task tags when present.",
	},

	"sink_record_send_rate": &inputs.FieldInfo{
		DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount,
		Desc: "Number of sink records sent per second by this Kafka Connect sink task, scoped by connector and task.",
	},

	"batch_size_avg": &inputs.FieldInfo{
		DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.SizeByte,
		Desc: "Current size of batch size avg in bytes, scoped by client_id, connector, and task tags when present.",
	},

	"offset_commit_completion_rate": &inputs.FieldInfo{
		DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount,
		Desc: "Rate for offset commit completion rate in events per second, scoped by client_id, connector, and task tags when present.",
	},

	"offset_commit_skip_total": &inputs.FieldInfo{
		DataType: inputs.Int, Type: inputs.Count, Unit: inputs.NCount,
		Desc: "Total number of offset commit skip, scoped by client_id, connector, and task tags when present.",
	},

	"sink_record_read_rate": &inputs.FieldInfo{
		DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount,
		Desc: "Rate for sink record read rate in events per second, scoped by client_id, connector, and task tags when present.",
	},

	"batch_size_max": &inputs.FieldInfo{
		DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.SizeByte,
		Desc: "Current size of batch size max in bytes, scoped by client_id, connector, and task tags when present.",
	},

	"partition_count": &inputs.FieldInfo{
		DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.NCount,
		Desc: "Current number of partition, scoped by client_id, connector, and task tags when present.",
	},

	"sink_record_active_count": &inputs.FieldInfo{
		DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.NCount,
		Desc: "Current number of sink record active, scoped by client_id, connector, and task tags when present.",
	},

	"sink_record_read_total": &inputs.FieldInfo{
		DataType: inputs.Int, Type: inputs.Count, Unit: inputs.NCount,
		Desc: "Total number of sink record read, scoped by client_id, connector, and task tags when present.",
	},

	"put_batch_max_time_ms": &inputs.FieldInfo{
		DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS,
		Desc: "Time spent in put batch max time ms, measured in milliseconds, scoped by client_id, connector, and task tags when present.",
	},

	"offset_commit_completion_total": &inputs.FieldInfo{
		DataType: inputs.Int, Type: inputs.Count, Unit: inputs.NCount,
		Desc: "Total number of offset commit completion, scoped by client_id, connector, and task tags when present.",
	},

	"sink_record_active_count_avg": &inputs.FieldInfo{
		DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.NCount,
		Desc: "Current value of sink record active count avg, scoped by client_id, connector, and task tags when present.",
	},

	"sink_record_active_count_max": &inputs.FieldInfo{
		DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.NCount,
		Desc: "Current value of sink record active count max, scoped by client_id, connector, and task tags when present.",
	},
}

var connectTags = map[string]interface{}{
	"jolokia_agent_url": &inputs.TagInfo{Desc: "Jolokia agent url path"},
	"type":              &inputs.TagInfo{Desc: "metric type"},
	"client_id":         &inputs.TagInfo{Desc: "client id"},
	"task":              &inputs.TagInfo{Desc: "task"},
	"connector":         &inputs.TagInfo{Desc: "connector"},
}

func (j *KafkaConnectMment) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name:   "kafka_connect",
		Cat:    point.Metric,
		Fields: connectFields,
		Tags:   connectTags,
		Desc:   "This metrics needs to be collected on the Connect instance",
		DescZh: "Kafka Connect 实例的连接器、任务、sink/source 记录和请求相关 JMX 指标。",
	}
}

//nolint:lll
var producerFields = map[string]interface{}{
	"io_wait_ratio": &inputs.FieldInfo{
		DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.PercentDecimal,
		Desc: "Fraction from 0 to 1 for io wait ratio, scoped by client_id.",
	},

	"io_waittime_total": &inputs.FieldInfo{
		DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.DurationNS,
		Desc: "Time spent in io waittime total, measured in nanoseconds, scoped by client_id.",
	},

	"record_send_total": &inputs.FieldInfo{
		DataType: inputs.Int, Type: inputs.Count, Unit: inputs.NCount,
		Desc: "Total number of records sent by this producer, scoped by client_id.",
	},

	"io_time_ns_avg": &inputs.FieldInfo{
		DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationNS,
		Desc: "Time spent in io time ns avg, measured in nanoseconds, scoped by client_id.",
	},

	"buffer_exhausted_rate": &inputs.FieldInfo{
		DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount,
		Desc: "Rate for buffer exhausted rate in events per second, scoped by client_id.",
	},

	"buffer_total_bytes": &inputs.FieldInfo{
		DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.SizeByte,
		Desc: "Total producer buffer memory in bytes available to this producer, scoped by client_id.",
	},

	"incoming_byte_total": &inputs.FieldInfo{
		DataType: inputs.Int, Type: inputs.Count, Unit: inputs.SizeByte,
		Desc: "Total bytes recorded for incoming byte total, scoped by client_id.",
	},

	"successful_reauthentication_total": &inputs.FieldInfo{
		DataType: inputs.Int, Type: inputs.Count, Unit: inputs.NCount,
		Desc: "Total number of successful reauthentication, scoped by client_id.",
	},

	"record_retry_total": &inputs.FieldInfo{
		DataType: inputs.Int, Type: inputs.Count, Unit: inputs.NCount,
		Desc: "Total number of record retry, scoped by client_id.",
	},

	"buffer_exhausted_total": &inputs.FieldInfo{
		DataType: inputs.Float, Type: inputs.Count, Unit: inputs.NCount,
		Desc: "Total number of buffer exhausted, scoped by client_id.",
	},

	"failed_reauthentication_rate": &inputs.FieldInfo{
		DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount,
		Desc: "Rate for failed reauthentication rate in events per second, scoped by client_id.",
	},

	"request_rate": &inputs.FieldInfo{
		DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.RequestsPerSec,
		Desc: "Rate for request rate in requests per second, scoped by client_id.",
	},

	"io_ratio": &inputs.FieldInfo{
		DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.PercentDecimal,
		Desc: "Fraction from 0 to 1 for io ratio, scoped by client_id.",
	},

	"io_wait_time_ns_avg": &inputs.FieldInfo{
		DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.DurationNS,
		Desc: "Time spent in io wait time ns avg, measured in nanoseconds, scoped by client_id.",
	},

	"metadata_age": &inputs.FieldInfo{
		DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount,
		Desc: "Current value of metadata age, scoped by client_id.",
	},

	"network_io_rate": &inputs.FieldInfo{
		DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount,
		Desc: "Rate for network io rate in events per second, scoped by client_id.",
	},

	"connection_close_rate": &inputs.FieldInfo{
		DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount,
		Desc: "Rate for connection close rate in events per second, scoped by client_id.",
	},

	"connection_count": &inputs.FieldInfo{
		DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.NCount,
		Desc: "Current number of connection, scoped by client_id.",
	},

	"produce_throttle_time_max": &inputs.FieldInfo{
		DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.DurationMS,
		Desc: "Time spent in produce throttle time max, measured in milliseconds, scoped by client_id.",
	},

	"successful_authentication_total": &inputs.FieldInfo{
		DataType: inputs.Int, Type: inputs.Count, Unit: inputs.NCount,
		Desc: "Total number of successful authentication, scoped by client_id.",
	},

	"request_total": &inputs.FieldInfo{
		DataType: inputs.Int, Type: inputs.Count, Unit: inputs.NCount,
		Desc: "Total number of request, scoped by client_id.",
	},

	"successful_authentication_no_reauth_total": &inputs.FieldInfo{
		DataType: inputs.Int, Type: inputs.Count, Unit: inputs.NCount,
		Desc: "Total number of successful authentication no reauth, scoped by client_id.",
	},

	"bufferpool_wait_ratio": &inputs.FieldInfo{
		DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.PercentDecimal,
		Desc: "Fraction from 0 to 1 for bufferpool wait ratio, scoped by client_id.",
	},

	"incoming_byte_rate": &inputs.FieldInfo{
		DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.BytesPerSec,
		Desc: "Number of bytes received per second by this producer client, scoped by client_id.",
	},

	"request_size_avg": &inputs.FieldInfo{
		DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.SizeByte,
		Desc: "Current size of request size avg in bytes, scoped by client_id.",
	},

	"select_total": &inputs.FieldInfo{
		DataType: inputs.Int, Type: inputs.Count, Unit: inputs.NCount,
		Desc: "Total number of select, scoped by client_id.",
	},

	"connection_close_total": &inputs.FieldInfo{
		DataType: inputs.Int, Type: inputs.Count, Unit: inputs.NCount,
		Desc: "Total number of connection close, scoped by client_id.",
	},

	"network_io_total": &inputs.FieldInfo{
		DataType: inputs.Int, Type: inputs.Count, Unit: inputs.NCount,
		Desc: "Total number of network io, scoped by client_id.",
	},

	"response_rate": &inputs.FieldInfo{
		DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount,
		Desc: "Rate for response rate in events per second, scoped by client_id.",
	},

	"select_rate": &inputs.FieldInfo{
		DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount,
		Desc: "Rate for select rate in events per second, scoped by client_id.",
	},

	"record_send_rate": &inputs.FieldInfo{
		DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount,
		Desc: "Number of records sent per second by this producer, scoped by client_id.",
	},

	"successful_authentication_rate": &inputs.FieldInfo{
		DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount,
		Desc: "Rate for successful authentication rate in events per second, scoped by client_id.",
	},

	"record_error_rate": &inputs.FieldInfo{
		DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount,
		Desc: "Rate for record error rate in events per second, scoped by client_id.",
	},

	"response_total": &inputs.FieldInfo{
		DataType: inputs.Int, Type: inputs.Count, Unit: inputs.NCount,
		Desc: "Total number of response, scoped by client_id.",
	},

	"produce_throttle_time_avg": &inputs.FieldInfo{
		DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.DurationMS,
		Desc: "Time spent in produce throttle time avg, measured in milliseconds, scoped by client_id.",
	},

	"record_error_total": &inputs.FieldInfo{
		DataType: inputs.Int, Type: inputs.Count, Unit: inputs.NCount,
		Desc: "Total number of record error, scoped by client_id.",
	},

	"requests_in_flight": &inputs.FieldInfo{
		DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.NCount,
		Desc: "Current value of requests in flight, scoped by client_id.",
	},

	"failed_authentication_rate": &inputs.FieldInfo{
		DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount,
		Desc: "Rate for failed authentication rate in events per second, scoped by client_id.",
	},

	"buffer_available_bytes": &inputs.FieldInfo{
		DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.SizeByte,
		Desc: "Current size of buffer available bytes in bytes, scoped by client_id.",
	},

	"iotime_total": &inputs.FieldInfo{
		DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.DurationNS,
		Desc: "Time spent in iotime total, measured in nanoseconds, scoped by client_id.",
	},

	"successful_reauthentication_rate": &inputs.FieldInfo{
		DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount,
		Desc: "Rate for successful reauthentication rate in events per second, scoped by client_id.",
	},

	"failed_authentication_total": &inputs.FieldInfo{
		DataType: inputs.Int, Type: inputs.Count, Unit: inputs.NCount,
		Desc: "Total number of failed authentication, scoped by client_id.",
	},

	"failed_reauthentication_total": &inputs.FieldInfo{
		DataType: inputs.Int, Type: inputs.Count, Unit: inputs.NCount,
		Desc: "Total number of failed reauthentication, scoped by client_id.",
	},

	"record_retry_rate": &inputs.FieldInfo{
		DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount,
		Desc: "Rate for record retry rate in events per second, scoped by client_id.",
	},

	"batch_split_total": &inputs.FieldInfo{
		DataType: inputs.Int, Type: inputs.Count, Unit: inputs.NCount,
		Desc: "Total number of batch split, scoped by client_id.",
	},

	"connection_creation_total": &inputs.FieldInfo{
		DataType: inputs.Int, Type: inputs.Count, Unit: inputs.NCount,
		Desc: "Total number of connection creation, scoped by client_id.",
	},

	"request_size_max": &inputs.FieldInfo{
		DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.SizeByte,
		Desc: "Current size of request size max in bytes, scoped by client_id.",
	},

	"bufferpool_wait_time_total": &inputs.FieldInfo{
		DataType: inputs.Int, Type: inputs.Count, Unit: inputs.NCount,
		Desc: "Total number of bufferpool wait time, scoped by client_id.",
	},

	"waiting_threads": &inputs.FieldInfo{
		DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.NCount,
		Desc: "Current value of waiting threads, scoped by client_id.",
	},

	"batch_split_rate": &inputs.FieldInfo{
		DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount,
		Desc: "Rate for batch split rate in events per second, scoped by client_id.",
	},

	"connection_creation_rate": &inputs.FieldInfo{
		DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount,
		Desc: "Rate for connection creation rate in events per second, scoped by client_id.",
	},

	"outgoing_byte_rate": &inputs.FieldInfo{
		DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.BytesPerSec,
		Desc: "Rate for outgoing byte rate in bytes per second, scoped by client_id.",
	},

	"outgoing_byte_total": &inputs.FieldInfo{
		DataType: inputs.Int, Type: inputs.Count, Unit: inputs.SizeByte,
		Desc: "Total bytes recorded for outgoing byte total, scoped by client_id.",
	},

	"count": &inputs.FieldInfo{
		DataType: inputs.Int, Type: inputs.Count, Unit: inputs.NCount,
		Desc: "Total event count, scoped by client_id.",
	},

	"commit_id": &inputs.FieldInfo{
		DataType: inputs.String, Type: inputs.String, Unit: inputs.NoUnit,
		Desc: "Identifier string for commit id, scoped by client_id.",
	},

	"start_time_ms": &inputs.FieldInfo{
		DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.TimestampMS,
		Desc: "Unix epoch timestamp in milliseconds for start time ms, scoped by client_id.",
	},

	"version": &inputs.FieldInfo{
		DataType: inputs.String, Type: inputs.String, Unit: inputs.NoUnit,
		Desc: "Version string for version, scoped by client_id.",
	},
}

var producerTags = map[string]interface{}{
	"jolokia_agent_url": &inputs.TagInfo{Desc: "Jolokia agent url path"},
	"type":              &inputs.TagInfo{Desc: "metric type"},
	"client_id":         &inputs.TagInfo{Desc: "client id"},
}

func (j *KafkaProducerMment) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name:   "kafka_producer",
		Cat:    point.Metric,
		Fields: producerFields,
		Tags:   producerTags,
		Desc:   "This metrics needs to be collected on the Producer instance",
		DescZh: "Kafka producer 客户端 JMX 指标，包含缓冲区、请求、记录、字节、IO、压缩和重新认证统计。",
	}
}

//nolint:lll
var consumerFields = map[string]interface{}{
	"count": &inputs.FieldInfo{
		DataType: inputs.Int, Type: inputs.Count, Unit: inputs.NCount,
		Desc: "Total event count, scoped by client_id.",
	},

	"bytes_consumed_rate": &inputs.FieldInfo{
		DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.BytesPerSec,
		Desc: "Number of bytes consumed per second by this consumer, scoped by client_id.",
	},

	"bytes_consumed_total": &inputs.FieldInfo{
		DataType: inputs.Int, Type: inputs.Count, Unit: inputs.SizeByte,
		Desc: "Total bytes recorded for bytes consumed total, scoped by client_id.",
	},

	"fetch_latency_avg": &inputs.FieldInfo{
		DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS,
		Desc: "Average consumer fetch request latency in milliseconds, scoped by client_id.",
	},

	"fetch_throttle_time_avg": &inputs.FieldInfo{
		DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS,
		Desc: "Time spent in fetch throttle time avg, measured in milliseconds, scoped by client_id.",
	},

	"fetch_total": &inputs.FieldInfo{
		DataType: inputs.Float, Type: inputs.Count, Unit: inputs.NCount,
		Desc: "Total number of fetch, scoped by client_id.",
	},

	"records_consumed_rate": &inputs.FieldInfo{
		DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount,
		Desc: "Number of records consumed per second by this consumer, scoped by client_id.",
	},

	"fetch_latency_max": &inputs.FieldInfo{
		DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS,
		Desc: "Time spent in fetch latency max, measured in milliseconds, scoped by client_id.",
	},

	"fetch_rate": &inputs.FieldInfo{
		DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.RequestsPerSec,
		Desc: "Number of consumer fetch requests per second, scoped by client_id.",
	},

	"fetch_throttle_time_max": &inputs.FieldInfo{
		DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS,
		Desc: "Time spent in fetch throttle time max, measured in milliseconds, scoped by client_id.",
	},

	"records_consumed_total": &inputs.FieldInfo{
		DataType: inputs.Float, Type: inputs.Count, Unit: inputs.NCount,
		Desc: "Total number of records consumed, scoped by client_id.",
	},

	"commit_total": &inputs.FieldInfo{
		DataType: inputs.Int, Type: inputs.Count, Unit: inputs.NCount,
		Desc: "Total number of commit, scoped by client_id.",
	},

	"join_rate": &inputs.FieldInfo{
		DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount,
		Desc: "Rate for join rate in events per second, scoped by client_id.",
	},

	"join_total": &inputs.FieldInfo{
		DataType: inputs.Int, Type: inputs.Count, Unit: inputs.NCount,
		Desc: "Total number of join, scoped by client_id.",
	},

	"rebalance_rate_per_hour": &inputs.FieldInfo{
		DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount,
		Desc: "Rate for rebalance rate per hour in events per second, scoped by client_id.",
	},

	"rebalance_total": &inputs.FieldInfo{
		DataType: inputs.Int, Type: inputs.Count, Unit: inputs.NCount,
		Desc: "Total number of rebalance, scoped by client_id.",
	},

	"failed_rebalance_rate_per_hour": &inputs.FieldInfo{
		DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount,
		Desc: "Rate for failed rebalance rate per hour in events per second, scoped by client_id.",
	},

	"heartbeat_rate": &inputs.FieldInfo{
		DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount,
		Desc: "Rate for heartbeat rate in events per second, scoped by client_id.",
	},

	"last_rebalance_seconds_ago": &inputs.FieldInfo{
		DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationSecond,
		Desc: "Time spent in last rebalance seconds ago, measured in seconds, scoped by client_id.",
	},

	"sync_rate": &inputs.FieldInfo{
		DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount,
		Desc: "Rate for sync rate in events per second, scoped by client_id.",
	},

	"sync_total": &inputs.FieldInfo{
		DataType: inputs.Int, Type: inputs.Count, Unit: inputs.NCount,
		Desc: "Total number of sync, scoped by client_id.",
	},

	"assigned_partitions": &inputs.FieldInfo{
		DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.NCount,
		Desc: "Current number of assigned partitions, scoped by client_id.",
	},

	"commit_rate": &inputs.FieldInfo{
		DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount,
		Desc: "Rate for commit rate in events per second, scoped by client_id.",
	},

	"heartbeat_response_time_max": &inputs.FieldInfo{
		DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS,
		Desc: "Time spent in heartbeat response time max, measured in milliseconds, scoped by client_id.",
	},

	"last_heartbeat_seconds_ago": &inputs.FieldInfo{
		DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationSecond,
		Desc: "Time spent in last heartbeat seconds ago, measured in seconds, scoped by client_id.",
	},

	"failed_rebalance_total": &inputs.FieldInfo{
		DataType: inputs.Float, Type: inputs.Count, Unit: inputs.NCount,
		Desc: "Total number of failed rebalance, scoped by client_id.",
	},

	"heartbeat_total": &inputs.FieldInfo{
		DataType: inputs.Int, Type: inputs.Count, Unit: inputs.NCount,
		Desc: "Total number of heartbeat, scoped by client_id.",
	},

	"rebalance_latency_total": &inputs.FieldInfo{
		DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.DurationMS,
		Desc: "Time spent in rebalance latency total, measured in milliseconds, scoped by client_id.",
	},

	"commit_id": &inputs.FieldInfo{
		DataType: inputs.String, Type: inputs.String, Unit: inputs.NoUnit,
		Desc: "Identifier string for commit id, scoped by client_id.",
	},

	"start_time_ms": &inputs.FieldInfo{
		DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.TimestampMS,
		Desc: "Unix epoch timestamp in milliseconds for start time ms, scoped by client_id.",
	},

	"version": &inputs.FieldInfo{
		DataType: inputs.String, Type: inputs.String, Unit: inputs.NoUnit,
		Desc: "Version string for version, scoped by client_id.",
	},

	"io_waittime_total": &inputs.FieldInfo{
		DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.DurationNS,
		Desc: "Time spent in io waittime total, measured in nanoseconds, scoped by client_id.",
	},

	"connection_creation_rate": &inputs.FieldInfo{
		DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount,
		Desc: "Rate for connection creation rate in events per second, scoped by client_id.",
	},

	"connection_creation_total": &inputs.FieldInfo{
		DataType: inputs.Int, Type: inputs.Count, Unit: inputs.NCount,
		Desc: "Total number of connection creation, scoped by client_id.",
	},

	"successful_authentication_rate": &inputs.FieldInfo{
		DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount,
		Desc: "Rate for successful authentication rate in events per second, scoped by client_id.",
	},

	"failed_authentication_rate": &inputs.FieldInfo{
		DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount,
		Desc: "Rate for failed authentication rate in events per second, scoped by client_id.",
	},

	"failed_authentication_total": &inputs.FieldInfo{
		DataType: inputs.Int, Type: inputs.Count, Unit: inputs.NCount,
		Desc: "Total number of failed authentication, scoped by client_id.",
	},

	"failed_reauthentication_total": &inputs.FieldInfo{
		DataType: inputs.Int, Type: inputs.Count, Unit: inputs.NCount,
		Desc: "Total number of failed reauthentication, scoped by client_id.",
	},

	"io_wait_time_ns_avg": &inputs.FieldInfo{
		DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.DurationNS,
		Desc: "Time spent in io wait time ns avg, measured in nanoseconds, scoped by client_id.",
	},

	"network_io_total": &inputs.FieldInfo{
		DataType: inputs.Int, Type: inputs.Count, Unit: inputs.NCount,
		Desc: "Total number of network io, scoped by client_id.",
	},

	"request_size_avg": &inputs.FieldInfo{
		DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.SizeByte,
		Desc: "Current size of request size avg in bytes, scoped by client_id.",
	},

	"successful_authentication_no_reauth_total": &inputs.FieldInfo{
		DataType: inputs.Int, Type: inputs.Count, Unit: inputs.NCount,
		Desc: "Total number of successful authentication no reauth, scoped by client_id.",
	},

	"connection_close_rate": &inputs.FieldInfo{
		DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount,
		Desc: "Rate for connection close rate in events per second, scoped by client_id.",
	},

	"io_time_ns_avg": &inputs.FieldInfo{
		DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.DurationNS,
		Desc: "Time spent in io time ns avg, measured in nanoseconds, scoped by client_id.",
	},

	"outgoing_byte_total": &inputs.FieldInfo{
		DataType: inputs.Float, Type: inputs.Count, Unit: inputs.SizeByte,
		Desc: "Total bytes recorded for outgoing byte total, scoped by client_id.",
	},

	"request_size_max": &inputs.FieldInfo{
		DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.SizeByte,
		Desc: "Current size of request size max in bytes, scoped by client_id.",
	},

	"response_rate": &inputs.FieldInfo{
		DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount,
		Desc: "Rate for response rate in events per second, scoped by client_id.",
	},

	"successful_authentication_total": &inputs.FieldInfo{
		DataType: inputs.Int, Type: inputs.Count, Unit: inputs.NCount,
		Desc: "Total number of successful authentication, scoped by client_id.",
	},

	"successful_reauthentication_total": &inputs.FieldInfo{
		DataType: inputs.Int, Type: inputs.Count, Unit: inputs.NCount,
		Desc: "Total number of successful reauthentication, scoped by client_id.",
	},

	"connection_count": &inputs.FieldInfo{
		DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.NCount,
		Desc: "Current number of connection, scoped by client_id.",
	},

	"io_ratio": &inputs.FieldInfo{
		DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.PercentDecimal,
		Desc: "Fraction from 0 to 1 for io ratio, scoped by client_id.",
	},

	"request_total": &inputs.FieldInfo{
		DataType: inputs.Int, Type: inputs.Count, Unit: inputs.NCount,
		Desc: "Total number of request, scoped by client_id.",
	},

	"response_total": &inputs.FieldInfo{
		DataType: inputs.Int, Type: inputs.Count, Unit: inputs.NCount,
		Desc: "Total number of response, scoped by client_id.",
	},

	"incoming_byte_total": &inputs.FieldInfo{
		DataType: inputs.Float, Type: inputs.Count, Unit: inputs.SizeByte,
		Desc: "Total bytes recorded for incoming byte total, scoped by client_id.",
	},

	"iotime_total": &inputs.FieldInfo{
		DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.DurationNS,
		Desc: "Time spent in iotime total, measured in nanoseconds, scoped by client_id.",
	},

	"last_poll_seconds_ago": &inputs.FieldInfo{
		DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.DurationSecond,
		Desc: "Time spent in last poll seconds ago, measured in seconds, scoped by client_id.",
	},

	"outgoing_byte_rate": &inputs.FieldInfo{
		DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.BytesPerSec,
		Desc: "Rate for outgoing byte rate in bytes per second, scoped by client_id.",
	},

	"connection_close_total": &inputs.FieldInfo{
		DataType: inputs.Int, Type: inputs.Count, Unit: inputs.NCount,
		Desc: "Total number of connection close, scoped by client_id.",
	},

	"failed_reauthentication_rate": &inputs.FieldInfo{
		DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount,
		Desc: "Rate for failed reauthentication rate in events per second, scoped by client_id.",
	},

	"incoming_byte_rate": &inputs.FieldInfo{
		DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.BytesPerSec,
		Desc: "Rate for incoming byte rate in bytes per second, scoped by client_id.",
	},

	"network_io_rate": &inputs.FieldInfo{
		DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount,
		Desc: "Rate for network io rate in events per second, scoped by client_id.",
	},

	"select_rate": &inputs.FieldInfo{
		DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount,
		Desc: "Rate for select rate in events per second, scoped by client_id.",
	},

	"io_wait_ratio": &inputs.FieldInfo{
		DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.PercentDecimal,
		Desc: "Fraction from 0 to 1 for io wait ratio, scoped by client_id.",
	},

	"request_rate": &inputs.FieldInfo{
		DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.RequestsPerSec,
		Desc: "Rate for request rate in requests per second, scoped by client_id.",
	},

	"select_total": &inputs.FieldInfo{
		DataType: inputs.Int, Type: inputs.Count, Unit: inputs.NCount,
		Desc: "Total number of select, scoped by client_id.",
	},

	"successful_reauthentication_rate": &inputs.FieldInfo{
		DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount,
		Desc: "Rate for successful reauthentication rate in events per second, scoped by client_id.",
	},
}

var consumerTags = map[string]interface{}{
	"jolokia_agent_url": &inputs.TagInfo{Desc: "Jolokia agent url path"},
	"type":              &inputs.TagInfo{Desc: "metric type"},
	"client_id":         &inputs.TagInfo{Desc: "client id"},
}

func (j *KafkaConsumerMment) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name:   "kafka_consumer",
		Cat:    point.Metric,
		Fields: consumerFields,
		Tags:   consumerTags,
		Desc:   "This metrics needs to be collected on the Consumer instance",
		DescZh: "Kafka consumer 客户端 JMX 指标，包含 fetch、记录、字节、延迟、提交、poll、心跳和协调器统计。",
	}
}

//nolint:lll
var logFields = map[string]interface{}{
	"OfflineLogDirectoryCount": &inputs.FieldInfo{
		DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.NCount,
		Desc: "Current number of offline Log Directory Count, scoped by host and Jolokia agent.",
	},

	"cleaner_recopy_percent": &inputs.FieldInfo{
		DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.Percent,
		Desc: "Percentage for cleaner recopy percent as a value from 0 to 100, scoped by host and Jolokia agent.",
	},

	"max_compaction_delay_secs": &inputs.FieldInfo{
		DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.DurationSecond,
		Desc: "Time spent in max compaction delay secs, measured in seconds, scoped by host and Jolokia agent.",
	},

	"max_clean_time_secs": &inputs.FieldInfo{
		DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.DurationSecond,
		Desc: "Time spent in max clean time secs, measured in seconds, scoped by host and Jolokia agent.",
	},

	"DeadThreadCount": &inputs.FieldInfo{
		DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.NCount,
		Desc: "Current number of dead Thread Count, scoped by host and Jolokia agent.",
	},

	"max_buffer_utilization_percent": &inputs.FieldInfo{
		DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.Percent,
		Desc: "Percentage for max buffer utilization percent as a value from 0 to 100, scoped by host and Jolokia agent.",
	},
}

var logTags = map[string]interface{}{
	"jolokia_agent_url": &inputs.TagInfo{Desc: "Jolokia agent url path"},
	"type":              &inputs.TagInfo{Desc: "metric type"},
}

func (j *KafkaLogMment) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name:   "kafka_log",
		Cat:    point.Metric,
		Desc:   "Kafka broker log manager JMX metrics, including log size, flush, recovery, cleaner, and log segment state.",
		DescZh: "Kafka broker 日志管理 JMX 指标，包含日志大小、flush、恢复、清理器和日志段状态。",
		Fields: logFields,
		Tags:   logTags,
	}
}

//nolint:lll
var networkFields = map[string]interface{}{
	"NetworkProcessorAvgIdlePercent": &inputs.FieldInfo{
		DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.PercentDecimal,
		Desc: "Fraction from 0 to 1 for network Processor Avg Idle Percent, scoped by host and Jolokia agent.",
	},

	"ExpiredConnectionsKilledCount": &inputs.FieldInfo{
		DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.NCount,
		Desc: "Current number of expired Connections Killed Count, scoped by host and Jolokia agent.",
	},

	"ControlPlaneExpiredConnectionsKilledCount": &inputs.FieldInfo{
		DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.NCount,
		Desc: "Current number of control Plane Expired Connections Killed Count, scoped by host and Jolokia agent.",
	},

	"MemoryPoolAvailable": &inputs.FieldInfo{
		DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.SizeByte,
		Desc: "Current size of memory Pool Available in bytes, scoped by host and Jolokia agent.",
	},

	"MemoryPoolUsed": &inputs.FieldInfo{
		DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.SizeByte,
		Desc: "Current size of memory Pool Used in bytes, scoped by host and Jolokia agent.",
	},
}

var networkTags = map[string]interface{}{
	"jolokia_agent_url": &inputs.TagInfo{Desc: "Jolokia agent url path"},
	"type":              &inputs.TagInfo{Desc: "metric type"},
}

func (j *KafkaNetworkMment) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name:   "kafka_network",
		Cat:    point.Metric,
		Desc:   "Kafka broker network processor JMX metrics, including idle percent, expired connections, and memory pool usage.",
		DescZh: "Kafka broker 网络处理器 JMX 指标，包含空闲比例、过期连接和内存池使用情况。",
		Fields: networkFields,
		Tags:   networkTags,
	}
}

//nolint:lll
var requestHandlerFields = map[string]interface{}{
	"RequestHandlerAvgIdlePercent.RateUnit": &inputs.FieldInfo{
		DataType: inputs.String, Type: inputs.String, Unit: inputs.NoUnit,
		Desc: "Kafka-reported rate unit label for request Handler Avg Idle Percent, scoped by host and Jolokia agent.",
	},

	"RequestHandlerAvgIdlePercent.EventType": &inputs.FieldInfo{
		DataType: inputs.String, Type: inputs.String, Unit: inputs.NoUnit,
		Desc: "Kafka-reported event type label for request Handler Avg Idle Percent, scoped by host and Jolokia agent.",
	},

	"RequestHandlerAvgIdlePercent.FifteenMinuteRate": &inputs.FieldInfo{
		DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.PercentDecimal,
		Desc: "Fraction from 0 to 1 for request Handler Avg Idle Percent, scoped by host and Jolokia agent.",
	},

	"RequestHandlerAvgIdlePercent.OneMinuteRate": &inputs.FieldInfo{
		DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.PercentDecimal,
		Desc: "Fraction from 0 to 1 for request Handler Avg Idle Percent, scoped by host and Jolokia agent.",
	},

	"RequestHandlerAvgIdlePercent.Count": &inputs.FieldInfo{
		DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.PercentDecimal,
		Desc: "Fraction from 0 to 1 for request Handler Avg Idle Percent, scoped by host and Jolokia agent.",
	},

	"RequestHandlerAvgIdlePercent.FiveMinuteRate": &inputs.FieldInfo{
		DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.PercentDecimal,
		Desc: "Fraction from 0 to 1 for request Handler Avg Idle Percent, scoped by host and Jolokia agent.",
	},

	"RequestHandlerAvgIdlePercent.MeanRate": &inputs.FieldInfo{
		DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.PercentDecimal,
		Desc: "Fraction from 0 to 1 for request Handler Avg Idle Percent, scoped by host and Jolokia agent.",
	},
}

var requestHandlerTags = map[string]interface{}{
	"jolokia_agent_url": &inputs.TagInfo{Desc: "Jolokia agent url path"},
}

func (j *KafkaRequestHandlerMment) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name:   "kafka_request_handler",
		Cat:    point.Metric,
		Desc:   "Kafka request handler JMX metrics, including request handler idle percentage and reported rate metadata.",
		DescZh: "Kafka 请求处理器 JMX 指标，包含请求处理器空闲比例及 Kafka 上报的速率元数据。",
		Fields: requestHandlerFields,
		Tags:   requestHandlerTags,
	}
}

//nolint:lll
var zooKeeperFields = map[string]interface{}{
	"ZooKeeperRequestLatencyMs.Mean": &inputs.FieldInfo{
		DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS,
		Desc: "Average time spent in zoo Keeper Request Latency, measured in milliseconds, scoped by Jolokia agent.",
	},

	"ZooKeeperRequestLatencyMs.98thPercentile": &inputs.FieldInfo{
		DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS,
		Desc: "98th percentile of time spent in zoo Keeper Request Latency, measured in milliseconds, scoped by Jolokia agent.",
	},

	"ZooKeeperRequestLatencyMs.75thPercentile": &inputs.FieldInfo{
		DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS,
		Desc: "75th percentile of time spent in zoo Keeper Request Latency, measured in milliseconds, scoped by Jolokia agent.",
	},

	"ZooKeeperRequestLatencyMs.95thPercentile": &inputs.FieldInfo{
		DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS,
		Desc: "95th percentile of time spent in zoo Keeper Request Latency, measured in milliseconds, scoped by Jolokia agent.",
	},

	"ZooKeeperRequestLatencyMs.Min": &inputs.FieldInfo{
		DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS,
		Desc: "Minimum time spent in zoo Keeper Request Latency, measured in milliseconds, scoped by Jolokia agent.",
	},

	"ZooKeeperRequestLatencyMs.999thPercentile": &inputs.FieldInfo{
		DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS,
		Desc: "99.9th percentile of time spent in zoo Keeper Request Latency, measured in milliseconds, scoped by Jolokia agent.",
	},

	"ZooKeeperRequestLatencyMs.Max": &inputs.FieldInfo{
		DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS,
		Desc: "Maximum time spent in zoo Keeper Request Latency, measured in milliseconds, scoped by Jolokia agent.",
	},

	"ZooKeeperRequestLatencyMs.StdDev": &inputs.FieldInfo{
		DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS,
		Desc: "Standard deviation of time spent in zoo Keeper Request Latency, measured in milliseconds, scoped by Jolokia agent.",
	},

	"ZooKeeperRequestLatencyMs.99thPercentile": &inputs.FieldInfo{
		DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS,
		Desc: "99th percentile of time spent in zoo Keeper Request Latency, measured in milliseconds, scoped by Jolokia agent.",
	},

	"ZooKeeperRequestLatencyMs.Count": &inputs.FieldInfo{
		DataType: inputs.Int, Type: inputs.Count, Unit: inputs.NCount,
		Desc: "Total number of zoo Keeper Request Latency, scoped by Jolokia agent.",
	},

	"ZooKeeperRequestLatencyMs.50thPercentile": &inputs.FieldInfo{
		DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS,
		Desc: "50th percentile of time spent in zoo Keeper Request Latency, measured in milliseconds, scoped by Jolokia agent.",
	},
}

var zooKeeperTags = map[string]interface{}{
	"jolokia_agent_url": &inputs.TagInfo{Desc: "Jolokia agent url path"},
}

func (j *KafkaZooKeeperMment) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name:   "kafka_zookeeper",
		Cat:    point.Metric,
		Desc:   "Kafka ZooKeeper client JMX metrics, including request latency, session expiration, and connection state statistics.",
		DescZh: "Kafka ZooKeeper 客户端 JMX 指标，包含请求延迟、会话过期和连接状态统计。",
		Fields: zooKeeperFields,
		Tags:   zooKeeperTags,
	}
}

//nolint:lll
var controllerFields = map[string]interface{}{
	"AutoLeaderBalanceRateAndTimeMs.50thPercentile": &inputs.FieldInfo{
		DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS,
		Desc: "50th percentile of time spent in auto Leader Balance Rate And Time, measured in milliseconds, scoped by controller broker.",
	},

	"AutoLeaderBalanceRateAndTimeMs.75thPercentile": &inputs.FieldInfo{
		DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS,
		Desc: "75th percentile of time spent in auto Leader Balance Rate And Time, measured in milliseconds, scoped by controller broker.",
	},

	"AutoLeaderBalanceRateAndTimeMs.95thPercentile": &inputs.FieldInfo{
		DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS,
		Desc: "95th percentile of time spent in auto Leader Balance Rate And Time, measured in milliseconds, scoped by controller broker.",
	},

	"AutoLeaderBalanceRateAndTimeMs.98thPercentile": &inputs.FieldInfo{
		DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS,
		Desc: "98th percentile of time spent in auto Leader Balance Rate And Time, measured in milliseconds, scoped by controller broker.",
	},

	"AutoLeaderBalanceRateAndTimeMs.999thPercentile": &inputs.FieldInfo{
		DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS,
		Desc: "99.9th percentile of time spent in auto Leader Balance Rate And Time, measured in milliseconds, scoped by controller broker.",
	},

	"AutoLeaderBalanceRateAndTimeMs.99thPercentile": &inputs.FieldInfo{
		DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS,
		Desc: "99th percentile of time spent in auto Leader Balance Rate And Time, measured in milliseconds, scoped by controller broker.",
	},

	"AutoLeaderBalanceRateAndTimeMs.FifteenMinuteRate": &inputs.FieldInfo{
		DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount,
		Desc: "Fifteen-minute moving average rate for auto Leader Balance Rate And Time in events per second, scoped by controller broker.",
	},

	"AutoLeaderBalanceRateAndTimeMs.FiveMinuteRate": &inputs.FieldInfo{
		DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount,
		Desc: "Five-minute moving average rate for auto Leader Balance Rate And Time in events per second, scoped by controller broker.",
	},

	"AutoLeaderBalanceRateAndTimeMs.Max": &inputs.FieldInfo{
		DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS,
		Desc: "Maximum time spent in auto Leader Balance Rate And Time, measured in milliseconds, scoped by controller broker.",
	},

	"AutoLeaderBalanceRateAndTimeMs.Mean": &inputs.FieldInfo{
		DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS,
		Desc: "Mean auto leader balance latency in milliseconds, scoped by the controller broker.",
	},

	"AutoLeaderBalanceRateAndTimeMs.MeanRate": &inputs.FieldInfo{
		DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount,
		Desc: "Mean rate for auto Leader Balance Rate And Time in events per second, scoped by controller broker.",
	},

	"AutoLeaderBalanceRateAndTimeMs.Min": &inputs.FieldInfo{
		DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS,
		Desc: "Minimum time spent in auto Leader Balance Rate And Time, measured in milliseconds, scoped by controller broker.",
	},

	"AutoLeaderBalanceRateAndTimeMs.OneMinuteRate": &inputs.FieldInfo{
		DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount,
		Desc: "One-minute moving average rate for auto Leader Balance Rate And Time in events per second, scoped by controller broker.",
	},

	"AutoLeaderBalanceRateAndTimeMs.StdDev": &inputs.FieldInfo{
		DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS,
		Desc: "Standard deviation of time spent in auto Leader Balance Rate And Time, measured in milliseconds, scoped by controller broker.",
	},

	"AutoLeaderBalanceRateAndTimeMs.Count": &inputs.FieldInfo{
		DataType: inputs.Int, Type: inputs.Count, Unit: inputs.NCount,
		Desc: "Total number of auto Leader Balance Rate And Time, scoped by controller broker.",
	},

	"AutoLeaderBalanceRateAndTimeMs.LatencyUnit": &inputs.FieldInfo{
		DataType: inputs.String, Type: inputs.String, Unit: inputs.NoUnit,
		Desc: "String unit label reported by Kafka for AutoLeaderBalanceRateAndTimeMs latency values, scoped by the controller broker.",
	},

	"AutoLeaderBalanceRateAndTimeMs.EventType": &inputs.FieldInfo{
		DataType: inputs.String, Type: inputs.String, Unit: inputs.NoUnit,
		Desc: "String event type label reported by Kafka for AutoLeaderBalanceRateAndTimeMs, scoped by the controller broker.",
	},

	"AutoLeaderBalanceRateAndTimeMs.RateUnit": &inputs.FieldInfo{
		DataType: inputs.String, Type: inputs.String, Unit: inputs.NoUnit,
		Desc: "String rate unit label reported by Kafka for AutoLeaderBalanceRateAndTimeMs rate values, scoped by the controller broker.",
	},

	"ControlledShutdownRateAndTimeMs.50thPercentile": &inputs.FieldInfo{
		DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS,
		Desc: "50th percentile of time spent in controlled Shutdown Rate And Time, measured in milliseconds, scoped by controller broker.",
	},

	"ControlledShutdownRateAndTimeMs.75thPercentile": &inputs.FieldInfo{
		DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS,
		Desc: "75th percentile of time spent in controlled Shutdown Rate And Time, measured in milliseconds, scoped by controller broker.",
	},

	"ControlledShutdownRateAndTimeMs.95thPercentile": &inputs.FieldInfo{
		DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS,
		Desc: "95th percentile of time spent in controlled Shutdown Rate And Time, measured in milliseconds, scoped by controller broker.",
	},

	"ControlledShutdownRateAndTimeMs.98thPercentile": &inputs.FieldInfo{
		DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS,
		Desc: "98th percentile of time spent in controlled Shutdown Rate And Time, measured in milliseconds, scoped by controller broker.",
	},

	"ControlledShutdownRateAndTimeMs.999thPercentile": &inputs.FieldInfo{
		DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS,
		Desc: "99.9th percentile of time spent in controlled Shutdown Rate And Time, measured in milliseconds, scoped by controller broker.",
	},

	"ControlledShutdownRateAndTimeMs.99thPercentile": &inputs.FieldInfo{
		DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS,
		Desc: "99th percentile of time spent in controlled Shutdown Rate And Time, measured in milliseconds, scoped by controller broker.",
	},

	"ControlledShutdownRateAndTimeMs.FifteenMinuteRate": &inputs.FieldInfo{
		DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount,
		Desc: "Fifteen-minute moving average rate for controlled Shutdown Rate And Time in events per second, scoped by controller broker.",
	},

	"ControlledShutdownRateAndTimeMs.FiveMinuteRate": &inputs.FieldInfo{
		DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount,
		Desc: "Five-minute moving average rate for controlled Shutdown Rate And Time in events per second, scoped by controller broker.",
	},

	"ControlledShutdownRateAndTimeMs.Max": &inputs.FieldInfo{
		DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS,
		Desc: "Maximum time spent in controlled Shutdown Rate And Time, measured in milliseconds, scoped by controller broker.",
	},

	"ControlledShutdownRateAndTimeMs.Mean": &inputs.FieldInfo{
		DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS,
		Desc: "Average time spent in controlled Shutdown Rate And Time, measured in milliseconds, scoped by controller broker.",
	},

	"ControlledShutdownRateAndTimeMs.MeanRate": &inputs.FieldInfo{
		DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount,
		Desc: "Mean rate for controlled Shutdown Rate And Time in events per second, scoped by controller broker.",
	},

	"ControlledShutdownRateAndTimeMs.Min": &inputs.FieldInfo{
		DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS,
		Desc: "Minimum time spent in controlled Shutdown Rate And Time, measured in milliseconds, scoped by controller broker.",
	},

	"ControlledShutdownRateAndTimeMs.OneMinuteRate": &inputs.FieldInfo{
		DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount,
		Desc: "One-minute moving average rate for controlled Shutdown Rate And Time in events per second, scoped by controller broker.",
	},

	"ControlledShutdownRateAndTimeMs.StdDev": &inputs.FieldInfo{
		DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS,
		Desc: "Standard deviation of time spent in controlled Shutdown Rate And Time, measured in milliseconds, scoped by controller broker.",
	},

	"ControlledShutdownRateAndTimeMs.Count": &inputs.FieldInfo{
		DataType: inputs.Int, Type: inputs.Count, Unit: inputs.NCount,
		Desc: "Total number of controlled Shutdown Rate And Time, scoped by controller broker.",
	},

	"ControlledShutdownRateAndTimeMs.LatencyUnit": &inputs.FieldInfo{
		DataType: inputs.String, Type: inputs.String, Unit: inputs.NoUnit,
		Desc: "Kafka-reported latency unit label for controlled Shutdown Rate And Time, scoped by controller broker.",
	},

	"ControlledShutdownRateAndTimeMs.EventType": &inputs.FieldInfo{
		DataType: inputs.String, Type: inputs.String, Unit: inputs.NoUnit,
		Desc: "Kafka-reported event type label for controlled Shutdown Rate And Time, scoped by controller broker.",
	},

	"ControlledShutdownRateAndTimeMs.RateUnit": &inputs.FieldInfo{
		DataType: inputs.String, Type: inputs.String, Unit: inputs.NoUnit,
		Desc: "Kafka-reported rate unit label for controlled Shutdown Rate And Time, scoped by controller broker.",
	},

	"ControllerChangeRateAndTimeMs.50thPercentile": &inputs.FieldInfo{
		DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS,
		Desc: "50th percentile of time spent in controller Change Rate And Time, measured in milliseconds, scoped by controller broker.",
	},

	"ControllerChangeRateAndTimeMs.75thPercentile": &inputs.FieldInfo{
		DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS,
		Desc: "75th percentile of time spent in controller Change Rate And Time, measured in milliseconds, scoped by controller broker.",
	},

	"ControllerChangeRateAndTimeMs.95thPercentile": &inputs.FieldInfo{
		DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS,
		Desc: "95th percentile of time spent in controller Change Rate And Time, measured in milliseconds, scoped by controller broker.",
	},

	"ControllerChangeRateAndTimeMs.98thPercentile": &inputs.FieldInfo{
		DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS,
		Desc: "98th percentile of time spent in controller Change Rate And Time, measured in milliseconds, scoped by controller broker.",
	},

	"ControllerChangeRateAndTimeMs.999thPercentile": &inputs.FieldInfo{
		DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS,
		Desc: "99.9th percentile of time spent in controller Change Rate And Time, measured in milliseconds, scoped by controller broker.",
	},

	"ControllerChangeRateAndTimeMs.99thPercentile": &inputs.FieldInfo{
		DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS,
		Desc: "99th percentile of time spent in controller Change Rate And Time, measured in milliseconds, scoped by controller broker.",
	},

	"ControllerChangeRateAndTimeMs.FifteenMinuteRate": &inputs.FieldInfo{
		DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount,
		Desc: "Fifteen-minute moving average rate for controller Change Rate And Time in events per second, scoped by controller broker.",
	},

	"ControllerChangeRateAndTimeMs.FiveMinuteRate": &inputs.FieldInfo{
		DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount,
		Desc: "Five-minute moving average rate for controller Change Rate And Time in events per second, scoped by controller broker.",
	},

	"ControllerChangeRateAndTimeMs.Max": &inputs.FieldInfo{
		DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS,
		Desc: "Maximum time spent in controller Change Rate And Time, measured in milliseconds, scoped by controller broker.",
	},

	"ControllerChangeRateAndTimeMs.Mean": &inputs.FieldInfo{
		DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS,
		Desc: "Average time spent in controller Change Rate And Time, measured in milliseconds, scoped by controller broker.",
	},

	"ControllerChangeRateAndTimeMs.MeanRate": &inputs.FieldInfo{
		DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount,
		Desc: "Mean rate for controller Change Rate And Time in events per second, scoped by controller broker.",
	},

	"ControllerChangeRateAndTimeMs.Min": &inputs.FieldInfo{
		DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS,
		Desc: "Minimum time spent in controller Change Rate And Time, measured in milliseconds, scoped by controller broker.",
	},

	"ControllerChangeRateAndTimeMs.OneMinuteRate": &inputs.FieldInfo{
		DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount,
		Desc: "One-minute moving average rate for controller Change Rate And Time in events per second, scoped by controller broker.",
	},

	"ControllerChangeRateAndTimeMs.StdDev": &inputs.FieldInfo{
		DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS,
		Desc: "Standard deviation of time spent in controller Change Rate And Time, measured in milliseconds, scoped by controller broker.",
	},

	"ControllerChangeRateAndTimeMs.Count": &inputs.FieldInfo{
		DataType: inputs.Int, Type: inputs.Count, Unit: inputs.NCount,
		Desc: "Total number of controller Change Rate And Time, scoped by controller broker.",
	},

	"ControllerChangeRateAndTimeMs.LatencyUnit": &inputs.FieldInfo{
		DataType: inputs.String, Type: inputs.String, Unit: inputs.NoUnit,
		Desc: "Kafka-reported latency unit label for controller Change Rate And Time, scoped by controller broker.",
	},

	"ControllerChangeRateAndTimeMs.EventType": &inputs.FieldInfo{
		DataType: inputs.String, Type: inputs.String, Unit: inputs.NoUnit,
		Desc: "Kafka-reported event type label for controller Change Rate And Time, scoped by controller broker.",
	},

	"ControllerChangeRateAndTimeMs.RateUnit": &inputs.FieldInfo{
		DataType: inputs.String, Type: inputs.String, Unit: inputs.NoUnit,
		Desc: "Kafka-reported rate unit label for controller Change Rate And Time, scoped by controller broker.",
	},

	"ControllerShutdownRateAndTimeMs.50thPercentile": &inputs.FieldInfo{
		DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS,
		Desc: "50th percentile of time spent in controller Shutdown Rate And Time, measured in milliseconds, scoped by controller broker.",
	},

	"ControllerShutdownRateAndTimeMs.75thPercentile": &inputs.FieldInfo{
		DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS,
		Desc: "75th percentile of time spent in controller Shutdown Rate And Time, measured in milliseconds, scoped by controller broker.",
	},

	"ControllerShutdownRateAndTimeMs.95thPercentile": &inputs.FieldInfo{
		DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS,
		Desc: "95th percentile of time spent in controller Shutdown Rate And Time, measured in milliseconds, scoped by controller broker.",
	},

	"ControllerShutdownRateAndTimeMs.98thPercentile": &inputs.FieldInfo{
		DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS,
		Desc: "98th percentile of time spent in controller Shutdown Rate And Time, measured in milliseconds, scoped by controller broker.",
	},

	"ControllerShutdownRateAndTimeMs.999thPercentile": &inputs.FieldInfo{
		DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS,
		Desc: "99.9th percentile of time spent in controller Shutdown Rate And Time, measured in milliseconds, scoped by controller broker.",
	},

	"ControllerShutdownRateAndTimeMs.99thPercentile": &inputs.FieldInfo{
		DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS,
		Desc: "99th percentile of time spent in controller Shutdown Rate And Time, measured in milliseconds, scoped by controller broker.",
	},

	"ControllerShutdownRateAndTimeMs.FifteenMinuteRate": &inputs.FieldInfo{
		DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount,
		Desc: "Fifteen-minute moving average rate for controller Shutdown Rate And Time in events per second, scoped by controller broker.",
	},

	"ControllerShutdownRateAndTimeMs.FiveMinuteRate": &inputs.FieldInfo{
		DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount,
		Desc: "Five-minute moving average rate for controller Shutdown Rate And Time in events per second, scoped by controller broker.",
	},

	"ControllerShutdownRateAndTimeMs.Max": &inputs.FieldInfo{
		DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS,
		Desc: "Maximum time spent in controller Shutdown Rate And Time, measured in milliseconds, scoped by controller broker.",
	},

	"ControllerShutdownRateAndTimeMs.Mean": &inputs.FieldInfo{
		DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS,
		Desc: "Average time spent in controller Shutdown Rate And Time, measured in milliseconds, scoped by controller broker.",
	},

	"ControllerShutdownRateAndTimeMs.MeanRate": &inputs.FieldInfo{
		DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount,
		Desc: "Mean rate for controller Shutdown Rate And Time in events per second, scoped by controller broker.",
	},

	"ControllerShutdownRateAndTimeMs.Min": &inputs.FieldInfo{
		DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS,
		Desc: "Minimum time spent in controller Shutdown Rate And Time, measured in milliseconds, scoped by controller broker.",
	},

	"ControllerShutdownRateAndTimeMs.OneMinuteRate": &inputs.FieldInfo{
		DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount,
		Desc: "One-minute moving average rate for controller Shutdown Rate And Time in events per second, scoped by controller broker.",
	},

	"ControllerShutdownRateAndTimeMs.StdDev": &inputs.FieldInfo{
		DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS,
		Desc: "Standard deviation of time spent in controller Shutdown Rate And Time, measured in milliseconds, scoped by controller broker.",
	},

	"ControllerShutdownRateAndTimeMs.Count": &inputs.FieldInfo{
		DataType: inputs.Int, Type: inputs.Count, Unit: inputs.NCount,
		Desc: "Total number of controller Shutdown Rate And Time, scoped by controller broker.",
	},

	"ControllerShutdownRateAndTimeMs.LatencyUnit": &inputs.FieldInfo{
		DataType: inputs.String, Type: inputs.String, Unit: inputs.NoUnit,
		Desc: "Kafka-reported latency unit label for controller Shutdown Rate And Time, scoped by controller broker.",
	},

	"ControllerShutdownRateAndTimeMs.EventType": &inputs.FieldInfo{
		DataType: inputs.String, Type: inputs.String, Unit: inputs.NoUnit,
		Desc: "Kafka-reported event type label for controller Shutdown Rate And Time, scoped by controller broker.",
	},

	"ControllerShutdownRateAndTimeMs.RateUnit": &inputs.FieldInfo{
		DataType: inputs.String, Type: inputs.String, Unit: inputs.NoUnit,
		Desc: "Kafka-reported rate unit label for controller Shutdown Rate And Time, scoped by controller broker.",
	},

	"IsrChangeRateAndTimeMs.50thPercentile": &inputs.FieldInfo{
		DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS,
		Desc: "50th percentile of time spent in isr Change Rate And Time, measured in milliseconds, scoped by controller broker.",
	},

	"IsrChangeRateAndTimeMs.75thPercentile": &inputs.FieldInfo{
		DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS,
		Desc: "75th percentile of time spent in isr Change Rate And Time, measured in milliseconds, scoped by controller broker.",
	},

	"IsrChangeRateAndTimeMs.95thPercentile": &inputs.FieldInfo{
		DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS,
		Desc: "95th percentile of time spent in isr Change Rate And Time, measured in milliseconds, scoped by controller broker.",
	},

	"IsrChangeRateAndTimeMs.98thPercentile": &inputs.FieldInfo{
		DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS,
		Desc: "98th percentile of time spent in isr Change Rate And Time, measured in milliseconds, scoped by controller broker.",
	},

	"IsrChangeRateAndTimeMs.999thPercentile": &inputs.FieldInfo{
		DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS,
		Desc: "99.9th percentile of time spent in isr Change Rate And Time, measured in milliseconds, scoped by controller broker.",
	},

	"IsrChangeRateAndTimeMs.99thPercentile": &inputs.FieldInfo{
		DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS,
		Desc: "99th percentile of time spent in isr Change Rate And Time, measured in milliseconds, scoped by controller broker.",
	},

	"IsrChangeRateAndTimeMs.FifteenMinuteRate": &inputs.FieldInfo{
		DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount,
		Desc: "Fifteen-minute moving average rate for isr Change Rate And Time in events per second, scoped by controller broker.",
	},

	"IsrChangeRateAndTimeMs.FiveMinuteRate": &inputs.FieldInfo{
		DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount,
		Desc: "Five-minute moving average rate for isr Change Rate And Time in events per second, scoped by controller broker.",
	},

	"IsrChangeRateAndTimeMs.Max": &inputs.FieldInfo{
		DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS,
		Desc: "Maximum time spent in isr Change Rate And Time, measured in milliseconds, scoped by controller broker.",
	},

	"IsrChangeRateAndTimeMs.Mean": &inputs.FieldInfo{
		DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS,
		Desc: "Average time spent in isr Change Rate And Time, measured in milliseconds, scoped by controller broker.",
	},

	"IsrChangeRateAndTimeMs.MeanRate": &inputs.FieldInfo{
		DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount,
		Desc: "Mean rate for isr Change Rate And Time in events per second, scoped by controller broker.",
	},

	"IsrChangeRateAndTimeMs.Min": &inputs.FieldInfo{
		DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS,
		Desc: "Minimum time spent in isr Change Rate And Time, measured in milliseconds, scoped by controller broker.",
	},

	"IsrChangeRateAndTimeMs.OneMinuteRate": &inputs.FieldInfo{
		DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount,
		Desc: "One-minute moving average rate for isr Change Rate And Time in events per second, scoped by controller broker.",
	},

	"IsrChangeRateAndTimeMs.StdDev": &inputs.FieldInfo{
		DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS,
		Desc: "Standard deviation of time spent in isr Change Rate And Time, measured in milliseconds, scoped by controller broker.",
	},

	"IsrChangeRateAndTimeMs.Count": &inputs.FieldInfo{
		DataType: inputs.Int, Type: inputs.Count, Unit: inputs.NCount,
		Desc: "Total number of isr Change Rate And Time, scoped by controller broker.",
	},

	"IsrChangeRateAndTimeMs.LatencyUnit": &inputs.FieldInfo{
		DataType: inputs.String, Type: inputs.String, Unit: inputs.NoUnit,
		Desc: "Kafka-reported latency unit label for isr Change Rate And Time, scoped by controller broker.",
	},

	"IsrChangeRateAndTimeMs.EventType": &inputs.FieldInfo{
		DataType: inputs.String, Type: inputs.String, Unit: inputs.NoUnit,
		Desc: "Kafka-reported event type label for isr Change Rate And Time, scoped by controller broker.",
	},

	"IsrChangeRateAndTimeMs.RateUnit": &inputs.FieldInfo{
		DataType: inputs.String, Type: inputs.String, Unit: inputs.NoUnit,
		Desc: "Kafka-reported rate unit label for isr Change Rate And Time, scoped by controller broker.",
	},

	"LeaderAndIsrResponseReceivedRateAndTimeMs.50thPercentile": &inputs.FieldInfo{
		DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS,
		Desc: "50th percentile of time spent in leader And Isr Response Received Rate And Time, measured in milliseconds, scoped by controller broker.",
	},

	"LeaderAndIsrResponseReceivedRateAndTimeMs.75thPercentile": &inputs.FieldInfo{
		DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS,
		Desc: "75th percentile of time spent in leader And Isr Response Received Rate And Time, measured in milliseconds, scoped by controller broker.",
	},

	"LeaderAndIsrResponseReceivedRateAndTimeMs.95thPercentile": &inputs.FieldInfo{
		DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS,
		Desc: "95th percentile of time spent in leader And Isr Response Received Rate And Time, measured in milliseconds, scoped by controller broker.",
	},

	"LeaderAndIsrResponseReceivedRateAndTimeMs.98thPercentile": &inputs.FieldInfo{
		DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS,
		Desc: "98th percentile of time spent in leader And Isr Response Received Rate And Time, measured in milliseconds, scoped by controller broker.",
	},

	"LeaderAndIsrResponseReceivedRateAndTimeMs.999thPercentile": &inputs.FieldInfo{
		DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS,
		Desc: "99.9th percentile of time spent in leader And Isr Response Received Rate And Time, measured in milliseconds, scoped by controller broker.",
	},

	"LeaderAndIsrResponseReceivedRateAndTimeMs.99thPercentile": &inputs.FieldInfo{
		DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS,
		Desc: "99th percentile of time spent in leader And Isr Response Received Rate And Time, measured in milliseconds, scoped by controller broker.",
	},

	"LeaderAndIsrResponseReceivedRateAndTimeMs.FifteenMinuteRate": &inputs.FieldInfo{
		DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount,
		Desc: "Fifteen-minute moving average rate for leader And Isr Response Received Rate And Time in events per second, scoped by controller broker.",
	},

	"LeaderAndIsrResponseReceivedRateAndTimeMs.FiveMinuteRate": &inputs.FieldInfo{
		DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount,
		Desc: "Five-minute moving average rate for leader And Isr Response Received Rate And Time in events per second, scoped by controller broker.",
	},

	"LeaderAndIsrResponseReceivedRateAndTimeMs.Max": &inputs.FieldInfo{
		DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS,
		Desc: "Maximum time spent in leader And Isr Response Received Rate And Time, measured in milliseconds, scoped by controller broker.",
	},

	"LeaderAndIsrResponseReceivedRateAndTimeMs.Mean": &inputs.FieldInfo{
		DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS,
		Desc: "Average time spent in leader And Isr Response Received Rate And Time, measured in milliseconds, scoped by controller broker.",
	},

	"LeaderAndIsrResponseReceivedRateAndTimeMs.MeanRate": &inputs.FieldInfo{
		DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount,
		Desc: "Mean rate for leader And Isr Response Received Rate And Time in events per second, scoped by controller broker.",
	},

	"LeaderAndIsrResponseReceivedRateAndTimeMs.Min": &inputs.FieldInfo{
		DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS,
		Desc: "Minimum time spent in leader And Isr Response Received Rate And Time, measured in milliseconds, scoped by controller broker.",
	},

	"LeaderAndIsrResponseReceivedRateAndTimeMs.OneMinuteRate": &inputs.FieldInfo{
		DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount,
		Desc: "One-minute moving average rate for leader And Isr Response Received Rate And Time in events per second, scoped by controller broker.",
	},

	"LeaderAndIsrResponseReceivedRateAndTimeMs.StdDev": &inputs.FieldInfo{
		DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS,
		Desc: "Standard deviation of time spent in leader And Isr Response Received Rate And Time, measured in milliseconds, scoped by controller broker.",
	},

	"LeaderAndIsrResponseReceivedRateAndTimeMs.Count": &inputs.FieldInfo{
		DataType: inputs.Int, Type: inputs.Count, Unit: inputs.NCount,
		Desc: "Total number of leader And Isr Response Received Rate And Time, scoped by controller broker.",
	},

	"LeaderAndIsrResponseReceivedRateAndTimeMs.LatencyUnit": &inputs.FieldInfo{
		DataType: inputs.String, Type: inputs.String, Unit: inputs.NoUnit,
		Desc: "Kafka-reported latency unit label for leader And Isr Response Received Rate And Time, scoped by controller broker.",
	},

	"LeaderAndIsrResponseReceivedRateAndTimeMs.EventType": &inputs.FieldInfo{
		DataType: inputs.String, Type: inputs.String, Unit: inputs.NoUnit,
		Desc: "Kafka-reported event type label for leader And Isr Response Received Rate And Time, scoped by controller broker.",
	},

	"LeaderAndIsrResponseReceivedRateAndTimeMs.RateUnit": &inputs.FieldInfo{
		DataType: inputs.String, Type: inputs.String, Unit: inputs.NoUnit,
		Desc: "Kafka-reported rate unit label for leader And Isr Response Received Rate And Time, scoped by controller broker.",
	},

	"LeaderElectionRateAndTimeMs.50thPercentile": &inputs.FieldInfo{
		DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS,
		Desc: "50th percentile of time spent in leader Election Rate And Time, measured in milliseconds, scoped by controller broker.",
	},

	"LeaderElectionRateAndTimeMs.75thPercentile": &inputs.FieldInfo{
		DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS,
		Desc: "75th percentile of time spent in leader Election Rate And Time, measured in milliseconds, scoped by controller broker.",
	},

	"LeaderElectionRateAndTimeMs.95thPercentile": &inputs.FieldInfo{
		DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS,
		Desc: "95th percentile of time spent in leader Election Rate And Time, measured in milliseconds, scoped by controller broker.",
	},

	"LeaderElectionRateAndTimeMs.98thPercentile": &inputs.FieldInfo{
		DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS,
		Desc: "98th percentile of time spent in leader Election Rate And Time, measured in milliseconds, scoped by controller broker.",
	},

	"LeaderElectionRateAndTimeMs.999thPercentile": &inputs.FieldInfo{
		DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS,
		Desc: "99.9th percentile of time spent in leader Election Rate And Time, measured in milliseconds, scoped by controller broker.",
	},

	"LeaderElectionRateAndTimeMs.99thPercentile": &inputs.FieldInfo{
		DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS,
		Desc: "99th percentile of time spent in leader Election Rate And Time, measured in milliseconds, scoped by controller broker.",
	},

	"LeaderElectionRateAndTimeMs.FifteenMinuteRate": &inputs.FieldInfo{
		DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount,
		Desc: "Fifteen-minute moving average rate for leader Election Rate And Time in events per second, scoped by controller broker.",
	},

	"LeaderElectionRateAndTimeMs.FiveMinuteRate": &inputs.FieldInfo{
		DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount,
		Desc: "Five-minute moving average rate for leader Election Rate And Time in events per second, scoped by controller broker.",
	},

	"LeaderElectionRateAndTimeMs.Max": &inputs.FieldInfo{
		DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS,
		Desc: "Maximum time spent in leader Election Rate And Time, measured in milliseconds, scoped by controller broker.",
	},

	"LeaderElectionRateAndTimeMs.Mean": &inputs.FieldInfo{
		DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS,
		Desc: "Average time spent in leader Election Rate And Time, measured in milliseconds, scoped by controller broker.",
	},

	"LeaderElectionRateAndTimeMs.MeanRate": &inputs.FieldInfo{
		DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount,
		Desc: "Mean rate for leader Election Rate And Time in events per second, scoped by controller broker.",
	},

	"LeaderElectionRateAndTimeMs.Min": &inputs.FieldInfo{
		DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS,
		Desc: "Minimum time spent in leader Election Rate And Time, measured in milliseconds, scoped by controller broker.",
	},

	"LeaderElectionRateAndTimeMs.OneMinuteRate": &inputs.FieldInfo{
		DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount,
		Desc: "One-minute moving average rate for leader Election Rate And Time in events per second, scoped by controller broker.",
	},

	"LeaderElectionRateAndTimeMs.StdDev": &inputs.FieldInfo{
		DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS,
		Desc: "Standard deviation of time spent in leader Election Rate And Time, measured in milliseconds, scoped by controller broker.",
	},

	"LeaderElectionRateAndTimeMs.Count": &inputs.FieldInfo{
		DataType: inputs.Int, Type: inputs.Count, Unit: inputs.NCount,
		Desc: "Total number of leader Election Rate And Time, scoped by controller broker.",
	},

	"LeaderElectionRateAndTimeMs.LatencyUnit": &inputs.FieldInfo{
		DataType: inputs.String, Type: inputs.String, Unit: inputs.NoUnit,
		Desc: "Kafka-reported latency unit label for leader Election Rate And Time, scoped by controller broker.",
	},

	"LeaderElectionRateAndTimeMs.EventType": &inputs.FieldInfo{
		DataType: inputs.String, Type: inputs.String, Unit: inputs.NoUnit,
		Desc: "Kafka-reported event type label for leader Election Rate And Time, scoped by controller broker.",
	},

	"LeaderElectionRateAndTimeMs.RateUnit": &inputs.FieldInfo{
		DataType: inputs.String, Type: inputs.String, Unit: inputs.NoUnit,
		Desc: "Kafka-reported rate unit label for leader Election Rate And Time, scoped by controller broker.",
	},

	"ListPartitionReassignmentRateAndTimeMs.50thPercentile": &inputs.FieldInfo{
		DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS,
		Desc: "50th percentile of time spent in list Partition Reassignment Rate And Time, measured in milliseconds, scoped by controller broker.",
	},

	"ListPartitionReassignmentRateAndTimeMs.75thPercentile": &inputs.FieldInfo{
		DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS,
		Desc: "75th percentile of time spent in list Partition Reassignment Rate And Time, measured in milliseconds, scoped by controller broker.",
	},

	"ListPartitionReassignmentRateAndTimeMs.95thPercentile": &inputs.FieldInfo{
		DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS,
		Desc: "95th percentile of time spent in list Partition Reassignment Rate And Time, measured in milliseconds, scoped by controller broker.",
	},

	"ListPartitionReassignmentRateAndTimeMs.98thPercentile": &inputs.FieldInfo{
		DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS,
		Desc: "98th percentile of time spent in list Partition Reassignment Rate And Time, measured in milliseconds, scoped by controller broker.",
	},

	"ListPartitionReassignmentRateAndTimeMs.999thPercentile": &inputs.FieldInfo{
		DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS,
		Desc: "99.9th percentile of time spent in list Partition Reassignment Rate And Time, measured in milliseconds, scoped by controller broker.",
	},

	"ListPartitionReassignmentRateAndTimeMs.99thPercentile": &inputs.FieldInfo{
		DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS,
		Desc: "99th percentile of time spent in list Partition Reassignment Rate And Time, measured in milliseconds, scoped by controller broker.",
	},

	"ListPartitionReassignmentRateAndTimeMs.FifteenMinuteRate": &inputs.FieldInfo{
		DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount,
		Desc: "Fifteen-minute moving average rate for list Partition Reassignment Rate And Time in events per second, scoped by controller broker.",
	},

	"ListPartitionReassignmentRateAndTimeMs.FiveMinuteRate": &inputs.FieldInfo{
		DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount,
		Desc: "Five-minute moving average rate for list Partition Reassignment Rate And Time in events per second, scoped by controller broker.",
	},

	"ListPartitionReassignmentRateAndTimeMs.Max": &inputs.FieldInfo{
		DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS,
		Desc: "Maximum time spent in list Partition Reassignment Rate And Time, measured in milliseconds, scoped by controller broker.",
	},

	"ListPartitionReassignmentRateAndTimeMs.Mean": &inputs.FieldInfo{
		DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS,
		Desc: "Average time spent in list Partition Reassignment Rate And Time, measured in milliseconds, scoped by controller broker.",
	},

	"ListPartitionReassignmentRateAndTimeMs.MeanRate": &inputs.FieldInfo{
		DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount,
		Desc: "Mean rate for list Partition Reassignment Rate And Time in events per second, scoped by controller broker.",
	},

	"ListPartitionReassignmentRateAndTimeMs.Min": &inputs.FieldInfo{
		DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS,
		Desc: "Minimum time spent in list Partition Reassignment Rate And Time, measured in milliseconds, scoped by controller broker.",
	},

	"ListPartitionReassignmentRateAndTimeMs.OneMinuteRate": &inputs.FieldInfo{
		DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount,
		Desc: "One-minute moving average rate for list Partition Reassignment Rate And Time in events per second, scoped by controller broker.",
	},

	"ListPartitionReassignmentRateAndTimeMs.StdDev": &inputs.FieldInfo{
		DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS,
		Desc: "Standard deviation of time spent in list Partition Reassignment Rate And Time, measured in milliseconds, scoped by controller broker.",
	},

	"ListPartitionReassignmentRateAndTimeMs.Count": &inputs.FieldInfo{
		DataType: inputs.Int, Type: inputs.Count, Unit: inputs.NCount,
		Desc: "Total number of list Partition Reassignment Rate And Time, scoped by controller broker.",
	},

	"ListPartitionReassignmentRateAndTimeMs.LatencyUnit": &inputs.FieldInfo{
		DataType: inputs.String, Type: inputs.String, Unit: inputs.NoUnit,
		Desc: "Kafka-reported latency unit label for list Partition Reassignment Rate And Time, scoped by controller broker.",
	},

	"ListPartitionReassignmentRateAndTimeMs.EventType": &inputs.FieldInfo{
		DataType: inputs.String, Type: inputs.String, Unit: inputs.NoUnit,
		Desc: "Kafka-reported event type label for list Partition Reassignment Rate And Time, scoped by controller broker.",
	},

	"ListPartitionReassignmentRateAndTimeMs.RateUnit": &inputs.FieldInfo{
		DataType: inputs.String, Type: inputs.String, Unit: inputs.NoUnit,
		Desc: "Kafka-reported rate unit label for list Partition Reassignment Rate And Time, scoped by controller broker.",
	},

	"LogDirChangeRateAndTimeMs.50thPercentile": &inputs.FieldInfo{
		DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS,
		Desc: "50th percentile of time spent in log Dir Change Rate And Time, measured in milliseconds, scoped by controller broker.",
	},

	"LogDirChangeRateAndTimeMs.75thPercentile": &inputs.FieldInfo{
		DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS,
		Desc: "75th percentile of time spent in log Dir Change Rate And Time, measured in milliseconds, scoped by controller broker.",
	},

	"LogDirChangeRateAndTimeMs.95thPercentile": &inputs.FieldInfo{
		DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS,
		Desc: "95th percentile of time spent in log Dir Change Rate And Time, measured in milliseconds, scoped by controller broker.",
	},

	"LogDirChangeRateAndTimeMs.98thPercentile": &inputs.FieldInfo{
		DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS,
		Desc: "98th percentile of time spent in log Dir Change Rate And Time, measured in milliseconds, scoped by controller broker.",
	},

	"LogDirChangeRateAndTimeMs.999thPercentile": &inputs.FieldInfo{
		DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS,
		Desc: "99.9th percentile of time spent in log Dir Change Rate And Time, measured in milliseconds, scoped by controller broker.",
	},

	"LogDirChangeRateAndTimeMs.99thPercentile": &inputs.FieldInfo{
		DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS,
		Desc: "99th percentile of time spent in log Dir Change Rate And Time, measured in milliseconds, scoped by controller broker.",
	},

	"LogDirChangeRateAndTimeMs.FifteenMinuteRate": &inputs.FieldInfo{
		DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount,
		Desc: "Fifteen-minute moving average rate for log Dir Change Rate And Time in events per second, scoped by controller broker.",
	},

	"LogDirChangeRateAndTimeMs.FiveMinuteRate": &inputs.FieldInfo{
		DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount,
		Desc: "Five-minute moving average rate for log Dir Change Rate And Time in events per second, scoped by controller broker.",
	},

	"LogDirChangeRateAndTimeMs.Max": &inputs.FieldInfo{
		DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS,
		Desc: "Maximum time spent in log Dir Change Rate And Time, measured in milliseconds, scoped by controller broker.",
	},

	"LogDirChangeRateAndTimeMs.Mean": &inputs.FieldInfo{
		DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS,
		Desc: "Average time spent in log Dir Change Rate And Time, measured in milliseconds, scoped by controller broker.",
	},

	"LogDirChangeRateAndTimeMs.MeanRate": &inputs.FieldInfo{
		DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount,
		Desc: "Mean rate for log Dir Change Rate And Time in events per second, scoped by controller broker.",
	},

	"LogDirChangeRateAndTimeMs.Min": &inputs.FieldInfo{
		DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS,
		Desc: "Minimum time spent in log Dir Change Rate And Time, measured in milliseconds, scoped by controller broker.",
	},

	"LogDirChangeRateAndTimeMs.OneMinuteRate": &inputs.FieldInfo{
		DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount,
		Desc: "One-minute moving average rate for log Dir Change Rate And Time in events per second, scoped by controller broker.",
	},

	"LogDirChangeRateAndTimeMs.StdDev": &inputs.FieldInfo{
		DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS,
		Desc: "Standard deviation of time spent in log Dir Change Rate And Time, measured in milliseconds, scoped by controller broker.",
	},

	"LogDirChangeRateAndTimeMs.Count": &inputs.FieldInfo{
		DataType: inputs.Int, Type: inputs.Count, Unit: inputs.NCount,
		Desc: "Total number of log Dir Change Rate And Time, scoped by controller broker.",
	},

	"LogDirChangeRateAndTimeMs.LatencyUnit": &inputs.FieldInfo{
		DataType: inputs.String, Type: inputs.String, Unit: inputs.NoUnit,
		Desc: "Kafka-reported latency unit label for log Dir Change Rate And Time, scoped by controller broker.",
	},

	"LogDirChangeRateAndTimeMs.EventType": &inputs.FieldInfo{
		DataType: inputs.String, Type: inputs.String, Unit: inputs.NoUnit,
		Desc: "Kafka-reported event type label for log Dir Change Rate And Time, scoped by controller broker.",
	},

	"LogDirChangeRateAndTimeMs.RateUnit": &inputs.FieldInfo{
		DataType: inputs.String, Type: inputs.String, Unit: inputs.NoUnit,
		Desc: "Kafka-reported rate unit label for log Dir Change Rate And Time, scoped by controller broker.",
	},

	"ManualLeaderBalanceRateAndTimeMs.50thPercentile": &inputs.FieldInfo{
		DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS,
		Desc: "50th percentile of time spent in manual Leader Balance Rate And Time, measured in milliseconds, scoped by controller broker.",
	},

	"ManualLeaderBalanceRateAndTimeMs.75thPercentile": &inputs.FieldInfo{
		DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS,
		Desc: "75th percentile of time spent in manual Leader Balance Rate And Time, measured in milliseconds, scoped by controller broker.",
	},

	"ManualLeaderBalanceRateAndTimeMs.95thPercentile": &inputs.FieldInfo{
		DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS,
		Desc: "95th percentile of time spent in manual Leader Balance Rate And Time, measured in milliseconds, scoped by controller broker.",
	},

	"ManualLeaderBalanceRateAndTimeMs.98thPercentile": &inputs.FieldInfo{
		DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS,
		Desc: "98th percentile of time spent in manual Leader Balance Rate And Time, measured in milliseconds, scoped by controller broker.",
	},

	"ManualLeaderBalanceRateAndTimeMs.999thPercentile": &inputs.FieldInfo{
		DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS,
		Desc: "99.9th percentile of time spent in manual Leader Balance Rate And Time, measured in milliseconds, scoped by controller broker.",
	},

	"ManualLeaderBalanceRateAndTimeMs.99thPercentile": &inputs.FieldInfo{
		DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS,
		Desc: "99th percentile of time spent in manual Leader Balance Rate And Time, measured in milliseconds, scoped by controller broker.",
	},

	"ManualLeaderBalanceRateAndTimeMs.FifteenMinuteRate": &inputs.FieldInfo{
		DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount,
		Desc: "Fifteen-minute moving average rate for manual Leader Balance Rate And Time in events per second, scoped by controller broker.",
	},

	"ManualLeaderBalanceRateAndTimeMs.FiveMinuteRate": &inputs.FieldInfo{
		DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount,
		Desc: "Five-minute moving average rate for manual Leader Balance Rate And Time in events per second, scoped by controller broker.",
	},

	"ManualLeaderBalanceRateAndTimeMs.Max": &inputs.FieldInfo{
		DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS,
		Desc: "Maximum time spent in manual Leader Balance Rate And Time, measured in milliseconds, scoped by controller broker.",
	},

	"ManualLeaderBalanceRateAndTimeMs.Mean": &inputs.FieldInfo{
		DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS,
		Desc: "Average time spent in manual Leader Balance Rate And Time, measured in milliseconds, scoped by controller broker.",
	},

	"ManualLeaderBalanceRateAndTimeMs.MeanRate": &inputs.FieldInfo{
		DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount,
		Desc: "Mean rate for manual Leader Balance Rate And Time in events per second, scoped by controller broker.",
	},

	"ManualLeaderBalanceRateAndTimeMs.Min": &inputs.FieldInfo{
		DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS,
		Desc: "Minimum time spent in manual Leader Balance Rate And Time, measured in milliseconds, scoped by controller broker.",
	},

	"ManualLeaderBalanceRateAndTimeMs.OneMinuteRate": &inputs.FieldInfo{
		DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount,
		Desc: "One-minute moving average rate for manual Leader Balance Rate And Time in events per second, scoped by controller broker.",
	},

	"ManualLeaderBalanceRateAndTimeMs.StdDev": &inputs.FieldInfo{
		DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS,
		Desc: "Standard deviation of time spent in manual Leader Balance Rate And Time, measured in milliseconds, scoped by controller broker.",
	},

	"ManualLeaderBalanceRateAndTimeMs.Count": &inputs.FieldInfo{
		DataType: inputs.Int, Type: inputs.Count, Unit: inputs.NCount,
		Desc: "Total number of manual Leader Balance Rate And Time, scoped by controller broker.",
	},

	"ManualLeaderBalanceRateAndTimeMs.LatencyUnit": &inputs.FieldInfo{
		DataType: inputs.String, Type: inputs.String, Unit: inputs.NoUnit,
		Desc: "Kafka-reported latency unit label for manual Leader Balance Rate And Time, scoped by controller broker.",
	},

	"ManualLeaderBalanceRateAndTimeMs.EventType": &inputs.FieldInfo{
		DataType: inputs.String, Type: inputs.String, Unit: inputs.NoUnit,
		Desc: "Kafka-reported event type label for manual Leader Balance Rate And Time, scoped by controller broker.",
	},

	"ManualLeaderBalanceRateAndTimeMs.RateUnit": &inputs.FieldInfo{
		DataType: inputs.String, Type: inputs.String, Unit: inputs.NoUnit,
		Desc: "Kafka-reported rate unit label for manual Leader Balance Rate And Time, scoped by controller broker.",
	},

	"PartitionReassignmentRateAndTimeMs.50thPercentile": &inputs.FieldInfo{
		DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS,
		Desc: "50th percentile of time spent in partition Reassignment Rate And Time, measured in milliseconds, scoped by controller broker.",
	},

	"PartitionReassignmentRateAndTimeMs.75thPercentile": &inputs.FieldInfo{
		DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS,
		Desc: "75th percentile of time spent in partition Reassignment Rate And Time, measured in milliseconds, scoped by controller broker.",
	},

	"PartitionReassignmentRateAndTimeMs.95thPercentile": &inputs.FieldInfo{
		DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS,
		Desc: "95th percentile of time spent in partition Reassignment Rate And Time, measured in milliseconds, scoped by controller broker.",
	},

	"PartitionReassignmentRateAndTimeMs.98thPercentile": &inputs.FieldInfo{
		DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS,
		Desc: "98th percentile of time spent in partition Reassignment Rate And Time, measured in milliseconds, scoped by controller broker.",
	},

	"PartitionReassignmentRateAndTimeMs.999thPercentile": &inputs.FieldInfo{
		DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS,
		Desc: "99.9th percentile of time spent in partition Reassignment Rate And Time, measured in milliseconds, scoped by controller broker.",
	},

	"PartitionReassignmentRateAndTimeMs.99thPercentile": &inputs.FieldInfo{
		DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS,
		Desc: "99th percentile of time spent in partition Reassignment Rate And Time, measured in milliseconds, scoped by controller broker.",
	},

	"PartitionReassignmentRateAndTimeMs.FifteenMinuteRate": &inputs.FieldInfo{
		DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount,
		Desc: "Fifteen-minute moving average rate for partition Reassignment Rate And Time in events per second, scoped by controller broker.",
	},

	"PartitionReassignmentRateAndTimeMs.FiveMinuteRate": &inputs.FieldInfo{
		DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount,
		Desc: "Five-minute moving average rate for partition Reassignment Rate And Time in events per second, scoped by controller broker.",
	},

	"PartitionReassignmentRateAndTimeMs.Max": &inputs.FieldInfo{
		DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS,
		Desc: "Maximum time spent in partition Reassignment Rate And Time, measured in milliseconds, scoped by controller broker.",
	},

	"PartitionReassignmentRateAndTimeMs.Mean": &inputs.FieldInfo{
		DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS,
		Desc: "Average time spent in partition Reassignment Rate And Time, measured in milliseconds, scoped by controller broker.",
	},

	"PartitionReassignmentRateAndTimeMs.MeanRate": &inputs.FieldInfo{
		DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount,
		Desc: "Mean rate for partition Reassignment Rate And Time in events per second, scoped by controller broker.",
	},

	"PartitionReassignmentRateAndTimeMs.Min": &inputs.FieldInfo{
		DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS,
		Desc: "Minimum time spent in partition Reassignment Rate And Time, measured in milliseconds, scoped by controller broker.",
	},

	"PartitionReassignmentRateAndTimeMs.OneMinuteRate": &inputs.FieldInfo{
		DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount,
		Desc: "One-minute moving average rate for partition Reassignment Rate And Time in events per second, scoped by controller broker.",
	},

	"PartitionReassignmentRateAndTimeMs.StdDev": &inputs.FieldInfo{
		DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS,
		Desc: "Standard deviation of time spent in partition Reassignment Rate And Time, measured in milliseconds, scoped by controller broker.",
	},

	"PartitionReassignmentRateAndTimeMs.Count": &inputs.FieldInfo{
		DataType: inputs.Int, Type: inputs.Count, Unit: inputs.NCount,
		Desc: "Total number of partition Reassignment Rate And Time, scoped by controller broker.",
	},

	"PartitionReassignmentRateAndTimeMs.LatencyUnit": &inputs.FieldInfo{
		DataType: inputs.String, Type: inputs.String, Unit: inputs.NoUnit,
		Desc: "Kafka-reported latency unit label for partition Reassignment Rate And Time, scoped by controller broker.",
	},

	"PartitionReassignmentRateAndTimeMs.EventType": &inputs.FieldInfo{
		DataType: inputs.String, Type: inputs.String, Unit: inputs.NoUnit,
		Desc: "Kafka-reported event type label for partition Reassignment Rate And Time, scoped by controller broker.",
	},

	"PartitionReassignmentRateAndTimeMs.RateUnit": &inputs.FieldInfo{
		DataType: inputs.String, Type: inputs.String, Unit: inputs.NoUnit,
		Desc: "Kafka-reported rate unit label for partition Reassignment Rate And Time, scoped by controller broker.",
	},

	"TopicChangeRateAndTimeMs.50thPercentile": &inputs.FieldInfo{
		DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS,
		Desc: "50th percentile of time spent in topic Change Rate And Time, measured in milliseconds, scoped by controller broker.",
	},

	"TopicChangeRateAndTimeMs.75thPercentile": &inputs.FieldInfo{
		DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS,
		Desc: "75th percentile of time spent in topic Change Rate And Time, measured in milliseconds, scoped by controller broker.",
	},

	"TopicChangeRateAndTimeMs.95thPercentile": &inputs.FieldInfo{
		DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS,
		Desc: "95th percentile of time spent in topic Change Rate And Time, measured in milliseconds, scoped by controller broker.",
	},

	"TopicChangeRateAndTimeMs.98thPercentile": &inputs.FieldInfo{
		DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS,
		Desc: "98th percentile of time spent in topic Change Rate And Time, measured in milliseconds, scoped by controller broker.",
	},

	"TopicChangeRateAndTimeMs.999thPercentile": &inputs.FieldInfo{
		DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS,
		Desc: "99.9th percentile of time spent in topic Change Rate And Time, measured in milliseconds, scoped by controller broker.",
	},

	"TopicChangeRateAndTimeMs.99thPercentile": &inputs.FieldInfo{
		DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS,
		Desc: "99th percentile of time spent in topic Change Rate And Time, measured in milliseconds, scoped by controller broker.",
	},

	"TopicChangeRateAndTimeMs.FifteenMinuteRate": &inputs.FieldInfo{
		DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount,
		Desc: "Fifteen-minute moving average rate for topic Change Rate And Time in events per second, scoped by controller broker.",
	},

	"TopicChangeRateAndTimeMs.FiveMinuteRate": &inputs.FieldInfo{
		DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount,
		Desc: "Five-minute moving average rate for topic Change Rate And Time in events per second, scoped by controller broker.",
	},

	"TopicChangeRateAndTimeMs.Max": &inputs.FieldInfo{
		DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS,
		Desc: "Maximum time spent in topic Change Rate And Time, measured in milliseconds, scoped by controller broker.",
	},

	"TopicChangeRateAndTimeMs.Mean": &inputs.FieldInfo{
		DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS,
		Desc: "Average time spent in topic Change Rate And Time, measured in milliseconds, scoped by controller broker.",
	},

	"TopicChangeRateAndTimeMs.MeanRate": &inputs.FieldInfo{
		DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount,
		Desc: "Mean rate for topic Change Rate And Time in events per second, scoped by controller broker.",
	},

	"TopicChangeRateAndTimeMs.Min": &inputs.FieldInfo{
		DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS,
		Desc: "Minimum time spent in topic Change Rate And Time, measured in milliseconds, scoped by controller broker.",
	},

	"TopicChangeRateAndTimeMs.OneMinuteRate": &inputs.FieldInfo{
		DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount,
		Desc: "One-minute moving average rate for topic Change Rate And Time in events per second, scoped by controller broker.",
	},

	"TopicChangeRateAndTimeMs.StdDev": &inputs.FieldInfo{
		DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS,
		Desc: "Standard deviation of time spent in topic Change Rate And Time, measured in milliseconds, scoped by controller broker.",
	},

	"TopicChangeRateAndTimeMs.Count": &inputs.FieldInfo{
		DataType: inputs.Int, Type: inputs.Count, Unit: inputs.NCount,
		Desc: "Total number of topic Change Rate And Time, scoped by controller broker.",
	},

	"TopicChangeRateAndTimeMs.LatencyUnit": &inputs.FieldInfo{
		DataType: inputs.String, Type: inputs.String, Unit: inputs.NoUnit,
		Desc: "Kafka-reported latency unit label for topic Change Rate And Time, scoped by controller broker.",
	},

	"TopicChangeRateAndTimeMs.EventType": &inputs.FieldInfo{
		DataType: inputs.String, Type: inputs.String, Unit: inputs.NoUnit,
		Desc: "Kafka-reported event type label for topic Change Rate And Time, scoped by controller broker.",
	},

	"TopicChangeRateAndTimeMs.RateUnit": &inputs.FieldInfo{
		DataType: inputs.String, Type: inputs.String, Unit: inputs.NoUnit,
		Desc: "Kafka-reported rate unit label for topic Change Rate And Time, scoped by controller broker.",
	},

	"TopicDeletionRateAndTimeMs.50thPercentile": &inputs.FieldInfo{
		DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS,
		Desc: "50th percentile of time spent in topic Deletion Rate And Time, measured in milliseconds, scoped by controller broker.",
	},

	"TopicDeletionRateAndTimeMs.75thPercentile": &inputs.FieldInfo{
		DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS,
		Desc: "75th percentile of time spent in topic Deletion Rate And Time, measured in milliseconds, scoped by controller broker.",
	},

	"TopicDeletionRateAndTimeMs.95thPercentile": &inputs.FieldInfo{
		DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS,
		Desc: "95th percentile of time spent in topic Deletion Rate And Time, measured in milliseconds, scoped by controller broker.",
	},

	"TopicDeletionRateAndTimeMs.98thPercentile": &inputs.FieldInfo{
		DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS,
		Desc: "98th percentile of time spent in topic Deletion Rate And Time, measured in milliseconds, scoped by controller broker.",
	},

	"TopicDeletionRateAndTimeMs.999thPercentile": &inputs.FieldInfo{
		DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS,
		Desc: "99.9th percentile of time spent in topic Deletion Rate And Time, measured in milliseconds, scoped by controller broker.",
	},

	"TopicDeletionRateAndTimeMs.99thPercentile": &inputs.FieldInfo{
		DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS,
		Desc: "99th percentile of time spent in topic Deletion Rate And Time, measured in milliseconds, scoped by controller broker.",
	},

	"TopicDeletionRateAndTimeMs.FifteenMinuteRate": &inputs.FieldInfo{
		DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount,
		Desc: "Fifteen-minute moving average rate for topic Deletion Rate And Time in events per second, scoped by controller broker.",
	},

	"TopicDeletionRateAndTimeMs.FiveMinuteRate": &inputs.FieldInfo{
		DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount,
		Desc: "Five-minute moving average rate for topic Deletion Rate And Time in events per second, scoped by controller broker.",
	},

	"TopicDeletionRateAndTimeMs.Max": &inputs.FieldInfo{
		DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS,
		Desc: "Maximum time spent in topic Deletion Rate And Time, measured in milliseconds, scoped by controller broker.",
	},

	"TopicDeletionRateAndTimeMs.Mean": &inputs.FieldInfo{
		DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS,
		Desc: "Average time spent in topic Deletion Rate And Time, measured in milliseconds, scoped by controller broker.",
	},

	"TopicDeletionRateAndTimeMs.MeanRate": &inputs.FieldInfo{
		DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount,
		Desc: "Mean rate for topic Deletion Rate And Time in events per second, scoped by controller broker.",
	},

	"TopicDeletionRateAndTimeMs.Min": &inputs.FieldInfo{
		DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS,
		Desc: "Minimum time spent in topic Deletion Rate And Time, measured in milliseconds, scoped by controller broker.",
	},

	"TopicDeletionRateAndTimeMs.OneMinuteRate": &inputs.FieldInfo{
		DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount,
		Desc: "One-minute moving average rate for topic Deletion Rate And Time in events per second, scoped by controller broker.",
	},

	"TopicDeletionRateAndTimeMs.StdDev": &inputs.FieldInfo{
		DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS,
		Desc: "Standard deviation of time spent in topic Deletion Rate And Time, measured in milliseconds, scoped by controller broker.",
	},

	"TopicDeletionRateAndTimeMs.Count": &inputs.FieldInfo{
		DataType: inputs.Int, Type: inputs.Count, Unit: inputs.NCount,
		Desc: "Total number of topic Deletion Rate And Time, scoped by controller broker.",
	},

	"TopicDeletionRateAndTimeMs.LatencyUnit": &inputs.FieldInfo{
		DataType: inputs.String, Type: inputs.String, Unit: inputs.NoUnit,
		Desc: "Kafka-reported latency unit label for topic Deletion Rate And Time, scoped by controller broker.",
	},

	"TopicDeletionRateAndTimeMs.EventType": &inputs.FieldInfo{
		DataType: inputs.String, Type: inputs.String, Unit: inputs.NoUnit,
		Desc: "Kafka-reported event type label for topic Deletion Rate And Time, scoped by controller broker.",
	},

	"TopicDeletionRateAndTimeMs.RateUnit": &inputs.FieldInfo{
		DataType: inputs.String, Type: inputs.String, Unit: inputs.NoUnit,
		Desc: "Kafka-reported rate unit label for topic Deletion Rate And Time, scoped by controller broker.",
	},

	"TopicUncleanLeaderElectionEnableRateAndTimeMs.50thPercentile": &inputs.FieldInfo{
		DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS,
		Desc: "50th percentile of time spent in topic Unclean Leader Election Enable Rate And Time, measured in milliseconds, scoped by controller broker.",
	},

	"TopicUncleanLeaderElectionEnableRateAndTimeMs.75thPercentile": &inputs.FieldInfo{
		DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS,
		Desc: "75th percentile of time spent in topic Unclean Leader Election Enable Rate And Time, measured in milliseconds, scoped by controller broker.",
	},

	"TopicUncleanLeaderElectionEnableRateAndTimeMs.95thPercentile": &inputs.FieldInfo{
		DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS,
		Desc: "95th percentile of time spent in topic Unclean Leader Election Enable Rate And Time, measured in milliseconds, scoped by controller broker.",
	},

	"TopicUncleanLeaderElectionEnableRateAndTimeMs.98thPercentile": &inputs.FieldInfo{
		DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS,
		Desc: "98th percentile of time spent in topic Unclean Leader Election Enable Rate And Time, measured in milliseconds, scoped by controller broker.",
	},

	"TopicUncleanLeaderElectionEnableRateAndTimeMs.999thPercentile": &inputs.FieldInfo{
		DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS,
		Desc: "99.9th percentile of time spent in topic Unclean Leader Election Enable Rate And Time, measured in milliseconds, scoped by controller broker.",
	},

	"TopicUncleanLeaderElectionEnableRateAndTimeMs.99thPercentile": &inputs.FieldInfo{
		DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS,
		Desc: "99th percentile of time spent in topic Unclean Leader Election Enable Rate And Time, measured in milliseconds, scoped by controller broker.",
	},

	"TopicUncleanLeaderElectionEnableRateAndTimeMs.FifteenMinuteRate": &inputs.FieldInfo{
		DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount,
		Desc: "Fifteen-minute moving average rate for topic Unclean Leader Election Enable Rate And Time in events per second, scoped by controller broker.",
	},

	"TopicUncleanLeaderElectionEnableRateAndTimeMs.FiveMinuteRate": &inputs.FieldInfo{
		DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount,
		Desc: "Five-minute moving average rate for topic Unclean Leader Election Enable Rate And Time in events per second, scoped by controller broker.",
	},

	"TopicUncleanLeaderElectionEnableRateAndTimeMs.Max": &inputs.FieldInfo{
		DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS,
		Desc: "Maximum time spent in topic Unclean Leader Election Enable Rate And Time, measured in milliseconds, scoped by controller broker.",
	},

	"TopicUncleanLeaderElectionEnableRateAndTimeMs.Mean": &inputs.FieldInfo{
		DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS,
		Desc: "Average time spent in topic Unclean Leader Election Enable Rate And Time, measured in milliseconds, scoped by controller broker.",
	},

	"TopicUncleanLeaderElectionEnableRateAndTimeMs.MeanRate": &inputs.FieldInfo{
		DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount,
		Desc: "Mean rate for topic Unclean Leader Election Enable Rate And Time in events per second, scoped by controller broker.",
	},

	"TopicUncleanLeaderElectionEnableRateAndTimeMs.Min": &inputs.FieldInfo{
		DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS,
		Desc: "Minimum time spent in topic Unclean Leader Election Enable Rate And Time, measured in milliseconds, scoped by controller broker.",
	},

	"TopicUncleanLeaderElectionEnableRateAndTimeMs.OneMinuteRate": &inputs.FieldInfo{
		DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount,
		Desc: "One-minute moving average rate for topic Unclean Leader Election Enable Rate And Time in events per second, scoped by controller broker.",
	},

	"TopicUncleanLeaderElectionEnableRateAndTimeMs.StdDev": &inputs.FieldInfo{
		DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS,
		Desc: "Standard deviation of time spent in topic Unclean Leader Election Enable Rate And Time, measured in milliseconds, scoped by controller broker.",
	},

	"TopicUncleanLeaderElectionEnableRateAndTimeMs.Count": &inputs.FieldInfo{
		DataType: inputs.Int, Type: inputs.Count, Unit: inputs.NCount,
		Desc: "Total number of topic Unclean Leader Election Enable Rate And Time, scoped by controller broker.",
	},

	"TopicUncleanLeaderElectionEnableRateAndTimeMs.LatencyUnit": &inputs.FieldInfo{
		DataType: inputs.String, Type: inputs.String, Unit: inputs.NoUnit,
		Desc: "Kafka-reported latency unit label for topic Unclean Leader Election Enable Rate And Time, scoped by controller broker.",
	},

	"TopicUncleanLeaderElectionEnableRateAndTimeMs.EventType": &inputs.FieldInfo{
		DataType: inputs.String, Type: inputs.String, Unit: inputs.NoUnit,
		Desc: "Kafka-reported event type label for topic Unclean Leader Election Enable Rate And Time, scoped by controller broker.",
	},

	"TopicUncleanLeaderElectionEnableRateAndTimeMs.RateUnit": &inputs.FieldInfo{
		DataType: inputs.String, Type: inputs.String, Unit: inputs.NoUnit,
		Desc: "Kafka-reported rate unit label for topic Unclean Leader Election Enable Rate And Time, scoped by controller broker.",
	},

	"UncleanLeaderElectionEnableRateAndTimeMs.50thPercentile": &inputs.FieldInfo{
		DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS,
		Desc: "50th percentile of time spent in unclean Leader Election Enable Rate And Time, measured in milliseconds, scoped by controller broker.",
	},

	"UncleanLeaderElectionEnableRateAndTimeMs.75thPercentile": &inputs.FieldInfo{
		DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS,
		Desc: "75th percentile of time spent in unclean Leader Election Enable Rate And Time, measured in milliseconds, scoped by controller broker.",
	},

	"UncleanLeaderElectionEnableRateAndTimeMs.95thPercentile": &inputs.FieldInfo{
		DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS,
		Desc: "95th percentile of time spent in unclean Leader Election Enable Rate And Time, measured in milliseconds, scoped by controller broker.",
	},

	"UncleanLeaderElectionEnableRateAndTimeMs.98thPercentile": &inputs.FieldInfo{
		DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS,
		Desc: "98th percentile of time spent in unclean Leader Election Enable Rate And Time, measured in milliseconds, scoped by controller broker.",
	},

	"UncleanLeaderElectionEnableRateAndTimeMs.999thPercentile": &inputs.FieldInfo{
		DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS,
		Desc: "99.9th percentile of time spent in unclean Leader Election Enable Rate And Time, measured in milliseconds, scoped by controller broker.",
	},

	"UncleanLeaderElectionEnableRateAndTimeMs.99thPercentile": &inputs.FieldInfo{
		DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS,
		Desc: "99th percentile of time spent in unclean Leader Election Enable Rate And Time, measured in milliseconds, scoped by controller broker.",
	},

	"UncleanLeaderElectionEnableRateAndTimeMs.FifteenMinuteRate": &inputs.FieldInfo{
		DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount,
		Desc: "Fifteen-minute moving average rate for unclean Leader Election Enable Rate And Time in events per second, scoped by controller broker.",
	},

	"UncleanLeaderElectionEnableRateAndTimeMs.FiveMinuteRate": &inputs.FieldInfo{
		DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount,
		Desc: "Five-minute moving average rate for unclean Leader Election Enable Rate And Time in events per second, scoped by controller broker.",
	},

	"UncleanLeaderElectionEnableRateAndTimeMs.Max": &inputs.FieldInfo{
		DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS,
		Desc: "Maximum time spent in unclean Leader Election Enable Rate And Time, measured in milliseconds, scoped by controller broker.",
	},

	"UncleanLeaderElectionEnableRateAndTimeMs.Mean": &inputs.FieldInfo{
		DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS,
		Desc: "Average time spent in unclean Leader Election Enable Rate And Time, measured in milliseconds, scoped by controller broker.",
	},

	"UncleanLeaderElectionEnableRateAndTimeMs.MeanRate": &inputs.FieldInfo{
		DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount,
		Desc: "Mean rate for unclean Leader Election Enable Rate And Time in events per second, scoped by controller broker.",
	},

	"UncleanLeaderElectionEnableRateAndTimeMs.Min": &inputs.FieldInfo{
		DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS,
		Desc: "Minimum time spent in unclean Leader Election Enable Rate And Time, measured in milliseconds, scoped by controller broker.",
	},

	"UncleanLeaderElectionEnableRateAndTimeMs.OneMinuteRate": &inputs.FieldInfo{
		DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount,
		Desc: "One-minute moving average rate for unclean Leader Election Enable Rate And Time in events per second, scoped by controller broker.",
	},

	"UncleanLeaderElectionEnableRateAndTimeMs.StdDev": &inputs.FieldInfo{
		DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS,
		Desc: "Standard deviation of time spent in unclean Leader Election Enable Rate And Time, measured in milliseconds, scoped by controller broker.",
	},

	"UncleanLeaderElectionEnableRateAndTimeMs.Count": &inputs.FieldInfo{
		DataType: inputs.Int, Type: inputs.Count, Unit: inputs.NCount,
		Desc: "Total number of unclean Leader Election Enable Rate And Time, scoped by controller broker.",
	},

	"UncleanLeaderElectionEnableRateAndTimeMs.LatencyUnit": &inputs.FieldInfo{
		DataType: inputs.String, Type: inputs.String, Unit: inputs.NoUnit,
		Desc: "Kafka-reported latency unit label for unclean Leader Election Enable Rate And Time, scoped by controller broker.",
	},

	"UncleanLeaderElectionEnableRateAndTimeMs.EventType": &inputs.FieldInfo{
		DataType: inputs.String, Type: inputs.String, Unit: inputs.NoUnit,
		Desc: "Kafka-reported event type label for unclean Leader Election Enable Rate And Time, scoped by controller broker.",
	},

	"UncleanLeaderElectionEnableRateAndTimeMs.RateUnit": &inputs.FieldInfo{
		DataType: inputs.String, Type: inputs.String, Unit: inputs.NoUnit,
		Desc: "Kafka-reported rate unit label for unclean Leader Election Enable Rate And Time, scoped by controller broker.",
	},

	"UpdateFeaturesRateAndTimeMs.50thPercentile": &inputs.FieldInfo{
		DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS,
		Desc: "50th percentile of time spent in update Features Rate And Time, measured in milliseconds, scoped by controller broker.",
	},

	"UpdateFeaturesRateAndTimeMs.75thPercentile": &inputs.FieldInfo{
		DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS,
		Desc: "75th percentile of time spent in update Features Rate And Time, measured in milliseconds, scoped by controller broker.",
	},

	"UpdateFeaturesRateAndTimeMs.95thPercentile": &inputs.FieldInfo{
		DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS,
		Desc: "95th percentile of time spent in update Features Rate And Time, measured in milliseconds, scoped by controller broker.",
	},

	"UpdateFeaturesRateAndTimeMs.98thPercentile": &inputs.FieldInfo{
		DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS,
		Desc: "98th percentile of time spent in update Features Rate And Time, measured in milliseconds, scoped by controller broker.",
	},

	"UpdateFeaturesRateAndTimeMs.999thPercentile": &inputs.FieldInfo{
		DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS,
		Desc: "99.9th percentile of time spent in update Features Rate And Time, measured in milliseconds, scoped by controller broker.",
	},

	"UpdateFeaturesRateAndTimeMs.99thPercentile": &inputs.FieldInfo{
		DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS,
		Desc: "99th percentile of time spent in update Features Rate And Time, measured in milliseconds, scoped by controller broker.",
	},

	"UpdateFeaturesRateAndTimeMs.FifteenMinuteRate": &inputs.FieldInfo{
		DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount,
		Desc: "Fifteen-minute moving average rate for update Features Rate And Time in events per second, scoped by controller broker.",
	},

	"UpdateFeaturesRateAndTimeMs.FiveMinuteRate": &inputs.FieldInfo{
		DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount,
		Desc: "Five-minute moving average rate for update Features Rate And Time in events per second, scoped by controller broker.",
	},

	"UpdateFeaturesRateAndTimeMs.Max": &inputs.FieldInfo{
		DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS,
		Desc: "Maximum time spent in update Features Rate And Time, measured in milliseconds, scoped by controller broker.",
	},

	"UpdateFeaturesRateAndTimeMs.Mean": &inputs.FieldInfo{
		DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS,
		Desc: "Average time spent in update Features Rate And Time, measured in milliseconds, scoped by controller broker.",
	},

	"UpdateFeaturesRateAndTimeMs.MeanRate": &inputs.FieldInfo{
		DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount,
		Desc: "Mean rate for update Features Rate And Time in events per second, scoped by controller broker.",
	},

	"UpdateFeaturesRateAndTimeMs.Min": &inputs.FieldInfo{
		DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS,
		Desc: "Minimum time spent in update Features Rate And Time, measured in milliseconds, scoped by controller broker.",
	},

	"UpdateFeaturesRateAndTimeMs.OneMinuteRate": &inputs.FieldInfo{
		DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount,
		Desc: "One-minute moving average rate for update Features Rate And Time in events per second, scoped by controller broker.",
	},

	"UpdateFeaturesRateAndTimeMs.StdDev": &inputs.FieldInfo{
		DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS,
		Desc: "Standard deviation of time spent in update Features Rate And Time, measured in milliseconds, scoped by controller broker.",
	},

	"UpdateFeaturesRateAndTimeMs.Count": &inputs.FieldInfo{
		DataType: inputs.Int, Type: inputs.Count, Unit: inputs.NCount,
		Desc: "Total number of update Features Rate And Time, scoped by controller broker.",
	},

	"UpdateFeaturesRateAndTimeMs.LatencyUnit": &inputs.FieldInfo{
		DataType: inputs.String, Type: inputs.String, Unit: inputs.NoUnit,
		Desc: "Kafka-reported latency unit label for update Features Rate And Time, scoped by controller broker.",
	},

	"UpdateFeaturesRateAndTimeMs.EventType": &inputs.FieldInfo{
		DataType: inputs.String, Type: inputs.String, Unit: inputs.NoUnit,
		Desc: "Kafka-reported event type label for update Features Rate And Time, scoped by controller broker.",
	},

	"UpdateFeaturesRateAndTimeMs.RateUnit": &inputs.FieldInfo{
		DataType: inputs.String, Type: inputs.String, Unit: inputs.NoUnit,
		Desc: "Kafka-reported rate unit label for update Features Rate And Time, scoped by controller broker.",
	},

	"UncleanLeaderElectionsPerSec.50thPercentile": &inputs.FieldInfo{
		DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount,
		Desc: "Rate for unclean Leader Elections Per Sec in events per second, scoped by controller broker.",
	},

	"UncleanLeaderElectionsPerSec.75thPercentile": &inputs.FieldInfo{
		DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount,
		Desc: "Rate for unclean Leader Elections Per Sec in events per second, scoped by controller broker.",
	},

	"UncleanLeaderElectionsPerSec.95thPercentile": &inputs.FieldInfo{
		DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount,
		Desc: "Rate for unclean Leader Elections Per Sec in events per second, scoped by controller broker.",
	},

	"UncleanLeaderElectionsPerSec.98thPercentile": &inputs.FieldInfo{
		DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount,
		Desc: "Rate for unclean Leader Elections Per Sec in events per second, scoped by controller broker.",
	},

	"UncleanLeaderElectionsPerSec.999thPercentile": &inputs.FieldInfo{
		DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount,
		Desc: "Rate for unclean Leader Elections Per Sec in events per second, scoped by controller broker.",
	},

	"UncleanLeaderElectionsPerSec.99thPercentile": &inputs.FieldInfo{
		DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount,
		Desc: "Rate for unclean Leader Elections Per Sec in events per second, scoped by controller broker.",
	},

	"UncleanLeaderElectionsPerSec.FifteenMinuteRate": &inputs.FieldInfo{
		DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount,
		Desc: "Fifteen-minute moving average rate for unclean Leader Elections Per Sec in events per second, scoped by controller broker.",
	},

	"UncleanLeaderElectionsPerSec.FiveMinuteRate": &inputs.FieldInfo{
		DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount,
		Desc: "Five-minute moving average rate for unclean Leader Elections Per Sec in events per second, scoped by controller broker.",
	},

	"UncleanLeaderElectionsPerSec.Max": &inputs.FieldInfo{
		DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount,
		Desc: "Rate for unclean Leader Elections Per Sec in events per second, scoped by controller broker.",
	},

	"UncleanLeaderElectionsPerSec.Mean": &inputs.FieldInfo{
		DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount,
		Desc: "Rate for unclean Leader Elections Per Sec in events per second, scoped by controller broker.",
	},

	"UncleanLeaderElectionsPerSec.MeanRate": &inputs.FieldInfo{
		DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount,
		Desc: "Mean rate for unclean Leader Elections Per Sec in events per second, scoped by controller broker.",
	},

	"UncleanLeaderElectionsPerSec.Min": &inputs.FieldInfo{
		DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount,
		Desc: "Rate for unclean Leader Elections Per Sec in events per second, scoped by controller broker.",
	},

	"UncleanLeaderElectionsPerSec.OneMinuteRate": &inputs.FieldInfo{
		DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount,
		Desc: "One-minute moving average rate for unclean Leader Elections Per Sec in events per second, scoped by controller broker.",
	},

	"UncleanLeaderElectionsPerSec.StdDev": &inputs.FieldInfo{
		DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount,
		Desc: "Rate for unclean Leader Elections Per Sec in events per second, scoped by controller broker.",
	},

	"UncleanLeaderElectionsPerSec.Count": &inputs.FieldInfo{
		DataType: inputs.Int, Type: inputs.Count, Unit: inputs.NCount,
		Desc: "Total number of unclean Leader Elections Per Sec, scoped by controller broker.",
	},

	"UncleanLeaderElectionsPerSec.LatencyUnit": &inputs.FieldInfo{
		DataType: inputs.String, Type: inputs.String, Unit: inputs.NoUnit,
		Desc: "Kafka-reported latency unit label for unclean Leader Elections Per Sec, scoped by controller broker.",
	},

	"UncleanLeaderElectionsPerSec.EventType": &inputs.FieldInfo{
		DataType: inputs.String, Type: inputs.String, Unit: inputs.NoUnit,
		Desc: "Kafka-reported event type label for unclean Leader Elections Per Sec, scoped by controller broker.",
	},

	"UncleanLeaderElectionsPerSec.RateUnit": &inputs.FieldInfo{
		DataType: inputs.String, Type: inputs.String, Unit: inputs.NoUnit,
		Desc: "Kafka-reported rate unit label for unclean Leader Elections Per Sec, scoped by controller broker.",
	},

	"EventQueueTimeMs.50thPercentile": &inputs.FieldInfo{
		DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS,
		Desc: "50th percentile of time spent in event Queue Time, measured in milliseconds, scoped by controller broker.",
	},

	"EventQueueTimeMs.75thPercentile": &inputs.FieldInfo{
		DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS,
		Desc: "75th percentile of time spent in event Queue Time, measured in milliseconds, scoped by controller broker.",
	},

	"EventQueueTimeMs.95thPercentile": &inputs.FieldInfo{
		DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS,
		Desc: "95th percentile of time spent in event Queue Time, measured in milliseconds, scoped by controller broker.",
	},

	"EventQueueTimeMs.98thPercentile": &inputs.FieldInfo{
		DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS,
		Desc: "98th percentile of time spent in event Queue Time, measured in milliseconds, scoped by controller broker.",
	},

	"EventQueueTimeMs.999thPercentile": &inputs.FieldInfo{
		DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS,
		Desc: "99.9th percentile of time spent in event Queue Time, measured in milliseconds, scoped by controller broker.",
	},

	"EventQueueTimeMs.99thPercentile": &inputs.FieldInfo{
		DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS,
		Desc: "99th percentile of time spent in event Queue Time, measured in milliseconds, scoped by controller broker.",
	},

	"EventQueueTimeMs.FifteenMinuteRate": &inputs.FieldInfo{
		DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount,
		Desc: "Fifteen-minute moving average rate for event Queue Time in events per second, scoped by controller broker.",
	},

	"EventQueueTimeMs.FiveMinuteRate": &inputs.FieldInfo{
		DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount,
		Desc: "Five-minute moving average rate for event Queue Time in events per second, scoped by controller broker.",
	},

	"EventQueueTimeMs.Max": &inputs.FieldInfo{
		DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS,
		Desc: "Maximum time spent in event Queue Time, measured in milliseconds, scoped by controller broker.",
	},

	"EventQueueTimeMs.Mean": &inputs.FieldInfo{
		DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS,
		Desc: "Average time spent in event Queue Time, measured in milliseconds, scoped by controller broker.",
	},

	"EventQueueTimeMs.MeanRate": &inputs.FieldInfo{
		DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount,
		Desc: "Mean rate for event Queue Time in events per second, scoped by controller broker.",
	},

	"EventQueueTimeMs.Min": &inputs.FieldInfo{
		DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS,
		Desc: "Minimum time spent in event Queue Time, measured in milliseconds, scoped by controller broker.",
	},

	"EventQueueTimeMs.OneMinuteRate": &inputs.FieldInfo{
		DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount,
		Desc: "One-minute moving average rate for event Queue Time in events per second, scoped by controller broker.",
	},

	"EventQueueTimeMs.StdDev": &inputs.FieldInfo{
		DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS,
		Desc: "Standard deviation of time spent in event Queue Time, measured in milliseconds, scoped by controller broker.",
	},

	"EventQueueTimeMs.Count": &inputs.FieldInfo{
		DataType: inputs.Int, Type: inputs.Count, Unit: inputs.NCount,
		Desc: "Total number of event Queue Time, scoped by controller broker.",
	},

	"EventQueueTimeMs.LatencyUnit": &inputs.FieldInfo{
		DataType: inputs.String, Type: inputs.String, Unit: inputs.NoUnit,
		Desc: "Kafka-reported latency unit label for event Queue Time, scoped by controller broker.",
	},

	"EventQueueTimeMs.EventType": &inputs.FieldInfo{
		DataType: inputs.String, Type: inputs.String, Unit: inputs.NoUnit,
		Desc: "Kafka-reported event type label for event Queue Time, scoped by controller broker.",
	},

	"EventQueueTimeMs.RateUnit": &inputs.FieldInfo{
		DataType: inputs.String, Type: inputs.String, Unit: inputs.NoUnit,
		Desc: "Kafka-reported rate unit label for event Queue Time, scoped by controller broker.",
	},

	"GlobalPartitionCount.Value": &inputs.FieldInfo{
		DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.NCount,
		Desc: "Current number of global Partition Count, scoped by controller broker.",
	},

	"GlobalTopicCount.Value": &inputs.FieldInfo{
		DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.NCount,
		Desc: "Current number of global Topic Count, scoped by controller broker.",
	},

	"OfflinePartitionsCount.Value": &inputs.FieldInfo{
		DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.NCount,
		Desc: "Current number of offline Partitions Count, scoped by controller broker.",
	},

	"PreferredReplicaImbalanceCount.Value": &inputs.FieldInfo{
		DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.NCount,
		Desc: "Current number of preferred Replica Imbalance Count, scoped by controller broker.",
	},

	"ReplicasIneligibleToDeleteCount.Value": &inputs.FieldInfo{
		DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.NCount,
		Desc: "Current number of replicas Ineligible To Delete Count, scoped by controller broker.",
	},

	"ReplicasToDeleteCount.Value": &inputs.FieldInfo{
		DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.NCount,
		Desc: "Current number of replicas To Delete Count, scoped by controller broker.",
	},

	"TopicsIneligibleToDeleteCount.Value": &inputs.FieldInfo{
		DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.NCount,
		Desc: "Current number of topics Ineligible To Delete Count, scoped by controller broker.",
	},

	"TopicsToDeleteCount.Value": &inputs.FieldInfo{
		DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.NCount,
		Desc: "Current number of topics To Delete Count, scoped by controller broker.",
	},

	"ActiveControllerCount.Value": &inputs.FieldInfo{
		DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.NCount,
		Desc: "Current number of active Controller Count, scoped by controller broker.",
	},

	"ControllerState.Value": &inputs.FieldInfo{
		DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.NCount,
		Desc: "Current value of controller State, scoped by controller broker.",
	},

	"EventQueueSize.Value": &inputs.FieldInfo{
		DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.NCount,
		Desc: "Current queue depth for event Queue Size, scoped by controller broker.",
	},

	"TotalQueueSize.Value": &inputs.FieldInfo{
		DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.NCount,
		Desc: "Current queue depth for total Queue Size, scoped by controller broker.",
	},
}

var controllerTags = map[string]interface{}{
	"jolokia_agent_url": &inputs.TagInfo{Desc: "Jolokia agent url path"},
}

func (j *KafkaControllerMment) Info() *inputs.MeasurementInfo { //nolint:funlen
	return &inputs.MeasurementInfo{
		Name:   "kafka_controller",
		Cat:    point.Metric,
		Desc:   "In Kafka cluster mode, a unique controller node will be elected, and only the controller node will receive valid metrics.",
		DescZh: "Kafka controller JMX 指标，包含 leader 选举、ISR 收缩/扩展、分区 leader、离线分区和控制器事件队列；集群模式下仅 controller 节点上报有效指标。",
		Fields: controllerFields,
		Tags:   controllerTags,
	}
}

var replicationTags = map[string]interface{}{
	"jolokia_agent_url": &inputs.TagInfo{Desc: "Jolokia agent url path"},
}

//nolint:lll
var replicationFields = map[string]interface{}{
	"FailedIsrUpdatesPerSec.Count": &inputs.FieldInfo{
		DataType: inputs.Int, Type: inputs.Count, Unit: inputs.NCount,
		Desc: "Total number of failed Isr Updates Per Sec, scoped by host and Jolokia agent.",
	},

	"FailedIsrUpdatesPerSec.EventType": &inputs.FieldInfo{
		DataType: inputs.String, Type: inputs.String, Unit: inputs.NoUnit,
		Desc: "Kafka-reported event type label for failed Isr Updates Per Sec, scoped by host and Jolokia agent.",
	},

	"FailedIsrUpdatesPerSec.FifteenMinuteRate": &inputs.FieldInfo{
		DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount,
		Desc: "Fifteen-minute moving average rate for failed Isr Updates Per Sec in events per second, scoped by host and Jolokia agent.",
	},

	"FailedIsrUpdatesPerSec.FiveMinuteRate": &inputs.FieldInfo{
		DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount,
		Desc: "Five-minute moving average rate for failed Isr Updates Per Sec in events per second, scoped by host and Jolokia agent.",
	},

	"FailedIsrUpdatesPerSec.MeanRate": &inputs.FieldInfo{
		DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount,
		Desc: "Mean rate for failed Isr Updates Per Sec in events per second, scoped by host and Jolokia agent.",
	},

	"FailedIsrUpdatesPerSec.OneMinuteRate": &inputs.FieldInfo{
		DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount,
		Desc: "One-minute moving average rate for failed Isr Updates Per Sec in events per second, scoped by host and Jolokia agent.",
	},

	"FailedIsrUpdatesPerSec.RateUnit": &inputs.FieldInfo{
		DataType: inputs.String, Type: inputs.String, Unit: inputs.NoUnit,
		Desc: "Kafka-reported rate unit label for failed Isr Updates Per Sec, scoped by host and Jolokia agent.",
	},

	"IsrExpandsPerSec.Count": &inputs.FieldInfo{
		DataType: inputs.Int, Type: inputs.Count, Unit: inputs.NCount,
		Desc: "Total number of isr Expands Per Sec, scoped by host and Jolokia agent.",
	},

	"IsrExpandsPerSec.EventType": &inputs.FieldInfo{
		DataType: inputs.String, Type: inputs.String, Unit: inputs.NoUnit,
		Desc: "Kafka-reported event type label for isr Expands Per Sec, scoped by host and Jolokia agent.",
	},

	"IsrExpandsPerSec.FifteenMinuteRate": &inputs.FieldInfo{
		DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount,
		Desc: "Fifteen-minute moving average rate for isr Expands Per Sec in events per second, scoped by host and Jolokia agent.",
	},

	"IsrExpandsPerSec.FiveMinuteRate": &inputs.FieldInfo{
		DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount,
		Desc: "Five-minute moving average rate for isr Expands Per Sec in events per second, scoped by host and Jolokia agent.",
	},

	"IsrExpandsPerSec.MeanRate": &inputs.FieldInfo{
		DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount,
		Desc: "Mean rate for isr Expands Per Sec in events per second, scoped by host and Jolokia agent.",
	},

	"IsrExpandsPerSec.OneMinuteRate": &inputs.FieldInfo{
		DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount,
		Desc: "One-minute moving average rate for isr Expands Per Sec in events per second, scoped by host and Jolokia agent.",
	},

	"IsrExpandsPerSec.RateUnit": &inputs.FieldInfo{
		DataType: inputs.String, Type: inputs.String, Unit: inputs.NoUnit,
		Desc: "Kafka-reported rate unit label for isr Expands Per Sec, scoped by host and Jolokia agent.",
	},

	"IsrShrinksPerSec.Count": &inputs.FieldInfo{
		DataType: inputs.Int, Type: inputs.Count, Unit: inputs.NCount,
		Desc: "Total number of isr Shrinks Per Sec, scoped by host and Jolokia agent.",
	},

	"IsrShrinksPerSec.EventType": &inputs.FieldInfo{
		DataType: inputs.String, Type: inputs.String, Unit: inputs.NoUnit,
		Desc: "Kafka-reported event type label for isr Shrinks Per Sec, scoped by host and Jolokia agent.",
	},

	"IsrShrinksPerSec.FifteenMinuteRate": &inputs.FieldInfo{
		DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount,
		Desc: "Fifteen-minute moving average rate for isr Shrinks Per Sec in events per second, scoped by host and Jolokia agent.",
	},

	"IsrShrinksPerSec.FiveMinuteRate": &inputs.FieldInfo{
		DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount,
		Desc: "Five-minute moving average rate for isr Shrinks Per Sec in events per second, scoped by host and Jolokia agent.",
	},

	"IsrShrinksPerSec.MeanRate": &inputs.FieldInfo{
		DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount,
		Desc: "Mean rate for isr Shrinks Per Sec in events per second, scoped by host and Jolokia agent.",
	},

	"IsrShrinksPerSec.OneMinuteRate": &inputs.FieldInfo{
		DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount,
		Desc: "One-minute moving average rate for isr Shrinks Per Sec in events per second, scoped by host and Jolokia agent.",
	},

	"IsrShrinksPerSec.RateUnit": &inputs.FieldInfo{
		DataType: inputs.String, Type: inputs.String, Unit: inputs.NoUnit,
		Desc: "Kafka-reported rate unit label for isr Shrinks Per Sec, scoped by host and Jolokia agent.",
	},

	"AtMinIsrPartitionCount.Value": &inputs.FieldInfo{
		DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.NCount,
		Desc: "Current number of at Min Isr Partition Count, scoped by host and Jolokia agent.",
	},

	"LeaderCount.Value": &inputs.FieldInfo{
		DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.NCount,
		Desc: "Current number of leader Count, scoped by host and Jolokia agent.",
	},

	"OfflineReplicaCount.Value": &inputs.FieldInfo{
		DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.NCount,
		Desc: "Current number of offline Replica Count, scoped by host and Jolokia agent.",
	},

	"PartitionCount.Value": &inputs.FieldInfo{
		DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.NCount,
		Desc: "Current number of partition Count, scoped by host and Jolokia agent.",
	},

	"ReassigningPartitions.Value": &inputs.FieldInfo{
		DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.NCount,
		Desc: "Current number of reassigning Partitions, scoped by host and Jolokia agent.",
	},

	"UnderMinIsrPartitionCount.Value": &inputs.FieldInfo{
		DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.NCount,
		Desc: "Current number of under Min Isr Partition Count, scoped by host and Jolokia agent.",
	},

	"UnderReplicatedPartitions.Value": &inputs.FieldInfo{
		DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.NCount,
		Desc: "Current number of under Replicated Partitions, scoped by host and Jolokia agent.",
	},
}

func (j *KafkaReplicaMment) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name:   "kafka_replica_manager",
		Cat:    point.Metric,
		Desc:   "Kafka replica manager JMX metrics, including ISR, replica fetcher, partition, leader, and replication health counters.",
		DescZh: "Kafka replica manager JMX 指标，包含 ISR、replica fetcher、分区、leader 和复制健康计数。",
		Fields: replicationFields,
		Tags:   replicationTags,
	}
}

var purgatoryTags = map[string]interface{}{
	"jolokia_agent_url": &inputs.TagInfo{Desc: "Jolokia agent url path"},
}

//nolint:lll
var purgatoryFields = map[string]interface{}{
	"AlterAcls.NumDelayedOperations": &inputs.FieldInfo{
		DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.PercentDecimal,
		Desc: "Fraction from 0 to 1 for num Delayed Operations, scoped by delayedOperation.",
	},

	"AlterAcls.PurgatorySize": &inputs.FieldInfo{
		DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.NCount,
		Desc: "Current value of alter Acls, scoped by delayedOperation.",
	},

	"DeleteRecords.NumDelayedOperations": &inputs.FieldInfo{
		DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.PercentDecimal,
		Desc: "Fraction from 0 to 1 for num Delayed Operations, scoped by delayedOperation.",
	},

	"DeleteRecords.PurgatorySize": &inputs.FieldInfo{
		DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.NCount,
		Desc: "Current number of requests waiting in delete Records, scoped by delayedOperation.",
	},

	"ElectLeader.NumDelayedOperations": &inputs.FieldInfo{
		DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.PercentDecimal,
		Desc: "Fraction from 0 to 1 for num Delayed Operations, scoped by delayedOperation.",
	},

	"ElectLeader.PurgatorySize": &inputs.FieldInfo{
		DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.NCount,
		Desc: "Current value of elect Leader, scoped by delayedOperation.",
	},

	"Fetch.NumDelayedOperations": &inputs.FieldInfo{
		DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.PercentDecimal,
		Desc: "Fraction from 0 to 1 for num Delayed Operations, scoped by delayedOperation.",
	},

	"Fetch.PurgatorySize": &inputs.FieldInfo{
		DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.NCount,
		Desc: "Current value of fetch, scoped by delayedOperation.",
	},

	"Heartbeat.NumDelayedOperations": &inputs.FieldInfo{
		DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.PercentDecimal,
		Desc: "Fraction from 0 to 1 for num Delayed Operations, scoped by delayedOperation.",
	},

	"Heartbeat.PurgatorySize": &inputs.FieldInfo{
		DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.NCount,
		Desc: "Current value of heartbeat, scoped by delayedOperation.",
	},

	"Produce.NumDelayedOperations": &inputs.FieldInfo{
		DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.PercentDecimal,
		Desc: "Fraction from 0 to 1 for num Delayed Operations, scoped by delayedOperation.",
	},

	"Produce.PurgatorySize": &inputs.FieldInfo{
		DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.NCount,
		Desc: "Current value of produce, scoped by delayedOperation.",
	},

	"Rebalance.NumDelayedOperations": &inputs.FieldInfo{
		DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.PercentDecimal,
		Desc: "Fraction from 0 to 1 for num Delayed Operations, scoped by delayedOperation.",
	},

	"Rebalance.PurgatorySize": &inputs.FieldInfo{
		DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.NCount,
		Desc: "Current value of rebalance, scoped by delayedOperation.",
	},

	"topic.NumDelayedOperations": &inputs.FieldInfo{
		DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.PercentDecimal,
		Desc: "Fraction from 0 to 1 for num Delayed Operations, scoped by delayedOperation.",
	},

	"topic.PurgatorySize": &inputs.FieldInfo{
		DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.NCount,
		Desc: "Current value of topic, scoped by delayedOperation.",
	},
}

func (j *KafkaPurgatoryMment) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name:   "kafka_purgatory",
		Cat:    point.Metric,
		Desc:   "Kafka request purgatory JMX metrics, including delayed operation counts and purgatory sizes by operation type.",
		DescZh: "Kafka 请求 purgatory JMX 指标，按操作类型记录延迟操作数量和 purgatory 大小。",
		Fields: purgatoryFields,
		Tags:   purgatoryTags,
	}
}

//nolint:lll
var requestFields = map[string]interface{}{
	"LocalTimeMs.50thPercentile": &inputs.FieldInfo{
		DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS,
		Desc: "50th percentile of time spent in local Time, measured in milliseconds, scoped by request.",
	},

	"LocalTimeMs.75thPercentile": &inputs.FieldInfo{
		DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS,
		Desc: "75th percentile of time spent in local Time, measured in milliseconds, scoped by request.",
	},

	"LocalTimeMs.95thPercentile": &inputs.FieldInfo{
		DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS,
		Desc: "95th percentile of time spent in local Time, measured in milliseconds, scoped by request.",
	},

	"LocalTimeMs.98thPercentile": &inputs.FieldInfo{
		DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS,
		Desc: "98th percentile of time spent in local Time, measured in milliseconds, scoped by request.",
	},

	"LocalTimeMs.999thPercentile": &inputs.FieldInfo{
		DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS,
		Desc: "99.9th percentile of time spent in local Time, measured in milliseconds, scoped by request.",
	},

	"LocalTimeMs.99thPercentile": &inputs.FieldInfo{
		DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS,
		Desc: "99th percentile of time spent in local Time, measured in milliseconds, scoped by request.",
	},

	"LocalTimeMs.Max": &inputs.FieldInfo{
		DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS,
		Desc: "Maximum time spent in local Time, measured in milliseconds, scoped by request.",
	},

	"LocalTimeMs.Mean": &inputs.FieldInfo{
		DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS,
		Desc: "Mean broker request local processing latency in milliseconds, scoped by request.",
	},

	"LocalTimeMs.Min": &inputs.FieldInfo{
		DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS,
		Desc: "Minimum time spent in local Time, measured in milliseconds, scoped by request.",
	},

	"LocalTimeMs.StdDev": &inputs.FieldInfo{
		DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS,
		Desc: "Standard deviation of time spent in local Time, measured in milliseconds, scoped by request.",
	},

	"LocalTimeMs.Count": &inputs.FieldInfo{
		DataType: inputs.Int, Type: inputs.Count, Unit: inputs.NCount,
		Desc: "Total number of local Time, scoped by request.",
	},

	"RemoteTimeMs.50thPercentile": &inputs.FieldInfo{
		DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS,
		Desc: "50th percentile of time spent in remote Time, measured in milliseconds, scoped by request.",
	},

	"RemoteTimeMs.75thPercentile": &inputs.FieldInfo{
		DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS,
		Desc: "75th percentile of time spent in remote Time, measured in milliseconds, scoped by request.",
	},

	"RemoteTimeMs.95thPercentile": &inputs.FieldInfo{
		DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS,
		Desc: "95th percentile of time spent in remote Time, measured in milliseconds, scoped by request.",
	},

	"RemoteTimeMs.98thPercentile": &inputs.FieldInfo{
		DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS,
		Desc: "98th percentile of time spent in remote Time, measured in milliseconds, scoped by request.",
	},

	"RemoteTimeMs.999thPercentile": &inputs.FieldInfo{
		DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS,
		Desc: "99.9th percentile of time spent in remote Time, measured in milliseconds, scoped by request.",
	},

	"RemoteTimeMs.99thPercentile": &inputs.FieldInfo{
		DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS,
		Desc: "99th percentile of time spent in remote Time, measured in milliseconds, scoped by request.",
	},

	"RemoteTimeMs.Max": &inputs.FieldInfo{
		DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS,
		Desc: "Maximum time spent in remote Time, measured in milliseconds, scoped by request.",
	},

	"RemoteTimeMs.Mean": &inputs.FieldInfo{
		DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS,
		Desc: "Average time spent in remote Time, measured in milliseconds, scoped by request.",
	},

	"RemoteTimeMs.Min": &inputs.FieldInfo{
		DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS,
		Desc: "Minimum time spent in remote Time, measured in milliseconds, scoped by request.",
	},

	"RemoteTimeMs.StdDev": &inputs.FieldInfo{
		DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS,
		Desc: "Standard deviation of time spent in remote Time, measured in milliseconds, scoped by request.",
	},

	"RemoteTimeMs.Count": &inputs.FieldInfo{
		DataType: inputs.Int, Type: inputs.Count, Unit: inputs.NCount,
		Desc: "Total number of remote Time, scoped by request.",
	},

	"RequestBytes.50thPercentile": &inputs.FieldInfo{
		DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.SizeByte,
		Desc: "50th percentile size of request Bytes in bytes, scoped by request.",
	},

	"RequestBytes.75thPercentile": &inputs.FieldInfo{
		DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.SizeByte,
		Desc: "75th percentile size of request Bytes in bytes, scoped by request.",
	},

	"RequestBytes.95thPercentile": &inputs.FieldInfo{
		DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.SizeByte,
		Desc: "95th percentile size of request Bytes in bytes, scoped by request.",
	},

	"RequestBytes.98thPercentile": &inputs.FieldInfo{
		DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.SizeByte,
		Desc: "98th percentile size of request Bytes in bytes, scoped by request.",
	},

	"RequestBytes.999thPercentile": &inputs.FieldInfo{
		DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.SizeByte,
		Desc: "99.9th percentile size of request Bytes in bytes, scoped by request.",
	},

	"RequestBytes.99thPercentile": &inputs.FieldInfo{
		DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.SizeByte,
		Desc: "99th percentile size of request Bytes in bytes, scoped by request.",
	},

	"RequestBytes.Max": &inputs.FieldInfo{
		DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.SizeByte,
		Desc: "Maximum size of request Bytes in bytes, scoped by request.",
	},

	"RequestBytes.Mean": &inputs.FieldInfo{
		DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.SizeByte,
		Desc: "Mean broker request size in bytes, scoped by request.",
	},

	"RequestBytes.Min": &inputs.FieldInfo{
		DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.SizeByte,
		Desc: "Minimum size of request Bytes in bytes, scoped by request.",
	},

	"RequestBytes.StdDev": &inputs.FieldInfo{
		DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.SizeByte,
		Desc: "Current size of request Bytes in bytes, scoped by request.",
	},

	"RequestBytes.Count": &inputs.FieldInfo{
		DataType: inputs.Int, Type: inputs.Count, Unit: inputs.NCount,
		Desc: "Total number of request Bytes, scoped by request.",
	},

	"RequestQueueTimeMs.50thPercentile": &inputs.FieldInfo{
		DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS,
		Desc: "50th percentile of time spent in request Queue Time, measured in milliseconds, scoped by request.",
	},

	"RequestQueueTimeMs.75thPercentile": &inputs.FieldInfo{
		DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS,
		Desc: "75th percentile of time spent in request Queue Time, measured in milliseconds, scoped by request.",
	},

	"RequestQueueTimeMs.95thPercentile": &inputs.FieldInfo{
		DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS,
		Desc: "95th percentile of time spent in request Queue Time, measured in milliseconds, scoped by request.",
	},

	"RequestQueueTimeMs.98thPercentile": &inputs.FieldInfo{
		DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS,
		Desc: "98th percentile of time spent in request Queue Time, measured in milliseconds, scoped by request.",
	},

	"RequestQueueTimeMs.999thPercentile": &inputs.FieldInfo{
		DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS,
		Desc: "99.9th percentile of time spent in request Queue Time, measured in milliseconds, scoped by request.",
	},

	"RequestQueueTimeMs.99thPercentile": &inputs.FieldInfo{
		DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS,
		Desc: "99th percentile of time spent in request Queue Time, measured in milliseconds, scoped by request.",
	},

	"RequestQueueTimeMs.Max": &inputs.FieldInfo{
		DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS,
		Desc: "Maximum time spent in request Queue Time, measured in milliseconds, scoped by request.",
	},

	"RequestQueueTimeMs.Mean": &inputs.FieldInfo{
		DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS,
		Desc: "Average time spent in request Queue Time, measured in milliseconds, scoped by request.",
	},

	"RequestQueueTimeMs.Min": &inputs.FieldInfo{
		DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS,
		Desc: "Minimum time spent in request Queue Time, measured in milliseconds, scoped by request.",
	},

	"RequestQueueTimeMs.StdDev": &inputs.FieldInfo{
		DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS,
		Desc: "Standard deviation of time spent in request Queue Time, measured in milliseconds, scoped by request.",
	},

	"RequestQueueTimeMs.Count": &inputs.FieldInfo{
		DataType: inputs.Int, Type: inputs.Count, Unit: inputs.NCount,
		Desc: "Total number of request Queue Time, scoped by request.",
	},

	"ResponseQueueTimeMs.50thPercentile": &inputs.FieldInfo{
		DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS,
		Desc: "50th percentile of time spent in response Queue Time, measured in milliseconds, scoped by request.",
	},

	"ResponseQueueTimeMs.75thPercentile": &inputs.FieldInfo{
		DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS,
		Desc: "75th percentile of time spent in response Queue Time, measured in milliseconds, scoped by request.",
	},

	"ResponseQueueTimeMs.95thPercentile": &inputs.FieldInfo{
		DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS,
		Desc: "95th percentile of time spent in response Queue Time, measured in milliseconds, scoped by request.",
	},

	"ResponseQueueTimeMs.98thPercentile": &inputs.FieldInfo{
		DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS,
		Desc: "98th percentile of time spent in response Queue Time, measured in milliseconds, scoped by request.",
	},

	"ResponseQueueTimeMs.999thPercentile": &inputs.FieldInfo{
		DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS,
		Desc: "99.9th percentile of time spent in response Queue Time, measured in milliseconds, scoped by request.",
	},

	"ResponseQueueTimeMs.99thPercentile": &inputs.FieldInfo{
		DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS,
		Desc: "99th percentile of time spent in response Queue Time, measured in milliseconds, scoped by request.",
	},

	"ResponseQueueTimeMs.Max": &inputs.FieldInfo{
		DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS,
		Desc: "Maximum time spent in response Queue Time, measured in milliseconds, scoped by request.",
	},

	"ResponseQueueTimeMs.Mean": &inputs.FieldInfo{
		DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS,
		Desc: "Average time spent in response Queue Time, measured in milliseconds, scoped by request.",
	},

	"ResponseQueueTimeMs.Min": &inputs.FieldInfo{
		DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS,
		Desc: "Minimum time spent in response Queue Time, measured in milliseconds, scoped by request.",
	},

	"ResponseQueueTimeMs.StdDev": &inputs.FieldInfo{
		DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS,
		Desc: "Standard deviation of time spent in response Queue Time, measured in milliseconds, scoped by request.",
	},

	"ResponseQueueTimeMs.Count": &inputs.FieldInfo{
		DataType: inputs.Int, Type: inputs.Count, Unit: inputs.NCount,
		Desc: "Total number of response Queue Time, scoped by request.",
	},

	"ResponseSendTimeMs.50thPercentile": &inputs.FieldInfo{
		DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS,
		Desc: "50th percentile of time spent in response Send Time, measured in milliseconds, scoped by request.",
	},

	"ResponseSendTimeMs.75thPercentile": &inputs.FieldInfo{
		DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS,
		Desc: "75th percentile of time spent in response Send Time, measured in milliseconds, scoped by request.",
	},

	"ResponseSendTimeMs.95thPercentile": &inputs.FieldInfo{
		DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS,
		Desc: "95th percentile of time spent in response Send Time, measured in milliseconds, scoped by request.",
	},

	"ResponseSendTimeMs.98thPercentile": &inputs.FieldInfo{
		DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS,
		Desc: "98th percentile of time spent in response Send Time, measured in milliseconds, scoped by request.",
	},

	"ResponseSendTimeMs.999thPercentile": &inputs.FieldInfo{
		DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS,
		Desc: "99.9th percentile of time spent in response Send Time, measured in milliseconds, scoped by request.",
	},

	"ResponseSendTimeMs.99thPercentile": &inputs.FieldInfo{
		DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS,
		Desc: "99th percentile of time spent in response Send Time, measured in milliseconds, scoped by request.",
	},

	"ResponseSendTimeMs.Max": &inputs.FieldInfo{
		DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS,
		Desc: "Maximum time spent in response Send Time, measured in milliseconds, scoped by request.",
	},

	"ResponseSendTimeMs.Mean": &inputs.FieldInfo{
		DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS,
		Desc: "Average time spent in response Send Time, measured in milliseconds, scoped by request.",
	},

	"ResponseSendTimeMs.Min": &inputs.FieldInfo{
		DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS,
		Desc: "Minimum time spent in response Send Time, measured in milliseconds, scoped by request.",
	},

	"ResponseSendTimeMs.StdDev": &inputs.FieldInfo{
		DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS,
		Desc: "Standard deviation of time spent in response Send Time, measured in milliseconds, scoped by request.",
	},

	"ResponseSendTimeMs.Count": &inputs.FieldInfo{
		DataType: inputs.Int, Type: inputs.Count, Unit: inputs.NCount,
		Desc: "Total number of response Send Time, scoped by request.",
	},

	"ThrottleTimeMs.50thPercentile": &inputs.FieldInfo{
		DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS,
		Desc: "50th percentile of time spent in throttle Time, measured in milliseconds, scoped by request.",
	},

	"ThrottleTimeMs.75thPercentile": &inputs.FieldInfo{
		DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS,
		Desc: "75th percentile of time spent in throttle Time, measured in milliseconds, scoped by request.",
	},

	"ThrottleTimeMs.95thPercentile": &inputs.FieldInfo{
		DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS,
		Desc: "95th percentile of time spent in throttle Time, measured in milliseconds, scoped by request.",
	},

	"ThrottleTimeMs.98thPercentile": &inputs.FieldInfo{
		DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS,
		Desc: "98th percentile of time spent in throttle Time, measured in milliseconds, scoped by request.",
	},

	"ThrottleTimeMs.999thPercentile": &inputs.FieldInfo{
		DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS,
		Desc: "99.9th percentile of time spent in throttle Time, measured in milliseconds, scoped by request.",
	},

	"ThrottleTimeMs.99thPercentile": &inputs.FieldInfo{
		DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS,
		Desc: "99th percentile of time spent in throttle Time, measured in milliseconds, scoped by request.",
	},

	"ThrottleTimeMs.Max": &inputs.FieldInfo{
		DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS,
		Desc: "Maximum time spent in throttle Time, measured in milliseconds, scoped by request.",
	},

	"ThrottleTimeMs.Mean": &inputs.FieldInfo{
		DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS,
		Desc: "Average time spent in throttle Time, measured in milliseconds, scoped by request.",
	},

	"ThrottleTimeMs.Min": &inputs.FieldInfo{
		DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS,
		Desc: "Minimum time spent in throttle Time, measured in milliseconds, scoped by request.",
	},

	"ThrottleTimeMs.StdDev": &inputs.FieldInfo{
		DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS,
		Desc: "Standard deviation of time spent in throttle Time, measured in milliseconds, scoped by request.",
	},

	"ThrottleTimeMs.Count": &inputs.FieldInfo{
		DataType: inputs.Int, Type: inputs.Count, Unit: inputs.NCount,
		Desc: "Total number of throttle Time, scoped by request.",
	},

	"TotalTimeMs.50thPercentile": &inputs.FieldInfo{
		DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS,
		Desc: "50th percentile of time spent in total Time, measured in milliseconds, scoped by request.",
	},

	"TotalTimeMs.75thPercentile": &inputs.FieldInfo{
		DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS,
		Desc: "75th percentile of time spent in total Time, measured in milliseconds, scoped by request.",
	},

	"TotalTimeMs.95thPercentile": &inputs.FieldInfo{
		DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS,
		Desc: "95th percentile of time spent in total Time, measured in milliseconds, scoped by request.",
	},

	"TotalTimeMs.98thPercentile": &inputs.FieldInfo{
		DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS,
		Desc: "98th percentile of time spent in total Time, measured in milliseconds, scoped by request.",
	},

	"TotalTimeMs.999thPercentile": &inputs.FieldInfo{
		DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS,
		Desc: "99.9th percentile of time spent in total Time, measured in milliseconds, scoped by request.",
	},

	"TotalTimeMs.99thPercentile": &inputs.FieldInfo{
		DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS,
		Desc: "99th percentile of time spent in total Time, measured in milliseconds, scoped by request.",
	},

	"TotalTimeMs.Max": &inputs.FieldInfo{
		DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS,
		Desc: "Maximum time spent in total Time, measured in milliseconds, scoped by request.",
	},

	"TotalTimeMs.Mean": &inputs.FieldInfo{
		DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS,
		Desc: "Average time spent in total Time, measured in milliseconds, scoped by request.",
	},

	"TotalTimeMs.Min": &inputs.FieldInfo{
		DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS,
		Desc: "Minimum time spent in total Time, measured in milliseconds, scoped by request.",
	},

	"TotalTimeMs.StdDev": &inputs.FieldInfo{
		DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS,
		Desc: "Standard deviation of time spent in total Time, measured in milliseconds, scoped by request.",
	},

	"TotalTimeMs.Count": &inputs.FieldInfo{
		DataType: inputs.Int, Type: inputs.Count, Unit: inputs.NCount,
		Desc: "Cumulative value of total Time, scoped by request.",
	},
}

var requestTags = map[string]interface{}{
	"jolokia_agent_url": &inputs.TagInfo{Desc: "Jolokia agent url path"},
}

func (j *KafkaRequestMment) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name:   "kafka_request",
		Cat:    point.Metric,
		Desc:   "Kafka request timing JMX metrics, including local, remote, queue, response, throttle, and total request times by request type.",
		DescZh: "Kafka 请求耗时 JMX 指标，按请求类型记录 local、remote、queue、response、throttle 和 total 等耗时。",
		Fields: requestFields,
		Tags:   requestTags,
	}
}

//nolint:lll
var topicsFields = map[string]interface{}{
	"BytesInPerSec.Count": &inputs.FieldInfo{
		DataType: inputs.Int, Type: inputs.Count, Unit: inputs.SizeByte,
		Desc: "Total bytes recorded for bytes In Per Sec, scoped by host and Jolokia agent.",
	},

	"BytesInPerSec.EventType": &inputs.FieldInfo{
		DataType: inputs.String, Type: inputs.String, Unit: inputs.NoUnit,
		Desc: "String event type label reported by Kafka for cluster-wide BytesInPerSec.",
	},

	"BytesInPerSec.FifteenMinuteRate": &inputs.FieldInfo{
		DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.BytesPerSec,
		Desc: "Fifteen-minute moving average rate for bytes In Per Sec in bytes per second, scoped by host and Jolokia agent.",
	},

	"BytesInPerSec.FiveMinuteRate": &inputs.FieldInfo{
		DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.BytesPerSec,
		Desc: "Five-minute moving average rate for bytes In Per Sec in bytes per second, scoped by host and Jolokia agent.",
	},

	"BytesInPerSec.MeanRate": &inputs.FieldInfo{
		DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.BytesPerSec,
		Desc: "Mean number of bytes received per second across broker topics.",
	},

	"BytesInPerSec.OneMinuteRate": &inputs.FieldInfo{
		DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.BytesPerSec,
		Desc: "One-minute moving average rate for bytes In Per Sec in bytes per second, scoped by host and Jolokia agent.",
	},

	"BytesInPerSec.RateUnit": &inputs.FieldInfo{
		DataType: inputs.String, Type: inputs.String, Unit: inputs.NoUnit,
		Desc: "Kafka-reported rate unit label for bytes In Per Sec, scoped by host and Jolokia agent.",
	},

	"BytesOutPerSec.Count": &inputs.FieldInfo{
		DataType: inputs.Int, Type: inputs.Count, Unit: inputs.SizeByte,
		Desc: "Total bytes recorded for bytes Out Per Sec, scoped by host and Jolokia agent.",
	},

	"BytesOutPerSec.EventType": &inputs.FieldInfo{
		DataType: inputs.String, Type: inputs.String, Unit: inputs.NoUnit,
		Desc: "Kafka-reported event type label for bytes Out Per Sec, scoped by host and Jolokia agent.",
	},

	"BytesOutPerSec.FifteenMinuteRate": &inputs.FieldInfo{
		DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.BytesPerSec,
		Desc: "Fifteen-minute moving average rate for bytes Out Per Sec in bytes per second, scoped by host and Jolokia agent.",
	},

	"BytesOutPerSec.FiveMinuteRate": &inputs.FieldInfo{
		DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.BytesPerSec,
		Desc: "Five-minute moving average rate for bytes Out Per Sec in bytes per second, scoped by host and Jolokia agent.",
	},

	"BytesOutPerSec.MeanRate": &inputs.FieldInfo{
		DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.BytesPerSec,
		Desc: "Mean rate for bytes Out Per Sec in bytes per second, scoped by host and Jolokia agent.",
	},

	"BytesOutPerSec.OneMinuteRate": &inputs.FieldInfo{
		DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.BytesPerSec,
		Desc: "One-minute moving average rate for bytes Out Per Sec in bytes per second, scoped by host and Jolokia agent.",
	},

	"BytesOutPerSec.RateUnit": &inputs.FieldInfo{
		DataType: inputs.String, Type: inputs.String, Unit: inputs.NoUnit,
		Desc: "Kafka-reported rate unit label for bytes Out Per Sec, scoped by host and Jolokia agent.",
	},

	"BytesRejectedPerSec.Count": &inputs.FieldInfo{
		DataType: inputs.Int, Type: inputs.Count, Unit: inputs.SizeByte,
		Desc: "Total bytes recorded for bytes Rejected Per Sec, scoped by host and Jolokia agent.",
	},

	"BytesRejectedPerSec.EventType": &inputs.FieldInfo{
		DataType: inputs.String, Type: inputs.String, Unit: inputs.NoUnit,
		Desc: "Kafka-reported event type label for bytes Rejected Per Sec, scoped by host and Jolokia agent.",
	},

	"BytesRejectedPerSec.FifteenMinuteRate": &inputs.FieldInfo{
		DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.BytesPerSec,
		Desc: "Fifteen-minute moving average rate for bytes Rejected Per Sec in bytes per second, scoped by host and Jolokia agent.",
	},

	"BytesRejectedPerSec.FiveMinuteRate": &inputs.FieldInfo{
		DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.BytesPerSec,
		Desc: "Five-minute moving average rate for bytes Rejected Per Sec in bytes per second, scoped by host and Jolokia agent.",
	},

	"BytesRejectedPerSec.MeanRate": &inputs.FieldInfo{
		DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.BytesPerSec,
		Desc: "Mean rate for bytes Rejected Per Sec in bytes per second, scoped by host and Jolokia agent.",
	},

	"BytesRejectedPerSec.OneMinuteRate": &inputs.FieldInfo{
		DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.BytesPerSec,
		Desc: "One-minute moving average rate for bytes Rejected Per Sec in bytes per second, scoped by host and Jolokia agent.",
	},

	"BytesRejectedPerSec.RateUnit": &inputs.FieldInfo{
		DataType: inputs.String, Type: inputs.String, Unit: inputs.NoUnit,
		Desc: "Kafka-reported rate unit label for bytes Rejected Per Sec, scoped by host and Jolokia agent.",
	},

	"FailedFetchRequestsPerSec.Count": &inputs.FieldInfo{
		DataType: inputs.Int, Type: inputs.Count, Unit: inputs.NCount,
		Desc: "Total number of failed Fetch Requests Per Sec, scoped by host and Jolokia agent.",
	},

	"FailedFetchRequestsPerSec.EventType": &inputs.FieldInfo{
		DataType: inputs.String, Type: inputs.String, Unit: inputs.NoUnit,
		Desc: "Kafka-reported event type label for failed Fetch Requests Per Sec, scoped by host and Jolokia agent.",
	},

	"FailedFetchRequestsPerSec.FifteenMinuteRate": &inputs.FieldInfo{
		DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.RequestsPerSec,
		Desc: "Fifteen-minute moving average rate for failed Fetch Requests Per Sec in requests per second, scoped by host and Jolokia agent.",
	},

	"FailedFetchRequestsPerSec.FiveMinuteRate": &inputs.FieldInfo{
		DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.RequestsPerSec,
		Desc: "Five-minute moving average rate for failed Fetch Requests Per Sec in requests per second, scoped by host and Jolokia agent.",
	},

	"FailedFetchRequestsPerSec.MeanRate": &inputs.FieldInfo{
		DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.RequestsPerSec,
		Desc: "Mean rate for failed Fetch Requests Per Sec in requests per second, scoped by host and Jolokia agent.",
	},

	"FailedFetchRequestsPerSec.OneMinuteRate": &inputs.FieldInfo{
		DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.RequestsPerSec,
		Desc: "One-minute moving average rate for failed Fetch Requests Per Sec in requests per second, scoped by host and Jolokia agent.",
	},

	"FailedFetchRequestsPerSec.RateUnit": &inputs.FieldInfo{
		DataType: inputs.String, Type: inputs.String, Unit: inputs.NoUnit,
		Desc: "Kafka-reported rate unit label for failed Fetch Requests Per Sec, scoped by host and Jolokia agent.",
	},

	"FailedProduceRequestsPerSec.Count": &inputs.FieldInfo{
		DataType: inputs.Int, Type: inputs.Count, Unit: inputs.NCount,
		Desc: "Total number of failed Produce Requests Per Sec, scoped by host and Jolokia agent.",
	},

	"FailedProduceRequestsPerSec.EventType": &inputs.FieldInfo{
		DataType: inputs.String, Type: inputs.String, Unit: inputs.NoUnit,
		Desc: "Kafka-reported event type label for failed Produce Requests Per Sec, scoped by host and Jolokia agent.",
	},

	"FailedProduceRequestsPerSec.FifteenMinuteRate": &inputs.FieldInfo{
		DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.RequestsPerSec,
		Desc: "Fifteen-minute moving average rate for failed Produce Requests Per Sec in requests per second, scoped by host and Jolokia agent.",
	},

	"FailedProduceRequestsPerSec.FiveMinuteRate": &inputs.FieldInfo{
		DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.RequestsPerSec,
		Desc: "Five-minute moving average rate for failed Produce Requests Per Sec in requests per second, scoped by host and Jolokia agent.",
	},

	"FailedProduceRequestsPerSec.MeanRate": &inputs.FieldInfo{
		DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.RequestsPerSec,
		Desc: "Mean rate for failed Produce Requests Per Sec in requests per second, scoped by host and Jolokia agent.",
	},

	"FailedProduceRequestsPerSec.OneMinuteRate": &inputs.FieldInfo{
		DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.RequestsPerSec,
		Desc: "One-minute moving average rate for failed Produce Requests Per Sec in requests per second, scoped by host and Jolokia agent.",
	},

	"FailedProduceRequestsPerSec.RateUnit": &inputs.FieldInfo{
		DataType: inputs.String, Type: inputs.String, Unit: inputs.NoUnit,
		Desc: "Kafka-reported rate unit label for failed Produce Requests Per Sec, scoped by host and Jolokia agent.",
	},

	"FetchMessageConversionsPerSec.Count": &inputs.FieldInfo{
		DataType: inputs.Int, Type: inputs.Count, Unit: inputs.NCount,
		Desc: "Total number of fetch Message Conversions Per Sec, scoped by host and Jolokia agent.",
	},

	"FetchMessageConversionsPerSec.EventType": &inputs.FieldInfo{
		DataType: inputs.String, Type: inputs.String, Unit: inputs.NoUnit,
		Desc: "Kafka-reported event type label for fetch Message Conversions Per Sec, scoped by host and Jolokia agent.",
	},

	"FetchMessageConversionsPerSec.FifteenMinuteRate": &inputs.FieldInfo{
		DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount,
		Desc: "Fifteen-minute moving average rate for fetch Message Conversions Per Sec in events per second, scoped by host and Jolokia agent.",
	},

	"FetchMessageConversionsPerSec.FiveMinuteRate": &inputs.FieldInfo{
		DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount,
		Desc: "Five-minute moving average rate for fetch Message Conversions Per Sec in events per second, scoped by host and Jolokia agent.",
	},

	"FetchMessageConversionsPerSec.MeanRate": &inputs.FieldInfo{
		DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount,
		Desc: "Mean rate for fetch Message Conversions Per Sec in events per second, scoped by host and Jolokia agent.",
	},

	"FetchMessageConversionsPerSec.OneMinuteRate": &inputs.FieldInfo{
		DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount,
		Desc: "One-minute moving average rate for fetch Message Conversions Per Sec in events per second, scoped by host and Jolokia agent.",
	},

	"FetchMessageConversionsPerSec.RateUnit": &inputs.FieldInfo{
		DataType: inputs.String, Type: inputs.String, Unit: inputs.NoUnit,
		Desc: "Kafka-reported rate unit label for fetch Message Conversions Per Sec, scoped by host and Jolokia agent.",
	},

	"InvalidMagicNumberRecordsPerSec.Count": &inputs.FieldInfo{
		DataType: inputs.Int, Type: inputs.Count, Unit: inputs.NCount,
		Desc: "Total number of invalid Magic Number Records Per Sec, scoped by host and Jolokia agent.",
	},

	"InvalidMagicNumberRecordsPerSec.EventType": &inputs.FieldInfo{
		DataType: inputs.String, Type: inputs.String, Unit: inputs.NoUnit,
		Desc: "Kafka-reported event type label for invalid Magic Number Records Per Sec, scoped by host and Jolokia agent.",
	},

	"InvalidMagicNumberRecordsPerSec.FifteenMinuteRate": &inputs.FieldInfo{
		DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount,
		Desc: "Fifteen-minute moving average rate for invalid Magic Number Records Per Sec in events per second, scoped by host and Jolokia agent.",
	},

	"InvalidMagicNumberRecordsPerSec.FiveMinuteRate": &inputs.FieldInfo{
		DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount,
		Desc: "Five-minute moving average rate for invalid Magic Number Records Per Sec in events per second, scoped by host and Jolokia agent.",
	},

	"InvalidMagicNumberRecordsPerSec.MeanRate": &inputs.FieldInfo{
		DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount,
		Desc: "Mean rate for invalid Magic Number Records Per Sec in events per second, scoped by host and Jolokia agent.",
	},

	"InvalidMagicNumberRecordsPerSec.OneMinuteRate": &inputs.FieldInfo{
		DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount,
		Desc: "One-minute moving average rate for invalid Magic Number Records Per Sec in events per second, scoped by host and Jolokia agent.",
	},

	"InvalidMagicNumberRecordsPerSec.RateUnit": &inputs.FieldInfo{
		DataType: inputs.String, Type: inputs.String, Unit: inputs.NoUnit,
		Desc: "Kafka-reported rate unit label for invalid Magic Number Records Per Sec, scoped by host and Jolokia agent.",
	},

	"InvalidMessageCrcRecordsPerSec.Count": &inputs.FieldInfo{
		DataType: inputs.Int, Type: inputs.Count, Unit: inputs.NCount,
		Desc: "Total number of invalid Message Crc Records Per Sec, scoped by host and Jolokia agent.",
	},

	"InvalidMessageCrcRecordsPerSec.EventType": &inputs.FieldInfo{
		DataType: inputs.String, Type: inputs.String, Unit: inputs.NoUnit,
		Desc: "Kafka-reported event type label for invalid Message Crc Records Per Sec, scoped by host and Jolokia agent.",
	},

	"InvalidMessageCrcRecordsPerSec.FifteenMinuteRate": &inputs.FieldInfo{
		DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount,
		Desc: "Fifteen-minute moving average rate for invalid Message Crc Records Per Sec in events per second, scoped by host and Jolokia agent.",
	},

	"InvalidMessageCrcRecordsPerSec.FiveMinuteRate": &inputs.FieldInfo{
		DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount,
		Desc: "Five-minute moving average rate for invalid Message Crc Records Per Sec in events per second, scoped by host and Jolokia agent.",
	},

	"InvalidMessageCrcRecordsPerSec.MeanRate": &inputs.FieldInfo{
		DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount,
		Desc: "Mean rate for invalid Message Crc Records Per Sec in events per second, scoped by host and Jolokia agent.",
	},

	"InvalidMessageCrcRecordsPerSec.OneMinuteRate": &inputs.FieldInfo{
		DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount,
		Desc: "One-minute moving average rate for invalid Message Crc Records Per Sec in events per second, scoped by host and Jolokia agent.",
	},

	"InvalidMessageCrcRecordsPerSec.RateUnit": &inputs.FieldInfo{
		DataType: inputs.String, Type: inputs.String, Unit: inputs.NoUnit,
		Desc: "Kafka-reported rate unit label for invalid Message Crc Records Per Sec, scoped by host and Jolokia agent.",
	},

	"InvalidOffsetOrSequenceRecordsPerSec.Count": &inputs.FieldInfo{
		DataType: inputs.Int, Type: inputs.Count, Unit: inputs.NCount,
		Desc: "Total number of invalid Offset Or Sequence Records Per Sec, scoped by host and Jolokia agent.",
	},

	"InvalidOffsetOrSequenceRecordsPerSec.EventType": &inputs.FieldInfo{
		DataType: inputs.String, Type: inputs.String, Unit: inputs.NoUnit,
		Desc: "Kafka-reported event type label for invalid Offset Or Sequence Records Per Sec, scoped by host and Jolokia agent.",
	},

	"InvalidOffsetOrSequenceRecordsPerSec.FifteenMinuteRate": &inputs.FieldInfo{
		DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount,
		Desc: "Fifteen-minute moving average rate for invalid Offset Or Sequence Records Per Sec in events per second, scoped by host and Jolokia agent.",
	},

	"InvalidOffsetOrSequenceRecordsPerSec.FiveMinuteRate": &inputs.FieldInfo{
		DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount,
		Desc: "Five-minute moving average rate for invalid Offset Or Sequence Records Per Sec in events per second, scoped by host and Jolokia agent.",
	},

	"InvalidOffsetOrSequenceRecordsPerSec.MeanRate": &inputs.FieldInfo{
		DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount,
		Desc: "Mean rate for invalid Offset Or Sequence Records Per Sec in events per second, scoped by host and Jolokia agent.",
	},

	"InvalidOffsetOrSequenceRecordsPerSec.OneMinuteRate": &inputs.FieldInfo{
		DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount,
		Desc: "One-minute moving average rate for invalid Offset Or Sequence Records Per Sec in events per second, scoped by host and Jolokia agent.",
	},

	"InvalidOffsetOrSequenceRecordsPerSec.RateUnit": &inputs.FieldInfo{
		DataType: inputs.String, Type: inputs.String, Unit: inputs.NoUnit,
		Desc: "Kafka-reported rate unit label for invalid Offset Or Sequence Records Per Sec, scoped by host and Jolokia agent.",
	},

	"MessagesInPerSec.Count": &inputs.FieldInfo{
		DataType: inputs.Int, Type: inputs.Count, Unit: inputs.NCount,
		Desc: "Total number of messages In Per Sec, scoped by host and Jolokia agent.",
	},

	"MessagesInPerSec.EventType": &inputs.FieldInfo{
		DataType: inputs.String, Type: inputs.String, Unit: inputs.NoUnit,
		Desc: "Kafka-reported event type label for messages In Per Sec, scoped by host and Jolokia agent.",
	},

	"MessagesInPerSec.FifteenMinuteRate": &inputs.FieldInfo{
		DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount,
		Desc: "Fifteen-minute moving average rate for messages In Per Sec in events per second, scoped by host and Jolokia agent.",
	},

	"MessagesInPerSec.FiveMinuteRate": &inputs.FieldInfo{
		DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount,
		Desc: "Five-minute moving average rate for messages In Per Sec in events per second, scoped by host and Jolokia agent.",
	},

	"MessagesInPerSec.MeanRate": &inputs.FieldInfo{
		DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount,
		Desc: "Mean rate for messages In Per Sec in events per second, scoped by host and Jolokia agent.",
	},

	"MessagesInPerSec.OneMinuteRate": &inputs.FieldInfo{
		DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount,
		Desc: "One-minute rate of messages received per second across broker topics.",
	},

	"MessagesInPerSec.RateUnit": &inputs.FieldInfo{
		DataType: inputs.String, Type: inputs.String, Unit: inputs.NoUnit,
		Desc: "Kafka-reported rate unit label for messages In Per Sec, scoped by host and Jolokia agent.",
	},

	"NoKeyCompactedTopicRecordsPerSec.Count": &inputs.FieldInfo{
		DataType: inputs.Int, Type: inputs.Count, Unit: inputs.NCount,
		Desc: "Total number of no Key Compacted Topic Records Per Sec, scoped by host and Jolokia agent.",
	},

	"NoKeyCompactedTopicRecordsPerSec.EventType": &inputs.FieldInfo{
		DataType: inputs.String, Type: inputs.String, Unit: inputs.NoUnit,
		Desc: "Kafka-reported event type label for no Key Compacted Topic Records Per Sec, scoped by host and Jolokia agent.",
	},

	"NoKeyCompactedTopicRecordsPerSec.FifteenMinuteRate": &inputs.FieldInfo{
		DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount,
		Desc: "Fifteen-minute moving average rate for no Key Compacted Topic Records Per Sec in events per second, scoped by host and Jolokia agent.",
	},

	"NoKeyCompactedTopicRecordsPerSec.FiveMinuteRate": &inputs.FieldInfo{
		DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount,
		Desc: "Five-minute moving average rate for no Key Compacted Topic Records Per Sec in events per second, scoped by host and Jolokia agent.",
	},

	"NoKeyCompactedTopicRecordsPerSec.MeanRate": &inputs.FieldInfo{
		DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount,
		Desc: "Mean rate for no Key Compacted Topic Records Per Sec in events per second, scoped by host and Jolokia agent.",
	},

	"NoKeyCompactedTopicRecordsPerSec.OneMinuteRate": &inputs.FieldInfo{
		DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount,
		Desc: "One-minute moving average rate for no Key Compacted Topic Records Per Sec in events per second, scoped by host and Jolokia agent.",
	},

	"NoKeyCompactedTopicRecordsPerSec.RateUnit": &inputs.FieldInfo{
		DataType: inputs.String, Type: inputs.String, Unit: inputs.NoUnit,
		Desc: "Kafka-reported rate unit label for no Key Compacted Topic Records Per Sec, scoped by host and Jolokia agent.",
	},

	"ProduceMessageConversionsPerSec.Count": &inputs.FieldInfo{
		DataType: inputs.Int, Type: inputs.Count, Unit: inputs.NCount,
		Desc: "Total number of produce Message Conversions Per Sec, scoped by host and Jolokia agent.",
	},

	"ProduceMessageConversionsPerSec.EventType": &inputs.FieldInfo{
		DataType: inputs.String, Type: inputs.String, Unit: inputs.NoUnit,
		Desc: "Kafka-reported event type label for produce Message Conversions Per Sec, scoped by host and Jolokia agent.",
	},

	"ProduceMessageConversionsPerSec.FifteenMinuteRate": &inputs.FieldInfo{
		DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount,
		Desc: "Fifteen-minute moving average rate for produce Message Conversions Per Sec in events per second, scoped by host and Jolokia agent.",
	},

	"ProduceMessageConversionsPerSec.FiveMinuteRate": &inputs.FieldInfo{
		DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount,
		Desc: "Five-minute moving average rate for produce Message Conversions Per Sec in events per second, scoped by host and Jolokia agent.",
	},

	"ProduceMessageConversionsPerSec.MeanRate": &inputs.FieldInfo{
		DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount,
		Desc: "Mean rate for produce Message Conversions Per Sec in events per second, scoped by host and Jolokia agent.",
	},

	"ProduceMessageConversionsPerSec.OneMinuteRate": &inputs.FieldInfo{
		DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount,
		Desc: "One-minute moving average rate for produce Message Conversions Per Sec in events per second, scoped by host and Jolokia agent.",
	},

	"ProduceMessageConversionsPerSec.RateUnit": &inputs.FieldInfo{
		DataType: inputs.String, Type: inputs.String, Unit: inputs.NoUnit,
		Desc: "Kafka-reported rate unit label for produce Message Conversions Per Sec, scoped by host and Jolokia agent.",
	},

	"ReassignmentBytesInPerSec.Count": &inputs.FieldInfo{
		DataType: inputs.Int, Type: inputs.Count, Unit: inputs.SizeByte,
		Desc: "Total bytes recorded for reassignment Bytes In Per Sec, scoped by host and Jolokia agent.",
	},

	"ReassignmentBytesInPerSec.EventType": &inputs.FieldInfo{
		DataType: inputs.String, Type: inputs.String, Unit: inputs.NoUnit,
		Desc: "Kafka-reported event type label for reassignment Bytes In Per Sec, scoped by host and Jolokia agent.",
	},

	"ReassignmentBytesInPerSec.FifteenMinuteRate": &inputs.FieldInfo{
		DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.BytesPerSec,
		Desc: "Fifteen-minute moving average rate for reassignment Bytes In Per Sec in bytes per second, scoped by host and Jolokia agent.",
	},

	"ReassignmentBytesInPerSec.FiveMinuteRate": &inputs.FieldInfo{
		DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.BytesPerSec,
		Desc: "Five-minute moving average rate for reassignment Bytes In Per Sec in bytes per second, scoped by host and Jolokia agent.",
	},

	"ReassignmentBytesInPerSec.MeanRate": &inputs.FieldInfo{
		DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.BytesPerSec,
		Desc: "Mean rate for reassignment Bytes In Per Sec in bytes per second, scoped by host and Jolokia agent.",
	},

	"ReassignmentBytesInPerSec.OneMinuteRate": &inputs.FieldInfo{
		DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.BytesPerSec,
		Desc: "One-minute moving average rate for reassignment Bytes In Per Sec in bytes per second, scoped by host and Jolokia agent.",
	},

	"ReassignmentBytesInPerSec.RateUnit": &inputs.FieldInfo{
		DataType: inputs.String, Type: inputs.String, Unit: inputs.NoUnit,
		Desc: "Kafka-reported rate unit label for reassignment Bytes In Per Sec, scoped by host and Jolokia agent.",
	},

	"ReassignmentBytesOutPerSec.Count": &inputs.FieldInfo{
		DataType: inputs.Int, Type: inputs.Count, Unit: inputs.SizeByte,
		Desc: "Total bytes recorded for reassignment Bytes Out Per Sec, scoped by host and Jolokia agent.",
	},

	"ReassignmentBytesOutPerSec.EventType": &inputs.FieldInfo{
		DataType: inputs.String, Type: inputs.String, Unit: inputs.NoUnit,
		Desc: "Kafka-reported event type label for reassignment Bytes Out Per Sec, scoped by host and Jolokia agent.",
	},

	"ReassignmentBytesOutPerSec.FifteenMinuteRate": &inputs.FieldInfo{
		DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.BytesPerSec,
		Desc: "Fifteen-minute moving average rate for reassignment Bytes Out Per Sec in bytes per second, scoped by host and Jolokia agent.",
	},

	"ReassignmentBytesOutPerSec.FiveMinuteRate": &inputs.FieldInfo{
		DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.BytesPerSec,
		Desc: "Five-minute moving average rate for reassignment Bytes Out Per Sec in bytes per second, scoped by host and Jolokia agent.",
	},

	"ReassignmentBytesOutPerSec.MeanRate": &inputs.FieldInfo{
		DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.BytesPerSec,
		Desc: "Mean rate for reassignment Bytes Out Per Sec in bytes per second, scoped by host and Jolokia agent.",
	},

	"ReassignmentBytesOutPerSec.OneMinuteRate": &inputs.FieldInfo{
		DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.BytesPerSec,
		Desc: "One-minute moving average rate for reassignment Bytes Out Per Sec in bytes per second, scoped by host and Jolokia agent.",
	},

	"ReassignmentBytesOutPerSec.RateUnit": &inputs.FieldInfo{
		DataType: inputs.String, Type: inputs.String, Unit: inputs.NoUnit,
		Desc: "Kafka-reported rate unit label for reassignment Bytes Out Per Sec, scoped by host and Jolokia agent.",
	},

	"ReplicationBytesInPerSec.Count": &inputs.FieldInfo{
		DataType: inputs.Int, Type: inputs.Count, Unit: inputs.SizeByte,
		Desc: "Total bytes recorded for replication Bytes In Per Sec, scoped by host and Jolokia agent.",
	},

	"ReplicationBytesInPerSec.EventType": &inputs.FieldInfo{
		DataType: inputs.String, Type: inputs.String, Unit: inputs.NoUnit,
		Desc: "Kafka-reported event type label for replication Bytes In Per Sec, scoped by host and Jolokia agent.",
	},

	"ReplicationBytesInPerSec.FifteenMinuteRate": &inputs.FieldInfo{
		DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.BytesPerSec,
		Desc: "Fifteen-minute moving average rate for replication Bytes In Per Sec in bytes per second, scoped by host and Jolokia agent.",
	},

	"ReplicationBytesInPerSec.FiveMinuteRate": &inputs.FieldInfo{
		DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.BytesPerSec,
		Desc: "Five-minute moving average rate for replication Bytes In Per Sec in bytes per second, scoped by host and Jolokia agent.",
	},

	"ReplicationBytesInPerSec.MeanRate": &inputs.FieldInfo{
		DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.BytesPerSec,
		Desc: "Mean rate for replication Bytes In Per Sec in bytes per second, scoped by host and Jolokia agent.",
	},

	"ReplicationBytesInPerSec.OneMinuteRate": &inputs.FieldInfo{
		DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.BytesPerSec,
		Desc: "One-minute moving average rate for replication Bytes In Per Sec in bytes per second, scoped by host and Jolokia agent.",
	},

	"ReplicationBytesInPerSec.RateUnit": &inputs.FieldInfo{
		DataType: inputs.String, Type: inputs.String, Unit: inputs.NoUnit,
		Desc: "Kafka-reported rate unit label for replication Bytes In Per Sec, scoped by host and Jolokia agent.",
	},

	"ReplicationBytesOutPerSec.Count": &inputs.FieldInfo{
		DataType: inputs.Int, Type: inputs.Count, Unit: inputs.SizeByte,
		Desc: "Total bytes recorded for replication Bytes Out Per Sec, scoped by host and Jolokia agent.",
	},

	"ReplicationBytesOutPerSec.EventType": &inputs.FieldInfo{
		DataType: inputs.String, Type: inputs.String, Unit: inputs.NoUnit,
		Desc: "Kafka-reported event type label for replication Bytes Out Per Sec, scoped by host and Jolokia agent.",
	},

	"ReplicationBytesOutPerSec.FifteenMinuteRate": &inputs.FieldInfo{
		DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.BytesPerSec,
		Desc: "Fifteen-minute moving average rate for replication Bytes Out Per Sec in bytes per second, scoped by host and Jolokia agent.",
	},

	"ReplicationBytesOutPerSec.FiveMinuteRate": &inputs.FieldInfo{
		DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.BytesPerSec,
		Desc: "Five-minute moving average rate for replication Bytes Out Per Sec in bytes per second, scoped by host and Jolokia agent.",
	},

	"ReplicationBytesOutPerSec.MeanRate": &inputs.FieldInfo{
		DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.BytesPerSec,
		Desc: "Mean rate for replication Bytes Out Per Sec in bytes per second, scoped by host and Jolokia agent.",
	},

	"ReplicationBytesOutPerSec.OneMinuteRate": &inputs.FieldInfo{
		DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.BytesPerSec,
		Desc: "One-minute moving average rate for replication Bytes Out Per Sec in bytes per second, scoped by host and Jolokia agent.",
	},

	"ReplicationBytesOutPerSec.RateUnit": &inputs.FieldInfo{
		DataType: inputs.String, Type: inputs.String, Unit: inputs.NoUnit,
		Desc: "Kafka-reported rate unit label for replication Bytes Out Per Sec, scoped by host and Jolokia agent.",
	},

	"TotalFetchRequestsPerSec.Count": &inputs.FieldInfo{
		DataType: inputs.Int, Type: inputs.Count, Unit: inputs.NCount,
		Desc: "Cumulative value of total Fetch Requests Per Sec, scoped by host and Jolokia agent.",
	},

	"TotalFetchRequestsPerSec.EventType": &inputs.FieldInfo{
		DataType: inputs.String, Type: inputs.String, Unit: inputs.NoUnit,
		Desc: "Kafka-reported event type label for total Fetch Requests Per Sec, scoped by host and Jolokia agent.",
	},

	"TotalFetchRequestsPerSec.FifteenMinuteRate": &inputs.FieldInfo{
		DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.RequestsPerSec,
		Desc: "Fifteen-minute moving average rate for total Fetch Requests Per Sec in requests per second, scoped by host and Jolokia agent.",
	},

	"TotalFetchRequestsPerSec.FiveMinuteRate": &inputs.FieldInfo{
		DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.RequestsPerSec,
		Desc: "Five-minute moving average rate for total Fetch Requests Per Sec in requests per second, scoped by host and Jolokia agent.",
	},

	"TotalFetchRequestsPerSec.MeanRate": &inputs.FieldInfo{
		DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.RequestsPerSec,
		Desc: "Mean rate for total Fetch Requests Per Sec in requests per second, scoped by host and Jolokia agent.",
	},

	"TotalFetchRequestsPerSec.OneMinuteRate": &inputs.FieldInfo{
		DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.RequestsPerSec,
		Desc: "One-minute moving average rate for total Fetch Requests Per Sec in requests per second, scoped by host and Jolokia agent.",
	},

	"TotalFetchRequestsPerSec.RateUnit": &inputs.FieldInfo{
		DataType: inputs.String, Type: inputs.String, Unit: inputs.NoUnit,
		Desc: "Kafka-reported rate unit label for total Fetch Requests Per Sec, scoped by host and Jolokia agent.",
	},

	"TotalProduceRequestsPerSec.Count": &inputs.FieldInfo{
		DataType: inputs.Int, Type: inputs.Count, Unit: inputs.NCount,
		Desc: "Cumulative value of total Produce Requests Per Sec, scoped by host and Jolokia agent.",
	},

	"TotalProduceRequestsPerSec.EventType": &inputs.FieldInfo{
		DataType: inputs.String, Type: inputs.String, Unit: inputs.NoUnit,
		Desc: "Kafka-reported event type label for total Produce Requests Per Sec, scoped by host and Jolokia agent.",
	},

	"TotalProduceRequestsPerSec.FifteenMinuteRate": &inputs.FieldInfo{
		DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.RequestsPerSec,
		Desc: "Fifteen-minute moving average rate for total Produce Requests Per Sec in requests per second, scoped by host and Jolokia agent.",
	},

	"TotalProduceRequestsPerSec.FiveMinuteRate": &inputs.FieldInfo{
		DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.RequestsPerSec,
		Desc: "Five-minute moving average rate for total Produce Requests Per Sec in requests per second, scoped by host and Jolokia agent.",
	},

	"TotalProduceRequestsPerSec.MeanRate": &inputs.FieldInfo{
		DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.RequestsPerSec,
		Desc: "Mean rate for total Produce Requests Per Sec in requests per second, scoped by host and Jolokia agent.",
	},

	"TotalProduceRequestsPerSec.OneMinuteRate": &inputs.FieldInfo{
		DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.RequestsPerSec,
		Desc: "One-minute moving average rate for total Produce Requests Per Sec in requests per second, scoped by host and Jolokia agent.",
	},

	"TotalProduceRequestsPerSec.RateUnit": &inputs.FieldInfo{
		DataType: inputs.String, Type: inputs.String, Unit: inputs.NoUnit,
		Desc: "Kafka-reported rate unit label for total Produce Requests Per Sec, scoped by host and Jolokia agent.",
	},
}

var topicsTags = map[string]interface{}{
	"jolokia_agent_url": &inputs.TagInfo{Desc: "Jolokia agent url path"},
}

func (j *KafkaTopicsMment) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name:   "kafka_topics",
		Cat:    point.Metric,
		Desc:   "Kafka broker-wide topic JMX metrics, including cluster-level bytes, messages, fetch requests, and produce requests rates.",
		DescZh: "Kafka broker 维度 topic JMX 指标，包含集群级字节、消息、fetch 请求和 produce 请求速率。",
		Fields: topicsFields,
		Tags:   topicsTags,
	}
}

//nolint:lll
var topicFields = map[string]interface{}{
	"BytesInPerSec.Count": &inputs.FieldInfo{
		DataType: inputs.Int, Type: inputs.Count, Unit: inputs.SizeByte,
		Desc: "Total bytes recorded for bytes In Per Sec, scoped by topic.",
	},

	"BytesInPerSec.EventType": &inputs.FieldInfo{
		DataType: inputs.String, Type: inputs.String, Unit: inputs.NoUnit,
		Desc: "Kafka-reported event type label for bytes In Per Sec, scoped by topic.",
	},

	"BytesInPerSec.FifteenMinuteRate": &inputs.FieldInfo{
		DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.BytesPerSec,
		Desc: "Fifteen-minute moving average rate for bytes In Per Sec in bytes per second, scoped by topic.",
	},

	"BytesInPerSec.FiveMinuteRate": &inputs.FieldInfo{
		DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.BytesPerSec,
		Desc: "Five-minute moving average rate for bytes In Per Sec in bytes per second, scoped by topic.",
	},

	"BytesInPerSec.MeanRate": &inputs.FieldInfo{
		DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.BytesPerSec,
		Desc: "Mean number of bytes received per second for this topic, scoped by topic.",
	},

	"BytesInPerSec.OneMinuteRate": &inputs.FieldInfo{
		DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.BytesPerSec,
		Desc: "One-minute moving average rate for bytes In Per Sec in bytes per second, scoped by topic.",
	},

	"BytesInPerSec.RateUnit": &inputs.FieldInfo{
		DataType: inputs.String, Type: inputs.String, Unit: inputs.NoUnit,
		Desc: "String rate unit label reported by Kafka for BytesInPerSec, scoped by topic.",
	},

	"BytesOutPerSec.Count": &inputs.FieldInfo{
		DataType: inputs.Int, Type: inputs.Count, Unit: inputs.SizeByte,
		Desc: "Total bytes recorded for bytes Out Per Sec, scoped by topic.",
	},

	"BytesOutPerSec.EventType": &inputs.FieldInfo{
		DataType: inputs.String, Type: inputs.String, Unit: inputs.NoUnit,
		Desc: "Kafka-reported event type label for bytes Out Per Sec, scoped by topic.",
	},

	"BytesOutPerSec.FifteenMinuteRate": &inputs.FieldInfo{
		DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.BytesPerSec,
		Desc: "Fifteen-minute moving average rate for bytes Out Per Sec in bytes per second, scoped by topic.",
	},

	"BytesOutPerSec.FiveMinuteRate": &inputs.FieldInfo{
		DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.BytesPerSec,
		Desc: "Five-minute moving average rate for bytes Out Per Sec in bytes per second, scoped by topic.",
	},

	"BytesOutPerSec.MeanRate": &inputs.FieldInfo{
		DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.BytesPerSec,
		Desc: "Mean rate for bytes Out Per Sec in bytes per second, scoped by topic.",
	},

	"BytesOutPerSec.OneMinuteRate": &inputs.FieldInfo{
		DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.BytesPerSec,
		Desc: "One-minute moving average rate for bytes Out Per Sec in bytes per second, scoped by topic.",
	},

	"BytesOutPerSec.RateUnit": &inputs.FieldInfo{
		DataType: inputs.String, Type: inputs.String, Unit: inputs.NoUnit,
		Desc: "Kafka-reported rate unit label for bytes Out Per Sec, scoped by topic.",
	},

	"MessagesInPerSec.Count": &inputs.FieldInfo{
		DataType: inputs.Int, Type: inputs.Count, Unit: inputs.NCount,
		Desc: "Total number of messages In Per Sec, scoped by topic.",
	},

	"MessagesInPerSec.EventType": &inputs.FieldInfo{
		DataType: inputs.String, Type: inputs.String, Unit: inputs.NoUnit,
		Desc: "Kafka-reported event type label for messages In Per Sec, scoped by topic.",
	},

	"MessagesInPerSec.FifteenMinuteRate": &inputs.FieldInfo{
		DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount,
		Desc: "Fifteen-minute moving average rate for messages In Per Sec in events per second, scoped by topic.",
	},

	"MessagesInPerSec.FiveMinuteRate": &inputs.FieldInfo{
		DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount,
		Desc: "Five-minute moving average rate for messages In Per Sec in events per second, scoped by topic.",
	},

	"MessagesInPerSec.MeanRate": &inputs.FieldInfo{
		DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount,
		Desc: "Mean rate for messages In Per Sec in events per second, scoped by topic.",
	},

	"MessagesInPerSec.OneMinuteRate": &inputs.FieldInfo{
		DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount,
		Desc: "One-minute rate of messages received per second for this topic, scoped by topic.",
	},

	"MessagesInPerSec.RateUnit": &inputs.FieldInfo{
		DataType: inputs.String, Type: inputs.String, Unit: inputs.NoUnit,
		Desc: "Kafka-reported rate unit label for messages In Per Sec, scoped by topic.",
	},

	"TotalFetchRequestsPerSec.Count": &inputs.FieldInfo{
		DataType: inputs.Int, Type: inputs.Count, Unit: inputs.NCount,
		Desc: "Cumulative value of total Fetch Requests Per Sec, scoped by topic.",
	},

	"TotalFetchRequestsPerSec.EventType": &inputs.FieldInfo{
		DataType: inputs.String, Type: inputs.String, Unit: inputs.NoUnit,
		Desc: "Kafka-reported event type label for total Fetch Requests Per Sec, scoped by topic.",
	},

	"TotalFetchRequestsPerSec.FifteenMinuteRate": &inputs.FieldInfo{
		DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.RequestsPerSec,
		Desc: "Fifteen-minute moving average rate for total Fetch Requests Per Sec in requests per second, scoped by topic.",
	},

	"TotalFetchRequestsPerSec.FiveMinuteRate": &inputs.FieldInfo{
		DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.RequestsPerSec,
		Desc: "Five-minute moving average rate for total Fetch Requests Per Sec in requests per second, scoped by topic.",
	},

	"TotalFetchRequestsPerSec.MeanRate": &inputs.FieldInfo{
		DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.RequestsPerSec,
		Desc: "Mean rate for total Fetch Requests Per Sec in requests per second, scoped by topic.",
	},

	"TotalFetchRequestsPerSec.OneMinuteRate": &inputs.FieldInfo{
		DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.RequestsPerSec,
		Desc: "One-minute moving average rate for total Fetch Requests Per Sec in requests per second, scoped by topic.",
	},

	"TotalFetchRequestsPerSec.RateUnit": &inputs.FieldInfo{
		DataType: inputs.String, Type: inputs.String, Unit: inputs.NoUnit,
		Desc: "Kafka-reported rate unit label for total Fetch Requests Per Sec, scoped by topic.",
	},

	"TotalProduceRequestsPerSec.Count": &inputs.FieldInfo{
		DataType: inputs.Int, Type: inputs.Count, Unit: inputs.NCount,
		Desc: "Cumulative value of total Produce Requests Per Sec, scoped by topic.",
	},

	"TotalProduceRequestsPerSec.EventType": &inputs.FieldInfo{
		DataType: inputs.String, Type: inputs.String, Unit: inputs.NoUnit,
		Desc: "Kafka-reported event type label for total Produce Requests Per Sec, scoped by topic.",
	},

	"TotalProduceRequestsPerSec.FifteenMinuteRate": &inputs.FieldInfo{
		DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.RequestsPerSec,
		Desc: "Fifteen-minute moving average rate for total Produce Requests Per Sec in requests per second, scoped by topic.",
	},

	"TotalProduceRequestsPerSec.FiveMinuteRate": &inputs.FieldInfo{
		DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.RequestsPerSec,
		Desc: "Five-minute moving average rate for total Produce Requests Per Sec in requests per second, scoped by topic.",
	},

	"TotalProduceRequestsPerSec.MeanRate": &inputs.FieldInfo{
		DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.RequestsPerSec,
		Desc: "Mean rate for total Produce Requests Per Sec in requests per second, scoped by topic.",
	},

	"TotalProduceRequestsPerSec.OneMinuteRate": &inputs.FieldInfo{
		DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.RequestsPerSec,
		Desc: "One-minute moving average rate for total Produce Requests Per Sec in requests per second, scoped by topic.",
	},

	"TotalProduceRequestsPerSec.RateUnit": &inputs.FieldInfo{
		DataType: inputs.String, Type: inputs.String, Unit: inputs.NoUnit,
		Desc: "Kafka-reported rate unit label for total Produce Requests Per Sec, scoped by topic.",
	},
}

var topicTags = map[string]interface{}{
	"jolokia_agent_url": &inputs.TagInfo{Desc: "Jolokia agent url path"},
	"topic":             &inputs.TagInfo{Desc: "topic name"},
}

func (j *KafkaTopicMment) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name:   "kafka_topic",
		Cat:    point.Metric,
		Desc:   "Kafka per-topic JMX metrics, including bytes in/out, messages in, fetch requests, and produce requests rates.",
		DescZh: "Kafka topic 维度 JMX 指标，包含流入/流出字节、流入消息、fetch 请求和 produce 请求速率。",
		Tags:   topicTags,
		Fields: topicFields,
	}
}

//nolint:lll
var partitionFields = map[string]interface{}{
	"LogEndOffset": &inputs.FieldInfo{
		DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.NCount,
		Desc: "Current log end offset for this topic partition, scoped by topic and partition.",
	},

	"LogStartOffset": &inputs.FieldInfo{
		DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.NCount,
		Desc: "Current offset for log Start Offset, scoped by topic and partition.",
	},

	"NumLogSegments": &inputs.FieldInfo{
		DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.NCount,
		Desc: "Current number of num Log Segments, scoped by topic and partition.",
	},

	"Size": &inputs.FieldInfo{
		DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.SizeByte,
		Desc: "Current on-disk size in bytes for this topic partition log, scoped by topic and partition.",
	},

	"UnderReplicatedPartitions": &inputs.FieldInfo{
		DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.NCount,
		Desc: "Current number of under Replicated Partitions, scoped by topic and partition.",
	},
}

var partitionTags = map[string]interface{}{
	"jolokia_agent_url": &inputs.TagInfo{Desc: "Jolokia agent url path"},
	"partition":         &inputs.TagInfo{Desc: "partition number"},
	"topic":             &inputs.TagInfo{Desc: "topic name"},
}

func (j *KafkaPartitionMment) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name:   "kafka_partition",
		Cat:    point.Metric,
		Desc:   "Kafka per-partition JMX metrics, including log start/end offsets, segment count, log size, and under-replication state.",
		DescZh: "Kafka partition 维度 JMX 指标，包含日志起止 offset、日志段数量、日志大小和副本不足状态。",
		Tags:   partitionTags,
		Fields: partitionFields,
	}
}
