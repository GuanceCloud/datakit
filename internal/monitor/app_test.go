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
	t.Skip() // disable the test during CI

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
	t.Skip() // disable the test during CI

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
# HELP datakit_cpu_usage Datakit CPU usage(%)
# TYPE datakit_cpu_usage gauge
datakit_cpu_usage 0.3878504673581006
# HELP datakit_election_api Election API latency(in millisecond)
# TYPE datakit_election_api counter
datakit_election_api{namespace="default",status="defeat"} 186
# HELP datakit_election_input_count Datakit election input count
# TYPE datakit_election_input_count gauge
datakit_election_input_count{namespace="default"} 3
# HELP datakit_election_pause_total Input paused count when election failed
# TYPE datakit_election_pause_total counter
datakit_election_pause_total{id="tanbiaos-MacBook-Pro.local",namespace="default"} 3
# HELP datakit_election_status Datakit election status, if metric = 0, meas not elected, or the elected time(unix timestamp second)
# TYPE datakit_election_status gauge
datakit_election_status{elected_id="tan-vm",id="tanbiaos-MacBook-Pro.local",namespace="default",status="defeat"} 0
# HELP datakit_error_total total errors, only count on error source, not include error message
# TYPE datakit_error_total counter
datakit_error_total{category="metric",source="mysql"} 58
# HELP datakit_filter_last_update
# TYPE datakit_filter_last_update gauge
datakit_filter_last_update 1.679236376e+09
# HELP datakit_filter_latency Filter latency(us) of these filters
# TYPE datakit_filter_latency summary
datakit_filter_latency_sum{category="metric",filters="{ measurement =  'cpu'  and ( host notin [ 'ubt-dev-01' ] )}",source="remote"} 30088
datakit_filter_latency_count{category="metric",filters="{ measurement =  'cpu'  and ( host notin [ 'ubt-dev-01' ] )}",source="remote"} 458
# HELP datakit_filter_point_dropped_total Dropped points of filters
# TYPE datakit_filter_point_dropped_total counter
datakit_filter_point_dropped_total{category="metric",filters="{ measurement =  'cpu'  and ( host notin [ 'ubt-dev-01' ] )}",source="remote"} 57
# HELP datakit_filter_point_total Filter points of filters
# TYPE datakit_filter_point_total counter
datakit_filter_point_total{category="metric",filters="{ measurement =  'cpu'  and ( host notin [ 'ubt-dev-01' ] )}",source="remote"} 2047
# HELP datakit_filter_pull_latency Filter pull(remote) latency(ms)
# TYPE datakit_filter_pull_latency summary
datakit_filter_pull_latency_sum{status="ok"} 2873
datakit_filter_pull_latency_count{status="ok"} 58
# HELP datakit_filter_update_total Filters(remote) updated count
# TYPE datakit_filter_update_total counter
datakit_filter_update_total 1
# HELP datakit_gc_summary Datakit golang GC paused(nano-second)
# TYPE datakit_gc_summary summary
datakit_gc_summary_sum 9.8162621e+07
datakit_gc_summary_count 347
# HELP datakit_goroutine_alive alive goroutines
# TYPE datakit_goroutine_alive gauge
datakit_goroutine_alive{name="election"} 1
datakit_goroutine_alive{name="http"} 3
datakit_goroutine_alive{name="inputs"} 16
datakit_goroutine_alive{name="inputs_cpu"} 1
datakit_goroutine_alive{name="inputs_dialtesting"} 0
datakit_goroutine_alive{name="internal_cgroup"} 0
datakit_goroutine_alive{name="io"} 112
datakit_goroutine_alive{name="io_dnswatcher"} 1
datakit_goroutine_alive{name="pipeline_remote"} 1
# HELP datakit_goroutine_cost
# TYPE datakit_goroutine_cost summary
datakit_goroutine_cost_sum{name="inputs"} 3.7405016041e+10
datakit_goroutine_cost_count{name="inputs"} 1
datakit_goroutine_cost_sum{name="inputs_dialtesting"} 1.70000902459e+11
datakit_goroutine_cost_count{name="inputs_dialtesting"} 1
datakit_goroutine_cost_sum{name="internal_cgroup"} 165417
datakit_goroutine_cost_count{name="internal_cgroup"} 1
# HELP datakit_goroutine_group_total
# TYPE datakit_goroutine_group_total gauge
datakit_goroutine_group_total 23
# HELP datakit_goroutine_stopped stopped goroutines
# TYPE datakit_goroutine_stopped counter
datakit_goroutine_stopped{name="inputs"} 1
datakit_goroutine_stopped{name="inputs_dialtesting"} 1
datakit_goroutine_stopped{name="internal_cgroup"} 1
# HELP datakit_goroutines goroutine count within Datakit
# TYPE datakit_goroutines gauge
datakit_goroutines 173
# HELP datakit_heap_alloc Datakit memory heap bytes
# TYPE datakit_heap_alloc gauge
datakit_heap_alloc 3.4335536e+07
# HELP datakit_http_api_elapsed API request cost(in ms)
# TYPE datakit_http_api_elapsed summary
datakit_http_api_elapsed_sum{api="/metrics",method="GET",status="OK"} 782
datakit_http_api_elapsed_count{api="/metrics",method="GET",status="OK"} 191
# HELP datakit_http_api_elapsed_histogram API request cost(in ms) histogram
# TYPE datakit_http_api_elapsed_histogram histogram
datakit_http_api_elapsed_histogram_bucket{api="/metrics",method="GET",status="OK",le="10"} 191
datakit_http_api_elapsed_histogram_bucket{api="/metrics",method="GET",status="OK",le="100"} 191
datakit_http_api_elapsed_histogram_bucket{api="/metrics",method="GET",status="OK",le="1000"} 191
datakit_http_api_elapsed_histogram_bucket{api="/metrics",method="GET",status="OK",le="5000"} 191
datakit_http_api_elapsed_histogram_bucket{api="/metrics",method="GET",status="OK",le="30000"} 191
datakit_http_api_elapsed_histogram_bucket{api="/metrics",method="GET",status="OK",le="+Inf"} 191
datakit_http_api_elapsed_histogram_sum{api="/metrics",method="GET",status="OK"} 782
datakit_http_api_elapsed_histogram_count{api="/metrics",method="GET",status="OK"} 191
# HELP datakit_http_api_req_size API request body size
# TYPE datakit_http_api_req_size summary
datakit_http_api_req_size_sum{api="/metrics",method="GET",status="OK"} 15280
datakit_http_api_req_size_count{api="/metrics",method="GET",status="OK"} 191
# HELP datakit_http_api_total API request counter
# TYPE datakit_http_api_total counter
datakit_http_api_total{api="/metrics",method="GET",status="OK"} 191
# HELP datakit_input_collect_latency Input collect latency(us)
# TYPE datakit_input_collect_latency summary
datakit_input_collect_latency_sum{category="logging",name="dialtesting"} 0
datakit_input_collect_latency_count{category="logging",name="dialtesting"} 17
datakit_input_collect_latency_sum{category="metric",name="cpu"} 11549
datakit_input_collect_latency_count{category="metric",name="cpu"} 57
datakit_input_collect_latency_sum{category="metric",name="disk"} 31198
datakit_input_collect_latency_count{category="metric",name="disk"} 58
datakit_input_collect_latency_sum{category="metric",name="diskio"} 31492
datakit_input_collect_latency_count{category="metric",name="diskio"} 58
datakit_input_collect_latency_sum{category="metric",name="mem"} 6715
datakit_input_collect_latency_count{category="metric",name="mem"} 58
datakit_input_collect_latency_sum{category="metric",name="net"} 199199
datakit_input_collect_latency_count{category="metric",name="net"} 56
datakit_input_collect_latency_sum{category="metric",name="self"} 8.1887932e+07
datakit_input_collect_latency_count{category="metric",name="self"} 57
datakit_input_collect_latency_sum{category="metric",name="swap"} 3803
datakit_input_collect_latency_count{category="metric",name="swap"} 57
datakit_input_collect_latency_sum{category="metric",name="system"} 36535
datakit_input_collect_latency_count{category="metric",name="system"} 57
datakit_input_collect_latency_sum{category="object",name="host_processes/object"} 4.40945068e+08
datakit_input_collect_latency_count{category="object",name="host_processes/object"} 19
datakit_input_collect_latency_sum{category="object",name="hostobject"} 3.445273e+06
datakit_input_collect_latency_count{category="object",name="hostobject"} 58
# HELP datakit_inputs_instance input instance count
# TYPE datakit_inputs_instance gauge
datakit_inputs_instance{input="container"} 1
datakit_inputs_instance{input="cpu"} 1
datakit_inputs_instance{input="ddtrace"} 1
datakit_inputs_instance{input="dialtesting"} 1
datakit_inputs_instance{input="disk"} 1
datakit_inputs_instance{input="diskio"} 1
datakit_inputs_instance{input="host_processes"} 1
datakit_inputs_instance{input="hostobject"} 1
datakit_inputs_instance{input="logstreaming"} 1
datakit_inputs_instance{input="mem"} 1
datakit_inputs_instance{input="mongodb"} 1
datakit_inputs_instance{input="mysql"} 1
datakit_inputs_instance{input="net"} 1
datakit_inputs_instance{input="rum"} 1
datakit_inputs_instance{input="self"} 1
datakit_inputs_instance{input="swap"} 1
datakit_inputs_instance{input="system"} 1
# HELP datakit_io_chan_capacity io channel capacity
# TYPE datakit_io_chan_capacity gauge
datakit_io_chan_capacity{category="all-the-same"} 100
# HELP datakit_io_chan_usage io channel usage(length of the chan)
# TYPE datakit_io_chan_usage gauge
datakit_io_chan_usage{category="logging"} 0
datakit_io_chan_usage{category="metric"} 0
datakit_io_chan_usage{category="object"} 0
# HELP datakit_io_dataway_api_latency dataway HTTP request latency(ms) partitioned by HTTP API(method@url) and HTTP status
# TYPE datakit_io_dataway_api_latency summary
datakit_io_dataway_api_latency_sum{api="/v1/datakit/pull",status="200 OK"} 3603
datakit_io_dataway_api_latency_count{api="/v1/datakit/pull",status="200 OK"} 68
datakit_io_dataway_api_latency_sum{api="/v1/election",status="200 OK"} 9767
datakit_io_dataway_api_latency_count{api="/v1/election",status="200 OK"} 186
datakit_io_dataway_api_latency_sum{api="/v1/write/logging",status="200 OK"} 732
datakit_io_dataway_api_latency_count{api="/v1/write/logging",status="200 OK"} 17
datakit_io_dataway_api_latency_sum{api="/v1/write/metric",status="200 OK"} 19663
datakit_io_dataway_api_latency_count{api="/v1/write/metric",status="200 OK"} 397
datakit_io_dataway_api_latency_sum{api="/v1/write/object",status="200 OK"} 4204
datakit_io_dataway_api_latency_count{api="/v1/write/object",status="200 OK"} 57
# HELP datakit_io_dataway_api_request_total dataway HTTP request processed, partitioned by status code and HTTP API(method@url)
# TYPE datakit_io_dataway_api_request_total counter
datakit_io_dataway_api_request_total{api="/v1/datakit/pull",status="200 OK"} 68
datakit_io_dataway_api_request_total{api="/v1/election",status="200 OK"} 186
datakit_io_dataway_api_request_total{api="/v1/write/logging",status="200 OK"} 17
datakit_io_dataway_api_request_total{api="/v1/write/metric",status="200 OK"} 397
datakit_io_dataway_api_request_total{api="/v1/write/object",status="200 OK"} 57
# HELP datakit_io_dataway_point_bytes_total dataway uploaded points bytes, partitioned by category and pint send status(ok/failed/dropped)
# TYPE datakit_io_dataway_point_bytes_total counter
datakit_io_dataway_point_bytes_total{category="/v1/write/metric",status="200 OK"} 189701
datakit_io_dataway_point_bytes_total{category="/v1/write/object",status="200 OK"} 1.437387e+06
datakit_io_dataway_point_bytes_total{category="dynamicDatawayCategory",status="200 OK"} 9343
# HELP datakit_io_dataway_point_total dataway uploaded points, partitioned by category and send status(ok/failed/dropped)
# TYPE datakit_io_dataway_point_total counter
datakit_io_dataway_point_total{category="/v1/write/metric",status="ok"} 1976
datakit_io_dataway_point_total{category="/v1/write/object",status="ok"} 9994
datakit_io_dataway_point_total{category="dynamicDatawayCategory",status="ok"} 17
# HELP datakit_io_dataway_sink_total dataway sink count, partitioned by category.
# TYPE datakit_io_dataway_sink_total counter
datakit_io_dataway_sink_total{category="/v1/write/metric"} 397
datakit_io_dataway_sink_total{category="/v1/write/object"} 57
# HELP datakit_io_feed_point_total Input feed point total
# TYPE datakit_io_feed_point_total counter
datakit_io_feed_point_total{category="logging",name="dialtesting"} 17
datakit_io_feed_point_total{category="metric",name="cpu"} 57
datakit_io_feed_point_total{category="metric",name="disk"} 522
datakit_io_feed_point_total{category="metric",name="diskio"} 174
datakit_io_feed_point_total{category="metric",name="mem"} 58
datakit_io_feed_point_total{category="metric",name="net"} 1008
datakit_io_feed_point_total{category="metric",name="self"} 114
datakit_io_feed_point_total{category="metric",name="swap"} 57
datakit_io_feed_point_total{category="metric",name="system"} 57
datakit_io_feed_point_total{category="object",name="host_processes/object"} 9937
datakit_io_feed_point_total{category="object",name="hostobject"} 58
# HELP datakit_io_feed_total Input feed total
# TYPE datakit_io_feed_total counter
datakit_io_feed_total{category="logging",name="dialtesting"} 17
datakit_io_feed_total{category="metric",name="cpu"} 57
datakit_io_feed_total{category="metric",name="disk"} 58
datakit_io_feed_total{category="metric",name="diskio"} 58
datakit_io_feed_total{category="metric",name="mem"} 58
datakit_io_feed_total{category="metric",name="mysql"} 58
datakit_io_feed_total{category="metric",name="net"} 56
datakit_io_feed_total{category="metric",name="self"} 57
datakit_io_feed_total{category="metric",name="swap"} 57
datakit_io_feed_total{category="metric",name="system"} 57
datakit_io_feed_total{category="object",name="host_processes/object"} 19
datakit_io_feed_total{category="object",name="hostobject"} 58
# HELP datakit_io_input_filter_point_total Input filtered point total
# TYPE datakit_io_input_filter_point_total counter
datakit_io_input_filter_point_total{category="logging",name="dialtesting"} 0
datakit_io_input_filter_point_total{category="metric",name="cpu"} 57
datakit_io_input_filter_point_total{category="metric",name="disk"} 0
datakit_io_input_filter_point_total{category="metric",name="diskio"} 0
datakit_io_input_filter_point_total{category="metric",name="mem"} 0
datakit_io_input_filter_point_total{category="metric",name="net"} 0
datakit_io_input_filter_point_total{category="metric",name="self"} 0
datakit_io_input_filter_point_total{category="metric",name="swap"} 0
datakit_io_input_filter_point_total{category="metric",name="system"} 0
datakit_io_input_filter_point_total{category="object",name="host_processes/object"} 0
datakit_io_input_filter_point_total{category="object",name="hostobject"} 0
# HELP datakit_io_last_feed Input last feed time(unix timestamp in second)
# TYPE datakit_io_last_feed gauge
datakit_io_last_feed{category="logging",name="dialtesting"} 1.679236545e+09
datakit_io_last_feed{category="metric",name="cpu"} 1.679236951e+09
datakit_io_last_feed{category="metric",name="disk"} 1.679236948e+09
datakit_io_last_feed{category="metric",name="diskio"} 1.679236949e+09
datakit_io_last_feed{category="metric",name="mem"} 1.67923695e+09
datakit_io_last_feed{category="metric",name="mysql"} 1.679236946e+09
datakit_io_last_feed{category="metric",name="net"} 1.679236945e+09
datakit_io_last_feed{category="metric",name="self"} 1.679236942e+09
datakit_io_last_feed{category="metric",name="swap"} 1.679236948e+09
datakit_io_last_feed{category="metric",name="system"} 1.679236942e+09
datakit_io_last_feed{category="object",name="host_processes/object"} 1.679236944e+09
datakit_io_last_feed{category="object",name="hostobject"} 1.679236951e+09
# HELP datakit_io_queue_pts IO module queued(cached) points
# TYPE datakit_io_queue_pts gauge
datakit_io_queue_pts{category="/v1/write/logging"} 0
datakit_io_queue_pts{category="/v1/write/metric"} 0
datakit_io_queue_pts{category="/v1/write/object"} 1
# HELP datakit_last_err Datakit errors, these errors come from inputs or any sub modules
# TYPE datakit_last_err gauge
datakit_last_err{category="metric",error="dial tcp 10.100.1.2:3306: connect: no route to host",source="mysql"} 1.679236946e+09
# HELP datakit_open_files Datakit open files(only available on Linux)
# TYPE datakit_open_files gauge
datakit_open_files -1
# HELP datakit_start_time Datakit start time(unix timestamp second)
# TYPE datakit_start_time gauge
datakit_start_time 0
# HELP datakit_sys_alloc Datakit memory system bytes
# TYPE datakit_sys_alloc gauge
datakit_sys_alloc 6.5537032e+07
# HELP datakit_uptime Datakit uptime(second)
# TYPE datakit_uptime counter
datakit_uptime{auto_update="false",branch="1492-iss-dk-metrics",build_at="2023-03-19 14:32:29",cgroup="not ready",docker="false",hostname="tanbiaos-MacBook-Pro.local",os_arch="darwin/arm64",version="1.5.6-854-g1a84464321"} 575
`)
