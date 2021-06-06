package nginx

import (
	"net/http"
	"time"

	"gitlab.jiagouyun.com/cloudcare-tools/cliutils/logger"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"
)

var (
	inputName   = `nginx`
	l           = logger.DefaultSLogger(inputName)
	minInterval = time.Second
	maxInterval = time.Second * 30
	sample      = `
[[inputs.nginx]]
	url = "http://localhost/server_status"
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

	[inputs.nginx.log]
	#	files = ["/var/log/nginx/access.log","/var/log/nginx/error.log"]
	#	# grok pipeline script path
	#	pipeline = "nginx.p"
	[inputs.nginx.tags]
	# some_tag = "some_value"
	# more_tag = "some_other_value"
	# ...`

	pipelineCfg = `
add_pattern("date2", "%{YEAR}[./]%{MONTHNUM}[./]%{MONTHDAY} %{TIME}")

# access log
grok(_, "%{IPORHOST:client_ip} %{NOTSPACE:http_ident} %{NOTSPACE:http_auth} \\[%{HTTPDATE:time}\\] \"%{DATA:http_method} %{GREEDYDATA:http_url} HTTP/%{NUMBER:http_version}\" %{INT:status_code} %{INT:bytes}")

# access log
add_pattern("access_common", "%{IPORHOST:client_ip} %{NOTSPACE:http_ident} %{NOTSPACE:http_auth} \\[%{HTTPDATE:time}\\] \"%{DATA:http_method} %{GREEDYDATA:http_url} HTTP/%{NUMBER:http_version}\" %{INT:status_code} %{INT:bytes}")
grok(_, '%{access_common} "%{NOTSPACE:referrer}" "%{GREEDYDATA:agent}')
user_agent(agent)

# error log
grok(_, "%{date2:time} \\[%{LOGLEVEL:status}\\] %{GREEDYDATA:msg}, client: %{IPORHOST:client_ip}, server: %{IPORHOST:server}, request: \"%{DATA:http_method} %{GREEDYDATA:http_url} HTTP/%{NUMBER:http_version}\", (upstream: \"%{GREEDYDATA:upstream}\", )?host: \"%{IPORHOST:ip_or_host}\"")
grok(_, "%{date2:time} \\[%{LOGLEVEL:status}\\] %{GREEDYDATA:msg}, client: %{IPORHOST:client_ip}, server: %{IPORHOST:server}, request: \"%{GREEDYDATA:http_method} %{GREEDYDATA:http_url} HTTP/%{NUMBER:http_version}\", host: \"%{IPORHOST:ip_or_host}\"")
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

func (_ *Input) SampleConfig() string {
	return sample
}

func (_ *Input) Catalog() string {
	return inputName
}

func (_ *Input) PipelineConfig() map[string]string {
	pipelineMap := map[string]string{
		"nginx": pipelineCfg,
	}
	return pipelineMap
}

// TODO
func (*Input) RunPipeline() {
}

func (n *Input) Run() {
	l = logger.SLogger(inputName)
	l.Info("nginx start")
	n.Interval.Duration = datakit.ProtectedInterval(minInterval, maxInterval, n.Interval.Duration)

	if n.Log != nil {
		go func() {
			inputs.JoinPipelinePath(n.Log, "nginx.p")
			n.Log.Source = nginx
			n.Log.Tags = map[string]string{}
			for k, v := range n.Tags {
				n.Log.Tags[k] = v
			}
			tail, err := inputs.NewTailer(n.Log)
			if err != nil {
				l.Errorf("init tailf err:%s", err.Error())
				return
			}
			n.tail = tail
			tail.Run()
		}()
	}

	client, err := n.createHttpClient()
	if err != nil {
		l.Errorf("[error] nginx init client err:%s", err.Error())
		return
	}
	n.client = client

	tick := time.NewTicker(n.Interval.Duration)
	defer tick.Stop()

	for {
		select {
		case <-tick.C:
			n.getMetric()
			if len(n.collectCache) > 0 {
				err := inputs.FeedMeasurement(inputName, datakit.Metric, n.collectCache, &io.Option{CollectCost: time.Since(n.start)})
				n.collectCache = n.collectCache[:0]
				if err != nil {
					n.lastErr = err
					l.Errorf(err.Error())
					continue
				}
			}
			if n.lastErr != nil {
				io.FeedLastError(inputName, n.lastErr.Error())
				n.lastErr = nil
			}
		case <-datakit.Exit.Wait():
			if n.tail != nil {
				n.tail.Close()
				l.Info("nginx log exit")
			}
			l.Info("nginx exit")
			return
		}
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

func (n *Input) createHttpClient() (*http.Client, error) {
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

func (_ *Input) AvailableArchs() []string {
	return datakit.AllArch
}

func (n *Input) SampleMeasurement() []inputs.Measurement {
	return []inputs.Measurement{
		&NginxMeasurement{},
		&ServerZoneMeasurement{},
		&UpstreamZoneMeasurement{},
		&CacheZoneMeasurement{},
	}
}

func init() {
	inputs.Add(inputName, func() inputs.Input {
		s := &Input{
			Interval: datakit.Duration{Duration: time.Second * 10},
		}
		return s
	})
}
