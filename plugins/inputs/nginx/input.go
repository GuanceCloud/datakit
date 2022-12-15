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
	"time"

	"gitlab.jiagouyun.com/cloudcare-tools/cliutils"
	"gitlab.jiagouyun.com/cloudcare-tools/cliutils/logger"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/config"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/goroutine"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/tailer"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io/point"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"
)

var _ inputs.ElectionInput = (*Input)(nil)

var (
	inputName   = `nginx`
	l           = logger.DefaultSLogger(inputName)
	minInterval = time.Second
	maxInterval = time.Second * 30
	sample      = `
[[inputs.nginx]]
	url = "http://localhost:80/server_status"
	# ##(optional) collection interval, default is 30s
	# interval = "30s"
	use_vts = false
	## Optional TLS Config
	# tls_ca = "/xxx/ca.pem"
	# tls_cert = "/xxx/cert.cer"
	# tls_key = "/xxx/key.key"
	## Use TLS but skip chain & host verification
	insecure_skip_verify = false
	# HTTP response timeout (default: 5s)
	response_timeout = "20s"

    ## Set true to enable election
	election = true

	[inputs.nginx.log]
	#	files = ["/var/log/nginx/access.log","/var/log/nginx/error.log"]
	#	# grok pipeline script path
	#	pipeline = "nginx.p"
	[inputs.nginx.tags]
	# some_tag = "some_value"
	# more_tag = "some_other_value"
	# ...`

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

func (n *Input) ElectionEnabled() bool {
	return n.Election
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
func (n *Input) LogExamples() map[string]map[string]string {
	return map[string]map[string]string{
		"nginx": {
			"Nginx error log1": `2021/04/21 09:24:04 [alert] 7#7: *168 write() to "/var/log/nginx/access.log" failed (28: No space left on device) while logging request, client: 120.204.196.129, server: localhost, request: "GET / HTTP/1.1", host: "47.98.103.73"`,
			"Nginx error log2": `2021/04/29 16:24:38 [emerg] 50102#0: unexpected ";" in /usr/local/etc/nginx/nginx.conf:23`,
			"Nginx access log": `127.0.0.1 - - [24/Mar/2021:13:54:19 +0800] "GET /basic_status HTTP/1.1" 200 97 "-" "Mozilla/5.0 (Macintosh; Intel Mac OS X 11_1_0) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/89.0.4389.72 Safari/537.36"`,
		},
	}
}

func (n *Input) GetPipeline() []*tailer.Option {
	return []*tailer.Option{
		{
			Source:  inputName,
			Service: inputName,
			Pipeline: func() string {
				if n.Log != nil {
					return n.Log.Pipeline
				}
				return ""
			}(),
		},
	}
}

func (n *Input) RunPipeline() {
	if n.Log == nil || len(n.Log.Files) == 0 {
		return
	}

	if n.Log.Pipeline == "" {
		n.Log.Pipeline = "nginx.p" // use default
	}

	opt := &tailer.Option{
		Source:     inputName,
		Service:    inputName,
		Pipeline:   n.Log.Pipeline,
		GlobalTags: n.Tags,
		Done:       n.semStop.Wait(),
	}

	var err error
	n.tail, err = tailer.NewTailer(n.Log.Files, opt)
	if err != nil {
		l.Error(err)
		return
	}

	g := goroutine.NewGroup(goroutine.Option{Name: "inputs_nginx"})
	g.Go(func(ctx context.Context) error {
		n.tail.Start()
		return nil
	})
}

func (n *Input) Run() {
	l = logger.SLogger(inputName)
	l.Info("nginx start")

	tick := time.NewTicker(n.Interval.Duration)
	defer tick.Stop()

	for {
		if n.pause {
			l.Debugf("not leader, skipped")
		} else {
			mpts, err := n.Collect()
			if err != nil {
				l.Errorf("Collect failed: %v", err)
			} else {
				for category, points := range mpts {
					if len(points) > 0 {
						if err := io.Feed(inputName, category, points,
							&io.Option{CollectCost: time.Since(n.start)}); err != nil {
							l.Errorf(err.Error())
							io.FeedLastError(inputName, err.Error())
						} else {
							n.collectCache = n.collectCache[:0]
						}
					}
				}
			}
		}

		select {
		case <-datakit.Exit.Wait():
			n.exit()
			l.Info("nginx exit")
			return
		case <-n.semStop.Wait():
			n.exit()
			l.Info("nginx return")
			return
		case <-tick.C:
		case n.pause = <-n.pauseCh:
		}
	}
}

func (n *Input) exit() {
	if n.tail != nil {
		n.tail.Close()
		l.Info("nginx log exit")
	}
}

func (n *Input) Terminate() {
	if n.semStop != nil {
		n.semStop.Close()
	}
}

func (n *Input) getMetric() {
	n.start = time.Now()
	if n.UseVts {
		n.getVTSMetric()
	} else {
		n.getStubStatusModuleMetric()
	}
}

func (n *Input) createHTTPClient() (*http.Client, error) {
	tlsCfg, err := n.ClientConfig.TLSConfig()
	if err != nil {
		return nil, err
	}

	if n.ResponseTimeout.Duration < time.Second {
		n.ResponseTimeout.Duration = time.Second * 5
	}

	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: tlsCfg,
		},
		Timeout: n.ResponseTimeout.Duration,
	}

	return client, nil
}

func (*Input) AvailableArchs() []string {
	return datakit.AllOSWithElection
}

func (n *Input) SampleMeasurement() []inputs.Measurement {
	return []inputs.Measurement{
		&NginxMeasurement{},
		&ServerZoneMeasurement{},
		&UpstreamZoneMeasurement{},
		&CacheZoneMeasurement{},
	}
}

func (n *Input) Pause() error {
	tick := time.NewTicker(inputs.ElectionPauseTimeout)
	defer tick.Stop()
	select {
	case n.pauseCh <- true:
		return nil
	case <-tick.C:
		return fmt.Errorf("pause %s failed", inputName)
	}
}

func (n *Input) Resume() error {
	tick := time.NewTicker(inputs.ElectionResumeTimeout)
	defer tick.Stop()
	select {
	case n.pauseCh <- false:
		return nil
	case <-tick.C:
		return fmt.Errorf("resume %s failed", inputName)
	}
}

func (n *Input) setup() error {
	if n.Interval.Duration == 0 {
		n.Interval.Duration = config.ProtectedInterval(minInterval, maxInterval, n.Interval.Duration)
	}

	if n.client == nil {
		client, err := n.createHTTPClient()
		if err != nil {
			fmt.Printf("[error] nginx init client err:%s\n", err.Error())
			return err
		}
		n.client = client
	}

	return nil
}

func (n *Input) Collect() (map[string][]*point.Point, error) {
	if err := n.setup(); err != nil {
		return map[string][]*point.Point{}, err
	}

	n.getMetric()

	if len(n.collectCache) == 0 {
		return map[string][]*point.Point{}, fmt.Errorf("no points")
	}

	pts, err := inputs.GetPointsFromMeasurement(n.collectCache)
	if err != nil {
		return map[string][]*point.Point{}, err
	}

	mpts := make(map[string][]*point.Point)
	mpts[datakit.Metric] = pts

	return mpts, nil
}

func NewNginx() *Input {
	return &Input{
		Interval: datakit.Duration{Duration: time.Second * 10},
		pauseCh:  make(chan bool, inputs.ElectionPauseChannelLength),
		Election: true,

		semStop: cliutils.NewSem(),
	}
}

func init() { //nolint:gochecknoinits
	inputs.Add(inputName, func() inputs.Input {
		return NewNginx()
	})
}
