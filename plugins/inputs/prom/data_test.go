package prom

type MockPoint struct {
	Field string            `json:"field"`
	Value float64           `json:"value"`
	Tags  map[string]string `json:"tags"`
}

type MockMetric struct {
	Type   string      `json:"type"`
	Points []MockPoint `json:"points"`
}

var promJsonStr = `
{
	"node_network_transmit_packets_untyped": {
		"type": "untyped",
		"points": [
			{
				"tags": {"device": "XHC20"},
				"value": 0
			}
		]
	},
	"node_network_transmit_packets_total": {
		"type": "counter",
		"points": [
			{
				"tags": {"device": "XHC20"},
				"value": 0
			},
			{
				"tags": {"device": "awdl0"},
				"value": 5783
			},
			{
				"tags": {"device": "bridge0"},
				"value": 0
			},
			{
				"tags": {"device": "en0"},
				"value": 1.974653e+07
			}
		]
	},
	"node_scrape_collector_duration_seconds": {
		"type": "gauge",
		"points": [
			{
				"tags": {"collector": "boottime"},
				"value": 4.299e-05
			},
			{
				"tags": {"collector": "cpu"},
				"value": 0.000179088
			},
			{
				"tags": {"collector": "diskstats"},
				"value": 0.001072566
			},
			{
				"tags": {"collector": "filesystem"},
				"value": 0.000199123
			}
		]
	},
	"go_gc_duration_seconds": {
		"type": "summary",
		"points": [
			{
				"tags": {"quantile": "0"},
				"value": 2.9713e-05
			},
			{
				"tags": {"quantile": "0.25"},
				"value": 7.8652e-05
			},
			{
				"tags": {"quantile": "0.5"},
				"value": 0.000112799
			},
			{
				"tags": {"quantile": "0.75"},
				"value": 0.000178596
			},
			{
				"tags": {"quantile": "1"},
				"value": 0.003374893
			},
			{
				"field": "go_gc_duration_seconds_sum",
				"value": 0.149290958
			},
			{
				"field": "go_gc_duration_seconds_count",
				"value": 435
			}
		]
	},
	"confluence_request_duration_on_path": {
		"type": "histogram",
		"points": [
			{
				"field": "confluence_request_duration_on_path_bucket",
				"tags": {"path": "/rest", "le": "0.005"},
				"value": 0.0
			},
			{
				"field": "confluence_request_duration_on_path_bucket",
				"tags": {"path": "/rest", "le": "0.01"},
				"value": 4.0
			},
			{
				"field": "confluence_request_duration_on_path_bucket",
				"tags": {"path": "/rest", "le": "0.025"},
				"value": 5.0
			},
			{
				"field": "confluence_request_duration_on_path_bucket",
				"tags": {"path": "/rest", "le": "+Inf"},
				"value": 6.0
			},
			{
				"field": "confluence_request_duration_on_path_count",
				"tags": {"path": "/rest"},
				"value": 6.0
			},
			{
				"field": "confluence_request_duration_on_path_sum",
				"tags": {"path": "/rest"},
				"value": 2.336312921
			}
		]
	}
	
}
`

var promText = `
# HELP node_network_transmit_packets_untyped Network device statistic transmit_packets.
# TYPE node_network_transmit_packets_untyped untyped
node_network_transmit_packets_untyped{device="XHC20"} 0
# HELP node_network_transmit_packets_total Network device statistic transmit_packets.
# TYPE node_network_transmit_packets_total counter
node_network_transmit_packets_total{device="XHC20"} 0
node_network_transmit_packets_total{device="awdl0"} 5783
node_network_transmit_packets_total{device="bridge0"} 0
node_network_transmit_packets_total{device="en0"} 1.974653e+07
# HELP node_scrape_collector_duration_seconds node_exporter: Duration of a collector scrape.
# TYPE node_scrape_collector_duration_seconds gauge
node_scrape_collector_duration_seconds{collector="boottime"} 4.299e-05
node_scrape_collector_duration_seconds{collector="cpu"} 0.000179088
node_scrape_collector_duration_seconds{collector="diskstats"} 0.001072566
node_scrape_collector_duration_seconds{collector="filesystem"} 0.000199123
# HELP go_gc_duration_seconds A summary of the pause duration of garbage collection cycles.
# TYPE go_gc_duration_seconds summary
go_gc_duration_seconds{quantile="0"} 2.9713e-05
go_gc_duration_seconds{quantile="0.25"} 7.8652e-05
go_gc_duration_seconds{quantile="0.5"} 0.000112799
go_gc_duration_seconds{quantile="0.75"} 0.000178596
go_gc_duration_seconds{quantile="1"} 0.003374893
go_gc_duration_seconds_sum 0.149290958
go_gc_duration_seconds_count 435
# HELP confluence_request_duration_on_path Request duration on path
# TYPE confluence_request_duration_on_path histogram
confluence_request_duration_on_path_bucket{path="/rest",le="0.005",} 0.0
confluence_request_duration_on_path_bucket{path="/rest",le="0.01",} 4.0
confluence_request_duration_on_path_bucket{path="/rest",le="0.025",} 5.0
confluence_request_duration_on_path_bucket{path="/rest",le="+Inf",} 6.0
confluence_request_duration_on_path_count{path="/rest",} 6.0
confluence_request_duration_on_path_sum{path="/rest",} 2.336312921
`
