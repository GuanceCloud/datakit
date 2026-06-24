// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package apache

import (
	"time"

	dkio "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/io"
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

  ## Set true to enable election
  election = true

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
		"IdleWorkers":         idleWorkers,
		"BusyWorkers":         busyWorkers,
		"MaxWorkers":          maxWorkers,
		"CPULoad":             cpuLoad,
		"Uptime":              uptime,
		"TotalkBytes":         netBytes,
		"TotalAccesses":       netHits,
		"ConnsTotal":          connsTotal,
		"ConnsAsyncWriting":   connsAsyncWriting,
		"ConnsAsyncKeepAlive": connsAsyncKeepAlive,
		"ConnsAsyncClosing":   connsAsyncClosing,
		"Scoreboard":          "scoreboard",
	}
	tagMap = map[string]string{
		"ServerVersion": "server_version",
		"ServerMPM":     "server_mpm",
	}

	customObjectFeedName = dkio.FeedSource(inputName, "CO")
)

const (
	idleWorkers          = "idle_workers"
	busyWorkers          = "busy_workers"
	maxWorkers           = "max_workers"
	cpuLoad              = "cpu_load"
	uptime               = "uptime"
	netBytes             = "net_bytes"
	netHits              = "net_hits"
	connsTotal           = "conns_total"
	connsAsyncWriting    = "conns_async_writing"
	connsAsyncKeepAlive  = "conns_async_keep_alive"
	connsAsyncClosing    = "conns_async_closing"
	waitingForConnection = "waiting_for_connection"
	startingUp           = "starting_up"
	readingRequest       = "reading_request"
	sendingReply         = "sending_reply"
	keepAlive            = "keepalive"
	dnsLookup            = "dns_lookup"
	closingConnection    = "closing_connection"
	logging              = "logging"
	gracefullyFinishing  = "gracefully_finishing"
	idleCleanup          = "idle_cleanup"
	openSlot             = "open_slot"
	disabled             = "disabled"

	apacheTagURL           = "url"
	apacheTagHost          = "host"
	apacheTagServerVersion = "server_version"
	apacheTagServerMPM     = "server_mpm"
)
