package nginx

import (
	"net/http"
	"path/filepath"
	"time"

	"gitlab.jiagouyun.com/cloudcare-tools/cliutils/logger"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"
)

var (
	inputName = `nginx`
	l         = logger.DefaultSLogger(inputName)
	sample    = `
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
	#	files = []
	#[inputs.nginx.log.option]
	#	ignore = [""]
	#	# your logging source, if it's empty, use 'default'
	#	source = "nginx"
	#	# add service tag, if it's empty, use $source.
	#	service = ""
	#	# grok pipeline script path
	#	pipeline = "nginx.p"
	#	# optional status:
	#	#   "emerg","alert","critical","error","warning","info","debug","OK"
	#	ignore_status = []
	#	# read file from beginning
	
	#	# optional encodings:
	#	#    "utf-8", "utf-16le", "utf-16le", "gbk", "gb18030" or ""
	#	character_encoding = ""
	#	# The pattern should be a regexp. Note the use of '''this regexp'''
	#	# regexp link: https://golang.org/pkg/regexp/syntax/#hdr-Syntax
	#	match = '''^\S'''
`
	pipelineCfg = `
add_pattern("date2", "%{YEAR}[./]%{MONTHNUM}[./]%{MONTHDAY} %{TIME}")

# access log
grok(_, "%{IPORHOST:client_ip} %{NOTSPACE:http_ident} %{NOTSPACE:http_auth} \\[%{HTTPDATE:time}\\] \"%{DATA:http_method} %{GREEDYDATA:http_url} HTTP/%{NUMBER:http_version}\" %{INT:status_code} %{INT:bytes}")

# access log
add_pattern("access_common", "%{IPORHOST:client_ip} %{NOTSPACE:http_ident} %{NOTSPACE:http_auth} \\[%{HTTPDATE:time}\\] \"%{DATA:http_method} %{GREEDYDATA:http_url} HTTP/%{NUMBER:http_version}\" %{INT:status_code} %{INT:bytes}")
grok(_, '%{access_common} "%{NOTSPACE:referrer}" "%{GREEDYDATA:agent}')
user_agent(agent)

# error log
grok(_, "%{date2:time} \\[%{LOGLEVEL:status}\\] %{GREEDYDATA:msg}, client: %{IPORHOST:client_ip}, server: %{IPORHOST:server}, request: \"%{DATA:http_method} %{GREEDYDATA:http_url} HTTP/%{NUMBER:http_version}\", (upstream: \"%{GREEDYDATA:upstream}\", )?host: \"%{IPORHOST:host}\"")

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

func (n *Input) Run() {
	l.Info("nginx start")

	if n.Log != nil {
		go func() {
			if err := n.Log.Init(); err != nil {
				l.Errorf("nginx init tailf err:%s", err.Error())
				return
			}
			if n.Log.Option.Pipeline != "" {
				n.Log.Option.Pipeline = filepath.Join(datakit.PipelineDir, n.Log.Option.Pipeline)
			} else {
				n.Log.Option.Pipeline = filepath.Join(datakit.PipelineDir, "nginx.p")
			}
			n.Log.Run()
		}()
	}

	client, err := n.createHttpClient()
	if err != nil {
		l.Errorf("[error] nginx init client err:%s", err.Error())
		return
	}
	n.client = client
	if n.Interval.Duration == 0 {
		n.Interval.Duration = time.Second * 30
	}

	tick := time.NewTicker(n.Interval.Duration)
	defer tick.Stop()
	cleanCacheTick := time.NewTicker(time.Second * 5)
	defer cleanCacheTick.Stop()

	for {
		select {
		case <-tick.C:
			n.getMetric()
		case <-cleanCacheTick.C:
			if len(n.collectCache) > 0 {
				inputs.FeedMeasurement(inputName, io.Metric, n.collectCache, &io.Option{CollectCost: time.Since(n.start)})
				n.collectCache = n.collectCache[:]
			}
		case <-datakit.Exit.Wait():
			if n.Log != nil {
				n.Log.Close()
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

func (i *Input) AvailableArchs() []string {
	return datakit.UnknownArch
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
		s := &Input{}
		return s
	})
}
