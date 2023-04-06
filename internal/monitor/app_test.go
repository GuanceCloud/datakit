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
datakit_cpu_usage 100
# HELP datakit_data_overuse Does current workspace's data(metric/logging) usaguse(if 0 not beyond, or with a unix timestamp when overuse occurred)
# TYPE datakit_data_overuse gauge
datakit_data_overuse 0
# HELP datakit_dns_cost DNS IP lookup cost(ms)
# TYPE datakit_dns_cost summary
datakit_dns_cost_sum{domain="openway.guance.com",status="ok"} 1628.0006679999995
datakit_dns_cost_count{domain="openway.guance.com",status="ok"} 23
# HELP datakit_dns_domain_total DNS watched domain counter
# TYPE datakit_dns_domain_total counter
datakit_dns_domain_total 1
# HELP datakit_dns_ip_updated_total Domain IP updated counter
# TYPE datakit_dns_ip_updated_total counter
datakit_dns_ip_updated_total{domain="openway.guance.com"} 1
# HELP datakit_dns_watch_run_total watch run counter
# TYPE datakit_dns_watch_run_total counter
datakit_dns_watch_run_total{interval="1m0s"} 23
# HELP datakit_election Election latency(in millisecond)
# TYPE datakit_election summary
datakit_election_sum{namespace="default",status="success"} 121
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
datakit_election_status{elected_id="tanbiaos-MacBook-Pro.local",id="tanbiaos-MacBook-Pro.local",namespace="default",status="success"} 1.680605514e+09
# HELP datakit_filter_last_update filter last update time(in unix timestamp second)
# TYPE datakit_filter_last_update gauge
datakit_filter_last_update 1.680605511e+09
# HELP datakit_filter_latency Filter latency(us) of these filters
# TYPE datakit_filter_latency summary
datakit_filter_latency_sum{category="metric",filters="{ measurement =  'cpu'  and ( host notin [ 'ubt-dev-01' ] )}",source="remote"} 66097
datakit_filter_latency_count{category="metric",filters="{ measurement =  'cpu'  and ( host notin [ 'ubt-dev-01' ] )}",source="remote"} 952
# HELP datakit_filter_point_dropped_total Dropped points of filters
# TYPE datakit_filter_point_dropped_total counter
datakit_filter_point_dropped_total{category="metric",filters="{ measurement =  'cpu'  and ( host notin [ 'ubt-dev-01' ] )}",source="remote"} 136
# HELP datakit_filter_point_total Filter points of filters
# TYPE datakit_filter_point_total counter
datakit_filter_point_total{category="metric",filters="{ measurement =  'cpu'  and ( host notin [ 'ubt-dev-01' ] )}",source="remote"} 4769
# HELP datakit_filter_pull_latency Filter pull(remote) latency(ms)
# TYPE datakit_filter_pull_latency summary
datakit_filter_pull_latency_sum{status="ok"} 17752
datakit_filter_pull_latency_count{status="ok"} 137
# HELP datakit_filter_update_total Filters(remote) updated count
# TYPE datakit_filter_update_total counter
datakit_filter_update_total 1
# HELP datakit_gc_summary Datakit golang GC paused(nano-second)
# TYPE datakit_gc_summary summary
datakit_gc_summary_sum 1.32739182e+08
datakit_gc_summary_count 471
# HELP datakit_goroutine_alive alive goroutines
# TYPE datakit_goroutine_alive gauge
datakit_goroutine_alive{name="election"} 1
datakit_goroutine_alive{name="http"} 3
datakit_goroutine_alive{name="inputs"} 10
datakit_goroutine_alive{name="inputs_cpu"} 1
datakit_goroutine_alive{name="inputs_dialtesting"} 0
datakit_goroutine_alive{name="internal_cgroup"} 0
datakit_goroutine_alive{name="io"} 27
datakit_goroutine_alive{name="io_dnswatcher"} 1
datakit_goroutine_alive{name="pipeline_remote"} 1
# HELP datakit_goroutine_cost goroutine running time(in nanosecond)
# TYPE datakit_goroutine_cost summary
datakit_goroutine_cost_sum{name="inputs_dialtesting"} 1.70000524583e+11
datakit_goroutine_cost_count{name="inputs_dialtesting"} 1
datakit_goroutine_cost_sum{name="internal_cgroup"} 115375
datakit_goroutine_cost_count{name="internal_cgroup"} 1
# HELP datakit_goroutine_groups goroutine group count
# TYPE datakit_goroutine_groups gauge
datakit_goroutine_groups 23
# HELP datakit_goroutine_stopped_total stopped goroutines
# TYPE datakit_goroutine_stopped_total counter
datakit_goroutine_stopped_total{name="inputs_dialtesting"} 1
datakit_goroutine_stopped_total{name="internal_cgroup"} 1
# HELP datakit_goroutines goroutine count within Datakit
# TYPE datakit_goroutines gauge
datakit_goroutines 68
# HELP datakit_heap_alloc Datakit memory heap bytes
# TYPE datakit_heap_alloc gauge
datakit_heap_alloc 2.4783536e+07
# HELP datakit_http_api_elapsed API request cost(in ms)
# TYPE datakit_http_api_elapsed summary
datakit_http_api_elapsed_sum{api="/metrics",method="GET",status="OK"} 2242
datakit_http_api_elapsed_count{api="/metrics",method="GET",status="OK"} 296
# HELP datakit_http_api_elapsed_histogram API request cost(in ms) histogram
# TYPE datakit_http_api_elapsed_histogram histogram
datakit_http_api_elapsed_histogram_bucket{api="/metrics",method="GET",status="OK",le="10"} 290
datakit_http_api_elapsed_histogram_bucket{api="/metrics",method="GET",status="OK",le="100"} 296
datakit_http_api_elapsed_histogram_bucket{api="/metrics",method="GET",status="OK",le="1000"} 296
datakit_http_api_elapsed_histogram_bucket{api="/metrics",method="GET",status="OK",le="5000"} 296
datakit_http_api_elapsed_histogram_bucket{api="/metrics",method="GET",status="OK",le="30000"} 296
datakit_http_api_elapsed_histogram_bucket{api="/metrics",method="GET",status="OK",le="+Inf"} 296
datakit_http_api_elapsed_histogram_sum{api="/metrics",method="GET",status="OK"} 2237
datakit_http_api_elapsed_histogram_count{api="/metrics",method="GET",status="OK"} 296
# HELP datakit_http_api_req_size API request body size
# TYPE datakit_http_api_req_size summary
datakit_http_api_req_size_sum{api="/metrics",method="GET",status="OK"} 18648
datakit_http_api_req_size_count{api="/metrics",method="GET",status="OK"} 296
# HELP datakit_http_api_total API request counter
# TYPE datakit_http_api_total counter
datakit_http_api_total{api="/metrics",method="GET",status="OK"} 296
# HELP datakit_input_collect_latency Input collect latency(us)
# TYPE datakit_input_collect_latency summary
datakit_input_collect_latency_sum{category="logging",name="dialtesting"} 0
datakit_input_collect_latency_count{category="logging",name="dialtesting"} 17
datakit_input_collect_latency_sum{category="metric",name="cpu"} 16320
datakit_input_collect_latency_count{category="metric",name="cpu"} 136
datakit_input_collect_latency_sum{category="metric",name="disk"} 201757
datakit_input_collect_latency_count{category="metric",name="disk"} 137
datakit_input_collect_latency_sum{category="metric",name="diskio"} 134012
datakit_input_collect_latency_count{category="metric",name="diskio"} 136
datakit_input_collect_latency_sum{category="metric",name="mem"} 18935
datakit_input_collect_latency_count{category="metric",name="mem"} 136
datakit_input_collect_latency_sum{category="metric",name="net"} 1.023833e+06
datakit_input_collect_latency_count{category="metric",name="net"} 136
datakit_input_collect_latency_sum{category="metric",name="self"} 1.06419126e+08
datakit_input_collect_latency_count{category="metric",name="self"} 136
datakit_input_collect_latency_sum{category="metric",name="swap"} 17784
datakit_input_collect_latency_count{category="metric",name="swap"} 135
datakit_input_collect_latency_sum{category="object",name="host_processes/object"} 5.64992131e+08
datakit_input_collect_latency_count{category="object",name="host_processes/object"} 45
# HELP datakit_inputs_instance input instance count
# TYPE datakit_inputs_instance gauge
datakit_inputs_instance{input="cpu"} 1
datakit_inputs_instance{input="ddtrace"} 1
datakit_inputs_instance{input="dialtesting"} 1
datakit_inputs_instance{input="disk"} 1
datakit_inputs_instance{input="diskio"} 1
datakit_inputs_instance{input="host_processes"} 1
datakit_inputs_instance{input="mem"} 1
datakit_inputs_instance{input="net"} 1
datakit_inputs_instance{input="self"} 1
datakit_inputs_instance{input="swap"} 1
# HELP datakit_io_chan_capacity IO channel capacity
# TYPE datakit_io_chan_capacity gauge
datakit_io_chan_capacity{category="all-the-same"} 100
# HELP datakit_io_chan_usage IO channel usage(length of the chan)
# TYPE datakit_io_chan_usage gauge
datakit_io_chan_usage{category="logging"} 0
datakit_io_chan_usage{category="metric"} 0
datakit_io_chan_usage{category="object"} 0
# HELP datakit_io_dataway_api_latency dataway HTTP request latency(ms) partitioned by HTTP API(method@url) and HTTP status
# TYPE datakit_io_dataway_api_latency summary
datakit_io_dataway_api_latency_sum{api="/v1/datakit/pull",status="OK"} 22005
datakit_io_dataway_api_latency_count{api="/v1/datakit/pull",status="OK"} 160
datakit_io_dataway_api_latency_sum{api="/v1/election",status="OK"} 120
datakit_io_dataway_api_latency_count{api="/v1/election",status="OK"} 1
datakit_io_dataway_api_latency_sum{api="/v1/election/heartbeat",status="OK"} 68104
datakit_io_dataway_api_latency_count{api="/v1/election/heartbeat",status="OK"} 453
datakit_io_dataway_api_latency_sum{api="/v1/write/logging",status="OK"} 1700
datakit_io_dataway_api_latency_count{api="/v1/write/logging",status="OK"} 17
datakit_io_dataway_api_latency_sum{api="/v1/write/metric",status="OK"} 27892
datakit_io_dataway_api_latency_count{api="/v1/write/metric",status="OK"} 180
datakit_io_dataway_api_latency_sum{api="/v1/write/object",status="OK"} 11841
datakit_io_dataway_api_latency_count{api="/v1/write/object",status="OK"} 45
# HELP datakit_io_dataway_api_request_total dataway HTTP request processed, partitioned by status code and HTTP API(url path)
# TYPE datakit_io_dataway_api_request_total counter
datakit_io_dataway_api_request_total{api="/v1/datakit/pull",status="OK"} 160
datakit_io_dataway_api_request_total{api="/v1/election",status="OK"} 1
datakit_io_dataway_api_request_total{api="/v1/election/heartbeat",status="OK"} 453
datakit_io_dataway_api_request_total{api="/v1/write/logging",status="OK"} 17
datakit_io_dataway_api_request_total{api="/v1/write/metric",status="OK"} 180
datakit_io_dataway_api_request_total{api="/v1/write/object",status="OK"} 45
# HELP datakit_io_dataway_point_bytes_total dataway uploaded points bytes, partitioned by category and pint send status(HTTP status)
# TYPE datakit_io_dataway_point_bytes_total counter
datakit_io_dataway_point_bytes_total{category="dynamic_dw",status="OK"} 6600
datakit_io_dataway_point_bytes_total{category="metric",status="OK"} 222549
datakit_io_dataway_point_bytes_total{category="object",status="OK"} 2.06212e+06
# HELP datakit_io_dataway_point_total dataway uploaded points, partitioned by category and send status(HTTP status)
# TYPE datakit_io_dataway_point_total counter
datakit_io_dataway_point_total{category="dynamic_dw",status="OK"} 17
datakit_io_dataway_point_total{category="metric",status="OK"} 4572
datakit_io_dataway_point_total{category="object",status="OK"} 15688
# HELP datakit_io_feed_point_total Input feed point total
# TYPE datakit_io_feed_point_total counter
datakit_io_feed_point_total{category="logging",name="dialtesting"} 17
datakit_io_feed_point_total{category="metric",name="cpu"} 136
datakit_io_feed_point_total{category="metric",name="disk"} 1370
datakit_io_feed_point_total{category="metric",name="diskio"} 408
datakit_io_feed_point_total{category="metric",name="mem"} 136
datakit_io_feed_point_total{category="metric",name="net"} 2312
datakit_io_feed_point_total{category="metric",name="self"} 272
datakit_io_feed_point_total{category="metric",name="swap"} 135
datakit_io_feed_point_total{category="object",name="host_processes/object"} 15688
# HELP datakit_io_feed_total Input feed total
# TYPE datakit_io_feed_total counter
datakit_io_feed_total{category="logging",name="dialtesting"} 17
datakit_io_feed_total{category="metric",name="cpu"} 136
datakit_io_feed_total{category="metric",name="disk"} 137
datakit_io_feed_total{category="metric",name="diskio"} 136
datakit_io_feed_total{category="metric",name="mem"} 136
datakit_io_feed_total{category="metric",name="net"} 136
datakit_io_feed_total{category="metric",name="self"} 136
datakit_io_feed_total{category="metric",name="swap"} 135
datakit_io_feed_total{category="object",name="host_processes/object"} 45
# HELP datakit_io_flush_total IO flush total
# TYPE datakit_io_flush_total counter
datakit_io_flush_total{category="logging"} 17
datakit_io_flush_total{category="metric"} 180
datakit_io_flush_total{category="object"} 45
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
# HELP datakit_io_input_filter_point_total Input filtered point total
# TYPE datakit_io_input_filter_point_total counter
datakit_io_input_filter_point_total{category="logging",name="dialtesting"} 0
datakit_io_input_filter_point_total{category="metric",name="cpu"} 136
datakit_io_input_filter_point_total{category="metric",name="disk"} 0
datakit_io_input_filter_point_total{category="metric",name="diskio"} 0
datakit_io_input_filter_point_total{category="metric",name="mem"} 0
datakit_io_input_filter_point_total{category="metric",name="net"} 0
datakit_io_input_filter_point_total{category="metric",name="self"} 0
datakit_io_input_filter_point_total{category="metric",name="swap"} 0
datakit_io_input_filter_point_total{category="object",name="host_processes/object"} 0
# HELP datakit_io_last_feed Input last feed time(unix timestamp in second)
# TYPE datakit_io_last_feed gauge
datakit_io_last_feed{category="logging",name="dialtesting"} 1.680605678e+09
datakit_io_last_feed{category="metric",name="cpu"} 1.680606872e+09
datakit_io_last_feed{category="metric",name="disk"} 1.680606872e+09
datakit_io_last_feed{category="metric",name="diskio"} 1.68060687e+09
datakit_io_last_feed{category="metric",name="mem"} 1.680606869e+09
datakit_io_last_feed{category="metric",name="net"} 1.680606873e+09
datakit_io_last_feed{category="metric",name="self"} 1.680606864e+09
datakit_io_last_feed{category="metric",name="swap"} 1.680606867e+09
datakit_io_last_feed{category="object",name="host_processes/object"} 1.680606854e+09
# HELP datakit_io_queue_pts IO module queued(cached) points
# TYPE datakit_io_queue_pts gauge
datakit_io_queue_pts{category="logging"} 0
datakit_io_queue_pts{category="metric"} 18
datakit_io_queue_pts{category="object"} 365
# HELP datakit_open_files Datakit open files(only available on Linux)
# TYPE datakit_open_files gauge
datakit_open_files -1
# HELP datakit_pipeline_cost Pipeline total running time(ms)
# TYPE datakit_pipeline_cost summary
datakit_pipeline_cost_sum{category="metric",name="cpu.p",namespace="default"} 2.7660030000000004
datakit_pipeline_cost_count{category="metric",name="cpu.p",namespace="default"} 136
# HELP datakit_pipeline_point_total Pipeline processed total points
# TYPE datakit_pipeline_point_total counter
datakit_pipeline_point_total{category="metric",name="cpu.p",namespace="default"} 136
# HELP datakit_pipeline_update_time Pipeline last update time(unix timestamp)
# TYPE datakit_pipeline_update_time gauge
datakit_pipeline_update_time{category="logging",name="apache.p",namespace="default"} 1.68060551e+09
datakit_pipeline_update_time{category="logging",name="consul.p",namespace="default"} 1.68060551e+09
datakit_pipeline_update_time{category="logging",name="elasticsearch.p",namespace="default"} 1.68060551e+09
datakit_pipeline_update_time{category="logging",name="jenkins.p",namespace="default"} 1.68060551e+09
datakit_pipeline_update_time{category="logging",name="kafka.p",namespace="default"} 1.68060551e+09
datakit_pipeline_update_time{category="logging",name="mongodb.p",namespace="default"} 1.68060551e+09
datakit_pipeline_update_time{category="logging",name="mysql.p",namespace="default"} 1.68060551e+09
datakit_pipeline_update_time{category="logging",name="nginx.p",namespace="default"} 1.68060551e+09
datakit_pipeline_update_time{category="logging",name="postgresql.p",namespace="default"} 1.68060551e+09
datakit_pipeline_update_time{category="logging",name="rabbitmq.p",namespace="default"} 1.68060551e+09
datakit_pipeline_update_time{category="logging",name="redis.p",namespace="default"} 1.68060551e+09
datakit_pipeline_update_time{category="logging",name="solr.p",namespace="default"} 1.68060551e+09
datakit_pipeline_update_time{category="logging",name="sqlserver.p",namespace="default"} 1.68060551e+09
datakit_pipeline_update_time{category="logging",name="tdengine.p",namespace="default"} 1.68060551e+09
datakit_pipeline_update_time{category="logging",name="tomcat.p",namespace="default"} 1.68060551e+09
datakit_pipeline_update_time{category="metric",name="cpu.p",namespace="default"} 1.68060551e+09
# HELP datakit_sys_alloc Datakit memory system bytes
# TYPE datakit_sys_alloc gauge
datakit_sys_alloc 5.6820744e+07
# HELP datakit_uptime Datakit uptime(second)
# TYPE datakit_uptime gauge
datakit_uptime{auto_update="false",branch="1492-iss-dk-metrics",build_at="2023-04-04 10:49:04",cgroup="-",docker="false",hostname="tanbiaos-MacBook-Pro.local",os_arch="darwin/arm64",version="1.5.7-149-g84ad0bdc43"} 1363
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
diskcache_get_latency_sum{no_fallback_on_error="false",no_lock="false",no_pos="false",no_sync="false",path="/Users/tanbiao/datakit/cache/custom_object"} 275
diskcache_get_latency_count{no_fallback_on_error="false",no_lock="false",no_pos="false",no_sync="false",path="/Users/tanbiao/datakit/cache/custom_object"} 45
diskcache_get_latency_sum{no_fallback_on_error="false",no_lock="false",no_pos="false",no_sync="false",path="/Users/tanbiao/datakit/cache/keyevent"} 364
diskcache_get_latency_count{no_fallback_on_error="false",no_lock="false",no_pos="false",no_sync="false",path="/Users/tanbiao/datakit/cache/keyevent"} 45
diskcache_get_latency_sum{no_fallback_on_error="false",no_lock="false",no_pos="false",no_sync="false",path="/Users/tanbiao/datakit/cache/logging"} 1949
diskcache_get_latency_count{no_fallback_on_error="false",no_lock="false",no_pos="false",no_sync="false",path="/Users/tanbiao/datakit/cache/logging"} 180
diskcache_get_latency_sum{no_fallback_on_error="false",no_lock="false",no_pos="false",no_sync="false",path="/Users/tanbiao/datakit/cache/metric"} 938
diskcache_get_latency_count{no_fallback_on_error="false",no_lock="false",no_pos="false",no_sync="false",path="/Users/tanbiao/datakit/cache/metric"} 180
diskcache_get_latency_sum{no_fallback_on_error="false",no_lock="false",no_pos="false",no_sync="false",path="/Users/tanbiao/datakit/cache/network"} 15007
diskcache_get_latency_count{no_fallback_on_error="false",no_lock="false",no_pos="false",no_sync="false",path="/Users/tanbiao/datakit/cache/network"} 180
diskcache_get_latency_sum{no_fallback_on_error="false",no_lock="false",no_pos="false",no_sync="false",path="/Users/tanbiao/datakit/cache/object"} 526
diskcache_get_latency_count{no_fallback_on_error="false",no_lock="false",no_pos="false",no_sync="false",path="/Users/tanbiao/datakit/cache/object"} 45
diskcache_get_latency_sum{no_fallback_on_error="false",no_lock="false",no_pos="false",no_sync="false",path="/Users/tanbiao/datakit/cache/profiling"} 1551
diskcache_get_latency_count{no_fallback_on_error="false",no_lock="false",no_pos="false",no_sync="false",path="/Users/tanbiao/datakit/cache/profiling"} 45
diskcache_get_latency_sum{no_fallback_on_error="false",no_lock="false",no_pos="false",no_sync="false",path="/Users/tanbiao/datakit/cache/rum"} 994
diskcache_get_latency_count{no_fallback_on_error="false",no_lock="false",no_pos="false",no_sync="false",path="/Users/tanbiao/datakit/cache/rum"} 180
diskcache_get_latency_sum{no_fallback_on_error="false",no_lock="false",no_pos="false",no_sync="false",path="/Users/tanbiao/datakit/cache/security"} 822
diskcache_get_latency_count{no_fallback_on_error="false",no_lock="false",no_pos="false",no_sync="false",path="/Users/tanbiao/datakit/cache/security"} 45
diskcache_get_latency_sum{no_fallback_on_error="false",no_lock="false",no_pos="false",no_sync="false",path="/Users/tanbiao/datakit/cache/tracing"} 6494
diskcache_get_latency_count{no_fallback_on_error="false",no_lock="false",no_pos="false",no_sync="false",path="/Users/tanbiao/datakit/cache/tracing"} 180
# HELP diskcache_get_total cache Get() count
# TYPE diskcache_get_total counter
diskcache_get_total{no_fallback_on_error="false",no_lock="false",no_pos="false",no_sync="false",path="/Users/tanbiao/datakit/cache/custom_object"} 45
diskcache_get_total{no_fallback_on_error="false",no_lock="false",no_pos="false",no_sync="false",path="/Users/tanbiao/datakit/cache/keyevent"} 45
diskcache_get_total{no_fallback_on_error="false",no_lock="false",no_pos="false",no_sync="false",path="/Users/tanbiao/datakit/cache/logging"} 180
diskcache_get_total{no_fallback_on_error="false",no_lock="false",no_pos="false",no_sync="false",path="/Users/tanbiao/datakit/cache/metric"} 180
diskcache_get_total{no_fallback_on_error="false",no_lock="false",no_pos="false",no_sync="false",path="/Users/tanbiao/datakit/cache/network"} 180
diskcache_get_total{no_fallback_on_error="false",no_lock="false",no_pos="false",no_sync="false",path="/Users/tanbiao/datakit/cache/object"} 45
diskcache_get_total{no_fallback_on_error="false",no_lock="false",no_pos="false",no_sync="false",path="/Users/tanbiao/datakit/cache/profiling"} 45
diskcache_get_total{no_fallback_on_error="false",no_lock="false",no_pos="false",no_sync="false",path="/Users/tanbiao/datakit/cache/rum"} 180
diskcache_get_total{no_fallback_on_error="false",no_lock="false",no_pos="false",no_sync="false",path="/Users/tanbiao/datakit/cache/security"} 45
diskcache_get_total{no_fallback_on_error="false",no_lock="false",no_pos="false",no_sync="false",path="/Users/tanbiao/datakit/cache/tracing"} 180
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
diskcache_open_time{no_fallback_on_error="false",no_lock="false",no_pos="false",no_sync="false",path="/Users/tanbiao/datakit/cache/custom_object"} 1.68060551e+09
diskcache_open_time{no_fallback_on_error="false",no_lock="false",no_pos="false",no_sync="false",path="/Users/tanbiao/datakit/cache/dynamic_dw"} 1.68060551e+09
diskcache_open_time{no_fallback_on_error="false",no_lock="false",no_pos="false",no_sync="false",path="/Users/tanbiao/datakit/cache/keyevent"} 1.68060551e+09
diskcache_open_time{no_fallback_on_error="false",no_lock="false",no_pos="false",no_sync="false",path="/Users/tanbiao/datakit/cache/logging"} 1.68060551e+09
diskcache_open_time{no_fallback_on_error="false",no_lock="false",no_pos="false",no_sync="false",path="/Users/tanbiao/datakit/cache/metric"} 1.68060551e+09
diskcache_open_time{no_fallback_on_error="false",no_lock="false",no_pos="false",no_sync="false",path="/Users/tanbiao/datakit/cache/network"} 1.68060551e+09
diskcache_open_time{no_fallback_on_error="false",no_lock="false",no_pos="false",no_sync="false",path="/Users/tanbiao/datakit/cache/object"} 1.68060551e+09
diskcache_open_time{no_fallback_on_error="false",no_lock="false",no_pos="false",no_sync="false",path="/Users/tanbiao/datakit/cache/profiling"} 1.68060551e+09
diskcache_open_time{no_fallback_on_error="false",no_lock="false",no_pos="false",no_sync="false",path="/Users/tanbiao/datakit/cache/rum"} 1.68060551e+09
diskcache_open_time{no_fallback_on_error="false",no_lock="false",no_pos="false",no_sync="false",path="/Users/tanbiao/datakit/cache/security"} 1.68060551e+09
diskcache_open_time{no_fallback_on_error="false",no_lock="false",no_pos="false",no_sync="false",path="/Users/tanbiao/datakit/cache/tracing"} 1.68060551e+09
`)
