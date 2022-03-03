//nolint:lll
package kafka

import (
	"time"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"
)

type KafkaMeasurement struct {
	name   string
	tags   map[string]string
	fields map[string]interface{}
	ts     time.Time
}

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
//    refer to https://github.com/DataDog/integrations-core/blob/master/confluent_platform/metadata.csv
var connectFields = map[string]interface{}{
	"commit_id":                            &inputs.FieldInfo{DataType: inputs.String, Type: inputs.String, Unit: inputs.UnknownUnit, Desc: ""},
	"start_time_ms":                        &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.TimestampMS, Desc: ""},
	"version":                              &inputs.FieldInfo{DataType: inputs.String, Type: inputs.String, Unit: inputs.UnknownUnit, Desc: ""},
	"count":                                &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Count, Unit: inputs.NCount, Desc: ""},
	"connector_count":                      &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Count, Unit: inputs.NCount, Desc: ""},
	"connector_startup_success_percentage": &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
	"connector_startup_success_total":      &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Count, Unit: inputs.NCount, Desc: ""},
	"task_startup_failure_percentage":      &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
	"task_startup_success_percentage":      &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
	"task_startup_success_total":           &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Count, Unit: inputs.NCount, Desc: ""},
	"connector_startup_attempts_total":     &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Count, Unit: inputs.NCount, Desc: ""},
	"connector_startup_failure_percentage": &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
	"connector_startup_failure_total":      &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Count, Unit: inputs.NCount, Desc: ""},
	"task_count":                           &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Count, Unit: inputs.NCount, Desc: ""},
	"task_startup_attempts_total":          &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Count, Unit: inputs.NCount, Desc: ""},
	"task_startup_failure_total":           &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Count, Unit: inputs.NCount, Desc: ""},
	"connector_type":                       &inputs.FieldInfo{DataType: inputs.String, Type: inputs.String, Unit: inputs.UnknownUnit, Desc: ""},
	"connector_unassigned_task_count":      &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Count, Unit: inputs.NCount, Desc: ""},
	"connector_version":                    &inputs.FieldInfo{DataType: inputs.String, Type: inputs.String, Unit: inputs.UnknownUnit, Desc: ""},
	"status":                               &inputs.FieldInfo{DataType: inputs.String, Type: inputs.String, Unit: inputs.UnknownUnit, Desc: ""},
	"connector_class":                      &inputs.FieldInfo{DataType: inputs.String, Type: inputs.String, Unit: inputs.UnknownUnit, Desc: ""},
	"connector_failed_task_count":          &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Count, Unit: inputs.NCount, Desc: ""},
	"connector_paused_task_count":          &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Count, Unit: inputs.NCount, Desc: ""},
	"connector_total_task_count":           &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Count, Unit: inputs.NCount, Desc: ""},
	"connector_destroyed_task_count":       &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Count, Unit: inputs.NCount, Desc: ""},
	"connector_restarting_task_count":      &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Count, Unit: inputs.NCount, Desc: ""},
	"connector_running_task_count":         &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Count, Unit: inputs.NCount, Desc: ""},
	"total_records_skipped":                &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Count, Unit: inputs.NCount, Desc: ""},
	"total_retries":                        &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Count, Unit: inputs.NCount, Desc: ""},
	"offset_commit_failure_percentage":     &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
	"running_ratio":                        &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
	"source_record_poll_total":             &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Count, Unit: inputs.NCount, Desc: ""},
	"total_record_failures":                &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Count, Unit: inputs.NCount, Desc: ""},
	"last_error_timestamp":                 &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.TimestampMS, Desc: "The epoch timestamp when this task last encountered an error in millisecond."},
	"offset_commit_success_percentage":     &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
	"source_record_poll_rate":              &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
	"source_record_write_rate":             &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
	"total_errors_logged":                  &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Count, Unit: inputs.NCount, Desc: ""},
	"offset_commit_max_time_ms":            &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS, Desc: ""},
	"source_record_active_count":           &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Count, Unit: inputs.NCount, Desc: ""},
	"source_record_write_total":            &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Count, Unit: inputs.NCount, Desc: ""},
	"total_record_errors":                  &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Count, Unit: inputs.NCount, Desc: ""},
	"deadletterqueue_produce_failures":     &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Count, Unit: inputs.NCount, Desc: ""},
	"deadletterqueue_produce_requests":     &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Count, Unit: inputs.NCount, Desc: ""},
	"offset_commit_avg_time_ms":            &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS, Desc: ""},
	"pause_ratio":                          &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
	"offset_commit_skip_rate":              &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
	"sink_record_send_total":               &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Count, Unit: inputs.NCount, Desc: ""},
	"offset_commit_seq_no":                 &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
	"put_batch_avg_time_ms":                &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS, Desc: ""},
	"sink_record_send_rate":                &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
	"batch_size_avg":                       &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
	"offset_commit_completion_rate":        &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
	"offset_commit_skip_total":             &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Count, Unit: inputs.NCount, Desc: ""},
	"sink_record_read_rate":                &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
	"batch_size_max":                       &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
	"partition_count":                      &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Count, Unit: inputs.NCount, Desc: ""},
	"sink_record_active_count":             &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Count, Unit: inputs.NCount, Desc: ""},
	"sink_record_read_total":               &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Count, Unit: inputs.NCount, Desc: ""},
	"put_batch_max_time_ms":                &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS, Desc: ""},
	"offset_commit_completion_total":       &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Count, Unit: inputs.NCount, Desc: ""},
	"sink_record_active_count_avg":         &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Count, Unit: inputs.NCount, Desc: ""},
	"sink_record_active_count_max":         &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Count, Unit: inputs.NCount, Desc: ""},
}

var connectTags = map[string]interface{}{
	"jolokia_agent_url": inputs.TagInfo{Desc: "jolokia agent url path"},
	"type":              inputs.TagInfo{Desc: "metric type"},
	"client_id":         inputs.TagInfo{Desc: "client id"},
	"task":              inputs.TagInfo{Desc: "task"},
	"connector":         inputs.TagInfo{Desc: "connector"},
}

func (j *KafkaConnectMment) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name:   "kafka_connect",
		Fields: connectFields,
		Tags:   connectTags,
		Desc:   "该指标集需在 Connector 实例上采集",
	}
}

func (j *KafkaConnectMment) LineProto() (*io.Point, error) {
	return io.NewPoint(j.name, j.tags, j.fields, &io.PointOption{Category: datakit.Metric, Time: j.ts})
}

var producerFields = map[string]interface{}{
	"io_wait_ratio":                             &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
	"io_waittime_total":                         &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.DurationNS, Desc: ""},
	"record_send_total":                         &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Count, Unit: inputs.NCount, Desc: ""},
	"io_time_ns_avg":                            &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationNS, Desc: ""},
	"buffer_exhausted_rate":                     &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
	"buffer_total_bytes":                        &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
	"incoming_byte_total":                       &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
	"successful_reauthentication_total":         &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
	"record_retry_total":                        &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
	"buffer_exhausted_total":                    &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
	"failed_reauthentication_rate":              &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
	"request_rate":                              &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
	"io_ratio":                                  &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
	"io_wait_time_ns_avg":                       &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.DurationNS, Desc: ""},
	"metadata_age":                              &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
	"network_io_rate":                           &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
	"connection_close_rate":                     &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
	"connection_count":                          &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Count, Unit: inputs.NCount, Desc: ""},
	"produce_throttle_time_max":                 &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.DurationMS, Desc: ""},
	"successful_authentication_total":           &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Count, Unit: inputs.NCount, Desc: ""},
	"request_total":                             &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Count, Unit: inputs.NCount, Desc: ""},
	"successful_authentication_no_reauth_total": &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Count, Unit: inputs.NCount, Desc: ""},
	"bufferpool_wait_ratio":                     &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
	"incoming_byte_rate":                        &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
	"request_size_avg":                          &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
	"select_total":                              &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Count, Unit: inputs.NCount, Desc: ""},
	"connection_close_total":                    &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Count, Unit: inputs.NCount, Desc: ""},
	"network_io_total":                          &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Count, Unit: inputs.NCount, Desc: ""},
	"response_rate":                             &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
	"select_rate":                               &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
	"record_send_rate":                          &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
	"successful_authentication_rate":            &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
	"record_error_rate":                         &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
	"response_total":                            &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Count, Unit: inputs.NCount, Desc: ""},
	"produce_throttle_time_avg":                 &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.DurationMS, Desc: ""},
	"record_error_total":                        &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
	"requests_in_flight":                        &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Count, Unit: inputs.NCount, Desc: ""},
	"failed_authentication_rate":                &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
	"buffer_available_bytes":                    &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
	"iotime_total":                              &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.DurationNS, Desc: ""},
	"successful_reauthentication_rate":          &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
	"failed_authentication_total":               &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Count, Unit: inputs.NCount, Desc: ""},
	"failed_reauthentication_total":             &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Count, Unit: inputs.NCount, Desc: ""},
	"record_retry_rate":                         &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
	"batch_split_total":                         &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
	"connection_creation_total":                 &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Count, Unit: inputs.NCount, Desc: ""},
	"request_size_max":                          &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
	"bufferpool_wait_time_total":                &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Count, Unit: inputs.NCount, Desc: ""},
	"waiting_threads":                           &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Count, Unit: inputs.NCount, Desc: ""},
	"batch_split_rate":                          &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
	"connection_creation_rate":                  &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
	"outgoing_byte_rate":                        &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
	"outgoing_byte_total":                       &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Count, Unit: inputs.NCount, Desc: ""},
	"count":                                     &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Count, Unit: inputs.NCount, Desc: ""},
	"commit_id":                                 &inputs.FieldInfo{DataType: inputs.String, Type: inputs.String, Unit: inputs.UnknownUnit, Desc: ""},
	"start_time_ms":                             &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.TimestampMS, Desc: ""},
	"version":                                   &inputs.FieldInfo{DataType: inputs.String, Type: inputs.String, Unit: inputs.UnknownUnit, Desc: ""},
}

var producerTags = map[string]interface{}{
	"jolokia_agent_url": inputs.TagInfo{Desc: "jolokia agent url path"},
	"type":              inputs.TagInfo{Desc: "metric type"},
	"client_id":         inputs.TagInfo{Desc: "client id"},
}

func (j *KafkaProducerMment) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name:   "kafka_producer",
		Fields: producerFields,
		Tags:   producerTags,
		Desc:   "该指标集需在 Producer 实例上采集",
	}
}

func (j *KafkaProducerMment) LineProto() (*io.Point, error) {
	return io.NewPoint(j.name, j.tags, j.fields, &io.PointOption{Category: datakit.Metric, Time: j.ts})
}

var consumerFields = map[string]interface{}{
	"count":                                     &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Count, Unit: inputs.NCount, Desc: ""},
	"bytes_consumed_rate":                       &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
	"bytes_consumed_total":                      &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Count, Unit: inputs.NCount, Desc: ""},
	"fetch_latency_avg":                         &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
	"fetch_throttle_time_avg":                   &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
	"fetch_total":                               &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
	"records_consumed_rate":                     &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
	"fetch_latency_max":                         &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
	"fetch_rate":                                &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
	"fetch_throttle_time_max":                   &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
	"records_consumed_total":                    &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
	"commit_total":                              &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Count, Unit: inputs.NCount, Desc: ""},
	"join_rate":                                 &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
	"join_total":                                &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Count, Unit: inputs.NCount, Desc: ""},
	"rebalance_rate_per_hour":                   &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
	"rebalance_total":                           &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Count, Unit: inputs.NCount, Desc: ""},
	"failed_rebalance_rate_per_hour":            &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
	"heartbeat_rate":                            &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
	"last_rebalance_seconds_ago":                &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
	"sync_rate":                                 &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
	"sync_total":                                &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Count, Unit: inputs.UnknownUnit, Desc: ""},
	"assigned_partitions":                       &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Count, Unit: inputs.NCount, Desc: ""},
	"commit_rate":                               &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
	"heartbeat_response_time_max":               &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
	"last_heartbeat_seconds_ago":                &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
	"failed_rebalance_total":                    &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
	"heartbeat_total":                           &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Count, Unit: inputs.NCount, Desc: ""},
	"rebalance_latency_total":                   &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Count, Unit: inputs.NCount, Desc: ""},
	"commit_id":                                 &inputs.FieldInfo{DataType: inputs.String, Type: inputs.String, Unit: inputs.UnknownUnit, Desc: ""},
	"start_time_ms":                             &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.TimestampMS, Desc: ""},
	"version":                                   &inputs.FieldInfo{DataType: inputs.String, Type: inputs.String, Unit: inputs.UnknownUnit, Desc: ""},
	"io_waittime_total":                         &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.DurationNS, Desc: ""},
	"connection_creation_rate":                  &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
	"connection_creation_total":                 &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Count, Unit: inputs.NCount, Desc: ""},
	"successful_authentication_rate":            &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
	"failed_authentication_rate":                &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
	"failed_authentication_total":               &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Count, Unit: inputs.NCount, Desc: ""},
	"failed_reauthentication_total":             &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Count, Unit: inputs.NCount, Desc: ""},
	"io_wait_time_ns_avg":                       &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.DurationNS, Desc: ""},
	"network_io_total":                          &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Count, Unit: inputs.NCount, Desc: ""},
	"request_size_avg":                          &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
	"successful_authentication_no_reauth_total": &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Count, Unit: inputs.NCount, Desc: ""},
	"connection_close_rate":                     &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
	"io_time_ns_avg":                            &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.DurationNS, Desc: ""},
	"outgoing_byte_total":                       &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
	"request_size_max":                          &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
	"response_rate":                             &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
	"successful_authentication_total":           &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Count, Unit: inputs.NCount, Desc: ""},
	"successful_reauthentication_total":         &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Count, Unit: inputs.NCount, Desc: ""},
	"connection_count":                          &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Count, Unit: inputs.NCount, Desc: ""},
	"io_ratio":                                  &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
	"request_total":                             &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Count, Unit: inputs.NCount, Desc: ""},
	"response_total":                            &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Count, Unit: inputs.NCount, Desc: ""},
	"incoming_byte_total":                       &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
	"iotime_total":                              &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Count, Unit: inputs.NCount, Desc: ""},
	"last_poll_seconds_ago":                     &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Count, Unit: inputs.NCount, Desc: ""},
	"outgoing_byte_rate":                        &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
	"connection_close_total":                    &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Count, Unit: inputs.NCount, Desc: ""},
	"failed_reauthentication_rate":              &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
	"incoming_byte_rate":                        &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
	"network_io_rate":                           &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
	"select_rate":                               &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
	"io_wait_ratio":                             &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
	"request_rate":                              &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
	"select_total":                              &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Count, Unit: inputs.NCount, Desc: ""},
	"successful_reauthentication_rate":          &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
}

var consumerTags = map[string]interface{}{
	"jolokia_agent_url": inputs.TagInfo{Desc: "jolokia agent url path"},
	"type":              inputs.TagInfo{Desc: "metric type"},
	"client_id":         inputs.TagInfo{Desc: "client id"},
}

func (j *KafkaConsumerMment) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name:   "kafka_consumer",
		Fields: consumerFields,
		Tags:   consumerTags,
		Desc:   "该指标集需在 Consumer 实例上采集",
	}
}

func (j *KafkaConsumerMment) LineProto() (*io.Point, error) {
	return io.NewPoint(j.name, j.tags, j.fields, &io.PointOption{Category: datakit.Metric, Time: j.ts})
}

var logFields = map[string]interface{}{
	"OfflineLogDirectoryCount":       &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Count, Unit: inputs.NCount, Desc: ""},
	"cleaner_recopy_percent":         &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
	"max_compaction_delay_secs":      &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.DurationSecond, Desc: ""},
	"max_clean_time_secs":            &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.DurationSecond, Desc: ""},
	"DeadThreadCount":                &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Count, Unit: inputs.NCount, Desc: ""},
	"max_buffer_utilization_percent": &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
}

var logTags = map[string]interface{}{
	"jolokia_agent_url": inputs.TagInfo{Desc: "jolokia agent url path"},
	"type":              inputs.TagInfo{Desc: "metric type"},
}

func (j *KafkaLogMment) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name:   "kafka_log",
		Fields: logFields,
		Tags:   logTags,
	}
}

func (j *KafkaLogMment) LineProto() (*io.Point, error) {
	return io.NewPoint(j.name, j.tags, j.fields, &io.PointOption{Category: datakit.Metric, Time: j.ts})
}

var networkFields = map[string]interface{}{
	"NetworkProcessorAvgIdlePercent":            &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
	"ExpiredConnectionsKilledCount":             &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Count, Unit: inputs.NCount, Desc: ""},
	"ControlPlaneExpiredConnectionsKilledCount": &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Count, Unit: inputs.NCount, Desc: ""},
	"MemoryPoolAvailable":                       &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Count, Unit: inputs.NCount, Desc: ""},
	"MemoryPoolUsed":                            &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Count, Unit: inputs.NCount, Desc: ""},
}

var networkTags = map[string]interface{}{
	"jolokia_agent_url": inputs.TagInfo{Desc: "jolokia agent url path"},
	"type":              inputs.TagInfo{Desc: "metric type"},
}

func (j *KafkaNetworkMment) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name:   "kafka_network",
		Fields: networkFields,
		Tags:   networkTags,
	}
}

func (j *KafkaNetworkMment) LineProto() (*io.Point, error) {
	return io.NewPoint(j.name, j.tags, j.fields, &io.PointOption{Category: datakit.Metric, Time: j.ts})
}

var requestHandlerFields = map[string]interface{}{
	"RequestHandlerAvgIdlePercent.RateUnit":          &inputs.FieldInfo{DataType: inputs.String, Type: inputs.String, Unit: inputs.String, Desc: ""},
	"RequestHandlerAvgIdlePercent.EventType":         &inputs.FieldInfo{DataType: inputs.String, Type: inputs.String, Unit: inputs.String, Desc: ""},
	"RequestHandlerAvgIdlePercent.FifteenMinuteRate": &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
	"RequestHandlerAvgIdlePercent.OneMinuteRate":     &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
	"RequestHandlerAvgIdlePercent.Count":             &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Count, Unit: inputs.NCount, Desc: ""},
	"RequestHandlerAvgIdlePercent.FiveMinuteRate":    &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
	"RequestHandlerAvgIdlePercent.MeanRate":          &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
}

var requestHandlerTags = map[string]interface{}{
	"jolokia_agent_url": inputs.TagInfo{Desc: "jolokia agent url path"},
}

func (j *KafkaRequestHandlerMment) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name:   "kafka_request_handler",
		Fields: requestHandlerFields,
		Tags:   requestHandlerTags,
	}
}

func (j *KafkaRequestHandlerMment) LineProto() (*io.Point, error) {
	return io.NewPoint(j.name, j.tags, j.fields, &io.PointOption{Category: datakit.Metric, Time: j.ts})
}

var zooKeeperFields = map[string]interface{}{
	"ZooKeeperRequestLatencyMs.Mean":            &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS, Desc: ""},
	"ZooKeeperRequestLatencyMs.98thPercentile":  &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS, Desc: ""},
	"ZooKeeperRequestLatencyMs.75thPercentile":  &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS, Desc: ""},
	"ZooKeeperRequestLatencyMs.95thPercentile":  &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS, Desc: ""},
	"ZooKeeperRequestLatencyMs.Min":             &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS, Desc: ""},
	"ZooKeeperRequestLatencyMs.999thPercentile": &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS, Desc: ""},
	"ZooKeeperRequestLatencyMs.Max":             &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS, Desc: ""},
	"ZooKeeperRequestLatencyMs.StdDev":          &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS, Desc: ""},
	"ZooKeeperRequestLatencyMs.99thPercentile":  &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS, Desc: ""},
	"ZooKeeperRequestLatencyMs.Count":           &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Count, Unit: inputs.NCount, Desc: ""},
	"ZooKeeperRequestLatencyMs.50thPercentile":  &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS, Desc: ""},
}

var zooKeeperTags = map[string]interface{}{
	"jolokia_agent_url": inputs.TagInfo{Desc: "jolokia agent url path"},
}

func (j *KafkaZooKeeperMment) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name:   "kafka_zookeeper",
		Fields: zooKeeperFields,
		Tags:   zooKeeperTags,
	}
}

func (j *KafkaZooKeeperMment) LineProto() (*io.Point, error) {
	return io.NewPoint(j.name, j.tags, j.fields, &io.PointOption{Category: datakit.Metric, Time: j.ts})
}

func (j *KafkaMeasurement) LineProto() (*io.Point, error) {
	return io.NewPoint(j.name, j.tags, j.fields, &io.PointOption{Category: datakit.Metric, Time: j.ts})
}

var controllerFields = map[string]interface{}{
	"AutoLeaderBalanceRateAndTimeMs.50thPercentile":    &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
	"AutoLeaderBalanceRateAndTimeMs.75thPercentile":    &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
	"AutoLeaderBalanceRateAndTimeMs.95thPercentile":    &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
	"AutoLeaderBalanceRateAndTimeMs.98thPercentile":    &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
	"AutoLeaderBalanceRateAndTimeMs.999thPercentile":   &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
	"AutoLeaderBalanceRateAndTimeMs.99thPercentile":    &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
	"AutoLeaderBalanceRateAndTimeMs.FifteenMinuteRate": &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
	"AutoLeaderBalanceRateAndTimeMs.FiveMinuteRate":    &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
	"AutoLeaderBalanceRateAndTimeMs.Max":               &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
	"AutoLeaderBalanceRateAndTimeMs.Mean":              &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
	"AutoLeaderBalanceRateAndTimeMs.MeanRate":          &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
	"AutoLeaderBalanceRateAndTimeMs.Min":               &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
	"AutoLeaderBalanceRateAndTimeMs.OneMinuteRate":     &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
	"AutoLeaderBalanceRateAndTimeMs.StdDev":            &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
	"AutoLeaderBalanceRateAndTimeMs.Count":             &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Count, Unit: inputs.NCount, Desc: ""},
	"AutoLeaderBalanceRateAndTimeMs.LatencyUnit":       &inputs.FieldInfo{DataType: inputs.String, Type: inputs.String, Unit: inputs.UnknownUnit, Desc: ""},
	"AutoLeaderBalanceRateAndTimeMs.EventType":         &inputs.FieldInfo{DataType: inputs.String, Type: inputs.String, Unit: inputs.UnknownUnit, Desc: ""},
	"AutoLeaderBalanceRateAndTimeMs.RateUnit":          &inputs.FieldInfo{DataType: inputs.String, Type: inputs.String, Unit: inputs.UnknownUnit, Desc: ""},

	"ControlledShutdownRateAndTimeMs.50thPercentile":    &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
	"ControlledShutdownRateAndTimeMs.75thPercentile":    &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
	"ControlledShutdownRateAndTimeMs.95thPercentile":    &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
	"ControlledShutdownRateAndTimeMs.98thPercentile":    &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
	"ControlledShutdownRateAndTimeMs.999thPercentile":   &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
	"ControlledShutdownRateAndTimeMs.99thPercentile":    &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
	"ControlledShutdownRateAndTimeMs.FifteenMinuteRate": &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
	"ControlledShutdownRateAndTimeMs.FiveMinuteRate":    &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
	"ControlledShutdownRateAndTimeMs.Max":               &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
	"ControlledShutdownRateAndTimeMs.Mean":              &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
	"ControlledShutdownRateAndTimeMs.MeanRate":          &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
	"ControlledShutdownRateAndTimeMs.Min":               &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
	"ControlledShutdownRateAndTimeMs.OneMinuteRate":     &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
	"ControlledShutdownRateAndTimeMs.StdDev":            &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
	"ControlledShutdownRateAndTimeMs.Count":             &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Count, Unit: inputs.NCount, Desc: ""},
	"ControlledShutdownRateAndTimeMs.LatencyUnit":       &inputs.FieldInfo{DataType: inputs.String, Type: inputs.String, Unit: inputs.UnknownUnit, Desc: ""},
	"ControlledShutdownRateAndTimeMs.EventType":         &inputs.FieldInfo{DataType: inputs.String, Type: inputs.String, Unit: inputs.UnknownUnit, Desc: ""},
	"ControlledShutdownRateAndTimeMs.RateUnit":          &inputs.FieldInfo{DataType: inputs.String, Type: inputs.String, Unit: inputs.UnknownUnit, Desc: ""},

	"ControllerChangeRateAndTimeMs.50thPercentile":    &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
	"ControllerChangeRateAndTimeMs.75thPercentile":    &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
	"ControllerChangeRateAndTimeMs.95thPercentile":    &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
	"ControllerChangeRateAndTimeMs.98thPercentile":    &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
	"ControllerChangeRateAndTimeMs.999thPercentile":   &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
	"ControllerChangeRateAndTimeMs.99thPercentile":    &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
	"ControllerChangeRateAndTimeMs.FifteenMinuteRate": &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
	"ControllerChangeRateAndTimeMs.FiveMinuteRate":    &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
	"ControllerChangeRateAndTimeMs.Max":               &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
	"ControllerChangeRateAndTimeMs.Mean":              &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
	"ControllerChangeRateAndTimeMs.MeanRate":          &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
	"ControllerChangeRateAndTimeMs.Min":               &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
	"ControllerChangeRateAndTimeMs.OneMinuteRate":     &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
	"ControllerChangeRateAndTimeMs.StdDev":            &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
	"ControllerChangeRateAndTimeMs.Count":             &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Count, Unit: inputs.NCount, Desc: ""},
	"ControllerChangeRateAndTimeMs.LatencyUnit":       &inputs.FieldInfo{DataType: inputs.String, Type: inputs.String, Unit: inputs.UnknownUnit, Desc: ""},
	"ControllerChangeRateAndTimeMs.EventType":         &inputs.FieldInfo{DataType: inputs.String, Type: inputs.String, Unit: inputs.UnknownUnit, Desc: ""},
	"ControllerChangeRateAndTimeMs.RateUnit":          &inputs.FieldInfo{DataType: inputs.String, Type: inputs.String, Unit: inputs.UnknownUnit, Desc: ""},

	"ControllerShutdownRateAndTimeMs.50thPercentile":    &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
	"ControllerShutdownRateAndTimeMs.75thPercentile":    &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
	"ControllerShutdownRateAndTimeMs.95thPercentile":    &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
	"ControllerShutdownRateAndTimeMs.98thPercentile":    &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
	"ControllerShutdownRateAndTimeMs.999thPercentile":   &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
	"ControllerShutdownRateAndTimeMs.99thPercentile":    &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
	"ControllerShutdownRateAndTimeMs.FifteenMinuteRate": &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
	"ControllerShutdownRateAndTimeMs.FiveMinuteRate":    &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
	"ControllerShutdownRateAndTimeMs.Max":               &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
	"ControllerShutdownRateAndTimeMs.Mean":              &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
	"ControllerShutdownRateAndTimeMs.MeanRate":          &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
	"ControllerShutdownRateAndTimeMs.Min":               &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
	"ControllerShutdownRateAndTimeMs.OneMinuteRate":     &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
	"ControllerShutdownRateAndTimeMs.StdDev":            &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
	"ControllerShutdownRateAndTimeMs.Count":             &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Count, Unit: inputs.NCount, Desc: ""},
	"ControllerShutdownRateAndTimeMs.LatencyUnit":       &inputs.FieldInfo{DataType: inputs.String, Type: inputs.String, Unit: inputs.UnknownUnit, Desc: ""},
	"ControllerShutdownRateAndTimeMs.EventType":         &inputs.FieldInfo{DataType: inputs.String, Type: inputs.String, Unit: inputs.UnknownUnit, Desc: ""},
	"ControllerShutdownRateAndTimeMs.RateUnit":          &inputs.FieldInfo{DataType: inputs.String, Type: inputs.String, Unit: inputs.UnknownUnit, Desc: ""},

	"IsrChangeRateAndTimeMs.50thPercentile":    &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
	"IsrChangeRateAndTimeMs.75thPercentile":    &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
	"IsrChangeRateAndTimeMs.95thPercentile":    &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
	"IsrChangeRateAndTimeMs.98thPercentile":    &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
	"IsrChangeRateAndTimeMs.999thPercentile":   &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
	"IsrChangeRateAndTimeMs.99thPercentile":    &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
	"IsrChangeRateAndTimeMs.FifteenMinuteRate": &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
	"IsrChangeRateAndTimeMs.FiveMinuteRate":    &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
	"IsrChangeRateAndTimeMs.Max":               &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
	"IsrChangeRateAndTimeMs.Mean":              &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
	"IsrChangeRateAndTimeMs.MeanRate":          &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
	"IsrChangeRateAndTimeMs.Min":               &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
	"IsrChangeRateAndTimeMs.OneMinuteRate":     &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
	"IsrChangeRateAndTimeMs.StdDev":            &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
	"IsrChangeRateAndTimeMs.Count":             &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Count, Unit: inputs.NCount, Desc: ""},
	"IsrChangeRateAndTimeMs.LatencyUnit":       &inputs.FieldInfo{DataType: inputs.String, Type: inputs.String, Unit: inputs.UnknownUnit, Desc: ""},
	"IsrChangeRateAndTimeMs.EventType":         &inputs.FieldInfo{DataType: inputs.String, Type: inputs.String, Unit: inputs.UnknownUnit, Desc: ""},
	"IsrChangeRateAndTimeMs.RateUnit":          &inputs.FieldInfo{DataType: inputs.String, Type: inputs.String, Unit: inputs.UnknownUnit, Desc: ""},

	"LeaderAndIsrResponseReceivedRateAndTimeMs.50thPercentile":    &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
	"LeaderAndIsrResponseReceivedRateAndTimeMs.75thPercentile":    &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
	"LeaderAndIsrResponseReceivedRateAndTimeMs.95thPercentile":    &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
	"LeaderAndIsrResponseReceivedRateAndTimeMs.98thPercentile":    &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
	"LeaderAndIsrResponseReceivedRateAndTimeMs.999thPercentile":   &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
	"LeaderAndIsrResponseReceivedRateAndTimeMs.99thPercentile":    &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
	"LeaderAndIsrResponseReceivedRateAndTimeMs.FifteenMinuteRate": &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
	"LeaderAndIsrResponseReceivedRateAndTimeMs.FiveMinuteRate":    &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
	"LeaderAndIsrResponseReceivedRateAndTimeMs.Max":               &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
	"LeaderAndIsrResponseReceivedRateAndTimeMs.Mean":              &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
	"LeaderAndIsrResponseReceivedRateAndTimeMs.MeanRate":          &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
	"LeaderAndIsrResponseReceivedRateAndTimeMs.Min":               &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
	"LeaderAndIsrResponseReceivedRateAndTimeMs.OneMinuteRate":     &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
	"LeaderAndIsrResponseReceivedRateAndTimeMs.StdDev":            &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
	"LeaderAndIsrResponseReceivedRateAndTimeMs.Count":             &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Count, Unit: inputs.NCount, Desc: ""},
	"LeaderAndIsrResponseReceivedRateAndTimeMs.LatencyUnit":       &inputs.FieldInfo{DataType: inputs.String, Type: inputs.String, Unit: inputs.UnknownUnit, Desc: ""},
	"LeaderAndIsrResponseReceivedRateAndTimeMs.EventType":         &inputs.FieldInfo{DataType: inputs.String, Type: inputs.String, Unit: inputs.UnknownUnit, Desc: ""},
	"LeaderAndIsrResponseReceivedRateAndTimeMs.RateUnit":          &inputs.FieldInfo{DataType: inputs.String, Type: inputs.String, Unit: inputs.UnknownUnit, Desc: ""},

	"LeaderElectionRateAndTimeMs.50thPercentile":    &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
	"LeaderElectionRateAndTimeMs.75thPercentile":    &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
	"LeaderElectionRateAndTimeMs.95thPercentile":    &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
	"LeaderElectionRateAndTimeMs.98thPercentile":    &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
	"LeaderElectionRateAndTimeMs.999thPercentile":   &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
	"LeaderElectionRateAndTimeMs.99thPercentile":    &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
	"LeaderElectionRateAndTimeMs.FifteenMinuteRate": &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
	"LeaderElectionRateAndTimeMs.FiveMinuteRate":    &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
	"LeaderElectionRateAndTimeMs.Max":               &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
	"LeaderElectionRateAndTimeMs.Mean":              &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
	"LeaderElectionRateAndTimeMs.MeanRate":          &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
	"LeaderElectionRateAndTimeMs.Min":               &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
	"LeaderElectionRateAndTimeMs.OneMinuteRate":     &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
	"LeaderElectionRateAndTimeMs.StdDev":            &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
	"LeaderElectionRateAndTimeMs.Count":             &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Count, Unit: inputs.NCount, Desc: ""},
	"LeaderElectionRateAndTimeMs.LatencyUnit":       &inputs.FieldInfo{DataType: inputs.String, Type: inputs.String, Unit: inputs.UnknownUnit, Desc: ""},
	"LeaderElectionRateAndTimeMs.EventType":         &inputs.FieldInfo{DataType: inputs.String, Type: inputs.String, Unit: inputs.UnknownUnit, Desc: ""},
	"LeaderElectionRateAndTimeMs.RateUnit":          &inputs.FieldInfo{DataType: inputs.String, Type: inputs.String, Unit: inputs.UnknownUnit, Desc: ""},

	"ListPartitionReassignmentRateAndTimeMs.50thPercentile":    &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
	"ListPartitionReassignmentRateAndTimeMs.75thPercentile":    &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
	"ListPartitionReassignmentRateAndTimeMs.95thPercentile":    &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
	"ListPartitionReassignmentRateAndTimeMs.98thPercentile":    &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
	"ListPartitionReassignmentRateAndTimeMs.999thPercentile":   &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
	"ListPartitionReassignmentRateAndTimeMs.99thPercentile":    &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
	"ListPartitionReassignmentRateAndTimeMs.FifteenMinuteRate": &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
	"ListPartitionReassignmentRateAndTimeMs.FiveMinuteRate":    &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
	"ListPartitionReassignmentRateAndTimeMs.Max":               &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
	"ListPartitionReassignmentRateAndTimeMs.Mean":              &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
	"ListPartitionReassignmentRateAndTimeMs.MeanRate":          &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
	"ListPartitionReassignmentRateAndTimeMs.Min":               &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
	"ListPartitionReassignmentRateAndTimeMs.OneMinuteRate":     &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
	"ListPartitionReassignmentRateAndTimeMs.StdDev":            &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
	"ListPartitionReassignmentRateAndTimeMs.Count":             &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Count, Unit: inputs.NCount, Desc: ""},
	"ListPartitionReassignmentRateAndTimeMs.LatencyUnit":       &inputs.FieldInfo{DataType: inputs.String, Type: inputs.String, Unit: inputs.UnknownUnit, Desc: ""},
	"ListPartitionReassignmentRateAndTimeMs.EventType":         &inputs.FieldInfo{DataType: inputs.String, Type: inputs.String, Unit: inputs.UnknownUnit, Desc: ""},
	"ListPartitionReassignmentRateAndTimeMs.RateUnit":          &inputs.FieldInfo{DataType: inputs.String, Type: inputs.String, Unit: inputs.UnknownUnit, Desc: ""},

	"LogDirChangeRateAndTimeMs.50thPercentile":    &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
	"LogDirChangeRateAndTimeMs.75thPercentile":    &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
	"LogDirChangeRateAndTimeMs.95thPercentile":    &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
	"LogDirChangeRateAndTimeMs.98thPercentile":    &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
	"LogDirChangeRateAndTimeMs.999thPercentile":   &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
	"LogDirChangeRateAndTimeMs.99thPercentile":    &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
	"LogDirChangeRateAndTimeMs.FifteenMinuteRate": &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
	"LogDirChangeRateAndTimeMs.FiveMinuteRate":    &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
	"LogDirChangeRateAndTimeMs.Max":               &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
	"LogDirChangeRateAndTimeMs.Mean":              &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
	"LogDirChangeRateAndTimeMs.MeanRate":          &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
	"LogDirChangeRateAndTimeMs.Min":               &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
	"LogDirChangeRateAndTimeMs.OneMinuteRate":     &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
	"LogDirChangeRateAndTimeMs.StdDev":            &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
	"LogDirChangeRateAndTimeMs.Count":             &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Count, Unit: inputs.NCount, Desc: ""},
	"LogDirChangeRateAndTimeMs.LatencyUnit":       &inputs.FieldInfo{DataType: inputs.String, Type: inputs.String, Unit: inputs.UnknownUnit, Desc: ""},
	"LogDirChangeRateAndTimeMs.EventType":         &inputs.FieldInfo{DataType: inputs.String, Type: inputs.String, Unit: inputs.UnknownUnit, Desc: ""},
	"LogDirChangeRateAndTimeMs.RateUnit":          &inputs.FieldInfo{DataType: inputs.String, Type: inputs.String, Unit: inputs.UnknownUnit, Desc: ""},

	"ManualLeaderBalanceRateAndTimeMs.50thPercentile":    &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
	"ManualLeaderBalanceRateAndTimeMs.75thPercentile":    &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
	"ManualLeaderBalanceRateAndTimeMs.95thPercentile":    &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
	"ManualLeaderBalanceRateAndTimeMs.98thPercentile":    &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
	"ManualLeaderBalanceRateAndTimeMs.999thPercentile":   &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
	"ManualLeaderBalanceRateAndTimeMs.99thPercentile":    &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
	"ManualLeaderBalanceRateAndTimeMs.FifteenMinuteRate": &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
	"ManualLeaderBalanceRateAndTimeMs.FiveMinuteRate":    &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
	"ManualLeaderBalanceRateAndTimeMs.Max":               &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
	"ManualLeaderBalanceRateAndTimeMs.Mean":              &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
	"ManualLeaderBalanceRateAndTimeMs.MeanRate":          &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
	"ManualLeaderBalanceRateAndTimeMs.Min":               &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
	"ManualLeaderBalanceRateAndTimeMs.OneMinuteRate":     &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
	"ManualLeaderBalanceRateAndTimeMs.StdDev":            &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
	"ManualLeaderBalanceRateAndTimeMs.Count":             &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Count, Unit: inputs.NCount, Desc: ""},
	"ManualLeaderBalanceRateAndTimeMs.LatencyUnit":       &inputs.FieldInfo{DataType: inputs.String, Type: inputs.String, Unit: inputs.UnknownUnit, Desc: ""},
	"ManualLeaderBalanceRateAndTimeMs.EventType":         &inputs.FieldInfo{DataType: inputs.String, Type: inputs.String, Unit: inputs.UnknownUnit, Desc: ""},
	"ManualLeaderBalanceRateAndTimeMs.RateUnit":          &inputs.FieldInfo{DataType: inputs.String, Type: inputs.String, Unit: inputs.UnknownUnit, Desc: ""},

	"PartitionReassignmentRateAndTimeMs.50thPercentile":    &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
	"PartitionReassignmentRateAndTimeMs.75thPercentile":    &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
	"PartitionReassignmentRateAndTimeMs.95thPercentile":    &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
	"PartitionReassignmentRateAndTimeMs.98thPercentile":    &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
	"PartitionReassignmentRateAndTimeMs.999thPercentile":   &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
	"PartitionReassignmentRateAndTimeMs.99thPercentile":    &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
	"PartitionReassignmentRateAndTimeMs.FifteenMinuteRate": &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
	"PartitionReassignmentRateAndTimeMs.FiveMinuteRate":    &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
	"PartitionReassignmentRateAndTimeMs.Max":               &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
	"PartitionReassignmentRateAndTimeMs.Mean":              &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
	"PartitionReassignmentRateAndTimeMs.MeanRate":          &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
	"PartitionReassignmentRateAndTimeMs.Min":               &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
	"PartitionReassignmentRateAndTimeMs.OneMinuteRate":     &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
	"PartitionReassignmentRateAndTimeMs.StdDev":            &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
	"PartitionReassignmentRateAndTimeMs.Count":             &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Count, Unit: inputs.NCount, Desc: ""},
	"PartitionReassignmentRateAndTimeMs.LatencyUnit":       &inputs.FieldInfo{DataType: inputs.String, Type: inputs.String, Unit: inputs.UnknownUnit, Desc: ""},
	"PartitionReassignmentRateAndTimeMs.EventType":         &inputs.FieldInfo{DataType: inputs.String, Type: inputs.String, Unit: inputs.UnknownUnit, Desc: ""},
	"PartitionReassignmentRateAndTimeMs.RateUnit":          &inputs.FieldInfo{DataType: inputs.String, Type: inputs.String, Unit: inputs.UnknownUnit, Desc: ""},

	"TopicChangeRateAndTimeMs.50thPercentile":    &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
	"TopicChangeRateAndTimeMs.75thPercentile":    &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
	"TopicChangeRateAndTimeMs.95thPercentile":    &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
	"TopicChangeRateAndTimeMs.98thPercentile":    &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
	"TopicChangeRateAndTimeMs.999thPercentile":   &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
	"TopicChangeRateAndTimeMs.99thPercentile":    &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
	"TopicChangeRateAndTimeMs.FifteenMinuteRate": &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
	"TopicChangeRateAndTimeMs.FiveMinuteRate":    &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
	"TopicChangeRateAndTimeMs.Max":               &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
	"TopicChangeRateAndTimeMs.Mean":              &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
	"TopicChangeRateAndTimeMs.MeanRate":          &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
	"TopicChangeRateAndTimeMs.Min":               &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
	"TopicChangeRateAndTimeMs.OneMinuteRate":     &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
	"TopicChangeRateAndTimeMs.StdDev":            &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
	"TopicChangeRateAndTimeMs.Count":             &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Count, Unit: inputs.NCount, Desc: ""},
	"TopicChangeRateAndTimeMs.LatencyUnit":       &inputs.FieldInfo{DataType: inputs.String, Type: inputs.String, Unit: inputs.UnknownUnit, Desc: ""},
	"TopicChangeRateAndTimeMs.EventType":         &inputs.FieldInfo{DataType: inputs.String, Type: inputs.String, Unit: inputs.UnknownUnit, Desc: ""},
	"TopicChangeRateAndTimeMs.RateUnit":          &inputs.FieldInfo{DataType: inputs.String, Type: inputs.String, Unit: inputs.UnknownUnit, Desc: ""},

	"TopicDeletionRateAndTimeMs.50thPercentile":    &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
	"TopicDeletionRateAndTimeMs.75thPercentile":    &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
	"TopicDeletionRateAndTimeMs.95thPercentile":    &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
	"TopicDeletionRateAndTimeMs.98thPercentile":    &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
	"TopicDeletionRateAndTimeMs.999thPercentile":   &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
	"TopicDeletionRateAndTimeMs.99thPercentile":    &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
	"TopicDeletionRateAndTimeMs.FifteenMinuteRate": &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
	"TopicDeletionRateAndTimeMs.FiveMinuteRate":    &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
	"TopicDeletionRateAndTimeMs.Max":               &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
	"TopicDeletionRateAndTimeMs.Mean":              &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
	"TopicDeletionRateAndTimeMs.MeanRate":          &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
	"TopicDeletionRateAndTimeMs.Min":               &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
	"TopicDeletionRateAndTimeMs.OneMinuteRate":     &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
	"TopicDeletionRateAndTimeMs.StdDev":            &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
	"TopicDeletionRateAndTimeMs.Count":             &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Count, Unit: inputs.NCount, Desc: ""},
	"TopicDeletionRateAndTimeMs.LatencyUnit":       &inputs.FieldInfo{DataType: inputs.String, Type: inputs.String, Unit: inputs.UnknownUnit, Desc: ""},
	"TopicDeletionRateAndTimeMs.EventType":         &inputs.FieldInfo{DataType: inputs.String, Type: inputs.String, Unit: inputs.UnknownUnit, Desc: ""},
	"TopicDeletionRateAndTimeMs.RateUnit":          &inputs.FieldInfo{DataType: inputs.String, Type: inputs.String, Unit: inputs.UnknownUnit, Desc: ""},

	"TopicUncleanLeaderElectionEnableRateAndTimeMs.50thPercentile":    &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
	"TopicUncleanLeaderElectionEnableRateAndTimeMs.75thPercentile":    &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
	"TopicUncleanLeaderElectionEnableRateAndTimeMs.95thPercentile":    &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
	"TopicUncleanLeaderElectionEnableRateAndTimeMs.98thPercentile":    &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
	"TopicUncleanLeaderElectionEnableRateAndTimeMs.999thPercentile":   &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
	"TopicUncleanLeaderElectionEnableRateAndTimeMs.99thPercentile":    &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
	"TopicUncleanLeaderElectionEnableRateAndTimeMs.FifteenMinuteRate": &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
	"TopicUncleanLeaderElectionEnableRateAndTimeMs.FiveMinuteRate":    &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
	"TopicUncleanLeaderElectionEnableRateAndTimeMs.Max":               &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
	"TopicUncleanLeaderElectionEnableRateAndTimeMs.Mean":              &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
	"TopicUncleanLeaderElectionEnableRateAndTimeMs.MeanRate":          &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
	"TopicUncleanLeaderElectionEnableRateAndTimeMs.Min":               &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
	"TopicUncleanLeaderElectionEnableRateAndTimeMs.OneMinuteRate":     &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
	"TopicUncleanLeaderElectionEnableRateAndTimeMs.StdDev":            &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
	"TopicUncleanLeaderElectionEnableRateAndTimeMs.Count":             &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Count, Unit: inputs.NCount, Desc: ""},
	"TopicUncleanLeaderElectionEnableRateAndTimeMs.LatencyUnit":       &inputs.FieldInfo{DataType: inputs.String, Type: inputs.String, Unit: inputs.UnknownUnit, Desc: ""},
	"TopicUncleanLeaderElectionEnableRateAndTimeMs.EventType":         &inputs.FieldInfo{DataType: inputs.String, Type: inputs.String, Unit: inputs.UnknownUnit, Desc: ""},
	"TopicUncleanLeaderElectionEnableRateAndTimeMs.RateUnit":          &inputs.FieldInfo{DataType: inputs.String, Type: inputs.String, Unit: inputs.UnknownUnit, Desc: ""},

	"UncleanLeaderElectionEnableRateAndTimeMs.50thPercentile":    &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
	"UncleanLeaderElectionEnableRateAndTimeMs.75thPercentile":    &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
	"UncleanLeaderElectionEnableRateAndTimeMs.95thPercentile":    &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
	"UncleanLeaderElectionEnableRateAndTimeMs.98thPercentile":    &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
	"UncleanLeaderElectionEnableRateAndTimeMs.999thPercentile":   &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
	"UncleanLeaderElectionEnableRateAndTimeMs.99thPercentile":    &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
	"UncleanLeaderElectionEnableRateAndTimeMs.FifteenMinuteRate": &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
	"UncleanLeaderElectionEnableRateAndTimeMs.FiveMinuteRate":    &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
	"UncleanLeaderElectionEnableRateAndTimeMs.Max":               &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
	"UncleanLeaderElectionEnableRateAndTimeMs.Mean":              &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
	"UncleanLeaderElectionEnableRateAndTimeMs.MeanRate":          &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
	"UncleanLeaderElectionEnableRateAndTimeMs.Min":               &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
	"UncleanLeaderElectionEnableRateAndTimeMs.OneMinuteRate":     &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
	"UncleanLeaderElectionEnableRateAndTimeMs.StdDev":            &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
	"UncleanLeaderElectionEnableRateAndTimeMs.Count":             &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Count, Unit: inputs.NCount, Desc: ""},
	"UncleanLeaderElectionEnableRateAndTimeMs.LatencyUnit":       &inputs.FieldInfo{DataType: inputs.String, Type: inputs.String, Unit: inputs.UnknownUnit, Desc: ""},
	"UncleanLeaderElectionEnableRateAndTimeMs.EventType":         &inputs.FieldInfo{DataType: inputs.String, Type: inputs.String, Unit: inputs.UnknownUnit, Desc: ""},
	"UncleanLeaderElectionEnableRateAndTimeMs.RateUnit":          &inputs.FieldInfo{DataType: inputs.String, Type: inputs.String, Unit: inputs.UnknownUnit, Desc: ""},

	"UpdateFeaturesRateAndTimeMs.50thPercentile":    &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
	"UpdateFeaturesRateAndTimeMs.75thPercentile":    &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
	"UpdateFeaturesRateAndTimeMs.95thPercentile":    &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
	"UpdateFeaturesRateAndTimeMs.98thPercentile":    &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
	"UpdateFeaturesRateAndTimeMs.999thPercentile":   &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
	"UpdateFeaturesRateAndTimeMs.99thPercentile":    &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
	"UpdateFeaturesRateAndTimeMs.FifteenMinuteRate": &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
	"UpdateFeaturesRateAndTimeMs.FiveMinuteRate":    &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
	"UpdateFeaturesRateAndTimeMs.Max":               &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
	"UpdateFeaturesRateAndTimeMs.Mean":              &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
	"UpdateFeaturesRateAndTimeMs.MeanRate":          &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
	"UpdateFeaturesRateAndTimeMs.Min":               &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
	"UpdateFeaturesRateAndTimeMs.OneMinuteRate":     &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
	"UpdateFeaturesRateAndTimeMs.StdDev":            &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
	"UpdateFeaturesRateAndTimeMs.Count":             &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Count, Unit: inputs.NCount, Desc: ""},
	"UpdateFeaturesRateAndTimeMs.LatencyUnit":       &inputs.FieldInfo{DataType: inputs.String, Type: inputs.String, Unit: inputs.UnknownUnit, Desc: ""},
	"UpdateFeaturesRateAndTimeMs.EventType":         &inputs.FieldInfo{DataType: inputs.String, Type: inputs.String, Unit: inputs.UnknownUnit, Desc: ""},
	"UpdateFeaturesRateAndTimeMs.RateUnit":          &inputs.FieldInfo{DataType: inputs.String, Type: inputs.String, Unit: inputs.UnknownUnit, Desc: ""},

	"UncleanLeaderElectionsPerSec.50thPercentile":    &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
	"UncleanLeaderElectionsPerSec.75thPercentile":    &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
	"UncleanLeaderElectionsPerSec.95thPercentile":    &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
	"UncleanLeaderElectionsPerSec.98thPercentile":    &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
	"UncleanLeaderElectionsPerSec.999thPercentile":   &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
	"UncleanLeaderElectionsPerSec.99thPercentile":    &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
	"UncleanLeaderElectionsPerSec.FifteenMinuteRate": &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
	"UncleanLeaderElectionsPerSec.FiveMinuteRate":    &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
	"UncleanLeaderElectionsPerSec.Max":               &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
	"UncleanLeaderElectionsPerSec.Mean":              &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
	"UncleanLeaderElectionsPerSec.MeanRate":          &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
	"UncleanLeaderElectionsPerSec.Min":               &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
	"UncleanLeaderElectionsPerSec.OneMinuteRate":     &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
	"UncleanLeaderElectionsPerSec.StdDev":            &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
	"UncleanLeaderElectionsPerSec.Count":             &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Count, Unit: inputs.NCount, Desc: ""},
	"UncleanLeaderElectionsPerSec.LatencyUnit":       &inputs.FieldInfo{DataType: inputs.String, Type: inputs.String, Unit: inputs.UnknownUnit, Desc: ""},
	"UncleanLeaderElectionsPerSec.EventType":         &inputs.FieldInfo{DataType: inputs.String, Type: inputs.String, Unit: inputs.UnknownUnit, Desc: ""},
	"UncleanLeaderElectionsPerSec.RateUnit":          &inputs.FieldInfo{DataType: inputs.String, Type: inputs.String, Unit: inputs.UnknownUnit, Desc: ""},

	"EventQueueTimeMs.50thPercentile":    &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
	"EventQueueTimeMs.75thPercentile":    &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
	"EventQueueTimeMs.95thPercentile":    &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
	"EventQueueTimeMs.98thPercentile":    &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
	"EventQueueTimeMs.999thPercentile":   &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
	"EventQueueTimeMs.99thPercentile":    &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
	"EventQueueTimeMs.FifteenMinuteRate": &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
	"EventQueueTimeMs.FiveMinuteRate":    &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
	"EventQueueTimeMs.Max":               &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
	"EventQueueTimeMs.Mean":              &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
	"EventQueueTimeMs.MeanRate":          &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
	"EventQueueTimeMs.Min":               &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
	"EventQueueTimeMs.OneMinuteRate":     &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
	"EventQueueTimeMs.StdDev":            &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
	"EventQueueTimeMs.Count":             &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Count, Unit: inputs.NCount, Desc: ""},
	"EventQueueTimeMs.LatencyUnit":       &inputs.FieldInfo{DataType: inputs.String, Type: inputs.String, Unit: inputs.UnknownUnit, Desc: ""},
	"EventQueueTimeMs.EventType":         &inputs.FieldInfo{DataType: inputs.String, Type: inputs.String, Unit: inputs.UnknownUnit, Desc: ""},
	"EventQueueTimeMs.RateUnit":          &inputs.FieldInfo{DataType: inputs.String, Type: inputs.String, Unit: inputs.UnknownUnit, Desc: ""},

	"GlobalPartitionCount.Value":            &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Count, Unit: inputs.NCount, Desc: ""},
	"GlobalTopicCount.Value":                &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Count, Unit: inputs.NCount, Desc: ""},
	"OfflinePartitionsCount.Value":          &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Count, Unit: inputs.NCount, Desc: ""},
	"PreferredReplicaImbalanceCount.Value":  &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Count, Unit: inputs.NCount, Desc: ""},
	"ReplicasIneligibleToDeleteCount.Value": &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Count, Unit: inputs.NCount, Desc: ""},
	"ReplicasToDeleteCount.Value":           &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Count, Unit: inputs.NCount, Desc: ""},
	"TopicsIneligibleToDeleteCount.Value":   &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Count, Unit: inputs.NCount, Desc: ""},
	"TopicsToDeleteCount.Value":             &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Count, Unit: inputs.NCount, Desc: ""},
	"ActiveControllerCount.Value":           &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Count, Unit: inputs.NCount, Desc: ""},

	"ControllerState.Value": &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
	"EventQueueSize.Value":  &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
	"TotalQueueSize.Value":  &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
}

var controllerTags = map[string]interface{}{
	"jolokia_agent_url": inputs.TagInfo{Desc: "jolokia agent url path"},
}

func (j *KafkaControllerMment) LineProto() (*io.Point, error) {
	return io.NewPoint(j.name, j.tags, j.fields, &io.PointOption{Category: datakit.Metric, Time: j.ts})
}

func (j *KafkaControllerMment) Info() *inputs.MeasurementInfo { //nolint:funlen
	return &inputs.MeasurementInfo{
		Name:   "kafka_controller",
		Fields: controllerFields,
		Tags:   controllerTags,
	}
}

var replicationTags = map[string]interface{}{
	"jolokia_agent_url": inputs.TagInfo{Desc: "jolokia agent url path"},
}

var replicationFields = map[string]interface{}{
	"FailedIsrUpdatesPerSec.Count":             &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Count, Unit: inputs.NCount, Desc: ""},
	"FailedIsrUpdatesPerSec.EventType":         &inputs.FieldInfo{DataType: inputs.String, Type: inputs.String, Unit: inputs.UnknownUnit, Desc: ""},
	"FailedIsrUpdatesPerSec.FifteenMinuteRate": &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
	"FailedIsrUpdatesPerSec.FiveMinuteRate":    &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
	"FailedIsrUpdatesPerSec.MeanRate":          &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
	"FailedIsrUpdatesPerSec.OneMinuteRate":     &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
	"FailedIsrUpdatesPerSec.RateUnit":          &inputs.FieldInfo{DataType: inputs.String, Type: inputs.String, Unit: inputs.UnknownUnit, Desc: ""},

	"IsrExpandsPerSec.Count":             &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Count, Unit: inputs.NCount, Desc: ""},
	"IsrExpandsPerSec.EventType":         &inputs.FieldInfo{DataType: inputs.String, Type: inputs.String, Unit: inputs.UnknownUnit, Desc: ""},
	"IsrExpandsPerSec.FifteenMinuteRate": &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
	"IsrExpandsPerSec.FiveMinuteRate":    &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
	"IsrExpandsPerSec.MeanRate":          &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
	"IsrExpandsPerSec.OneMinuteRate":     &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
	"IsrExpandsPerSec.RateUnit":          &inputs.FieldInfo{DataType: inputs.String, Type: inputs.String, Unit: inputs.UnknownUnit, Desc: ""},

	"IsrShrinksPerSec.Count":             &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Count, Unit: inputs.NCount, Desc: ""},
	"IsrShrinksPerSec.EventType":         &inputs.FieldInfo{DataType: inputs.String, Type: inputs.String, Unit: inputs.UnknownUnit, Desc: ""},
	"IsrShrinksPerSec.FifteenMinuteRate": &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
	"IsrShrinksPerSec.FiveMinuteRate":    &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
	"IsrShrinksPerSec.MeanRate":          &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
	"IsrShrinksPerSec.OneMinuteRate":     &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
	"IsrShrinksPerSec.RateUnit":          &inputs.FieldInfo{DataType: inputs.String, Type: inputs.String, Unit: inputs.UnknownUnit, Desc: ""},

	"AtMinIsrPartitionCount.Value":    &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Count, Unit: inputs.NCount, Desc: ""},
	"LeaderCount.Value":               &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Count, Unit: inputs.NCount, Desc: ""},
	"OfflineReplicaCount.Value":       &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Count, Unit: inputs.NCount, Desc: ""},
	"PartitionCount.Value":            &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Count, Unit: inputs.NCount, Desc: ""},
	"ReassigningPartitions.Value":     &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Count, Unit: inputs.NCount, Desc: ""},
	"UnderMinIsrPartitionCount.Value": &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Count, Unit: inputs.NCount, Desc: ""},
	"UnderReplicatedPartitions.Value": &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Count, Unit: inputs.NCount, Desc: ""},
}

func (j *KafkaReplicaMment) LineProto() (*io.Point, error) {
	return io.NewPoint(j.name, j.tags, j.fields, &io.PointOption{Category: datakit.Metric, Time: j.ts})
}

func (j *KafkaReplicaMment) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name:   "kafka_replica_manager",
		Fields: replicationFields,
		Tags:   replicationTags,
	}
}

var purgatoryTags = map[string]interface{}{
	"jolokia_agent_url": inputs.TagInfo{Desc: "jolokia agent url path"},
}

var purgatoryFields = map[string]interface{}{
	"AlterAcls.NumDelayedOperations": &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
	"AlterAcls.PurgatorySize":        &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},

	"DeleteRecords.NumDelayedOperations": &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
	"DeleteRecords.PurgatorySize":        &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},

	"ElectLeader.NumDelayedOperations": &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
	"ElectLeader.PurgatorySize":        &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},

	"Fetch.NumDelayedOperations": &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
	"Fetch.PurgatorySize":        &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},

	"Heartbeat.NumDelayedOperations": &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
	"Heartbeat.PurgatorySize":        &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},

	"Produce.NumDelayedOperations": &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
	"Produce.PurgatorySize":        &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},

	"Rebalance.NumDelayedOperations": &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
	"Rebalance.PurgatorySize":        &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},

	"topic.NumDelayedOperations": &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
	"topic.PurgatorySize":        &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
}

func (j *KafkaPurgatoryMment) LineProto() (*io.Point, error) {
	return io.NewPoint(j.name, j.tags, j.fields, &io.PointOption{Category: datakit.Metric, Time: j.ts})
}

func (j *KafkaPurgatoryMment) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name:   "kafka_purgatory",
		Fields: purgatoryFields,
		Tags:   purgatoryTags,
	}
}

func (j *KafkaRequestMment) LineProto() (*io.Point, error) {
	return io.NewPoint(j.name, j.tags, j.fields, &io.PointOption{Category: datakit.Metric, Time: j.ts})
}

var requestFields = map[string]interface{}{
	"LocalTimeMs.50thPercentile":  &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS, Desc: ""},
	"LocalTimeMs.75thPercentile":  &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS, Desc: ""},
	"LocalTimeMs.95thPercentile":  &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS, Desc: ""},
	"LocalTimeMs.98thPercentile":  &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS, Desc: ""},
	"LocalTimeMs.999thPercentile": &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS, Desc: ""},
	"LocalTimeMs.99thPercentile":  &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS, Desc: ""},
	"LocalTimeMs.Max":             &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS, Desc: ""},
	"LocalTimeMs.Mean":            &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS, Desc: ""},
	"LocalTimeMs.Min":             &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS, Desc: ""},
	"LocalTimeMs.StdDev":          &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS, Desc: ""},
	"LocalTimeMs.Count":           &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Count, Unit: inputs.NCount, Desc: ""},

	"RemoteTimeMs.50thPercentile":  &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS, Desc: ""},
	"RemoteTimeMs.75thPercentile":  &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS, Desc: ""},
	"RemoteTimeMs.95thPercentile":  &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS, Desc: ""},
	"RemoteTimeMs.98thPercentile":  &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS, Desc: ""},
	"RemoteTimeMs.999thPercentile": &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS, Desc: ""},
	"RemoteTimeMs.99thPercentile":  &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS, Desc: ""},
	"RemoteTimeMs.Max":             &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS, Desc: ""},
	"RemoteTimeMs.Mean":            &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS, Desc: ""},
	"RemoteTimeMs.Min":             &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS, Desc: ""},
	"RemoteTimeMs.StdDev":          &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS, Desc: ""},
	"RemoteTimeMs.Count":           &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Count, Unit: inputs.NCount, Desc: ""},

	"RequestBytes.50thPercentile":  &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS, Desc: ""},
	"RequestBytes.75thPercentile":  &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS, Desc: ""},
	"RequestBytes.95thPercentile":  &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS, Desc: ""},
	"RequestBytes.98thPercentile":  &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS, Desc: ""},
	"RequestBytes.999thPercentile": &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS, Desc: ""},
	"RequestBytes.99thPercentile":  &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS, Desc: ""},
	"RequestBytes.Max":             &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS, Desc: ""},
	"RequestBytes.Mean":            &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS, Desc: ""},
	"RequestBytes.Min":             &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS, Desc: ""},
	"RequestBytes.StdDev":          &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS, Desc: ""},
	"RequestBytes.Count":           &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Count, Unit: inputs.NCount, Desc: ""},

	"RequestQueueTimeMs.50thPercentile":  &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS, Desc: ""},
	"RequestQueueTimeMs.75thPercentile":  &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS, Desc: ""},
	"RequestQueueTimeMs.95thPercentile":  &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS, Desc: ""},
	"RequestQueueTimeMs.98thPercentile":  &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS, Desc: ""},
	"RequestQueueTimeMs.999thPercentile": &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS, Desc: ""},
	"RequestQueueTimeMs.99thPercentile":  &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS, Desc: ""},
	"RequestQueueTimeMs.Max":             &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS, Desc: ""},
	"RequestQueueTimeMs.Mean":            &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS, Desc: ""},
	"RequestQueueTimeMs.Min":             &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS, Desc: ""},
	"RequestQueueTimeMs.StdDev":          &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS, Desc: ""},
	"RequestQueueTimeMs.Count":           &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Count, Unit: inputs.NCount, Desc: ""},

	"ResponseQueueTimeMs.50thPercentile":  &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS, Desc: ""},
	"ResponseQueueTimeMs.75thPercentile":  &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS, Desc: ""},
	"ResponseQueueTimeMs.95thPercentile":  &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS, Desc: ""},
	"ResponseQueueTimeMs.98thPercentile":  &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS, Desc: ""},
	"ResponseQueueTimeMs.999thPercentile": &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS, Desc: ""},
	"ResponseQueueTimeMs.99thPercentile":  &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS, Desc: ""},
	"ResponseQueueTimeMs.Max":             &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS, Desc: ""},
	"ResponseQueueTimeMs.Mean":            &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS, Desc: ""},
	"ResponseQueueTimeMs.Min":             &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS, Desc: ""},
	"ResponseQueueTimeMs.StdDev":          &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS, Desc: ""},
	"ResponseQueueTimeMs.Count":           &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Count, Unit: inputs.NCount, Desc: ""},

	"ResponseSendTimeMs.50thPercentile":  &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS, Desc: ""},
	"ResponseSendTimeMs.75thPercentile":  &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS, Desc: ""},
	"ResponseSendTimeMs.95thPercentile":  &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS, Desc: ""},
	"ResponseSendTimeMs.98thPercentile":  &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS, Desc: ""},
	"ResponseSendTimeMs.999thPercentile": &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS, Desc: ""},
	"ResponseSendTimeMs.99thPercentile":  &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS, Desc: ""},
	"ResponseSendTimeMs.Max":             &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS, Desc: ""},
	"ResponseSendTimeMs.Mean":            &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS, Desc: ""},
	"ResponseSendTimeMs.Min":             &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS, Desc: ""},
	"ResponseSendTimeMs.StdDev":          &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS, Desc: ""},
	"ResponseSendTimeMs.Count":           &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Count, Unit: inputs.NCount, Desc: ""},

	"ThrottleTimeMs.50thPercentile":  &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS, Desc: ""},
	"ThrottleTimeMs.75thPercentile":  &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS, Desc: ""},
	"ThrottleTimeMs.95thPercentile":  &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS, Desc: ""},
	"ThrottleTimeMs.98thPercentile":  &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS, Desc: ""},
	"ThrottleTimeMs.999thPercentile": &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS, Desc: ""},
	"ThrottleTimeMs.99thPercentile":  &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS, Desc: ""},
	"ThrottleTimeMs.Max":             &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS, Desc: ""},
	"ThrottleTimeMs.Mean":            &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS, Desc: ""},
	"ThrottleTimeMs.Min":             &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS, Desc: ""},
	"ThrottleTimeMs.StdDev":          &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS, Desc: ""},
	"ThrottleTimeMs.Count":           &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Count, Unit: inputs.NCount, Desc: ""},

	"TotalTimeMs.50thPercentile":  &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS, Desc: ""},
	"TotalTimeMs.75thPercentile":  &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS, Desc: ""},
	"TotalTimeMs.95thPercentile":  &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS, Desc: ""},
	"TotalTimeMs.98thPercentile":  &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS, Desc: ""},
	"TotalTimeMs.999thPercentile": &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS, Desc: ""},
	"TotalTimeMs.99thPercentile":  &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS, Desc: ""},
	"TotalTimeMs.Max":             &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS, Desc: ""},
	"TotalTimeMs.Mean":            &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS, Desc: ""},
	"TotalTimeMs.Min":             &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS, Desc: ""},
	"TotalTimeMs.StdDev":          &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS, Desc: ""},
	"TotalTimeMs.Count":           &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Count, Unit: inputs.NCount, Desc: ""},
}

var requestTags = map[string]interface{}{
	"jolokia_agent_url": inputs.TagInfo{Desc: "jolokia agent url path"},
}

func (j *KafkaRequestMment) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name:   "kafka_request",
		Fields: requestFields,
		Tags:   requestTags,
	}
}

func (j *KafkaTopicsMment) LineProto() (*io.Point, error) {
	return io.NewPoint(j.name, j.tags, j.fields, &io.PointOption{Category: datakit.Metric, Time: j.ts})
}

var topicsFields = map[string]interface{}{
	"BytesInPerSec.Count":             &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Count, Unit: inputs.NCount, Desc: ""},
	"BytesInPerSec.EventType":         &inputs.FieldInfo{DataType: inputs.String, Type: inputs.String, Unit: inputs.UnknownUnit, Desc: ""},
	"BytesInPerSec.FifteenMinuteRate": &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
	"BytesInPerSec.FiveMinuteRate":    &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
	"BytesInPerSec.MeanRate":          &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
	"BytesInPerSec.OneMinuteRate":     &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
	"BytesInPerSec.RateUnit":          &inputs.FieldInfo{DataType: inputs.String, Type: inputs.String, Unit: inputs.UnknownUnit, Desc: ""},

	"BytesOutPerSec.Count":             &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Count, Unit: inputs.NCount, Desc: ""},
	"BytesOutPerSec.EventType":         &inputs.FieldInfo{DataType: inputs.String, Type: inputs.String, Unit: inputs.UnknownUnit, Desc: ""},
	"BytesOutPerSec.FifteenMinuteRate": &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
	"BytesOutPerSec.FiveMinuteRate":    &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
	"BytesOutPerSec.MeanRate":          &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
	"BytesOutPerSec.OneMinuteRate":     &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
	"BytesOutPerSec.RateUnit":          &inputs.FieldInfo{DataType: inputs.String, Type: inputs.String, Unit: inputs.UnknownUnit, Desc: ""},

	"BytesRejectedPerSec.Count":             &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Count, Unit: inputs.NCount, Desc: ""},
	"BytesRejectedPerSec.EventType":         &inputs.FieldInfo{DataType: inputs.String, Type: inputs.String, Unit: inputs.UnknownUnit, Desc: ""},
	"BytesRejectedPerSec.FifteenMinuteRate": &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
	"BytesRejectedPerSec.FiveMinuteRate":    &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
	"BytesRejectedPerSec.MeanRate":          &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
	"BytesRejectedPerSec.OneMinuteRate":     &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
	"BytesRejectedPerSec.RateUnit":          &inputs.FieldInfo{DataType: inputs.String, Type: inputs.String, Unit: inputs.UnknownUnit, Desc: ""},

	"FailedFetchRequestsPerSec.Count":             &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Count, Unit: inputs.NCount, Desc: ""},
	"FailedFetchRequestsPerSec.EventType":         &inputs.FieldInfo{DataType: inputs.String, Type: inputs.String, Unit: inputs.UnknownUnit, Desc: ""},
	"FailedFetchRequestsPerSec.FifteenMinuteRate": &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
	"FailedFetchRequestsPerSec.FiveMinuteRate":    &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
	"FailedFetchRequestsPerSec.MeanRate":          &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
	"FailedFetchRequestsPerSec.OneMinuteRate":     &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
	"FailedFetchRequestsPerSec.RateUnit":          &inputs.FieldInfo{DataType: inputs.String, Type: inputs.String, Unit: inputs.UnknownUnit, Desc: ""},

	"FailedProduceRequestsPerSec.Count":             &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Count, Unit: inputs.NCount, Desc: ""},
	"FailedProduceRequestsPerSec.EventType":         &inputs.FieldInfo{DataType: inputs.String, Type: inputs.String, Unit: inputs.UnknownUnit, Desc: ""},
	"FailedProduceRequestsPerSec.FifteenMinuteRate": &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
	"FailedProduceRequestsPerSec.FiveMinuteRate":    &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
	"FailedProduceRequestsPerSec.MeanRate":          &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
	"FailedProduceRequestsPerSec.OneMinuteRate":     &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
	"FailedProduceRequestsPerSec.RateUnit":          &inputs.FieldInfo{DataType: inputs.String, Type: inputs.String, Unit: inputs.UnknownUnit, Desc: ""},

	"FetchMessageConversionsPerSec.Count":             &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Count, Unit: inputs.NCount, Desc: ""},
	"FetchMessageConversionsPerSec.EventType":         &inputs.FieldInfo{DataType: inputs.String, Type: inputs.String, Unit: inputs.UnknownUnit, Desc: ""},
	"FetchMessageConversionsPerSec.FifteenMinuteRate": &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
	"FetchMessageConversionsPerSec.FiveMinuteRate":    &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
	"FetchMessageConversionsPerSec.MeanRate":          &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
	"FetchMessageConversionsPerSec.OneMinuteRate":     &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
	"FetchMessageConversionsPerSec.RateUnit":          &inputs.FieldInfo{DataType: inputs.String, Type: inputs.String, Unit: inputs.UnknownUnit, Desc: ""},

	"InvalidMagicNumberRecordsPerSec.Count":             &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Count, Unit: inputs.NCount, Desc: ""},
	"InvalidMagicNumberRecordsPerSec.EventType":         &inputs.FieldInfo{DataType: inputs.String, Type: inputs.String, Unit: inputs.UnknownUnit, Desc: ""},
	"InvalidMagicNumberRecordsPerSec.FifteenMinuteRate": &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
	"InvalidMagicNumberRecordsPerSec.FiveMinuteRate":    &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
	"InvalidMagicNumberRecordsPerSec.MeanRate":          &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
	"InvalidMagicNumberRecordsPerSec.OneMinuteRate":     &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
	"InvalidMagicNumberRecordsPerSec.RateUnit":          &inputs.FieldInfo{DataType: inputs.String, Type: inputs.String, Unit: inputs.UnknownUnit, Desc: ""},

	"InvalidMessageCrcRecordsPerSec.Count":             &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Count, Unit: inputs.NCount, Desc: ""},
	"InvalidMessageCrcRecordsPerSec.EventType":         &inputs.FieldInfo{DataType: inputs.String, Type: inputs.String, Unit: inputs.UnknownUnit, Desc: ""},
	"InvalidMessageCrcRecordsPerSec.FifteenMinuteRate": &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
	"InvalidMessageCrcRecordsPerSec.FiveMinuteRate":    &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
	"InvalidMessageCrcRecordsPerSec.MeanRate":          &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
	"InvalidMessageCrcRecordsPerSec.OneMinuteRate":     &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
	"InvalidMessageCrcRecordsPerSec.RateUnit":          &inputs.FieldInfo{DataType: inputs.String, Type: inputs.String, Unit: inputs.UnknownUnit, Desc: ""},

	"InvalidOffsetOrSequenceRecordsPerSec.Count":             &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Count, Unit: inputs.NCount, Desc: ""},
	"InvalidOffsetOrSequenceRecordsPerSec.EventType":         &inputs.FieldInfo{DataType: inputs.String, Type: inputs.String, Unit: inputs.UnknownUnit, Desc: ""},
	"InvalidOffsetOrSequenceRecordsPerSec.FifteenMinuteRate": &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
	"InvalidOffsetOrSequenceRecordsPerSec.FiveMinuteRate":    &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
	"InvalidOffsetOrSequenceRecordsPerSec.MeanRate":          &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
	"InvalidOffsetOrSequenceRecordsPerSec.OneMinuteRate":     &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
	"InvalidOffsetOrSequenceRecordsPerSec.RateUnit":          &inputs.FieldInfo{DataType: inputs.String, Type: inputs.String, Unit: inputs.UnknownUnit, Desc: ""},

	"MessagesInPerSec.Count":             &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Count, Unit: inputs.NCount, Desc: ""},
	"MessagesInPerSec.EventType":         &inputs.FieldInfo{DataType: inputs.String, Type: inputs.String, Unit: inputs.UnknownUnit, Desc: ""},
	"MessagesInPerSec.FifteenMinuteRate": &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
	"MessagesInPerSec.FiveMinuteRate":    &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
	"MessagesInPerSec.MeanRate":          &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
	"MessagesInPerSec.OneMinuteRate":     &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
	"MessagesInPerSec.RateUnit":          &inputs.FieldInfo{DataType: inputs.String, Type: inputs.String, Unit: inputs.UnknownUnit, Desc: ""},

	"NoKeyCompactedTopicRecordsPerSec.Count":             &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Count, Unit: inputs.NCount, Desc: ""},
	"NoKeyCompactedTopicRecordsPerSec.EventType":         &inputs.FieldInfo{DataType: inputs.String, Type: inputs.String, Unit: inputs.UnknownUnit, Desc: ""},
	"NoKeyCompactedTopicRecordsPerSec.FifteenMinuteRate": &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
	"NoKeyCompactedTopicRecordsPerSec.FiveMinuteRate":    &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
	"NoKeyCompactedTopicRecordsPerSec.MeanRate":          &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
	"NoKeyCompactedTopicRecordsPerSec.OneMinuteRate":     &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
	"NoKeyCompactedTopicRecordsPerSec.RateUnit":          &inputs.FieldInfo{DataType: inputs.String, Type: inputs.String, Unit: inputs.UnknownUnit, Desc: ""},

	"ProduceMessageConversionsPerSec.Count":             &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Count, Unit: inputs.NCount, Desc: ""},
	"ProduceMessageConversionsPerSec.EventType":         &inputs.FieldInfo{DataType: inputs.String, Type: inputs.String, Unit: inputs.UnknownUnit, Desc: ""},
	"ProduceMessageConversionsPerSec.FifteenMinuteRate": &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
	"ProduceMessageConversionsPerSec.FiveMinuteRate":    &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
	"ProduceMessageConversionsPerSec.MeanRate":          &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
	"ProduceMessageConversionsPerSec.OneMinuteRate":     &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
	"ProduceMessageConversionsPerSec.RateUnit":          &inputs.FieldInfo{DataType: inputs.String, Type: inputs.String, Unit: inputs.UnknownUnit, Desc: ""},

	"ReassignmentBytesInPerSec.Count":             &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Count, Unit: inputs.NCount, Desc: ""},
	"ReassignmentBytesInPerSec.EventType":         &inputs.FieldInfo{DataType: inputs.String, Type: inputs.String, Unit: inputs.UnknownUnit, Desc: ""},
	"ReassignmentBytesInPerSec.FifteenMinuteRate": &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
	"ReassignmentBytesInPerSec.FiveMinuteRate":    &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
	"ReassignmentBytesInPerSec.MeanRate":          &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
	"ReassignmentBytesInPerSec.OneMinuteRate":     &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
	"ReassignmentBytesInPerSec.RateUnit":          &inputs.FieldInfo{DataType: inputs.String, Type: inputs.String, Unit: inputs.UnknownUnit, Desc: ""},

	"ReassignmentBytesOutPerSec.Count":             &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Count, Unit: inputs.NCount, Desc: ""},
	"ReassignmentBytesOutPerSec.EventType":         &inputs.FieldInfo{DataType: inputs.String, Type: inputs.String, Unit: inputs.UnknownUnit, Desc: ""},
	"ReassignmentBytesOutPerSec.FifteenMinuteRate": &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
	"ReassignmentBytesOutPerSec.FiveMinuteRate":    &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
	"ReassignmentBytesOutPerSec.MeanRate":          &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
	"ReassignmentBytesOutPerSec.OneMinuteRate":     &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
	"ReassignmentBytesOutPerSec.RateUnit":          &inputs.FieldInfo{DataType: inputs.String, Type: inputs.String, Unit: inputs.UnknownUnit, Desc: ""},

	"ReplicationBytesInPerSec.Count":             &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Count, Unit: inputs.NCount, Desc: ""},
	"ReplicationBytesInPerSec.EventType":         &inputs.FieldInfo{DataType: inputs.String, Type: inputs.String, Unit: inputs.UnknownUnit, Desc: ""},
	"ReplicationBytesInPerSec.FifteenMinuteRate": &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
	"ReplicationBytesInPerSec.FiveMinuteRate":    &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
	"ReplicationBytesInPerSec.MeanRate":          &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
	"ReplicationBytesInPerSec.OneMinuteRate":     &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
	"ReplicationBytesInPerSec.RateUnit":          &inputs.FieldInfo{DataType: inputs.String, Type: inputs.String, Unit: inputs.UnknownUnit, Desc: ""},

	"ReplicationBytesOutPerSec.Count":             &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Count, Unit: inputs.NCount, Desc: ""},
	"ReplicationBytesOutPerSec.EventType":         &inputs.FieldInfo{DataType: inputs.String, Type: inputs.String, Unit: inputs.UnknownUnit, Desc: ""},
	"ReplicationBytesOutPerSec.FifteenMinuteRate": &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
	"ReplicationBytesOutPerSec.FiveMinuteRate":    &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
	"ReplicationBytesOutPerSec.MeanRate":          &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
	"ReplicationBytesOutPerSec.OneMinuteRate":     &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
	"ReplicationBytesOutPerSec.RateUnit":          &inputs.FieldInfo{DataType: inputs.String, Type: inputs.String, Unit: inputs.UnknownUnit, Desc: ""},

	"TotalFetchRequestsPerSec.Count":             &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Count, Unit: inputs.NCount, Desc: ""},
	"TotalFetchRequestsPerSec.EventType":         &inputs.FieldInfo{DataType: inputs.String, Type: inputs.String, Unit: inputs.UnknownUnit, Desc: ""},
	"TotalFetchRequestsPerSec.FifteenMinuteRate": &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
	"TotalFetchRequestsPerSec.FiveMinuteRate":    &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
	"TotalFetchRequestsPerSec.MeanRate":          &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
	"TotalFetchRequestsPerSec.OneMinuteRate":     &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
	"TotalFetchRequestsPerSec.RateUnit":          &inputs.FieldInfo{DataType: inputs.String, Type: inputs.String, Unit: inputs.UnknownUnit, Desc: ""},

	"TotalProduceRequestsPerSec.Count":             &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Count, Unit: inputs.NCount, Desc: ""},
	"TotalProduceRequestsPerSec.EventType":         &inputs.FieldInfo{DataType: inputs.String, Type: inputs.String, Unit: inputs.UnknownUnit, Desc: ""},
	"TotalProduceRequestsPerSec.FifteenMinuteRate": &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
	"TotalProduceRequestsPerSec.FiveMinuteRate":    &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
	"TotalProduceRequestsPerSec.MeanRate":          &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
	"TotalProduceRequestsPerSec.OneMinuteRate":     &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
	"TotalProduceRequestsPerSec.RateUnit":          &inputs.FieldInfo{DataType: inputs.String, Type: inputs.String, Unit: inputs.UnknownUnit, Desc: ""},
}

var topicsTags = map[string]interface{}{
	"jolokia_agent_url": inputs.TagInfo{Desc: "jolokia agent url path"},
}

func (j *KafkaTopicsMment) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name:   "kafka_topics",
		Fields: topicsFields,
		Tags:   topicsTags,
	}
}

func (j *KafkaTopicMment) LineProto() (*io.Point, error) {
	return io.NewPoint(j.name, j.tags, j.fields, &io.PointOption{Category: datakit.Metric, Time: j.ts})
}

var topicFields = map[string]interface{}{
	"BytesInPerSec.Count":             &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Count, Unit: inputs.NCount, Desc: ""},
	"BytesInPerSec.EventType":         &inputs.FieldInfo{DataType: inputs.String, Type: inputs.String, Unit: inputs.UnknownUnit, Desc: ""},
	"BytesInPerSec.FifteenMinuteRate": &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
	"BytesInPerSec.FiveMinuteRate":    &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
	"BytesInPerSec.MeanRate":          &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
	"BytesInPerSec.OneMinuteRate":     &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
	"BytesInPerSec.RateUnit":          &inputs.FieldInfo{DataType: inputs.String, Type: inputs.String, Unit: inputs.UnknownUnit, Desc: ""},

	"BytesOutPerSec.Count":             &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Count, Unit: inputs.NCount, Desc: ""},
	"BytesOutPerSec.EventType":         &inputs.FieldInfo{DataType: inputs.String, Type: inputs.String, Unit: inputs.UnknownUnit, Desc: ""},
	"BytesOutPerSec.FifteenMinuteRate": &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
	"BytesOutPerSec.FiveMinuteRate":    &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
	"BytesOutPerSec.MeanRate":          &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
	"BytesOutPerSec.OneMinuteRate":     &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
	"BytesOutPerSec.RateUnit":          &inputs.FieldInfo{DataType: inputs.String, Type: inputs.String, Unit: inputs.UnknownUnit, Desc: ""},

	"MessagesInPerSec.Count":             &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Count, Unit: inputs.NCount, Desc: ""},
	"MessagesInPerSec.EventType":         &inputs.FieldInfo{DataType: inputs.String, Type: inputs.String, Unit: inputs.UnknownUnit, Desc: ""},
	"MessagesInPerSec.FifteenMinuteRate": &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
	"MessagesInPerSec.FiveMinuteRate":    &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
	"MessagesInPerSec.MeanRate":          &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
	"MessagesInPerSec.OneMinuteRate":     &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
	"MessagesInPerSec.RateUnit":          &inputs.FieldInfo{DataType: inputs.String, Type: inputs.String, Unit: inputs.UnknownUnit, Desc: ""},

	"TotalFetchRequestsPerSec.Count":             &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Count, Unit: inputs.NCount, Desc: ""},
	"TotalFetchRequestsPerSec.EventType":         &inputs.FieldInfo{DataType: inputs.String, Type: inputs.String, Unit: inputs.UnknownUnit, Desc: ""},
	"TotalFetchRequestsPerSec.FifteenMinuteRate": &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
	"TotalFetchRequestsPerSec.FiveMinuteRate":    &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
	"TotalFetchRequestsPerSec.MeanRate":          &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
	"TotalFetchRequestsPerSec.OneMinuteRate":     &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
	"TotalFetchRequestsPerSec.RateUnit":          &inputs.FieldInfo{DataType: inputs.String, Type: inputs.String, Unit: inputs.UnknownUnit, Desc: ""},

	"TotalProduceRequestsPerSec.Count":             &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Count, Unit: inputs.NCount, Desc: ""},
	"TotalProduceRequestsPerSec.EventType":         &inputs.FieldInfo{DataType: inputs.String, Type: inputs.String, Unit: inputs.UnknownUnit, Desc: ""},
	"TotalProduceRequestsPerSec.FifteenMinuteRate": &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
	"TotalProduceRequestsPerSec.FiveMinuteRate":    &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
	"TotalProduceRequestsPerSec.MeanRate":          &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
	"TotalProduceRequestsPerSec.OneMinuteRate":     &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
	"TotalProduceRequestsPerSec.RateUnit":          &inputs.FieldInfo{DataType: inputs.String, Type: inputs.String, Unit: inputs.UnknownUnit, Desc: ""},
}

var topicTags = map[string]interface{}{
	"jolokia_agent_url": inputs.TagInfo{Desc: "jolokia agent url path"},
	"topic":             inputs.TagInfo{Desc: "topic name"},
}

func (j *KafkaTopicMment) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name:   "kafka_topic",
		Tags:   topicTags,
		Fields: topicFields,
	}
}

func (j *KafkaPartitionMment) LineProto() (*io.Point, error) {
	return io.NewPoint(j.name, j.tags, j.fields, &io.PointOption{Category: datakit.Metric, Time: j.ts})
}

var partitionFields = map[string]interface{}{
	"LogEndOffset":              &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
	"LogStartOffset":            &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
	"NumLogSegments":            &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
	"Size":                      &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
	"UnderReplicatedPartitions": &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: ""},
}

var partitionTags = map[string]interface{}{
	"jolokia_agent_url": inputs.TagInfo{Desc: "jolokia agent url path"},
	"partition":         inputs.TagInfo{Desc: "partition number"},
	"topic":             inputs.TagInfo{Desc: "topic name"},
}

func (j *KafkaPartitionMment) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name:   "kafka_partition",
		Tags:   partitionTags,
		Fields: partitionFields,
	}
}
