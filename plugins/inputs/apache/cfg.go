package apache

import (
	"time"
)

const (
	inputName   = "apache"
	minInterval = time.Second
	maxInterval = time.Second * 30

	//nolint:lll
	sample = `
[[inputs.apache]]
  url = "http://127.0.0.1/server-status?auto"
  # ##(optional) collection interval, default is 30s
  # interval = "30s"

  # username = ""
  # password = ""

  ## Optional TLS Config
  # tls_ca = "/xxx/ca.pem"
  # tls_cert = "/xxx/cert.cer"
  # tls_key = "/xxx/key.key"
  ## Use TLS but skip chain & host verification
  insecure_skip_verify = false

  # [inputs.apache.log]
  # files = []
  # #grok pipeline script path
  # pipeline = "apache.p"

  [inputs.apache.tags]
  # some_tag = "some_value"
  # more_tag = "some_other_value"
  # ... `

	// 此处 ip_or_host 可能存在 `127.0.0.1:80 127.0.0.1` 和 `127.0.0.1`	，使用 GREEDYDATA.
	//nolint:lll
	pipeline = `
# access log
grok(_,"%{GREEDYDATA:ip_or_host} - - \\[%{HTTPDATE:time}\\] \"%{DATA:http_method} %{GREEDYDATA:http_url} HTTP/%{NUMBER:http_version}\" %{NUMBER:http_code} ")
grok(_,"%{GREEDYDATA:ip_or_host} - - \\[%{HTTPDATE:time}\\] \"-\" %{NUMBER:http_code} ")
default_time(time)
cast(http_code,"int")

# error log
grok(_,"\\[%{HTTPDERROR_DATE:time}\\] \\[%{GREEDYDATA:type}:%{GREEDYDATA:status}\\] \\[pid %{GREEDYDATA:pid}:tid %{GREEDYDATA:tid}\\] ")
grok(_,"\\[%{HTTPDERROR_DATE:time}\\] \\[%{GREEDYDATA:type}:%{GREEDYDATA:status}\\] \\[pid %{INT:pid}\\] ")
default_time(time)
`
)

var (
	filedMap = map[string]string{
		"IdleWorkers":         "idle_workers",
		"BusyWorkers":         "busy_workers",
		"CPULoad":             "cpu_load",
		"Uptime":              "uptime",
		"TotalkBytes":         "net_bytes",
		"TotalAccesses":       "net_hits",
		"ConnsTotal":          "conns_total",
		"ConnsAsyncWriting":   "conns_async_writing",
		"ConnsAsyncKeepAlive": "conns_async_keep_alive",
		"ConnsAsyncClosing":   "conns_async_closing",
		"Scoreboard":          "scoreboard",
	}
	tagMap = map[string]string{
		"ServerVersion": "server_version",
		"ServerMPM":     "server_mpm",
	}
)

// scoreboard metrics.
const (
	WaitingForConnection = "waiting_for_connection"
	StartingUp           = "starting_up"
	ReadingRequest       = "reading_request"
	SendingReply         = "sending_reply"
	KeepAlive            = "keepalive"
	DNSLookup            = "dns_lookup"
	ClosingConnection    = "closing_connection"
	Logging              = "logging"
	GracefullyFinishing  = "gracefully_finishing"
	IdleCleanup          = "idle_cleanup"
	OpenSlot             = "open_slot"
)
