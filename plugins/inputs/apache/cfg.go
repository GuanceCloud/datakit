package apache

import (
	"github.com/influxdata/telegraf/plugins/common/tls"
	"gitlab.jiagouyun.com/cloudcare-tools/cliutils/logger"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"
	"net/http"
	"time"
)

var (
	inputName   = "apache"
	l           = logger.DefaultSLogger(inputName)
	minInterval = time.Second
	maxInterval = time.Second * 30
	sample      = `
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

	[inputs.apache.log]
	#	files = []
	#	# grok pipeline script path
	#	pipeline = "nginx.p"
	[inputs.apache.tags]
	# a = "b"`

	//此处 ip_or_host 可能存在 `127.0.0.1:80 127.0.0.1` 和 `127.0.0.1`	，使用 GREEDYDATA
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
	}
	tagMap = map[string]string{
		"ServerVersion": "server_version",
		"ServerMPM":     "server_mpm",
	}
)

type Input struct {
	Url      string               `toml:"url"`
	Username string               `toml:"username,omitempty"`
	Password string               `toml:"password,omitempty"`
	Interval datakit.Duration     `toml:"interval,omitempty"`
	Tags     map[string]string    `toml:"tags,omitempty"`
	Log      *inputs.TailerOption `toml:"log"`

	tls.ClientConfig

	start        time.Time
	tail         *inputs.Tailer
	collectCache []inputs.Measurement
	client       *http.Client
	lastErr      error
}

type Measurement struct {
	name   string
	tags   map[string]string
	fields map[string]interface{}
	ts     time.Time
}

func (m *Measurement) LineProto() (*io.Point, error) {
	return io.MakePoint(m.name, m.tags, m.fields, m.ts)
}

func (m *Measurement) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name: inputName,
		Fields: map[string]interface{}{
			"idle_workers":           newCountFieldInfo("The number of idle workers"),
			"busy_workers":           newCountFieldInfo("The number of workers serving requests."),
			"cpu_load":               newOtherFieldInfo(inputs.Float, inputs.Gauge, inputs.Percent, "The percent of CPU used"),
			"uptime":                 newOtherFieldInfo(inputs.Int, inputs.Gauge, inputs.DurationSecond, "The amount of time the server has been running"),
			"net_bytes":              newOtherFieldInfo(inputs.Int, inputs.Gauge, inputs.SizeByte, "The total number of bytes served."),
			"net_hits":               newCountFieldInfo("The total number of requests performed"),
			"conns_total":            newCountFieldInfo("The total number of requests performed."),
			"conns_async_writing":    newCountFieldInfo("The number of asynchronous writes connections."),
			"conns_async_keep_alive": newCountFieldInfo("The number of asynchronous keep alive connections."),
			"conns_async_closing":    newCountFieldInfo("The number of asynchronous closing connections."),
		},
		Tags: map[string]interface{}{
			"url":            inputs.NewTagInfo("apache server status url"),
			"server_version": inputs.NewTagInfo("apache server version"),
			"server_mpm":     inputs.NewTagInfo("apache server Multi-Processing Module,prefork、worker and event"),
		},
	}
}

func newCountFieldInfo(desc string) *inputs.FieldInfo {
	return &inputs.FieldInfo{
		DataType: inputs.Int,
		Type:     inputs.Count,
		Unit:     inputs.NCount,
		Desc:     desc,
	}
}

func newOtherFieldInfo(datatype, Type, unit, desc string) *inputs.FieldInfo {
	return &inputs.FieldInfo{
		DataType: datatype,
		Type:     Type,
		Unit:     unit,
		Desc:     desc,
	}
}
