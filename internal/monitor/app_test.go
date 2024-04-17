// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package monitor

import (
	"net/http"
	"net/http/httptest"
	T "testing"
)

func TestApp(t *T.T) {
	t.Skip()
	t.Run("app", func(t *T.T) {
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(200)
			w.Write(metricsData)
		}))

		Start(WithHost("http", ts.Listener.Addr().String()), WithVerbose(true), WithMaxRun(1))
		t.Cleanup(func() {
			ts.Close()
		})
	})
}

func TestAppOnNilData(t *T.T) {
	t.Skip()

	emptyMetrics := []byte(`# HELP datakit_cpu_usage Datakit CPU usage(%)
# TYPE datakit_cpu_usage gauge
`)

	t.Run("app", func(t *T.T) {
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(200)
			w.Write(emptyMetrics)
		}))

		Start(WithHost("http", ts.Listener.Addr().String()), WithVerbose(true), WithMaxRun(1))
		t.Cleanup(func() {
			ts.Close()
		})
	})
}

var metricsData = []byte(`
# HELP datakit_cpu_cores Datakit CPU cores
# TYPE datakit_cpu_cores gauge
datakit_cpu_cores 10
# HELP datakit_cpu_usage Datakit CPU usage(%)
# TYPE datakit_cpu_usage gauge
datakit_cpu_usage 0.3665847825475623
# HELP datakit_data_overuse Does current workspace's data(metric/logging) usage(if 0 not beyond, or with a unix timestamp when overuse occurred)
# TYPE datakit_data_overuse gauge
datakit_data_overuse 0
# HELP datakit_dialtesting_task_check_cost_seconds Task check time
# TYPE datakit_dialtesting_task_check_cost_seconds summary
datakit_dialtesting_task_check_cost_seconds_sum{protocol="HTTP",region="default",status="FAIL"} 0.000238042
datakit_dialtesting_task_check_cost_seconds_count{protocol="HTTP",region="default",status="FAIL"} 1
datakit_dialtesting_task_check_cost_seconds_sum{protocol="HTTP",region="default",status="OK"} 0.0024858760000000006
datakit_dialtesting_task_check_cost_seconds_count{protocol="HTTP",region="default",status="OK"} 18
# HELP datakit_dialtesting_task_number The number of tasks
# TYPE datakit_dialtesting_task_number gauge
datakit_dialtesting_task_number{protocol="HTTP",region="default"} 1
# HELP datakit_dialtesting_task_run_cost_seconds Task run time
# TYPE datakit_dialtesting_task_run_cost_seconds summary
datakit_dialtesting_task_run_cost_seconds_sum{protocol="HTTP",region="default"} 1.7180489579999998
datakit_dialtesting_task_run_cost_seconds_count{protocol="HTTP",region="default"} 19
# HELP datakit_dialtesting_task_synchronized_total Task synchronized number
# TYPE datakit_dialtesting_task_synchronized_total counter
datakit_dialtesting_task_synchronized_total{protocol="HTTP",region=""} 1
# HELP datakit_dns_cost_seconds DNS IP lookup cost
# TYPE datakit_dns_cost_seconds summary
datakit_dns_cost_seconds_sum{domain="openway.guance.com",status="ok"} 0.025607542
datakit_dns_cost_seconds_count{domain="openway.guance.com",status="ok"} 4
# HELP datakit_dns_domain_total DNS watched domain counter
# TYPE datakit_dns_domain_total counter
datakit_dns_domain_total 1
# HELP datakit_dns_ip_updated_total Domain IP updated counter
# TYPE datakit_dns_ip_updated_total counter
datakit_dns_ip_updated_total{domain="openway.guance.com"} 1
# HELP datakit_dns_watch_run_total Watch run counter
# TYPE datakit_dns_watch_run_total counter
datakit_dns_watch_run_total{interval="1m0s"} 4
# HELP datakit_election_status Datakit election status, if metric = 0, meas not elected, or the elected time(unix timestamp second)
# TYPE datakit_election_status gauge
datakit_election_status{elected_id="<checking...>",id="tanbiaos-MacBook-Pro.local",namespace="default",status="disabled"} 0
# HELP datakit_error_total Total errors, only count on error source, not include error message
# TYPE datakit_error_total counter
datakit_error_total{category="object",source="hostobject"} 1
# HELP datakit_filter_last_update_timestamp_seconds Filter last update time
# TYPE datakit_filter_last_update_timestamp_seconds gauge
datakit_filter_last_update_timestamp_seconds 1.684823227e+09
# HELP datakit_filter_latency_seconds Filter latency of these filters
# TYPE datakit_filter_latency_seconds summary
datakit_filter_latency_seconds_sum{category="metric",filters="{ measurement =  'cpu'  and ( host notin [ 'ubt-dev-01' ] )}",source="remote"} 0.013598088999999997
datakit_filter_latency_seconds_count{category="metric",filters="{ measurement =  'cpu'  and ( host notin [ 'ubt-dev-01' ] )}",source="remote"} 118
# HELP datakit_filter_point_dropped_total Dropped points of filters
# TYPE datakit_filter_point_dropped_total counter
datakit_filter_point_dropped_total{category="metric",filters="{ measurement =  'cpu'  and ( host notin [ 'ubt-dev-01' ] )}",source="remote"} 18
# HELP datakit_filter_point_total Filter points of filters
# TYPE datakit_filter_point_total counter
datakit_filter_point_total{category="metric",filters="{ measurement =  'cpu'  and ( host notin [ 'ubt-dev-01' ] )}",source="remote"} 2311
# HELP datakit_filter_pull_latency_seconds Filter pull(remote) latency
# TYPE datakit_filter_pull_latency_seconds summary
datakit_filter_pull_latency_seconds_sum{status="ok"} 0.9842480429999999
datakit_filter_pull_latency_seconds_count{status="ok"} 19
# HELP datakit_filter_update_total Filters(remote) updated count
# TYPE datakit_filter_update_total counter
datakit_filter_update_total 1
# HELP datakit_gc_summary_seconds Datakit golang GC paused
# TYPE datakit_gc_summary_seconds summary
datakit_gc_summary_seconds_sum 0.007234254
datakit_gc_summary_seconds_count 23
# HELP datakit_goroutine_alive Alive Goroutines
# TYPE datakit_goroutine_alive gauge
datakit_goroutine_alive{name="http"} 3
datakit_goroutine_alive{name="inputs"} 11
datakit_goroutine_alive{name="inputs_cpu"} 1
datakit_goroutine_alive{name="inputs_dialtesting"} 1
datakit_goroutine_alive{name="internal_cgroup"} 0
datakit_goroutine_alive{name="io"} 27
datakit_goroutine_alive{name="io_dnswatcher"} 1
datakit_goroutine_alive{name="pipeline_remote"} 1
# HELP datakit_goroutine_cost_seconds Goroutine running duration
# TYPE datakit_goroutine_cost_seconds summary
datakit_goroutine_cost_seconds_sum{name="internal_cgroup"} 7.6292e-05
datakit_goroutine_cost_seconds_count{name="internal_cgroup"} 1
# HELP datakit_goroutine_groups Goroutine group count
# TYPE datakit_goroutine_groups gauge
datakit_goroutine_groups 22
# HELP datakit_goroutine_stopped_total Stopped Goroutines
# TYPE datakit_goroutine_stopped_total counter
datakit_goroutine_stopped_total{name="internal_cgroup"} 1
# HELP datakit_goroutines Goroutine count within Datakit
# TYPE datakit_goroutines gauge
datakit_goroutines 62
# HELP datakit_heap_alloc_bytes Datakit memory heap bytes
# TYPE datakit_heap_alloc_bytes gauge
datakit_heap_alloc_bytes 1.6564832e+07
# HELP datakit_http_api_elapsed_histogram_seconds API request cost histogram
# TYPE datakit_http_api_elapsed_histogram_seconds histogram
datakit_http_api_elapsed_histogram_seconds_bucket{api="/metrics",method="GET",status="OK",le="10"} 11
datakit_http_api_elapsed_histogram_seconds_bucket{api="/metrics",method="GET",status="OK",le="100"} 11
datakit_http_api_elapsed_histogram_seconds_bucket{api="/metrics",method="GET",status="OK",le="1000"} 11
datakit_http_api_elapsed_histogram_seconds_bucket{api="/metrics",method="GET",status="OK",le="5000"} 11
datakit_http_api_elapsed_histogram_seconds_bucket{api="/metrics",method="GET",status="OK",le="30000"} 11
datakit_http_api_elapsed_histogram_seconds_bucket{api="/metrics",method="GET",status="OK",le="+Inf"} 11
datakit_http_api_elapsed_histogram_seconds_sum{api="/metrics",method="GET",status="OK"} 11.057606917
datakit_http_api_elapsed_histogram_seconds_count{api="/metrics",method="GET",status="OK"} 11
# HELP datakit_http_api_elapsed_seconds API request cost
# TYPE datakit_http_api_elapsed_seconds summary
datakit_http_api_elapsed_seconds_sum{api="/metrics",method="GET",status="OK"} 11.057729251999998
datakit_http_api_elapsed_seconds_count{api="/metrics",method="GET",status="OK"} 11
# HELP datakit_http_api_req_size_bytes API request body size
# TYPE datakit_http_api_req_size_bytes summary
datakit_http_api_req_size_bytes_sum{api="/metrics",method="GET",status="OK"} 812
datakit_http_api_req_size_bytes_count{api="/metrics",method="GET",status="OK"} 11
# HELP datakit_http_api_total API request counter
# TYPE datakit_http_api_total counter
datakit_http_api_total{api="/metrics",method="GET",status="OK"} 11
# HELP datakit_input_collect_latency_seconds Input collect latency
# TYPE datakit_input_collect_latency_seconds summary
datakit_input_collect_latency_seconds_sum{category="logging",name="dialtesting"} 0
datakit_input_collect_latency_seconds_count{category="logging",name="dialtesting"} 19
datakit_input_collect_latency_seconds_sum{category="metric",name="cpu"} 0.003283752
datakit_input_collect_latency_seconds_count{category="metric",name="cpu"} 18
datakit_input_collect_latency_seconds_sum{category="metric",name="disk"} 0.056611001
datakit_input_collect_latency_seconds_count{category="metric",name="disk"} 19
datakit_input_collect_latency_seconds_sum{category="metric",name="diskio"} 0.028365121000000004
datakit_input_collect_latency_seconds_count{category="metric",name="diskio"} 19
datakit_input_collect_latency_seconds_sum{category="metric",name="mem"} 0.003911626
datakit_input_collect_latency_seconds_count{category="metric",name="mem"} 19
datakit_input_collect_latency_seconds_sum{category="metric",name="net"} 0.23410908299999997
datakit_input_collect_latency_seconds_count{category="metric",name="net"} 18
datakit_input_collect_latency_seconds_sum{category="metric",name="prom/dk-prom"} 7.107598209
datakit_input_collect_latency_seconds_count{category="metric",name="prom/dk-prom"} 7
datakit_input_collect_latency_seconds_sum{category="metric",name="swap"} 0.002507541
datakit_input_collect_latency_seconds_count{category="metric",name="swap"} 18
datakit_input_collect_latency_seconds_sum{category="object",name="host_processes/object"} 4.88061
datakit_input_collect_latency_seconds_count{category="object",name="host_processes/object"} 7
# HELP datakit_inputs_instance_total Input instance count
# TYPE datakit_inputs_instance_total gauge
datakit_inputs_instance_total{input="cpu"} 1
datakit_inputs_instance_total{input="ddtrace"} 1
datakit_inputs_instance_total{input="dialtesting"} 1
datakit_inputs_instance_total{input="disk"} 1
datakit_inputs_instance_total{input="diskio"} 1
datakit_inputs_instance_total{input="host_processes"} 1
datakit_inputs_instance_total{input="hostobject"} 1
datakit_inputs_instance_total{input="mem"} 1
datakit_inputs_instance_total{input="net"} 1
datakit_inputs_instance_total{input="prom"} 1
datakit_inputs_instance_total{input="swap"} 1
# HELP datakit_io_chan_capacity IO channel capacity
# TYPE datakit_io_chan_capacity gauge
datakit_io_chan_capacity{category="all-the-same"} 100
# HELP datakit_io_chan_usage IO channel usage(length of the channel)
# TYPE datakit_io_chan_usage gauge
datakit_io_chan_usage{category="dynamic_dw"} 0
datakit_io_chan_usage{category="metric"} 0
datakit_io_chan_usage{category="object"} 0
# HELP datakit_io_dataway_api_latency_seconds Dataway HTTP request latency partitioned by HTTP API(method@url) and HTTP status
# TYPE datakit_io_dataway_api_latency_seconds summary
datakit_io_dataway_api_latency_seconds_sum{api="/v1/check/token/tkn_2af4b19d7f5a489fa81f0fff7e63b588",status="OK"} 0.095935333
datakit_io_dataway_api_latency_seconds_count{api="/v1/check/token/tkn_2af4b19d7f5a489fa81f0fff7e63b588",status="OK"} 1
datakit_io_dataway_api_latency_seconds_sum{api="/v1/datakit/pull",status="OK"} 1.2907557090000001
datakit_io_dataway_api_latency_seconds_count{api="/v1/datakit/pull",status="OK"} 23
datakit_io_dataway_api_latency_seconds_sum{api="/v1/write/logging",status="OK"} 0.9108642889999999
datakit_io_dataway_api_latency_seconds_count{api="/v1/write/logging",status="OK"} 19
datakit_io_dataway_api_latency_seconds_sum{api="/v1/write/metric",status="OK"} 2.159057043
datakit_io_dataway_api_latency_seconds_count{api="/v1/write/metric",status="OK"} 24
datakit_io_dataway_api_latency_seconds_sum{api="/v1/write/object",status="OK"} 0.9048214579999999
datakit_io_dataway_api_latency_seconds_count{api="/v1/write/object",status="OK"} 6
# HELP datakit_io_dataway_api_request_total Dataway HTTP request processed, partitioned by status code and HTTP API(url path)
# TYPE datakit_io_dataway_api_request_total counter
datakit_io_dataway_api_request_total{api="/v1/check/token/tkn_2af4b19d7f5a489fa81f0fff7e63b588",status="OK"} 1
datakit_io_dataway_api_request_total{api="/v1/datakit/pull",status="OK"} 23
datakit_io_dataway_api_request_total{api="/v1/write/logging",status="OK"} 19
datakit_io_dataway_api_request_total{api="/v1/write/metric",status="OK"} 24
datakit_io_dataway_api_request_total{api="/v1/write/object",status="OK"} 6
# HELP datakit_io_dataway_not_sink_point_total Dataway not-Sinked points(condition or category not match)
# TYPE datakit_io_dataway_not_sink_point_total counter
datakit_io_dataway_not_sink_point_total{category="metric"} 1996
datakit_io_dataway_not_sink_point_total{category="object"} 2057
datakit_io_dataway_not_sink_point_total{category="unknown"} 19
# HELP datakit_io_dataway_point_bytes_total Dataway uploaded points bytes, partitioned by category and pint send status(HTTP status)
# TYPE datakit_io_dataway_point_bytes_total counter
datakit_io_dataway_point_bytes_total{category="dynamic_dw",enc="gzip",status="OK"} 7389
datakit_io_dataway_point_bytes_total{category="dynamic_dw",enc="gzip",status="total"} 7389
datakit_io_dataway_point_bytes_total{category="dynamic_dw",enc="raw",status="OK"} 10485
datakit_io_dataway_point_bytes_total{category="dynamic_dw",enc="raw",status="total"} 10485
datakit_io_dataway_point_bytes_total{category="metric",enc="gzip",status="OK"} 44836
datakit_io_dataway_point_bytes_total{category="metric",enc="gzip",status="total"} 44836
datakit_io_dataway_point_bytes_total{category="metric",enc="raw",status="OK"} 415614
datakit_io_dataway_point_bytes_total{category="metric",enc="raw",status="total"} 415614
datakit_io_dataway_point_bytes_total{category="object",enc="gzip",status="OK"} 279011
datakit_io_dataway_point_bytes_total{category="object",enc="gzip",status="total"} 279011
datakit_io_dataway_point_bytes_total{category="object",enc="raw",status="OK"} 2.256364e+06
datakit_io_dataway_point_bytes_total{category="object",enc="raw",status="total"} 2.256364e+06
# HELP datakit_io_dataway_point_total Dataway uploaded points, partitioned by category and send status(HTTP status)
# TYPE datakit_io_dataway_point_total counter
datakit_io_dataway_point_total{category="dynamic_dw",status="OK"} 19
datakit_io_dataway_point_total{category="dynamic_dw",status="total"} 19
datakit_io_dataway_point_total{category="metric",status="OK"} 1996
datakit_io_dataway_point_total{category="metric",status="total"} 1996
datakit_io_dataway_point_total{category="object",status="OK"} 2057
datakit_io_dataway_point_total{category="object",status="total"} 2057
# HELP datakit_io_dataway_sink_total Dataway Sinked count, partitioned by category.
# TYPE datakit_io_dataway_sink_total counter
datakit_io_dataway_sink_total{category="metric"} 24
datakit_io_dataway_sink_total{category="object"} 6
datakit_io_dataway_sink_total{category="unknown"} 19
# HELP datakit_io_feed_point_total Input feed point total
# TYPE datakit_io_feed_point_total counter
datakit_io_feed_point_total{category="logging",name="dialtesting"} 19
datakit_io_feed_point_total{category="metric",name="cpu"} 18
datakit_io_feed_point_total{category="metric",name="disk"} 190
datakit_io_feed_point_total{category="metric",name="diskio"} 19
datakit_io_feed_point_total{category="metric",name="mem"} 19
datakit_io_feed_point_total{category="metric",name="net"} 306
datakit_io_feed_point_total{category="metric",name="prom/dk-prom"} 1741
datakit_io_feed_point_total{category="metric",name="swap"} 18
datakit_io_feed_point_total{category="object",name="host_processes/object"} 2404
# HELP datakit_io_feed_total Input feed total
# TYPE datakit_io_feed_total counter
datakit_io_feed_total{category="logging",name="dialtesting"} 19
datakit_io_feed_total{category="metric",name="cpu"} 18
datakit_io_feed_total{category="metric",name="disk"} 19
datakit_io_feed_total{category="metric",name="diskio"} 19
datakit_io_feed_total{category="metric",name="mem"} 19
datakit_io_feed_total{category="metric",name="net"} 18
datakit_io_feed_total{category="metric",name="prom/dk-prom"} 7
datakit_io_feed_total{category="metric",name="swap"} 18
datakit_io_feed_total{category="object",name="host_processes/object"} 7
datakit_io_feed_total{category="object",name="hostobject"} 1
# HELP datakit_io_flush_total IO flush total
# TYPE datakit_io_flush_total counter
datakit_io_flush_total{category="logging"} 19
datakit_io_flush_total{category="metric"} 24
datakit_io_flush_total{category="object"} 6
# HELP datakit_io_flush_workers IO flush workers
# TYPE datakit_io_flush_workers gauge
datakit_io_flush_workers{category="custom_object"} 1
datakit_io_flush_workers{category="keyevent"} 1
datakit_io_flush_workers{category="logging"} 4
datakit_io_flush_workers{category="metric"} 4
datakit_io_flush_workers{category="network"} 4
datakit_io_flush_workers{category="object"} 1
datakit_io_flush_workers{category="profiling"} 1
datakit_io_flush_workers{category="rum"} 4
datakit_io_flush_workers{category="security"} 1
datakit_io_flush_workers{category="tracing"} 4
datakit_io_flush_workers{category="unknown"} 1
# HELP datakit_io_http_conn_idle_time_seconds Dataway HTTP connection idle time
# TYPE datakit_io_http_conn_idle_time_seconds summary
datakit_io_http_conn_idle_time_seconds_sum 177.790835169
datakit_io_http_conn_idle_time_seconds_count 73
# HELP datakit_io_http_conn_reused_from_idle_total Dataway HTTP connection reused from idle count
# TYPE datakit_io_http_conn_reused_from_idle_total counter
datakit_io_http_conn_reused_from_idle_total 43
# HELP datakit_io_http_connect_cost_seconds Dataway HTTP connect cost
# TYPE datakit_io_http_connect_cost_seconds summary
datakit_io_http_connect_cost_seconds_sum 0.497708669
datakit_io_http_connect_cost_seconds_count 73
# HELP datakit_io_http_dns_cost_seconds Dataway HTTP DNS cost
# TYPE datakit_io_http_dns_cost_seconds summary
datakit_io_http_dns_cost_seconds_sum 0.036188208
datakit_io_http_dns_cost_seconds_count 73
# HELP datakit_io_http_got_first_resp_byte_cost_seconds Dataway got first response byte cost
# TYPE datakit_io_http_got_first_resp_byte_cost_seconds summary
datakit_io_http_got_first_resp_byte_cost_seconds{quantile="0.5"} 38
datakit_io_http_got_first_resp_byte_cost_seconds{quantile="0.75"} 53
datakit_io_http_got_first_resp_byte_cost_seconds{quantile="0.95"} 92
datakit_io_http_got_first_resp_byte_cost_seconds_sum 3580
datakit_io_http_got_first_resp_byte_cost_seconds_count 73
# HELP datakit_io_http_tcp_conn_total Dataway HTTP TCP connection count
# TYPE datakit_io_http_tcp_conn_total counter
datakit_io_http_tcp_conn_total{remote="118.31.126.76:443",type="created"} 6
datakit_io_http_tcp_conn_total{remote="118.31.126.76:443",type="reused"} 30
datakit_io_http_tcp_conn_total{remote="47.110.144.10:443",type="created"} 3
datakit_io_http_tcp_conn_total{remote="47.110.144.10:443",type="reused"} 10
datakit_io_http_tcp_conn_total{remote="47.114.74.166:443",type="created"} 7
datakit_io_http_tcp_conn_total{remote="47.114.74.166:443",type="reused"} 17
# HELP datakit_io_http_tls_handshake_seconds Dataway TLS handshake cost
# TYPE datakit_io_http_tls_handshake_seconds summary
datakit_io_http_tls_handshake_seconds_sum 1.411738
datakit_io_http_tls_handshake_seconds_count 73
# HELP datakit_io_input_filter_point_total Input filtered point total
# TYPE datakit_io_input_filter_point_total counter
datakit_io_input_filter_point_total{category="logging",name="dialtesting"} 0
datakit_io_input_filter_point_total{category="metric",name="cpu"} 18
datakit_io_input_filter_point_total{category="metric",name="disk"} 0
datakit_io_input_filter_point_total{category="metric",name="diskio"} 0
datakit_io_input_filter_point_total{category="metric",name="mem"} 0
datakit_io_input_filter_point_total{category="metric",name="net"} 0
datakit_io_input_filter_point_total{category="metric",name="prom/dk-prom"} 0
datakit_io_input_filter_point_total{category="metric",name="swap"} 0
datakit_io_input_filter_point_total{category="object",name="host_processes/object"} 0
# HELP datakit_io_last_feed_timestamp_seconds Input last feed time(according to Datakit local time)
# TYPE datakit_io_last_feed_timestamp_seconds gauge
datakit_io_last_feed_timestamp_seconds{category="logging",name="dialtesting"} 1.684823416e+09
datakit_io_last_feed_timestamp_seconds{category="metric",name="cpu"} 1.684823409e+09
datakit_io_last_feed_timestamp_seconds{category="metric",name="disk"} 1.684823415e+09
datakit_io_last_feed_timestamp_seconds{category="metric",name="diskio"} 1.68482341e+09
datakit_io_last_feed_timestamp_seconds{category="metric",name="mem"} 1.684823412e+09
datakit_io_last_feed_timestamp_seconds{category="metric",name="net"} 1.684823408e+09
datakit_io_last_feed_timestamp_seconds{category="metric",name="prom/dk-prom"} 1.684823414e+09
datakit_io_last_feed_timestamp_seconds{category="metric",name="swap"} 1.684823408e+09
datakit_io_last_feed_timestamp_seconds{category="object",name="host_processes/object"} 1.684823415e+09
datakit_io_last_feed_timestamp_seconds{category="object",name="hostobject"} 1.684823229e+09
# HELP datakit_io_queue_points IO module queued(cached) points
# TYPE datakit_io_queue_points gauge
datakit_io_queue_points{category="logging"} 0
datakit_io_queue_points{category="metric"} 10
datakit_io_queue_points{category="object"} 347
# HELP datakit_last_err Datakit errors(when error occurred), these errors come from inputs or any sub modules
# TYPE datakit_last_err gauge
datakit_last_err{category="object",error="collector stats missing",source="hostobject"} 1.684823229e+09
# HELP datakit_open_files Datakit open files(only available on Linux)
# TYPE datakit_open_files gauge
datakit_open_files -1
# HELP datakit_pipeline_cost_seconds Pipeline total running time
# TYPE datakit_pipeline_cost_seconds summary
datakit_pipeline_cost_seconds_sum{category="metric",name="cpu.p",namespace="default"} 0.00046662600000000004
datakit_pipeline_cost_seconds_count{category="metric",name="cpu.p",namespace="default"} 18
datakit_pipeline_cost_seconds_sum{category="metric",name="disk.p",namespace="default"} 0.0002983749999999999
datakit_pipeline_cost_seconds_count{category="metric",name="disk.p",namespace="default"} 190
datakit_pipeline_cost_seconds_sum{category="object",name="host_processes.p",namespace="default"} 0.00037552800000001076
datakit_pipeline_cost_seconds_count{category="object",name="host_processes.p",namespace="default"} 2403
# HELP datakit_pipeline_last_update_timestamp_seconds Pipeline last update time
# TYPE datakit_pipeline_last_update_timestamp_seconds gauge
datakit_pipeline_last_update_timestamp_seconds{category="logging",name="apache.p",namespace="default"} 1.684823227e+09
datakit_pipeline_last_update_timestamp_seconds{category="logging",name="consul.p",namespace="default"} 1.684823227e+09
datakit_pipeline_last_update_timestamp_seconds{category="logging",name="elasticsearch.p",namespace="default"} 1.684823227e+09
datakit_pipeline_last_update_timestamp_seconds{category="logging",name="jenkins.p",namespace="default"} 1.684823227e+09
datakit_pipeline_last_update_timestamp_seconds{category="logging",name="kafka.p",namespace="default"} 1.684823227e+09
datakit_pipeline_last_update_timestamp_seconds{category="logging",name="mongodb.p",namespace="default"} 1.684823227e+09
datakit_pipeline_last_update_timestamp_seconds{category="logging",name="mysql.p",namespace="default"} 1.684823227e+09
datakit_pipeline_last_update_timestamp_seconds{category="logging",name="nginx.p",namespace="default"} 1.684823227e+09
datakit_pipeline_last_update_timestamp_seconds{category="logging",name="postgresql.p",namespace="default"} 1.684823227e+09
datakit_pipeline_last_update_timestamp_seconds{category="logging",name="rabbitmq.p",namespace="default"} 1.684823227e+09
datakit_pipeline_last_update_timestamp_seconds{category="logging",name="redis.p",namespace="default"} 1.684823227e+09
datakit_pipeline_last_update_timestamp_seconds{category="logging",name="solr.p",namespace="default"} 1.684823227e+09
datakit_pipeline_last_update_timestamp_seconds{category="logging",name="sqlserver.p",namespace="default"} 1.684823227e+09
datakit_pipeline_last_update_timestamp_seconds{category="logging",name="tdengine.p",namespace="default"} 1.684823227e+09
datakit_pipeline_last_update_timestamp_seconds{category="logging",name="tomcat.p",namespace="default"} 1.684823227e+09
datakit_pipeline_last_update_timestamp_seconds{category="metric",name="cpu.p",namespace="default"} 1.684823227e+09
datakit_pipeline_last_update_timestamp_seconds{category="metric",name="disk.p",namespace="default"} 1.684823227e+09
datakit_pipeline_last_update_timestamp_seconds{category="object",name="host_processes.p",namespace="default"} 1.684823227e+09
# HELP datakit_pipeline_point_total Pipeline processed total points
# TYPE datakit_pipeline_point_total counter
datakit_pipeline_point_total{category="metric",name="cpu.p",namespace="default"} 18
datakit_pipeline_point_total{category="metric",name="disk.p",namespace="default"} 190
datakit_pipeline_point_total{category="object",name="host_processes.p",namespace="default"} 2404
# HELP datakit_sys_alloc_bytes Datakit memory system bytes
# TYPE datakit_sys_alloc_bytes gauge
datakit_sys_alloc_bytes 6.0687368e+07
# HELP datakit_uptime_seconds Datakit uptime
# TYPE datakit_uptime_seconds gauge
datakit_uptime_seconds{auto_update="false",branch="1639-iss-prom-metric-naming",build_at="2023-05-23 06:25:48",cgroup="-",docker="false",hostname="tanbiaos-MacBook-Pro.local",os_arch="darwin/arm64",version="1.6.1-462-gefd7d29427"} 191
# HELP diskcache_batch_size data file size(in bytes)
# TYPE diskcache_batch_size gauge
diskcache_batch_size{no_fallback_on_error="false",no_lock="false",no_pos="false",no_sync="false",path="/Users/tanbiao/datakit/cache/custom_object"} 2.097152e+07
diskcache_batch_size{no_fallback_on_error="false",no_lock="false",no_pos="false",no_sync="false",path="/Users/tanbiao/datakit/cache/dynamic_dw"} 2.097152e+07
diskcache_batch_size{no_fallback_on_error="false",no_lock="false",no_pos="false",no_sync="false",path="/Users/tanbiao/datakit/cache/keyevent"} 2.097152e+07
diskcache_batch_size{no_fallback_on_error="false",no_lock="false",no_pos="false",no_sync="false",path="/Users/tanbiao/datakit/cache/logging"} 2.097152e+07
diskcache_batch_size{no_fallback_on_error="false",no_lock="false",no_pos="false",no_sync="false",path="/Users/tanbiao/datakit/cache/metric"} 2.097152e+07
diskcache_batch_size{no_fallback_on_error="false",no_lock="false",no_pos="false",no_sync="false",path="/Users/tanbiao/datakit/cache/network"} 2.097152e+07
diskcache_batch_size{no_fallback_on_error="false",no_lock="false",no_pos="false",no_sync="false",path="/Users/tanbiao/datakit/cache/object"} 2.097152e+07
diskcache_batch_size{no_fallback_on_error="false",no_lock="false",no_pos="false",no_sync="false",path="/Users/tanbiao/datakit/cache/profiling"} 2.097152e+07
diskcache_batch_size{no_fallback_on_error="false",no_lock="false",no_pos="false",no_sync="false",path="/Users/tanbiao/datakit/cache/rum"} 2.097152e+07
diskcache_batch_size{no_fallback_on_error="false",no_lock="false",no_pos="false",no_sync="false",path="/Users/tanbiao/datakit/cache/security"} 2.097152e+07
diskcache_batch_size{no_fallback_on_error="false",no_lock="false",no_pos="false",no_sync="false",path="/Users/tanbiao/datakit/cache/tracing"} 2.097152e+07
# HELP diskcache_capacity current capacity(in bytes)
# TYPE diskcache_capacity gauge
diskcache_capacity{no_fallback_on_error="false",no_lock="false",no_pos="false",no_sync="false",path="/Users/tanbiao/datakit/cache/custom_object"} 1.125899906842624e+15
diskcache_capacity{no_fallback_on_error="false",no_lock="false",no_pos="false",no_sync="false",path="/Users/tanbiao/datakit/cache/dynamic_dw"} 1.125899906842624e+15
diskcache_capacity{no_fallback_on_error="false",no_lock="false",no_pos="false",no_sync="false",path="/Users/tanbiao/datakit/cache/keyevent"} 1.125899906842624e+15
diskcache_capacity{no_fallback_on_error="false",no_lock="false",no_pos="false",no_sync="false",path="/Users/tanbiao/datakit/cache/logging"} 1.125899906842624e+15
diskcache_capacity{no_fallback_on_error="false",no_lock="false",no_pos="false",no_sync="false",path="/Users/tanbiao/datakit/cache/metric"} 1.125899906842624e+15
diskcache_capacity{no_fallback_on_error="false",no_lock="false",no_pos="false",no_sync="false",path="/Users/tanbiao/datakit/cache/network"} 1.125899906842624e+15
diskcache_capacity{no_fallback_on_error="false",no_lock="false",no_pos="false",no_sync="false",path="/Users/tanbiao/datakit/cache/object"} 1.125899906842624e+15
diskcache_capacity{no_fallback_on_error="false",no_lock="false",no_pos="false",no_sync="false",path="/Users/tanbiao/datakit/cache/profiling"} 1.125899906842624e+15
diskcache_capacity{no_fallback_on_error="false",no_lock="false",no_pos="false",no_sync="false",path="/Users/tanbiao/datakit/cache/rum"} 1.125899906842624e+15
diskcache_capacity{no_fallback_on_error="false",no_lock="false",no_pos="false",no_sync="false",path="/Users/tanbiao/datakit/cache/security"} 1.125899906842624e+15
diskcache_capacity{no_fallback_on_error="false",no_lock="false",no_pos="false",no_sync="false",path="/Users/tanbiao/datakit/cache/tracing"} 1.125899906842624e+15
# HELP diskcache_get_bytes_total cache Get() bytes count
# TYPE diskcache_get_bytes_total counter
diskcache_get_bytes_total{no_fallback_on_error="false",no_lock="false",no_pos="false",no_sync="false",path="/Users/tanbiao/datakit/cache/custom_object"} 0
diskcache_get_bytes_total{no_fallback_on_error="false",no_lock="false",no_pos="false",no_sync="false",path="/Users/tanbiao/datakit/cache/keyevent"} 0
diskcache_get_bytes_total{no_fallback_on_error="false",no_lock="false",no_pos="false",no_sync="false",path="/Users/tanbiao/datakit/cache/logging"} 0
diskcache_get_bytes_total{no_fallback_on_error="false",no_lock="false",no_pos="false",no_sync="false",path="/Users/tanbiao/datakit/cache/metric"} 0
diskcache_get_bytes_total{no_fallback_on_error="false",no_lock="false",no_pos="false",no_sync="false",path="/Users/tanbiao/datakit/cache/network"} 0
diskcache_get_bytes_total{no_fallback_on_error="false",no_lock="false",no_pos="false",no_sync="false",path="/Users/tanbiao/datakit/cache/object"} 0
diskcache_get_bytes_total{no_fallback_on_error="false",no_lock="false",no_pos="false",no_sync="false",path="/Users/tanbiao/datakit/cache/profiling"} 0
diskcache_get_bytes_total{no_fallback_on_error="false",no_lock="false",no_pos="false",no_sync="false",path="/Users/tanbiao/datakit/cache/rum"} 0
diskcache_get_bytes_total{no_fallback_on_error="false",no_lock="false",no_pos="false",no_sync="false",path="/Users/tanbiao/datakit/cache/security"} 0
diskcache_get_bytes_total{no_fallback_on_error="false",no_lock="false",no_pos="false",no_sync="false",path="/Users/tanbiao/datakit/cache/tracing"} 0
# HELP diskcache_get_latency Get() time cost(micro-second)
# TYPE diskcache_get_latency summary
diskcache_get_latency_sum{no_fallback_on_error="false",no_lock="false",no_pos="false",no_sync="false",path="/Users/tanbiao/datakit/cache/custom_object"} 95
diskcache_get_latency_count{no_fallback_on_error="false",no_lock="false",no_pos="false",no_sync="false",path="/Users/tanbiao/datakit/cache/custom_object"} 6
diskcache_get_latency_sum{no_fallback_on_error="false",no_lock="false",no_pos="false",no_sync="false",path="/Users/tanbiao/datakit/cache/keyevent"} 116
diskcache_get_latency_count{no_fallback_on_error="false",no_lock="false",no_pos="false",no_sync="false",path="/Users/tanbiao/datakit/cache/keyevent"} 6
diskcache_get_latency_sum{no_fallback_on_error="false",no_lock="false",no_pos="false",no_sync="false",path="/Users/tanbiao/datakit/cache/logging"} 171
diskcache_get_latency_count{no_fallback_on_error="false",no_lock="false",no_pos="false",no_sync="false",path="/Users/tanbiao/datakit/cache/logging"} 24
diskcache_get_latency_sum{no_fallback_on_error="false",no_lock="false",no_pos="false",no_sync="false",path="/Users/tanbiao/datakit/cache/metric"} 73
diskcache_get_latency_count{no_fallback_on_error="false",no_lock="false",no_pos="false",no_sync="false",path="/Users/tanbiao/datakit/cache/metric"} 24
diskcache_get_latency_sum{no_fallback_on_error="false",no_lock="false",no_pos="false",no_sync="false",path="/Users/tanbiao/datakit/cache/network"} 146
diskcache_get_latency_count{no_fallback_on_error="false",no_lock="false",no_pos="false",no_sync="false",path="/Users/tanbiao/datakit/cache/network"} 24
diskcache_get_latency_sum{no_fallback_on_error="false",no_lock="false",no_pos="false",no_sync="false",path="/Users/tanbiao/datakit/cache/object"} 109
diskcache_get_latency_count{no_fallback_on_error="false",no_lock="false",no_pos="false",no_sync="false",path="/Users/tanbiao/datakit/cache/object"} 6
diskcache_get_latency_sum{no_fallback_on_error="false",no_lock="false",no_pos="false",no_sync="false",path="/Users/tanbiao/datakit/cache/profiling"} 63
diskcache_get_latency_count{no_fallback_on_error="false",no_lock="false",no_pos="false",no_sync="false",path="/Users/tanbiao/datakit/cache/profiling"} 6
diskcache_get_latency_sum{no_fallback_on_error="false",no_lock="false",no_pos="false",no_sync="false",path="/Users/tanbiao/datakit/cache/rum"} 236
diskcache_get_latency_count{no_fallback_on_error="false",no_lock="false",no_pos="false",no_sync="false",path="/Users/tanbiao/datakit/cache/rum"} 24
diskcache_get_latency_sum{no_fallback_on_error="false",no_lock="false",no_pos="false",no_sync="false",path="/Users/tanbiao/datakit/cache/security"} 122
diskcache_get_latency_count{no_fallback_on_error="false",no_lock="false",no_pos="false",no_sync="false",path="/Users/tanbiao/datakit/cache/security"} 6
diskcache_get_latency_sum{no_fallback_on_error="false",no_lock="false",no_pos="false",no_sync="false",path="/Users/tanbiao/datakit/cache/tracing"} 127
diskcache_get_latency_count{no_fallback_on_error="false",no_lock="false",no_pos="false",no_sync="false",path="/Users/tanbiao/datakit/cache/tracing"} 24
# HELP diskcache_get_total cache Get() count
# TYPE diskcache_get_total counter
diskcache_get_total{no_fallback_on_error="false",no_lock="false",no_pos="false",no_sync="false",path="/Users/tanbiao/datakit/cache/custom_object"} 6
diskcache_get_total{no_fallback_on_error="false",no_lock="false",no_pos="false",no_sync="false",path="/Users/tanbiao/datakit/cache/keyevent"} 6
diskcache_get_total{no_fallback_on_error="false",no_lock="false",no_pos="false",no_sync="false",path="/Users/tanbiao/datakit/cache/logging"} 24
diskcache_get_total{no_fallback_on_error="false",no_lock="false",no_pos="false",no_sync="false",path="/Users/tanbiao/datakit/cache/metric"} 24
diskcache_get_total{no_fallback_on_error="false",no_lock="false",no_pos="false",no_sync="false",path="/Users/tanbiao/datakit/cache/network"} 24
diskcache_get_total{no_fallback_on_error="false",no_lock="false",no_pos="false",no_sync="false",path="/Users/tanbiao/datakit/cache/object"} 6
diskcache_get_total{no_fallback_on_error="false",no_lock="false",no_pos="false",no_sync="false",path="/Users/tanbiao/datakit/cache/profiling"} 6
diskcache_get_total{no_fallback_on_error="false",no_lock="false",no_pos="false",no_sync="false",path="/Users/tanbiao/datakit/cache/rum"} 24
diskcache_get_total{no_fallback_on_error="false",no_lock="false",no_pos="false",no_sync="false",path="/Users/tanbiao/datakit/cache/security"} 6
diskcache_get_total{no_fallback_on_error="false",no_lock="false",no_pos="false",no_sync="false",path="/Users/tanbiao/datakit/cache/tracing"} 24
# HELP diskcache_max_data max data to Put(in bytes), default 0
# TYPE diskcache_max_data gauge
diskcache_max_data{no_fallback_on_error="false",no_lock="false",no_pos="false",no_sync="false",path="/Users/tanbiao/datakit/cache/custom_object"} 0
diskcache_max_data{no_fallback_on_error="false",no_lock="false",no_pos="false",no_sync="false",path="/Users/tanbiao/datakit/cache/dynamic_dw"} 0
diskcache_max_data{no_fallback_on_error="false",no_lock="false",no_pos="false",no_sync="false",path="/Users/tanbiao/datakit/cache/keyevent"} 0
diskcache_max_data{no_fallback_on_error="false",no_lock="false",no_pos="false",no_sync="false",path="/Users/tanbiao/datakit/cache/logging"} 0
diskcache_max_data{no_fallback_on_error="false",no_lock="false",no_pos="false",no_sync="false",path="/Users/tanbiao/datakit/cache/metric"} 0
diskcache_max_data{no_fallback_on_error="false",no_lock="false",no_pos="false",no_sync="false",path="/Users/tanbiao/datakit/cache/network"} 0
diskcache_max_data{no_fallback_on_error="false",no_lock="false",no_pos="false",no_sync="false",path="/Users/tanbiao/datakit/cache/object"} 0
diskcache_max_data{no_fallback_on_error="false",no_lock="false",no_pos="false",no_sync="false",path="/Users/tanbiao/datakit/cache/profiling"} 0
diskcache_max_data{no_fallback_on_error="false",no_lock="false",no_pos="false",no_sync="false",path="/Users/tanbiao/datakit/cache/rum"} 0
diskcache_max_data{no_fallback_on_error="false",no_lock="false",no_pos="false",no_sync="false",path="/Users/tanbiao/datakit/cache/security"} 0
diskcache_max_data{no_fallback_on_error="false",no_lock="false",no_pos="false",no_sync="false",path="/Users/tanbiao/datakit/cache/tracing"} 0
# HELP diskcache_open_time current cache Open time in unix timestamp(second)
# TYPE diskcache_open_time gauge
diskcache_open_time{no_fallback_on_error="false",no_lock="false",no_pos="false",no_sync="false",path="/Users/tanbiao/datakit/cache/custom_object"} 1.684823227e+09
diskcache_open_time{no_fallback_on_error="false",no_lock="false",no_pos="false",no_sync="false",path="/Users/tanbiao/datakit/cache/dynamic_dw"} 1.684823227e+09
diskcache_open_time{no_fallback_on_error="false",no_lock="false",no_pos="false",no_sync="false",path="/Users/tanbiao/datakit/cache/keyevent"} 1.684823227e+09
diskcache_open_time{no_fallback_on_error="false",no_lock="false",no_pos="false",no_sync="false",path="/Users/tanbiao/datakit/cache/logging"} 1.684823227e+09
diskcache_open_time{no_fallback_on_error="false",no_lock="false",no_pos="false",no_sync="false",path="/Users/tanbiao/datakit/cache/metric"} 1.684823227e+09
diskcache_open_time{no_fallback_on_error="false",no_lock="false",no_pos="false",no_sync="false",path="/Users/tanbiao/datakit/cache/network"} 1.684823227e+09
diskcache_open_time{no_fallback_on_error="false",no_lock="false",no_pos="false",no_sync="false",path="/Users/tanbiao/datakit/cache/object"} 1.684823227e+09
diskcache_open_time{no_fallback_on_error="false",no_lock="false",no_pos="false",no_sync="false",path="/Users/tanbiao/datakit/cache/profiling"} 1.684823227e+09
diskcache_open_time{no_fallback_on_error="false",no_lock="false",no_pos="false",no_sync="false",path="/Users/tanbiao/datakit/cache/rum"} 1.684823227e+09
diskcache_open_time{no_fallback_on_error="false",no_lock="false",no_pos="false",no_sync="false",path="/Users/tanbiao/datakit/cache/security"} 1.684823227e+09
diskcache_open_time{no_fallback_on_error="false",no_lock="false",no_pos="false",no_sync="false",path="/Users/tanbiao/datakit/cache/tracing"} 1.684823227e+09
`)
