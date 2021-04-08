package nginx

import (
	"gitlab.jiagouyun.com/cloudcare-tools/cliutils/logger"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"
	"net/http"
	"time"
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
`
)

func (_ *Input) SampleConfig() string {
	return sample
}

func (_ *Input) Catalog() string {
	return inputName
}

func (n *Input) Run() {
	l.Info("nginx start")
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
