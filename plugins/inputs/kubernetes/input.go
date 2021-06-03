package kubernetes

import (
	"net/http"
	"time"

	"gitlab.jiagouyun.com/cloudcare-tools/cliutils/logger"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"
)

const (
	inputName = "kubernetes"

	sampleCfg = `
[inputs.kubernetes]
  # URL for the Kubernetes API, tls 6443 or proxy 8080
  kube_apiserver_url = "http://127.0.0.1:8080/metrics"

  ## Optional TLS Config
  # tls_ca = "/path/to/ca.pem"
  # tls_cert = "/path/to/cert.pem"
  # tls_key = "/path/to/key.pem"
  ## Use TLS but skip chain & host verification

  [inputs.kubernetes.tags]
    # tags1 = "value1"
`
)

var l = logger.DefaultSLogger(inputName)

func init() {
	inputs.Add(inputName, func() inputs.Input {
		return newInput()
	})
}

type Input struct {
	KubeAPIServerURL string            `toml:"kube_apiserver_url"`
	Tags             map[string]string `toml:"tags"`
	ClientConfig                       // tls config
	httpCli          *http.Client
}

func newInput() *Input {
	return &Input{
		Tags: make(map[string]string),
		httpCli: &http.Client{
			Timeout: 5 * time.Second,
		},
	}
}

func (*Input) SampleConfig() string {
	return sampleCfg
}

func (*Input) Catalog() string {
	return "kubernetes"
}

func (*Input) PipelineConfig() map[string]string {
	return nil
}

/*
func (*Input) AvailableArchs() []string {
	return []string{datakit.OSLinux}
} */

func (n *Input) SampleMeasurement() []inputs.Measurement {
	return nil
}

func (this *Input) Run() {
	// FIXME
	// 现在如果配置 tls 错误，将不使用 tls，而不是持续报错
	// 这是不严谨的

	tlsconfig, err := this.TLSConfig()
	if err != nil {
		l.Warn(err)
	}

	if tlsconfig != nil {
		this.httpCli.Transport = &http.Transport{
			TLSClientConfig: tlsconfig,
		}
	}

	tick := time.NewTicker(time.Second * 5)
	defer tick.Stop()

	for {
		select {
		case <-datakit.Exit.Wait():
			return

		case <-tick.C:
			this.gather()
		}
	}
}

func (this *Input) gather() {
	startTime := time.Now()

	pt, err := this.gatherMetrics()
	if err != nil {
		l.Error(err)
		return
	}

	cost := time.Since(startTime)

	if err := io.Feed(inputName, datakit.Metric, []*io.Point{pt}, &io.Option{CollectCost: cost}); err != nil {
		l.Error(err)
	}
}

func (this *Input) gatherMetrics() (*io.Point, error) {
	resp, err := http.Get(this.KubeAPIServerURL)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	fields, err := promTextToMetrics(resp.Body)
	if err != nil {
		return nil, err
	}

	return io.MakePoint(inputName, this.Tags, fields, time.Now())
}
