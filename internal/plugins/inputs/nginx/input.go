// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

// Package nginx collects NGINX metrics.
package nginx

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/GuanceCloud/cliutils"
	"github.com/GuanceCloud/cliutils/logger"
	"github.com/GuanceCloud/cliutils/point"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/config"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/goroutine"
	dkio "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/io"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/metrics"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/tailer"
)

var _ inputs.ElectionInput = (*Input)(nil)

var (
	inputName            = `nginx`
	customObjectFeedName = inputName + "/CO"

	l           = logger.DefaultSLogger(inputName)
	minInterval = time.Second * 10
	maxInterval = time.Second * 30
	sample      = `
[[inputs.nginx]]
  ## Nginx status URL.
  ## (Default) If not use with VTS, the formula is like this: "http://localhost:80/basic_status".
  ## If using with VTS, the formula is like this: "http://localhost:80/status/format/json".
  url = "http://localhost:80/basic_status"
  # If using Nginx Plus, this formula is like this: "http://localhost:8080/api/<api_version>".
  # Note: Nginx Plus not support VTS and should be used with http_stub_status_module (Default)
  # plus_api_url = "http://localhost:8080/api/9"

  ## Optional Can set ports as [<form>,<to>], Datakit will collect all ports.
  # ports = [80,80]

  ## Optional collection interval, default is 10s
  # interval = "30s"
  use_vts = false
  use_plus_api = false
  ## Optional TLS Config
  # tls_ca = "/xxx/ca.pem"
  # tls_cert = "/xxx/cert.cer"
  # tls_key = "/xxx/key.key"
  ## Use TLS but skip chain & host verification
  insecure_skip_verify = false
  ## HTTP response timeout (default: 5s)
  response_timeout = "20s"

  ## Set true to enable election
  election = true

# [inputs.nginx.log]
  # files = ["/var/log/nginx/access.log","/var/log/nginx/error.log"]
  ## grok pipeline script path
  # pipeline = "nginx.p"
# [inputs.nginx.tags]
  # some_tag = "some_value"
  # more_tag = "some_other_value"
`

	//nolint:lll
	pipelineCfg = `
add_pattern("date2", "%{YEAR}[./]%{MONTHNUM}[./]%{MONTHDAY} %{TIME}")

# access log
grok(_, "%{NOTSPACE:client_ip} %{NOTSPACE:http_ident} %{NOTSPACE:http_auth} \\[%{HTTPDATE:time}\\] \"%{DATA:http_method} %{GREEDYDATA:http_url} HTTP/%{NUMBER:http_version}\" %{INT:status_code} %{INT:bytes}")

# access log
add_pattern("access_common", "%{NOTSPACE:client_ip} %{NOTSPACE:http_ident} %{NOTSPACE:http_auth} \\[%{HTTPDATE:time}\\] \"%{DATA:http_method} %{GREEDYDATA:http_url} HTTP/%{NUMBER:http_version}\" %{INT:status_code} %{INT:bytes}")
grok(_, '%{access_common} "%{NOTSPACE:referrer}" "%{GREEDYDATA:agent}"')
user_agent(agent)

# error log
grok(_, "%{date2:time} \\[%{LOGLEVEL:status}\\] %{GREEDYDATA:msg}, client: %{NOTSPACE:client_ip}, server: %{NOTSPACE:server}, request: \"%{DATA:http_method} %{GREEDYDATA:http_url} HTTP/%{NUMBER:http_version}\", (upstream: \"%{GREEDYDATA:upstream}\", )?host: \"%{NOTSPACE:ip_or_host}\"")
grok(_, "%{date2:time} \\[%{LOGLEVEL:status}\\] %{GREEDYDATA:msg}, client: %{NOTSPACE:client_ip}, server: %{NOTSPACE:server}, request: \"%{GREEDYDATA:http_method} %{GREEDYDATA:http_url} HTTP/%{NUMBER:http_version}\", host: \"%{NOTSPACE:ip_or_host}\"")
grok(_,"%{date2:time} \\[%{LOGLEVEL:status}\\] %{GREEDYDATA:msg}")

group_in(status, ["warn", "notice"], "warning")
group_in(status, ["error", "crit", "alert", "emerg"], "error")

cast(status_code, "int")
cast(bytes, "int")

group_between(status_code, [200,299], "OK", status)
group_between(status_code, [300,399], "notice", status)
group_between(status_code, [400,499], "warning", status)
group_between(status_code, [500,599], "error", status)


nullif(http_ident, "-")
nullif(http_auth, "-")
nullif(upstream, "")
default_time(time)
`
)

func (ipt *Input) ElectionEnabled() bool {
	return ipt.Election
}

func (*Input) SampleConfig() string {
	return sample
}

func (*Input) Catalog() string {
	return inputName
}

func (*Input) PipelineConfig() map[string]string {
	pipelineMap := map[string]string{
		"nginx": pipelineCfg,
	}
	return pipelineMap
}

//nolint:lll
func (ipt *Input) LogExamples() map[string]map[string]string {
	return map[string]map[string]string{
		"nginx": {
			"Nginx error log1": `2021/04/21 09:24:04 [alert] 7#7: *168 write() to "/var/log/nginx/access.log" failed (28: No space left on device) while logging request, client: 120.204.196.129, server: localhost, request: "GET / HTTP/1.1", host: "47.98.103.73"`,
			"Nginx error log2": `2021/04/29 16:24:38 [emerg] 50102#0: unexpected ";" in /usr/local/etc/nginx/nginx.conf:23`,
			"Nginx access log": `127.0.0.1 - - [24/Mar/2021:13:54:19 +0800] "GET /basic_status HTTP/1.1" 200 97 "-" "Mozilla/5.0 (Macintosh; Intel Mac OS X 11_1_0) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/89.0.4389.72 Safari/537.36"`,
		},
	}
}

func (ipt *Input) GetPipeline() []tailer.Option {
	opts := []tailer.Option{
		tailer.WithSource(inputName),
		tailer.WithService(inputName),
	}
	if ipt.Log != nil {
		opts = append(opts, tailer.WithPipeline(ipt.Log.Pipeline))
	}
	return opts
}

func (ipt *Input) RunPipeline() {
	if ipt.Log == nil || len(ipt.Log.Files) == 0 {
		return
	}

	opts := []tailer.Option{
		tailer.WithSource(inputName),
		tailer.WithService(inputName),
		tailer.WithPipeline(ipt.Log.Pipeline),
		tailer.WithMaxMultilineLength(int64(float64(config.Cfg.Dataway.MaxRawBodySize) * 0.8)),
		tailer.WithGlobalTags(inputs.MergeTags(ipt.Tagger.HostTags(), ipt.Tags, "")),
		tailer.EnableDebugFields(config.Cfg.EnableDebugFields),
	}

	var err error
	ipt.tail, err = tailer.NewTailer(ipt.Log.Files, opts...)
	if err != nil {
		l.Error(err)
		return
	}

	g := goroutine.NewGroup(goroutine.Option{Name: "inputs_nginx"})
	g.Go(func(ctx context.Context) error {
		ipt.tail.Start()
		return nil
	})
}

func (ipt *Input) Run() {
	if err := ipt.setup(); err != nil {
		ipt.FeedErrUpMetric()
		ipt.FeedCoByErr(err)
		return
	}

	tick := time.NewTicker(ipt.Interval)
	defer tick.Stop()

	if ipt.Election {
		ipt.mergedTags = inputs.MergeTags(ipt.Tagger.ElectionTags(), ipt.Tags, ipt.URL)
	} else {
		ipt.mergedTags = inputs.MergeTags(ipt.Tagger.HostTags(), ipt.Tags, ipt.URL)
	}

	l.Infof("merged tags: %+#v", ipt.mergedTags)

	lastTS := time.Now()
	for {
		if ipt.pause {
			l.Debugf("not leader, skipped")
		} else {
			ipt.setUpState()
			ipt.FeedCoPts()

			ipt.collect(lastTS.UnixNano())
			if len(ipt.collectCache) > 0 {
				if err := ipt.feeder.FeedV2(point.Metric, ipt.collectCache,
					dkio.WithCollectCost(time.Since(lastTS)),
					dkio.WithElection(ipt.Election),
					dkio.WithInputName(inputName),
				); err != nil {
					ipt.feeder.FeedLastError(err.Error(),
						metrics.WithLastErrorInput(inputName),
					)
					l.Errorf("feed : %s", err)
				} else {
					ipt.collectCache = ipt.collectCache[:0]
				}
			} else {
				l.Warn("collect nil points")

				ipt.setErrUpState()
			}

			ipt.FeedUpMetric(lastTS)
		}

		select {
		case <-datakit.Exit.Wait():
			ipt.exit()
			l.Info("nginx exit")
			return
		case <-ipt.semStop.Wait():
			ipt.exit()
			l.Info("nginx return")
			return
		case tt := <-tick.C:
			nextts := inputs.AlignTimeMillSec(tt, lastTS.UnixMilli(), ipt.Interval.Milliseconds())
			lastTS = time.UnixMilli(nextts)
		case ipt.pause = <-ipt.pauseCh:
		}
	}
}

func (ipt *Input) exit() {
	if ipt.tail != nil {
		ipt.tail.Close()
		l.Info("nginx log exit")
	}
}

func (ipt *Input) Terminate() {
	if ipt.semStop != nil {
		ipt.semStop.Close()
	}
}

func (ipt *Input) getMetric(alignTS int64) {
	for i := ipt.Ports[0]; i <= ipt.Ports[1]; i++ {
		if ipt.UsePlusAPI { //nolint
			ipt.getPlusMetric(alignTS)
			ipt.getStubStatusModuleMetric(i, alignTS)
		} else if ipt.UseVts {
			ipt.getVTSMetric(i, alignTS)
		} else {
			ipt.getStubStatusModuleMetric(i, alignTS)
		}
	}
}

func (ipt *Input) createHTTPClient() (*http.Client, error) {
	tlsCfg, err := ipt.ClientConfig.TLSConfig()
	if err != nil {
		return nil, err
	}

	if ipt.ResponseTimeout < time.Second {
		ipt.ResponseTimeout = time.Second * 5
	}

	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: tlsCfg,
		},
		Timeout: ipt.ResponseTimeout,
	}

	return client, nil
}

func (*Input) AvailableArchs() []string {
	return datakit.AllOSWithElection
}

func (ipt *Input) SampleMeasurement() []inputs.Measurement {
	return []inputs.Measurement{
		&NginxMeasurement{},
		&ServerZoneMeasurement{},
		&UpstreamZoneMeasurement{},
		&CacheZoneMeasurement{},
		&LocationZoneMeasurement{},
		&customerObjectMeasurement{},
	}
}

func (ipt *Input) Pause() error {
	tick := time.NewTicker(inputs.ElectionPauseTimeout)
	defer tick.Stop()
	select {
	case ipt.pauseCh <- true:
		return nil
	case <-tick.C:
		return fmt.Errorf("pause %s failed", inputName)
	}
}

func (ipt *Input) Resume() error {
	tick := time.NewTicker(inputs.ElectionResumeTimeout)
	defer tick.Stop()
	select {
	case ipt.pauseCh <- false:
		return nil
	case <-tick.C:
		return fmt.Errorf("resume %s failed", inputName)
	}
}

func (ipt *Input) setup() error {
	l = logger.SLogger(inputName)
	l.Info("nginx start")

	if err := ipt.checkPortsAndURL(); err != nil {
		l.Errorf("checkPortsAndURL : %v", err)
		return err
	}

	ipt.Interval = config.ProtectedInterval(minInterval, maxInterval, ipt.Interval)

	client, err := ipt.createHTTPClient()
	if err != nil {
		l.Errorf("[error] nginx init client err:%s\n", err.Error())
		return err
	}
	ipt.client = client

	return nil
}

func (ipt *Input) checkPortsAndURL() error {
	u, err := url.ParseRequestURI(ipt.URL)
	if err != nil {
		return err
	}
	scheme := u.Scheme
	hostname := u.Hostname()
	port := u.Port()
	if port == "" {
		port = "80"
	}

	if ipt.Ports[0] == 0 && ipt.Ports[1] == 0 {
		p, err := strconv.Atoi(port)
		if err != nil {
			return err
		}
		ipt.Ports = [2]int{p, p}
	}
	if ipt.Ports[0] < 1 || ipt.Ports[0] > 65535 || ipt.Ports[1] < 1 || ipt.Ports[1] > 65535 {
		return fmt.Errorf("ports error, now is: %v", ipt.Ports)
	}
	if ipt.Ports[0] > ipt.Ports[1] {
		return fmt.Errorf("ports error, now is: %v", ipt.Ports)
	}

	ipt.host = scheme + "://" + hostname
	ipt.path = u.Path

	return nil
}

func (ipt *Input) collect(alignTS int64) {
	ipt.getMetric(alignTS)
}

func defaultInput() *Input {
	return &Input{
		Interval: time.Second * 10,
		pauseCh:  make(chan bool, inputs.ElectionPauseChannelLength),
		Election: true,
		feeder:   dkio.DefaultFeeder(),
		semStop:  cliutils.NewSem(),
		Tagger:   datakit.DefaultGlobalTagger(),
	}
}

func init() { //nolint:gochecknoinits
	inputs.Add(inputName, func() inputs.Input {
		return defaultInput()
	})
}
