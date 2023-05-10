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
			w.Write(metrics)
		}))

		Start(WithHost(ts.Listener.Addr().String()), WithVerbose(true), WithMaxRun(1))
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

		Start(WithHost(ts.Listener.Addr().String()), WithVerbose(true), WithMaxRun(1))
		t.Cleanup(func() {
			ts.Close()
		})
	})
}

var metrics = []byte(`
# HELP datakit_cpu_cores Datakit CPU cores
# TYPE datakit_cpu_cores gauge
datakit_cpu_cores 10
# HELP datakit_cpu_usage Datakit CPU usage(%)
# TYPE datakit_cpu_usage gauge
datakit_cpu_usage 0.6977714719970705
# HELP datakit_data_overuse Does current workspace's data(metric/logging) usaguse(if 0 not beyond, or with a unix timestamp when overuse occurred)
# TYPE datakit_data_overuse gauge
datakit_data_overuse 0
# HELP datakit_dns_cost DNS IP lookup cost(ms)
# TYPE datakit_dns_cost summary
datakit_dns_cost_sum{domain="openway.guance.com",status="ok"} 610.278291
datakit_dns_cost_count{domain="openway.guance.com",status="ok"} 2
# HELP datakit_dns_domain_total DNS watched domain counter
# TYPE datakit_dns_domain_total counter
datakit_dns_domain_total 1
# HELP datakit_dns_ip_updated_total Domain IP updated counter
# TYPE datakit_dns_ip_updated_total counter
datakit_dns_ip_updated_total{domain="openway.guance.com"} 1
# HELP datakit_dns_watch_run_total watch run counter
# TYPE datakit_dns_watch_run_total counter
datakit_dns_watch_run_total{interval="1m0s"} 2
# HELP datakit_election Election latency(in millisecond)
# TYPE datakit_election summary
datakit_election_sum{namespace="default",status="success"} 1159
datakit_election_count{namespace="default",status="success"} 1
# HELP datakit_election_inputs Datakit election input count
# TYPE datakit_election_inputs gauge
datakit_election_inputs{namespace="default"} 0
# HELP datakit_election_pause_total Input paused count when election failed
# TYPE datakit_election_pause_total counter
datakit_election_pause_total{id="tanbiaos-MacBook-Pro.local",namespace="default"} 0
# HELP datakit_election_resume_total Input resume count when election OK
# TYPE datakit_election_resume_total counter
datakit_election_resume_total{id="tanbiaos-MacBook-Pro.local",namespace="default"} 0
# HELP datakit_election_status Datakit election status, if metric = 0, meas not elected, or the elected time(unix timestamp second)
# TYPE datakit_election_status gauge
datakit_election_status{elected_id="tanbiaos-MacBook-Pro.local",id="tanbiaos-MacBook-Pro.local",namespace="default",status="success"} 1.683606085e+09
# HELP datakit_filter_last_update filter last update time(in unix timestamp second)
# TYPE datakit_filter_last_update gauge
datakit_filter_last_update 1.683606081e+09
# HELP datakit_filter_latency Filter latency(us) of these filters
# TYPE datakit_filter_latency summary
datakit_filter_latency_sum{category="metric",filters="{ measurement =  'cpu'  and ( host notin [ 'ubt-dev-01' ] )}",source="remote"} 1515
datakit_filter_latency_count{category="metric",filters="{ measurement =  'cpu'  and ( host notin [ 'ubt-dev-01' ] )}",source="remote"} 41
# HELP datakit_filter_point_dropped_total Dropped points of filters
# TYPE datakit_filter_point_dropped_total counter
datakit_filter_point_dropped_total{category="metric",filters="{ measurement =  'cpu'  and ( host notin [ 'ubt-dev-01' ] )}",source="remote"} 6
# HELP datakit_filter_point_total Filter points of filters
# TYPE datakit_filter_point_total counter
datakit_filter_point_total{category="metric",filters="{ measurement =  'cpu'  and ( host notin [ 'ubt-dev-01' ] )}",source="remote"} 224
# HELP datakit_filter_pull_latency Filter pull(remote) latency(ms)
# TYPE datakit_filter_pull_latency summary
datakit_filter_pull_latency_sum{status="ok"} 3138
datakit_filter_pull_latency_count{status="ok"} 8
# HELP datakit_filter_update_total Filters(remote) updated count
# TYPE datakit_filter_update_total counter
datakit_filter_update_total 1
# HELP datakit_gc_summary Datakit golang GC paused(nano-second)
# TYPE datakit_gc_summary summary
datakit_gc_summary_sum 3.124167e+06
datakit_gc_summary_count 14
# HELP datakit_goroutine_alive alive goroutines
# TYPE datakit_goroutine_alive gauge
datakit_goroutine_alive{name="election"} 1
datakit_goroutine_alive{name="http"} 3
datakit_goroutine_alive{name="inputs"} 10
datakit_goroutine_alive{name="inputs_cpu"} 1
datakit_goroutine_alive{name="inputs_dialtesting"} 1
datakit_goroutine_alive{name="internal_cgroup"} 0
datakit_goroutine_alive{name="io"} 27
datakit_goroutine_alive{name="io_dnswatcher"} 1
datakit_goroutine_alive{name="pipeline_remote"} 1
# HELP datakit_goroutine_cost goroutine running time(in nanosecond)
# TYPE datakit_goroutine_cost summary
datakit_goroutine_cost_sum{name="internal_cgroup"} 73291
datakit_goroutine_cost_count{name="internal_cgroup"} 1
# HELP datakit_goroutine_groups goroutine group count
# TYPE datakit_goroutine_groups gauge
datakit_goroutine_groups 23
# HELP datakit_goroutine_stopped_total stopped goroutines
# TYPE datakit_goroutine_stopped_total counter
datakit_goroutine_stopped_total{name="internal_cgroup"} 1
# HELP datakit_goroutines goroutine count within Datakit
# TYPE datakit_goroutines gauge
datakit_goroutines 60
# HELP datakit_heap_alloc Datakit memory heap bytes
# TYPE datakit_heap_alloc gauge
datakit_heap_alloc 3.211148e+07
# HELP datakit_http_api_elapsed API request cost(in ms)
# TYPE datakit_http_api_elapsed summary
datakit_http_api_elapsed_sum{api="/metrics",method="GET",status="OK"} 27057
datakit_http_api_elapsed_count{api="/metrics",method="GET",status="OK"} 27
# HELP datakit_http_api_elapsed_histogram API request cost(in ms) histogram
# TYPE datakit_http_api_elapsed_histogram histogram
datakit_http_api_elapsed_histogram_bucket{api="/metrics",method="GET",status="OK",le="10"} 0
datakit_http_api_elapsed_histogram_bucket{api="/metrics",method="GET",status="OK",le="100"} 0
datakit_http_api_elapsed_histogram_bucket{api="/metrics",method="GET",status="OK",le="1000"} 0
datakit_http_api_elapsed_histogram_bucket{api="/metrics",method="GET",status="OK",le="5000"} 27
datakit_http_api_elapsed_histogram_bucket{api="/metrics",method="GET",status="OK",le="30000"} 27
datakit_http_api_elapsed_histogram_bucket{api="/metrics",method="GET",status="OK",le="+Inf"} 27
datakit_http_api_elapsed_histogram_sum{api="/metrics",method="GET",status="OK"} 27057
datakit_http_api_elapsed_histogram_count{api="/metrics",method="GET",status="OK"} 27
# HELP datakit_http_api_req_size API request body size
# TYPE datakit_http_api_req_size summary
datakit_http_api_req_size_sum{api="/metrics",method="GET",status="OK"} 2076
datakit_http_api_req_size_count{api="/metrics",method="GET",status="OK"} 27
# HELP datakit_http_api_total API request counter
# TYPE datakit_http_api_total counter
datakit_http_api_total{api="/metrics",method="GET",status="OK"} 27
# HELP datakit_input_collect_latency Input collect latency(us)
# TYPE datakit_input_collect_latency summary
datakit_input_collect_latency_sum{category="logging",name="dialtesting"} 0
datakit_input_collect_latency_count{category="logging",name="dialtesting"} 7
datakit_input_collect_latency_sum{category="metric",name="cpu"} 321
datakit_input_collect_latency_count{category="metric",name="cpu"} 6
datakit_input_collect_latency_sum{category="metric",name="disk"} 12245
datakit_input_collect_latency_count{category="metric",name="disk"} 8
datakit_input_collect_latency_sum{category="metric",name="diskio"} 3844
datakit_input_collect_latency_count{category="metric",name="diskio"} 7
datakit_input_collect_latency_sum{category="metric",name="mem"} 419
datakit_input_collect_latency_count{category="metric",name="mem"} 7
datakit_input_collect_latency_sum{category="metric",name="net"} 35434
datakit_input_collect_latency_count{category="metric",name="net"} 6
datakit_input_collect_latency_sum{category="metric",name="swap"} 223
datakit_input_collect_latency_count{category="metric",name="swap"} 7
datakit_input_collect_latency_sum{category="object",name="host_processes/object"} 3.2720667e+07
datakit_input_collect_latency_count{category="object",name="host_processes/object"} 3
datakit_input_collect_latency_sum{category="object",name="hostobject"} 3.539677e+06
datakit_input_collect_latency_count{category="object",name="hostobject"} 1
# HELP datakit_inputs_instance input instance count
# TYPE datakit_inputs_instance gauge
datakit_inputs_instance{input="cpu"} 1
datakit_inputs_instance{input="ddtrace"} 1
datakit_inputs_instance{input="dialtesting"} 1
datakit_inputs_instance{input="disk"} 1
datakit_inputs_instance{input="diskio"} 1
datakit_inputs_instance{input="host_processes"} 1
datakit_inputs_instance{input="hostobject"} 1
datakit_inputs_instance{input="mem"} 1
datakit_inputs_instance{input="net"} 1
datakit_inputs_instance{input="swap"} 1
# HELP datakit_io_chan_capacity IO channel capacity
# TYPE datakit_io_chan_capacity gauge
datakit_io_chan_capacity{category="all-the-same"} 100
# HELP datakit_io_chan_usage IO channel usage(length of the chan)
# TYPE datakit_io_chan_usage gauge
datakit_io_chan_usage{category="dynamic_dw"} 0
datakit_io_chan_usage{category="metric"} 0
datakit_io_chan_usage{category="object"} 0
# HELP datakit_io_dataway_api_latency dataway HTTP request latency(ms) partitioned by HTTP API(method@url) and HTTP status
# TYPE datakit_io_dataway_api_latency summary
datakit_io_dataway_api_latency_sum{api="/v1/datakit/pull",status="OK"} 5110
datakit_io_dataway_api_latency_count{api="/v1/datakit/pull",status="OK"} 10
datakit_io_dataway_api_latency_sum{api="/v1/election",status="OK"} 1159
datakit_io_dataway_api_latency_count{api="/v1/election",status="OK"} 1
datakit_io_dataway_api_latency_sum{api="/v1/election/heartbeat",status="OK"} 3126
datakit_io_dataway_api_latency_count{api="/v1/election/heartbeat",status="OK"} 23
datakit_io_dataway_api_latency_sum{api="/v1/write/logging",status="OK"} 627
datakit_io_dataway_api_latency_count{api="/v1/write/logging",status="OK"} 7
datakit_io_dataway_api_latency_sum{api="/v1/write/metric",status="OK"} 1627
datakit_io_dataway_api_latency_count{api="/v1/write/metric",status="OK"} 8
datakit_io_dataway_api_latency_sum{api="/v1/write/object",status="OK"} 1311
datakit_io_dataway_api_latency_count{api="/v1/write/object",status="OK"} 2
# HELP datakit_io_dataway_api_request_total dataway HTTP request processed, partitioned by status code and HTTP API(url path)
# TYPE datakit_io_dataway_api_request_total counter
datakit_io_dataway_api_request_total{api="/v1/datakit/pull",status="OK"} 10
datakit_io_dataway_api_request_total{api="/v1/election",status="OK"} 1
datakit_io_dataway_api_request_total{api="/v1/election/heartbeat",status="OK"} 23
datakit_io_dataway_api_request_total{api="/v1/write/logging",status="OK"} 7
datakit_io_dataway_api_request_total{api="/v1/write/metric",status="OK"} 8
datakit_io_dataway_api_request_total{api="/v1/write/object",status="OK"} 2
# HELP datakit_io_dataway_not_sink_point_total dataway not-sinked points(condition or category not match)
# TYPE datakit_io_dataway_not_sink_point_total counter
datakit_io_dataway_not_sink_point_total{category="metric"} 174
datakit_io_dataway_not_sink_point_total{category="object"} 723
datakit_io_dataway_not_sink_point_total{category="unknown"} 7
# HELP datakit_io_dataway_point_bytes_total dataway uploaded points bytes, partitioned by category and pint send status(HTTP status)
# TYPE datakit_io_dataway_point_bytes_total counter
datakit_io_dataway_point_bytes_total{category="dynamic_dw",enc="gzip",status="OK"} 2719
datakit_io_dataway_point_bytes_total{category="dynamic_dw",enc="gzip",status="total"} 2719
datakit_io_dataway_point_bytes_total{category="dynamic_dw",enc="raw",status="OK"} 3874
datakit_io_dataway_point_bytes_total{category="dynamic_dw",enc="raw",status="total"} 3874
datakit_io_dataway_point_bytes_total{category="metric",enc="gzip",status="OK"} 6345
datakit_io_dataway_point_bytes_total{category="metric",enc="gzip",status="total"} 6345
datakit_io_dataway_point_bytes_total{category="metric",enc="raw",status="OK"} 46331
datakit_io_dataway_point_bytes_total{category="metric",enc="raw",status="total"} 46331
datakit_io_dataway_point_bytes_total{category="object",enc="gzip",status="OK"} 98251
datakit_io_dataway_point_bytes_total{category="object",enc="gzip",status="total"} 98251
datakit_io_dataway_point_bytes_total{category="object",enc="raw",status="OK"} 817070
datakit_io_dataway_point_bytes_total{category="object",enc="raw",status="total"} 817070
# HELP datakit_io_dataway_point_total dataway uploaded points, partitioned by category and send status(HTTP status)
# TYPE datakit_io_dataway_point_total counter
datakit_io_dataway_point_total{category="dynamic_dw",status="OK"} 7
datakit_io_dataway_point_total{category="dynamic_dw",status="total"} 7
datakit_io_dataway_point_total{category="metric",status="OK"} 174
datakit_io_dataway_point_total{category="metric",status="total"} 174
datakit_io_dataway_point_total{category="object",status="OK"} 723
datakit_io_dataway_point_total{category="object",status="total"} 723
# HELP datakit_io_dataway_sink_total dataway sink count, partitioned by category.
# TYPE datakit_io_dataway_sink_total counter
datakit_io_dataway_sink_total{category="metric"} 8
datakit_io_dataway_sink_total{category="object"} 2
datakit_io_dataway_sink_total{category="unknown"} 7
# HELP datakit_io_feed_point_total Input feed point total
# TYPE datakit_io_feed_point_total counter
datakit_io_feed_point_total{category="logging",name="dialtesting"} 7
datakit_io_feed_point_total{category="metric",name="cpu"} 6
datakit_io_feed_point_total{category="metric",name="disk"} 88
datakit_io_feed_point_total{category="metric",name="diskio"} 14
datakit_io_feed_point_total{category="metric",name="mem"} 7
datakit_io_feed_point_total{category="metric",name="net"} 102
datakit_io_feed_point_total{category="metric",name="swap"} 7
datakit_io_feed_point_total{category="object",name="host_processes/object"} 1084
datakit_io_feed_point_total{category="object",name="hostobject"} 1
# HELP datakit_io_feed_total Input feed total
# TYPE datakit_io_feed_total counter
datakit_io_feed_total{category="logging",name="dialtesting"} 7
datakit_io_feed_total{category="metric",name="cpu"} 6
datakit_io_feed_total{category="metric",name="disk"} 8
datakit_io_feed_total{category="metric",name="diskio"} 7
datakit_io_feed_total{category="metric",name="mem"} 7
datakit_io_feed_total{category="metric",name="net"} 6
datakit_io_feed_total{category="metric",name="swap"} 7
datakit_io_feed_total{category="object",name="host_processes/object"} 3
datakit_io_feed_total{category="object",name="hostobject"} 1
# HELP datakit_io_flush_total IO flush total
# TYPE datakit_io_flush_total counter
datakit_io_flush_total{category="logging"} 7
datakit_io_flush_total{category="metric"} 8
datakit_io_flush_total{category="object"} 2
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
# HELP datakit_io_http_conn_idle_time Dataway HTTP connection idle time(ms)
# TYPE datakit_io_http_conn_idle_time summary
datakit_io_http_conn_idle_time_sum 62773.994417000016
datakit_io_http_conn_idle_time_count 51
# HELP datakit_io_http_conn_reused_from_idle_total Dataway HTTP connection reused from idle count
# TYPE datakit_io_http_conn_reused_from_idle_total counter
datakit_io_http_conn_reused_from_idle_total 39
# HELP datakit_io_http_connect_cost Dataway HTTP connect cost(ms)
# TYPE datakit_io_http_connect_cost summary
datakit_io_http_connect_cost_sum 1169.848958
datakit_io_http_connect_cost_count 51
# HELP datakit_io_http_dns_cost Dataway HTTP DNS cost(ms)
# TYPE datakit_io_http_dns_cost summary
datakit_io_http_dns_cost_sum 1162.241541
datakit_io_http_dns_cost_count 51
# HELP datakit_io_http_got_first_resp_byte_cost Dataway got first response byte cost(ms)
# TYPE datakit_io_http_got_first_resp_byte_cost summary
datakit_io_http_got_first_resp_byte_cost{quantile="0.5"} 74.92675
datakit_io_http_got_first_resp_byte_cost{quantile="0.75"} 149.235791
datakit_io_http_got_first_resp_byte_cost{quantile="0.95"} 570.713042
datakit_io_http_got_first_resp_byte_cost_sum 7630.258953999998
datakit_io_http_got_first_resp_byte_cost_count 51
# HELP datakit_io_http_tcp_conn_total Dataway HTTP TCP connection count
# TYPE datakit_io_http_tcp_conn_total counter
datakit_io_http_tcp_conn_total{remote="118.31.126.76:443",type="created"} 5
datakit_io_http_tcp_conn_total{remote="47.110.144.10:443",type="created"} 2
datakit_io_http_tcp_conn_total{remote="47.110.144.10:443",type="reused"} 33
datakit_io_http_tcp_conn_total{remote="47.114.74.166:443",type="created"} 2
datakit_io_http_tcp_conn_total{remote="47.114.74.166:443",type="reused"} 9
# HELP datakit_io_http_tls_handshake Dataway TLS handshake cost(ms)
# TYPE datakit_io_http_tls_handshake summary
datakit_io_http_tls_handshake_sum 3106.7237509999995
datakit_io_http_tls_handshake_count 51
# HELP datakit_io_input_filter_point_total Input filtered point total
# TYPE datakit_io_input_filter_point_total counter
datakit_io_input_filter_point_total{category="logging",name="dialtesting"} 0
datakit_io_input_filter_point_total{category="metric",name="cpu"} 6
datakit_io_input_filter_point_total{category="metric",name="disk"} 0
datakit_io_input_filter_point_total{category="metric",name="diskio"} 0
datakit_io_input_filter_point_total{category="metric",name="mem"} 0
datakit_io_input_filter_point_total{category="metric",name="net"} 0
datakit_io_input_filter_point_total{category="metric",name="swap"} 0
datakit_io_input_filter_point_total{category="object",name="host_processes/object"} 0
datakit_io_input_filter_point_total{category="object",name="hostobject"} 0
# HELP datakit_io_last_feed Input last feed time(unix timestamp in second)
# TYPE datakit_io_last_feed gauge
datakit_io_last_feed{category="logging",name="dialtesting"} 1.683606149e+09
datakit_io_last_feed{category="metric",name="cpu"} 1.683606146e+09
datakit_io_last_feed{category="metric",name="disk"} 1.683606152e+09
datakit_io_last_feed{category="metric",name="diskio"} 1.683606146e+09
datakit_io_last_feed{category="metric",name="mem"} 1.683606147e+09
datakit_io_last_feed{category="metric",name="net"} 1.683606148e+09
datakit_io_last_feed{category="metric",name="swap"} 1.683606151e+09
datakit_io_last_feed{category="object",name="host_processes/object"} 1.683606153e+09
datakit_io_last_feed{category="object",name="hostobject"} 1.683606086e+09
# HELP datakit_io_queue_pts IO module queued(cached) points
# TYPE datakit_io_queue_pts gauge
datakit_io_queue_pts{category="logging"} 0
datakit_io_queue_pts{category="metric"} 11
datakit_io_queue_pts{category="object"} 362
# HELP datakit_open_files Datakit open files(only available on Linux)
# TYPE datakit_open_files gauge
datakit_open_files -1
# HELP datakit_pipeline_cost Pipeline total running time(ms)
# TYPE datakit_pipeline_cost summary
datakit_pipeline_cost_sum{category="metric",name="cpu.p",namespace="default"} 0.035082
datakit_pipeline_cost_count{category="metric",name="cpu.p",namespace="default"} 6
datakit_pipeline_cost_sum{category="metric",name="disk.p",namespace="default"} 0.13529099999999997
datakit_pipeline_cost_count{category="metric",name="disk.p",namespace="default"} 88
datakit_pipeline_cost_sum{category="object",name="host_processes.p",namespace="default"} 0.15977199999999914
datakit_pipeline_cost_count{category="object",name="host_processes.p",namespace="default"} 1084
# HELP datakit_pipeline_point_total Pipeline processed total points
# TYPE datakit_pipeline_point_total counter
datakit_pipeline_point_total{category="metric",name="cpu.p",namespace="default"} 6
datakit_pipeline_point_total{category="metric",name="disk.p",namespace="default"} 88
datakit_pipeline_point_total{category="object",name="host_processes.p",namespace="default"} 1084
# HELP datakit_pipeline_update_time Pipeline last update time(unix timestamp)
# TYPE datakit_pipeline_update_time gauge
datakit_pipeline_update_time{category="logging",name="apache.p",namespace="default"} 1.68360608e+09
datakit_pipeline_update_time{category="logging",name="consul.p",namespace="default"} 1.68360608e+09
datakit_pipeline_update_time{category="logging",name="elasticsearch.p",namespace="default"} 1.68360608e+09
datakit_pipeline_update_time{category="logging",name="jenkins.p",namespace="default"} 1.68360608e+09
datakit_pipeline_update_time{category="logging",name="kafka.p",namespace="default"} 1.68360608e+09
datakit_pipeline_update_time{category="logging",name="mongodb.p",namespace="default"} 1.68360608e+09
datakit_pipeline_update_time{category="logging",name="mysql.p",namespace="default"} 1.68360608e+09
datakit_pipeline_update_time{category="logging",name="nginx.p",namespace="default"} 1.68360608e+09
datakit_pipeline_update_time{category="logging",name="postgresql.p",namespace="default"} 1.68360608e+09
datakit_pipeline_update_time{category="logging",name="rabbitmq.p",namespace="default"} 1.68360608e+09
datakit_pipeline_update_time{category="logging",name="redis.p",namespace="default"} 1.68360608e+09
datakit_pipeline_update_time{category="logging",name="solr.p",namespace="default"} 1.68360608e+09
datakit_pipeline_update_time{category="logging",name="sqlserver.p",namespace="default"} 1.68360608e+09
datakit_pipeline_update_time{category="logging",name="tdengine.p",namespace="default"} 1.68360608e+09
datakit_pipeline_update_time{category="logging",name="tomcat.p",namespace="default"} 1.68360608e+09
datakit_pipeline_update_time{category="metric",name="cpu.p",namespace="default"} 1.68360608e+09
datakit_pipeline_update_time{category="metric",name="disk.p",namespace="default"} 1.68360608e+09
datakit_pipeline_update_time{category="object",name="host_processes.p",namespace="default"} 1.68360608e+09
# HELP datakit_sys_alloc Datakit memory system bytes
# TYPE datakit_sys_alloc gauge
datakit_sys_alloc 6.47506e+07
# HELP datakit_uptime Datakit uptime(second)
# TYPE datakit_uptime gauge
datakit_uptime{auto_update="false",branch="fix-snat",build_at="2023-05-09 04:20:34",cgroup="-",docker="false",hostname="tanbiaos-MacBook-Pro.local",os_arch="darwin/arm64",version="1.6.0-389-ga366f8d2f9"} 78
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
diskcache_get_latency_sum{no_fallback_on_error="false",no_lock="false",no_pos="false",no_sync="false",path="/Users/tanbiao/datakit/cache/custom_object"} 40
diskcache_get_latency_count{no_fallback_on_error="false",no_lock="false",no_pos="false",no_sync="false",path="/Users/tanbiao/datakit/cache/custom_object"} 2
diskcache_get_latency_sum{no_fallback_on_error="false",no_lock="false",no_pos="false",no_sync="false",path="/Users/tanbiao/datakit/cache/keyevent"} 6
diskcache_get_latency_count{no_fallback_on_error="false",no_lock="false",no_pos="false",no_sync="false",path="/Users/tanbiao/datakit/cache/keyevent"} 2
diskcache_get_latency_sum{no_fallback_on_error="false",no_lock="false",no_pos="false",no_sync="false",path="/Users/tanbiao/datakit/cache/logging"} 136
diskcache_get_latency_count{no_fallback_on_error="false",no_lock="false",no_pos="false",no_sync="false",path="/Users/tanbiao/datakit/cache/logging"} 8
diskcache_get_latency_sum{no_fallback_on_error="false",no_lock="false",no_pos="false",no_sync="false",path="/Users/tanbiao/datakit/cache/metric"} 23
diskcache_get_latency_count{no_fallback_on_error="false",no_lock="false",no_pos="false",no_sync="false",path="/Users/tanbiao/datakit/cache/metric"} 8
diskcache_get_latency_sum{no_fallback_on_error="false",no_lock="false",no_pos="false",no_sync="false",path="/Users/tanbiao/datakit/cache/network"} 46
diskcache_get_latency_count{no_fallback_on_error="false",no_lock="false",no_pos="false",no_sync="false",path="/Users/tanbiao/datakit/cache/network"} 8
diskcache_get_latency_sum{no_fallback_on_error="false",no_lock="false",no_pos="false",no_sync="false",path="/Users/tanbiao/datakit/cache/object"} 9
diskcache_get_latency_count{no_fallback_on_error="false",no_lock="false",no_pos="false",no_sync="false",path="/Users/tanbiao/datakit/cache/object"} 2
diskcache_get_latency_sum{no_fallback_on_error="false",no_lock="false",no_pos="false",no_sync="false",path="/Users/tanbiao/datakit/cache/profiling"} 32
diskcache_get_latency_count{no_fallback_on_error="false",no_lock="false",no_pos="false",no_sync="false",path="/Users/tanbiao/datakit/cache/profiling"} 2
diskcache_get_latency_sum{no_fallback_on_error="false",no_lock="false",no_pos="false",no_sync="false",path="/Users/tanbiao/datakit/cache/rum"} 30
diskcache_get_latency_count{no_fallback_on_error="false",no_lock="false",no_pos="false",no_sync="false",path="/Users/tanbiao/datakit/cache/rum"} 8
diskcache_get_latency_sum{no_fallback_on_error="false",no_lock="false",no_pos="false",no_sync="false",path="/Users/tanbiao/datakit/cache/security"} 121
diskcache_get_latency_count{no_fallback_on_error="false",no_lock="false",no_pos="false",no_sync="false",path="/Users/tanbiao/datakit/cache/security"} 2
diskcache_get_latency_sum{no_fallback_on_error="false",no_lock="false",no_pos="false",no_sync="false",path="/Users/tanbiao/datakit/cache/tracing"} 121
diskcache_get_latency_count{no_fallback_on_error="false",no_lock="false",no_pos="false",no_sync="false",path="/Users/tanbiao/datakit/cache/tracing"} 8
# HELP diskcache_get_total cache Get() count
# TYPE diskcache_get_total counter
diskcache_get_total{no_fallback_on_error="false",no_lock="false",no_pos="false",no_sync="false",path="/Users/tanbiao/datakit/cache/custom_object"} 2
diskcache_get_total{no_fallback_on_error="false",no_lock="false",no_pos="false",no_sync="false",path="/Users/tanbiao/datakit/cache/keyevent"} 2
diskcache_get_total{no_fallback_on_error="false",no_lock="false",no_pos="false",no_sync="false",path="/Users/tanbiao/datakit/cache/logging"} 8
diskcache_get_total{no_fallback_on_error="false",no_lock="false",no_pos="false",no_sync="false",path="/Users/tanbiao/datakit/cache/metric"} 8
diskcache_get_total{no_fallback_on_error="false",no_lock="false",no_pos="false",no_sync="false",path="/Users/tanbiao/datakit/cache/network"} 8
diskcache_get_total{no_fallback_on_error="false",no_lock="false",no_pos="false",no_sync="false",path="/Users/tanbiao/datakit/cache/object"} 2
diskcache_get_total{no_fallback_on_error="false",no_lock="false",no_pos="false",no_sync="false",path="/Users/tanbiao/datakit/cache/profiling"} 2
diskcache_get_total{no_fallback_on_error="false",no_lock="false",no_pos="false",no_sync="false",path="/Users/tanbiao/datakit/cache/rum"} 8
diskcache_get_total{no_fallback_on_error="false",no_lock="false",no_pos="false",no_sync="false",path="/Users/tanbiao/datakit/cache/security"} 2
diskcache_get_total{no_fallback_on_error="false",no_lock="false",no_pos="false",no_sync="false",path="/Users/tanbiao/datakit/cache/tracing"} 8
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
diskcache_open_time{no_fallback_on_error="false",no_lock="false",no_pos="false",no_sync="false",path="/Users/tanbiao/datakit/cache/custom_object"} 1.68360608e+09
diskcache_open_time{no_fallback_on_error="false",no_lock="false",no_pos="false",no_sync="false",path="/Users/tanbiao/datakit/cache/dynamic_dw"} 1.68360608e+09
diskcache_open_time{no_fallback_on_error="false",no_lock="false",no_pos="false",no_sync="false",path="/Users/tanbiao/datakit/cache/keyevent"} 1.68360608e+09
diskcache_open_time{no_fallback_on_error="false",no_lock="false",no_pos="false",no_sync="false",path="/Users/tanbiao/datakit/cache/logging"} 1.68360608e+09
diskcache_open_time{no_fallback_on_error="false",no_lock="false",no_pos="false",no_sync="false",path="/Users/tanbiao/datakit/cache/metric"} 1.68360608e+09
diskcache_open_time{no_fallback_on_error="false",no_lock="false",no_pos="false",no_sync="false",path="/Users/tanbiao/datakit/cache/network"} 1.68360608e+09
diskcache_open_time{no_fallback_on_error="false",no_lock="false",no_pos="false",no_sync="false",path="/Users/tanbiao/datakit/cache/object"} 1.68360608e+09
diskcache_open_time{no_fallback_on_error="false",no_lock="false",no_pos="false",no_sync="false",path="/Users/tanbiao/datakit/cache/profiling"} 1.68360608e+09
diskcache_open_time{no_fallback_on_error="false",no_lock="false",no_pos="false",no_sync="false",path="/Users/tanbiao/datakit/cache/rum"} 1.68360608e+09
diskcache_open_time{no_fallback_on_error="false",no_lock="false",no_pos="false",no_sync="false",path="/Users/tanbiao/datakit/cache/security"} 1.68360608e+09
diskcache_open_time{no_fallback_on_error="false",no_lock="false",no_pos="false",no_sync="false",path="/Users/tanbiao/datakit/cache/tracing"} 1.68360608e+09
`)
