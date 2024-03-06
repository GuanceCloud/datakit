// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

// Package healthcheck monitors the process/tcp/http health status
package healthcheck

import (
	"encoding/json"
	"regexp"
	"time"

	"github.com/GuanceCloud/cliutils"
	"github.com/GuanceCloud/cliutils/logger"
	"github.com/GuanceCloud/cliutils/point"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/config"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/export/doc"
	dkio "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/io"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs"
)

var (
	l                                = logger.DefaultSLogger(inputName)
	minMetricInterval                = time.Second * 10
	maxMetricInterval                = time.Minute * 60
	_                 inputs.ReadEnv = (*Input)(nil)
)

type process struct {
	Names      []string `toml:"names" json:"names"`             // process names
	NamesRegex []string `toml:"names_regex" json:"names_regex"` // process names regex
	MinRunTime string   `toml:"min_run_time" json:"min_run_time"`

	namesRegex []*regexp.Regexp
	minRunTime time.Duration
	processes  map[int32]*processInfo
}

type tcp struct {
	HostPorts         []string `toml:"host_ports" json:"host_ports"`
	ConnectionTimeOut string   `toml:"connection_timeout" json:"connection_timeout"`

	connectionTimeout time.Duration
}

type http struct {
	HTTPURLs          []string          `toml:"http_urls" json:"http_urls"`
	Method            string            `toml:"method" json:"method"`
	ExpectStatus      int               `toml:"expect_status" json:"expect_status"`
	Timeout           string            `toml:"timeout" json:"timeout"`
	IgnoreInsecureTLS bool              `toml:"ignore_insecure_tls" json:"ignore_insecure_tls"`
	Headers           map[string]string `toml:"headers" json:"headers"`
}

type Input struct {
	Interval string            `toml:"interval" json:"interval"`
	Process  []*process        `toml:"process" json:"process"`
	TCP      []*tcp            `toml:"tcp" json:"tcp"`
	HTTP     []*http           `toml:"http" json:"http"`
	Tags     map[string]string `toml:"tags" json:"tags"`

	semStop      *cliutils.Sem // start stop signal
	feeder       dkio.Feeder
	collectCache []*point.Point
	tagger       datakit.GlobalTagger
	mergedTags   map[string]string
	collectFuncs map[string]func() error
	tcp          []*tcp
	http         []*http
	process      []*process
}

func (*Input) Catalog() string { return category }

func (*Input) SampleConfig() string { return sampleConfig }

func (*Input) AvailableArchs() []string { return datakit.AllOS }

func (*Input) SampleMeasurement() []inputs.Measurement {
	return []inputs.Measurement{&ProcessMetric{}, &TCPMetric{}, &HTTPMetric{}}
}

func (ipt *Input) Collect() error {
	ipt.collectCache = make([]*point.Point, 0)

	for k, f := range ipt.collectFuncs {
		if err := f(); err != nil {
			l.Warnf("check %s failed: %s", k, err.Error())
		}
	}

	return nil
}

func (ipt *Input) initConfig() {
	ipt.mergedTags = inputs.MergeTags(ipt.tagger.HostTags(), ipt.Tags, "")
	for _, process := range ipt.Process {
		for _, v := range process.NamesRegex {
			if r, err := regexp.Compile(v); err != nil {
				l.Warnf("regexp compile(%s) error: %s, ignored", v, err.Error())
			} else {
				process.namesRegex = append(process.namesRegex, r)
			}
		}

		// parse min run time and set default
		if process.MinRunTime != "" {
			if du, err := time.ParseDuration(process.MinRunTime); err != nil {
				l.Warnf("parse duration error: %s, using default %s", err.Error(), defaultMinRunTime)
			} else {
				process.minRunTime = du
			}
		}
		if process.minRunTime == 0 {
			process.minRunTime = 10 * time.Minute
		}

		if len(process.namesRegex) == 0 && len(process.Names) == 0 {
			continue
		}

		ipt.process = append(ipt.process, process)
	}

	for _, tcp := range ipt.TCP {
		// parse connection timeout and set default
		if tcp.ConnectionTimeOut != "" {
			if du, err := time.ParseDuration(tcp.ConnectionTimeOut); err != nil {
				l.Warnf("parse duration error: %s", err.Error())
			} else {
				tcp.connectionTimeout = du
			}
		}
		if tcp.connectionTimeout == 0 {
			tcp.connectionTimeout = 3 * time.Second
		}

		tcp.HostPorts = filterEmptyValues(tcp.HostPorts)

		if len(tcp.HostPorts) == 0 {
			continue
		}

		ipt.tcp = append(ipt.tcp, tcp)
	}

	for _, http := range ipt.HTTP {
		// set default method
		if http.Method == "" {
			http.Method = "GET"
		}

		// set default expect status
		if http.ExpectStatus == 0 {
			http.ExpectStatus = 200
		}

		// remove empty url
		http.HTTPURLs = filterEmptyValues(http.HTTPURLs)

		if len(http.HTTPURLs) == 0 {
			continue
		}

		ipt.http = append(ipt.http, http)
	}

	ipt.collectFuncs = make(map[string]func() error, 0)
	if len(ipt.process) > 0 {
		ipt.collectFuncs["process"] = ipt.collectProcess
	}

	if len(ipt.tcp) > 0 {
		ipt.collectFuncs["tcp"] = ipt.collectTCP
	}

	if len(ipt.http) > 0 {
		ipt.collectFuncs["http"] = ipt.collectHTTP
	}
}

func (ipt *Input) Run() {
	l = logger.SLogger(inputName)
	l.Info("host_healthcheck start...")

	duration, err := time.ParseDuration(ipt.Interval)
	if err != nil {
		l.Error("invalid interval, %s", err.Error())
		return
	} else if duration <= 0 {
		l.Error("invalid interval, cannot be less than zero")
		return
	}

	ipt.initConfig()

	duration = config.ProtectedInterval(minMetricInterval, maxMetricInterval, duration)

	tick := time.NewTicker(duration)
	defer tick.Stop()

	for {
		start := time.Now()
		if err := ipt.Collect(); err != nil {
			ipt.feeder.FeedLastError(err.Error(),
				dkio.WithLastErrorInput(inputName),
			)
			l.Error(err)
		}

		if len(ipt.collectCache) > 0 {
			if err := ipt.feeder.Feed(inputName, point.Metric, ipt.collectCache,
				&dkio.Option{CollectCost: time.Since(start)}); err != nil {
				ipt.feeder.FeedLastError(err.Error(),
					dkio.WithLastErrorInput(inputName),
					dkio.WithLastErrorCategory(point.Metric),
				)
				l.Errorf("feed measurement: %s", err)
			}
		}

		select {
		case <-tick.C:
		case <-datakit.Exit.Wait():
			l.Info("host_healthcheck exit")
			return

		case <-ipt.semStop.Wait():
			l.Info("host_healthcheck return")
			return
		}
	}
}

func (ipt *Input) GetENVDoc() []*inputs.ENVInfo {
	// nolint:lll
	infos := []*inputs.ENVInfo{
		{FieldName: "Interval"},
		{FieldName: "Process", Type: doc.JSON, Example: `[{"names":["nginx","mysql"],"min_run_time":"10m"}]`, Desc: "Check process", DescZh: "检查处理器"},
		{FieldName: "TCP", Type: doc.JSON, Example: `[{"host_ports":["10.100.1.2:3369","192.168.1.2:6379"],"connection_timeout":"3s"}]`, Desc: "Check TCP", DescZh: "检查 TCP"},
		{FieldName: "HTTP", Type: doc.JSON, Example: `[{"http_urls":["http://local-ip:port/path/to/api?arg1=x&arg2=y"],"method":"GET","expect_status":200,"timeout":"30s","ignore_insecure_tls":false,"headers":{"Header1":"header-value-1","Hedaer2":"header-value-2"}}]`, Desc: "Check HTTP", DescZh: "检查 HTTP"},
		{FieldName: "Tags", Type: doc.JSON, Example: `{"some_tag":"some_value","more_tag":"some_other_value"}`},
	}

	return doc.SetENVDoc("ENV_INPUT_HEALTHCHECK_", infos)
}

// ReadEnv support envs：
//
//		ENV_INPUT_HEALTHCHECK_INTERVAL : string
//		ENV_INPUT_HEALTHCHECK_PROCESS : JSON string
//		ENV_INPUT_HEALTHCHECK_TCP : JSON string
//		ENV_INPUT_HEALTHCHECK_HTTP : JSON string
//		ENV_INPUT_HEALTHCHECK_TAGS : JSON string

func (ipt *Input) ReadEnv(envs map[string]string) {
	if str, ok := envs["ENV_INPUT_HEALTHCHECK_INTERVAL"]; ok {
		ipt.Interval = str
	}

	if value, ok := envs["ENV_INPUT_HEALTHCHECK_PROCESS"]; ok {
		conf := []*process{}
		if err := json.Unmarshal([]byte(value), &conf); err != nil {
			l.Warnf("parse ENV_INPUT_HEALTHCHECK_PROCESS=%s failed, %s", value, err.Error())
		} else {
			ipt.Process = conf
		}
	}

	if value, ok := envs["ENV_INPUT_HEALTHCHECK_TCP"]; ok {
		conf := []*tcp{}
		if err := json.Unmarshal([]byte(value), &conf); err != nil {
			l.Warnf("parse ENV_INPUT_HEALTHCHECK_TCP=%s failed, %s", value, err.Error())
		} else {
			ipt.TCP = conf
		}
	}

	if value, ok := envs["ENV_INPUT_HEALTHCHECK_HTTP"]; ok {
		conf := []*http{}
		if err := json.Unmarshal([]byte(value), &conf); err != nil {
			l.Warnf("parse ENV_INPUT_HEALTHCHECK_HTTP=%s failed, %s", value, err.Error())
		} else {
			ipt.HTTP = conf
		}
	}

	if value, ok := envs["ENV_INPUT_HEALTHCHECK_TAGS"]; ok {
		var tags map[string]string
		if err := json.Unmarshal([]byte(value), &tags); err != nil {
			l.Warnf("parse ENV_INPUT_HEALTHCHECK_TAGS=%s failed: %s", value, err.Error())
		} else {
			ipt.Tags = tags
		}
	}
}

func filterEmptyValues(list []string) []string {
	filterList := []string{}

	for _, v := range list {
		if v == "" {
			continue
		}
		filterList = append(filterList, v)
	}

	return filterList
}

func (ipt *Input) Terminate() {
	if ipt.semStop != nil {
		ipt.semStop.Close()
	}
}

func defaultInput() *Input {
	return &Input{
		semStop: cliutils.NewSem(),
		Tags:    make(map[string]string),
		feeder:  dkio.DefaultFeeder(),
		tagger:  datakit.DefaultGlobalTagger(),
	}
}

func init() { //nolint:gochecknoinits
	inputs.Add(inputName, func() inputs.Input {
		return defaultInput()
	})
}
