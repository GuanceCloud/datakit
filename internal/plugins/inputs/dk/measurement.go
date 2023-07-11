// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package dk

import (
	"fmt"

	dkpt "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/io/point"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs"
)

func (*Input) SampleMeasurement() []inputs.Measurement {
	return []inputs.Measurement{meas{}}
}

type meas struct{}

// LineProto not implemented.
func (meas) LineProto() (*dkpt.Point, error) {
	return nil, fmt.Errorf("dk not implement interface LineProto()")
}

//nolint:lll
func (meas) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name: "dk",
		Fields: map[string]interface{}{
			"datakit_dns_domain_total":                         &inputs.FieldInfo{Type: inputs.Count, DataType: inputs.Float, Desc: "DNS watched domain counter"},
			"datakit_dns_ip_updated_total":                     &inputs.FieldInfo{Type: inputs.Count, DataType: inputs.Float, Desc: "Domain IP updated counter"},
			"datakit_dns_watch_run_total":                      &inputs.FieldInfo{Type: inputs.Count, DataType: inputs.Float, Desc: "Watch run counter"},
			"datakit_dns_cost_seconds":                         &inputs.FieldInfo{Type: inputs.Summary, DataType: inputs.Float, Desc: "DNS IP lookup cost"},
			"datakit_election_pause_total":                     &inputs.FieldInfo{Type: inputs.Count, DataType: inputs.Float, Desc: "Input paused count when election failed"},
			"datakit_election_resume_total":                    &inputs.FieldInfo{Type: inputs.Count, DataType: inputs.Float, Desc: "Input resume count when election OK"},
			"datakit_election_status":                          &inputs.FieldInfo{Type: inputs.Gauge, DataType: inputs.Float, Desc: "Datakit election status, if metric = 0, meas not elected, or the elected time(unix timestamp second)"},
			"datakit_election_inputs":                          &inputs.FieldInfo{Type: inputs.Gauge, DataType: inputs.Float, Desc: "Datakit election input count"},
			"datakit_election_seconds":                         &inputs.FieldInfo{Type: inputs.Summary, DataType: inputs.Float, Desc: "Election latency"},
			"datakit_goroutine_alive":                          &inputs.FieldInfo{Type: inputs.Gauge, DataType: inputs.Float, Desc: "Alive Goroutines"},
			"datakit_goroutine_stopped_total":                  &inputs.FieldInfo{Type: inputs.Count, DataType: inputs.Float, Desc: "Stopped Goroutines"},
			"datakit_goroutine_groups":                         &inputs.FieldInfo{Type: inputs.Gauge, DataType: inputs.Float, Desc: "Goroutine group count"},
			"datakit_goroutine_cost_seconds":                   &inputs.FieldInfo{Type: inputs.Summary, DataType: inputs.Float, Desc: "Goroutine running duration"},
			"datakit_http_api_elapsed_seconds":                 &inputs.FieldInfo{Type: inputs.Summary, DataType: inputs.Float, Desc: "API request cost"},
			"datakit_http_api_req_size_bytes":                  &inputs.FieldInfo{Type: inputs.Summary, DataType: inputs.Float, Desc: "API request body size"},
			"datakit_http_api_total":                           &inputs.FieldInfo{Type: inputs.Count, DataType: inputs.Float, Desc: "API request counter"},
			"datakit_httpcli_tcp_conn_total":                   &inputs.FieldInfo{Type: inputs.Count, DataType: inputs.Float, Desc: "HTTP TCP connection count"},
			"datakit_httpcli_conn_reused_from_idle_total":      &inputs.FieldInfo{Type: inputs.Count, DataType: inputs.Float, Desc: "HTTP connection reused from idle count"},
			"datakit_httpcli_conn_idle_time_seconds":           &inputs.FieldInfo{Type: inputs.Summary, DataType: inputs.Float, Desc: "HTTP connection idle time"},
			"datakit_httpcli_dns_cost_seconds":                 &inputs.FieldInfo{Type: inputs.Summary, DataType: inputs.Float, Desc: "HTTP DNS cost"},
			"datakit_httpcli_tls_handshake_seconds":            &inputs.FieldInfo{Type: inputs.Summary, DataType: inputs.Float, Desc: "HTTP TLS handshake cost"},
			"datakit_httpcli_http_connect_cost_seconds":        &inputs.FieldInfo{Type: inputs.Summary, DataType: inputs.Float, Desc: "HTTP connect cost"},
			"datakit_httpcli_got_first_resp_byte_cost_seconds": &inputs.FieldInfo{Type: inputs.Summary, DataType: inputs.Float, Desc: "Got first response byte cost"},
			"datakit_io_http_retry_total":                      &inputs.FieldInfo{Type: inputs.Count, DataType: inputs.Float, Desc: "Dataway HTTP retried count"},
			"datakit_io_dataway_sink_total":                    &inputs.FieldInfo{Type: inputs.Count, DataType: inputs.Float, Desc: "Dataway Sinked count, partitioned by category."},
			"datakit_io_dataway_not_sink_point_total":          &inputs.FieldInfo{Type: inputs.Count, DataType: inputs.Float, Desc: "Dataway not-Sinked points(condition or category not match)"},
			"datakit_io_dataway_sink_point_total":              &inputs.FieldInfo{Type: inputs.Count, DataType: inputs.Float, Desc: "Dataway Sinked points, partitioned by category and point send status(ok/failed/dropped)"},
			"datakit_io_flush_failcache_bytes":                 &inputs.FieldInfo{Type: inputs.Summary, DataType: inputs.Float, Desc: "IO flush fail-cache bytes(in gzip) summary"},
			"datakit_io_dataway_point_total":                   &inputs.FieldInfo{Type: inputs.Count, DataType: inputs.Float, Desc: "Dataway uploaded points, partitioned by category and send status(HTTP status)"},
			"datakit_io_dataway_point_bytes_total":             &inputs.FieldInfo{Type: inputs.Count, DataType: inputs.Float, Desc: "Dataway uploaded points bytes, partitioned by category and pint send status(HTTP status)"},
			"datakit_io_dataway_api_latency_seconds":           &inputs.FieldInfo{Type: inputs.Summary, DataType: inputs.Float, Desc: "Dataway HTTP request latency partitioned by HTTP API(method@url) and HTTP status"},
			"datakit_filter_last_update_timestamp_seconds":     &inputs.FieldInfo{Type: inputs.Gauge, DataType: inputs.Float, Desc: "Filter last update time"},
			"datakit_filter_point_total":                       &inputs.FieldInfo{Type: inputs.Count, DataType: inputs.Float, Desc: "Filter points of filters"},
			"datakit_filter_point_dropped_total":               &inputs.FieldInfo{Type: inputs.Count, DataType: inputs.Float, Desc: "Dropped points of filters"},
			"datakit_filter_pull_latency_seconds":              &inputs.FieldInfo{Type: inputs.Summary, DataType: inputs.Float, Desc: "Filter pull(remote) latency"},
			"datakit_filter_latency_seconds":                   &inputs.FieldInfo{Type: inputs.Summary, DataType: inputs.Float, Desc: "Filter latency of these filters"},
			"datakit_filter_update_total":                      &inputs.FieldInfo{Type: inputs.Count, DataType: inputs.Float, Desc: "Filters(remote) updated count"},
			"datakit_error_total":                              &inputs.FieldInfo{Type: inputs.Count, DataType: inputs.Float, Desc: "Total errors, only count on error source, not include error message"},
			"datakit_io_feed_point_total":                      &inputs.FieldInfo{Type: inputs.Count, DataType: inputs.Float, Desc: "Input feed point total"},
			"datakit_io_input_filter_point_total":              &inputs.FieldInfo{Type: inputs.Count, DataType: inputs.Float, Desc: "Input filtered point total"},
			"datakit_io_feed_total":                            &inputs.FieldInfo{Type: inputs.Count, DataType: inputs.Float, Desc: "Input feed total"},
			"datakit_io_last_feed_timestamp_seconds":           &inputs.FieldInfo{Type: inputs.Gauge, DataType: inputs.Float, Desc: "Input last feed time(according to Datakit local time)"},
			"datakit_input_collect_latency_seconds":            &inputs.FieldInfo{Type: inputs.Summary, DataType: inputs.Float, Desc: "Input collect latency"},
			"datakit_io_chan_usage":                            &inputs.FieldInfo{Type: inputs.Gauge, DataType: inputs.Float, Desc: "IO channel usage(length of the channel)"},
			"datakit_io_chan_capacity":                         &inputs.FieldInfo{Type: inputs.Gauge, DataType: inputs.Float, Desc: "IO channel capacity"},
			"datakit_io_feed_cost_seconds":                     &inputs.FieldInfo{Type: inputs.Summary, DataType: inputs.Float, Desc: "IO feed waiting(on block mode) seconds"},
			"datakit_io_feed_drop_point_total":                 &inputs.FieldInfo{Type: inputs.Count, DataType: inputs.Float, Desc: "IO feed drop(on non-block mode) points"},
			"datakit_io_flush_workers":                         &inputs.FieldInfo{Type: inputs.Gauge, DataType: inputs.Float, Desc: "IO flush workers"},
			"datakit_io_flush_total":                           &inputs.FieldInfo{Type: inputs.Count, DataType: inputs.Float, Desc: "IO flush total"},
			"datakit_io_queue_points":                          &inputs.FieldInfo{Type: inputs.Gauge, DataType: inputs.Float, Desc: "IO module queued(cached) points"},
			"datakit_last_err":                                 &inputs.FieldInfo{Type: inputs.Gauge, DataType: inputs.Float, Desc: "Datakit errors(when error occurred), these errors come from inputs or any sub modules"},
			"datakit_goroutines":                               &inputs.FieldInfo{Type: inputs.Gauge, DataType: inputs.Float, Desc: "Goroutine count within Datakit"},
			"datakit_heap_alloc_bytes":                         &inputs.FieldInfo{Type: inputs.Gauge, DataType: inputs.Float, Desc: "Datakit memory heap bytes"},
			"datakit_sys_alloc_bytes":                          &inputs.FieldInfo{Type: inputs.Gauge, DataType: inputs.Float, Desc: "Datakit memory system bytes"},
			"datakit_cpu_usage":                                &inputs.FieldInfo{Type: inputs.Gauge, DataType: inputs.Float, Desc: "Datakit CPU usage(%)"},
			"datakit_open_files":                               &inputs.FieldInfo{Type: inputs.Gauge, DataType: inputs.Float, Desc: "Datakit open files(only available on Linux)"},
			"datakit_cpu_cores":                                &inputs.FieldInfo{Type: inputs.Gauge, DataType: inputs.Float, Desc: "Datakit CPU cores"},
			"datakit_uptime_seconds":                           &inputs.FieldInfo{Type: inputs.Gauge, DataType: inputs.Float, Desc: "Datakit uptime"},
			"datakit_data_overuse":                             &inputs.FieldInfo{Type: inputs.Gauge, DataType: inputs.Float, Desc: "Does current workspace's data(metric/logging) usage(if 0 not beyond, or with a unix timestamp when overuse occurred)"},
			"datakit_process_ctx_switch_total":                 &inputs.FieldInfo{Type: inputs.Count, DataType: inputs.Float, Desc: "Datakit process context switch count(Linux only)"},
			"datakit_process_io_count_total":                   &inputs.FieldInfo{Type: inputs.Count, DataType: inputs.Float, Desc: "Datakit process IO count"},
			"datakit_process_io_bytes_total":                   &inputs.FieldInfo{Type: inputs.Count, DataType: inputs.Float, Desc: "Datakit process IO bytes count"},
			"datakit_pipeline_point_total":                     &inputs.FieldInfo{Type: inputs.Count, DataType: inputs.Float, Desc: "Pipeline processed total points"},
			"datakit_pipeline_drop_point_total":                &inputs.FieldInfo{Type: inputs.Count, DataType: inputs.Float, Desc: "Pipeline total dropped points"},
			"datakit_pipeline_error_point_total":               &inputs.FieldInfo{Type: inputs.Count, DataType: inputs.Float, Desc: "Pipeline processed total error points"},
			"datakit_pipeline_cost_seconds":                    &inputs.FieldInfo{Type: inputs.Summary, DataType: inputs.Float, Desc: "Pipeline total running time"},
			"datakit_pipeline_last_update_timestamp_seconds":   &inputs.FieldInfo{Type: inputs.Gauge, DataType: inputs.Float, Desc: "Pipeline last update time"},
			"datakit_dialtesting_task_number":                  &inputs.FieldInfo{Type: inputs.Gauge, DataType: inputs.Float, Desc: "The number of tasks"},
			"datakit_dialtesting_pull_cost_seconds":            &inputs.FieldInfo{Type: inputs.Summary, DataType: inputs.Float, Desc: "Time cost to pull tasks"},
			"datakit_dialtesting_task_synchronized_total":      &inputs.FieldInfo{Type: inputs.Count, DataType: inputs.Float, Desc: "Task synchronized number"},
			"datakit_dialtesting_task_invalid_total":           &inputs.FieldInfo{Type: inputs.Count, DataType: inputs.Float, Desc: "Invalid task number"},
			"datakit_dialtesting_task_check_cost_seconds":      &inputs.FieldInfo{Type: inputs.Summary, DataType: inputs.Float, Desc: "Task check time"},
			"datakit_dialtesting_task_run_cost_seconds":        &inputs.FieldInfo{Type: inputs.Summary, DataType: inputs.Float, Desc: "Task run time"},
			"datakit_kafkamq_consumer_message_total":           &inputs.FieldInfo{Type: inputs.Count, DataType: inputs.Float, Desc: "Kafka consumer message numbers from Datakit start"},
			"datakit_kafkamq_group_election_total":             &inputs.FieldInfo{Type: inputs.Count, DataType: inputs.Float, Desc: "Kafka group election count"},
			"datakit_inputs_instance":                          &inputs.FieldInfo{Type: inputs.Gauge, DataType: inputs.Float, Desc: "Input instance count"},
			"datakit_inputs_crash_total":                       &inputs.FieldInfo{Type: inputs.Count, DataType: inputs.Float, Desc: "Input crash count"},
			"datakit_prom_collect_points":                      &inputs.FieldInfo{Type: inputs.Summary, DataType: inputs.Float, Desc: "Total number of prom collection points"},
			"datakit_prom_http_get_bytes":                      &inputs.FieldInfo{Type: inputs.Summary, DataType: inputs.Float, Desc: "HTTP get bytes"},
			"datakit_prom_http_latency_in_second":              &inputs.FieldInfo{Type: inputs.Summary, DataType: inputs.Float, Desc: "HTTP latency(in second)"},
		},
		Tags: nil,
	}
}
