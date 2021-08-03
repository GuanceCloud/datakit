package kubernetes

import (
	"fmt"
	"time"

	"gitlab.jiagouyun.com/cloudcare-tools/cliutils/logger"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/net"
	timex "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/time"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"
)

var (
	inputName = "kubernetes"
	catalog   = "kubernetes"

	l = logger.DefaultSLogger("kubernetes")
)

const (
	defaultMetricInterval = time.Minute * 1
	defaultObjectInterval = time.Minute * 5
)

type Input struct {
	URL      string `toml:"url"`
	Interval string `toml:"interval"`

	BearerToken        string `toml:"bearer_token"`
	BearerTokenString  string `toml:"bearer_token_string"`
	TLSCA              string `toml:"tls_ca"`
	TLSCert            string `toml:"tls_cert"`
	TLSKey             string `toml:"tls_key"`
	InsecureSkipVerify bool   `toml:"insecure_skip_verify"`

	Tags map[string]string `toml:"tags"`

	DepercatedTimeout string `toml:"timeout"`

	client       *client
	resourceList []resource

	chPause chan bool
	pause   bool
}

func (this *Input) Run() {
	l = logger.SLogger(inputName)

	if this.setup() {
		return
	}

	metricTick := time.NewTicker(func() time.Duration {
		dur, err := timex.ParseDuration(this.Interval)
		if err != nil || dur < defaultMetricInterval {
			l.Debug("use default metric interval: 60s")
			return defaultMetricInterval
		}
		return dur
	}())
	defer metricTick.Stop()

	objectTick := time.NewTicker(defaultObjectInterval)
	defer objectTick.Stop()

	// 首先运行一次采集
	this.gatherMetric()
	this.gatherObject()

	for {
		select {
		case <-metricTick.C:
			if this.pause {
				l.Debugf("not leader, skipped")
				continue
			}
			this.gatherMetric()

		case <-objectTick.C:
			if this.pause {
				l.Debugf("not leader, skipped")
				continue
			}
			this.gatherObject()

			l.Debugf("exec discovery server")
			if err := this.collectPodsExporter(); err != nil {
				l.Errorf("%s discovery exec error %v", err)
				io.FeedLastError(inputName, err.Error())
			}

		case <-datakit.Exit.Wait():
			l.Info("kubernetes exit")
			return

		case this.pause = <-this.chPause:
			// nil
		}
	}
}

func (this *Input) setup() bool {
	for {
		select {
		case <-datakit.Exit.Wait():
			l.Info("exit")
			return true
		default:
			// nil
		}
		time.Sleep(time.Second)

		if err := this.buildClient(); err != nil {
			l.Error(err)
			continue
		}

		break
	}

	this.buildResources()

	return false
}

func (this *Input) buildClient() error {
	var cli *client
	var err error

	if this.URL == "" {
		return fmt.Errorf("invalid k8s url, cannot be empty")
	}

	if this.BearerToken != "" {
		cli, err = newClientFromBearerToken(this.URL, this.BearerToken)
		if err != nil {
			return err
		}
		l.Debug("use bearer_token")
		goto end
	}

	if this.BearerTokenString != "" {
		cli, err = newClientFromBearerTokenString(this.URL, this.BearerTokenString)
		if err != nil {
			return err
		}
		l.Debug("use bearer_token string")
		goto end
	}

	if this.TLSCA != "" {
		t := net.TlsClientConfig{
			CaCerts: func() []string {
				if this.TLSCA == "" {
					return nil
				}
				return []string{this.TLSCA}
			}(),
			Cert:               this.TLSCert,
			CertKey:            this.TLSKey,
			InsecureSkipVerify: this.InsecureSkipVerify,
		}

		cli, err = newClientFromTLS(this.URL, &t)
		if err != nil {
			return err
		}
		l.Debug("use tls config")
		goto end
	}
end:
	if cli != nil {
		this.client = cli
		return nil
	}

	return fmt.Errorf("failed of build client")
}

func (this *Input) buildResources() {
	this.resourceList = []resource{
		// metric
		&kubernetesMetric{client: this.client, tags: this.Tags},
		// object
		&cluster{client: this.client, tags: this.Tags},
		&deployment{client: this.client, tags: this.Tags},
		&replicaSet{client: this.client, tags: this.Tags},
		&service{client: this.client, tags: this.Tags},
		&node{client: this.client, tags: this.Tags},
		&job{client: this.client, tags: this.Tags},
		&cronJob{client: this.client, tags: this.Tags},
		// &pod{client: this.client, tags: this.Tags},
	}
}

func (this *Input) gatherObject() {
	if len(this.resourceList) < 2 {
		return
	}
	for _, resource := range this.resourceList[1:] {
		resource.Gather()
	}
}

func (this *Input) gatherMetric() {
	if len(this.resourceList) == 0 {
		return
	}
	this.resourceList[0].Gather()
}

func (this *Input) Pause() error {
	tick := time.NewTicker(time.Second * 5)
	select {
	case this.chPause <- true:
		return nil
	case <-tick.C:
		return fmt.Errorf("pause %s failed", inputName)
	}
}

func (this *Input) Resume() error {
	tick := time.NewTicker(time.Second * 5)
	select {
	case this.chPause <- false:
		return nil
	case <-tick.C:
		return fmt.Errorf("resume %s failed", inputName)
	}
}

func (*Input) Catalog() string { return catalog }

func (*Input) SampleConfig() string { return sampleCfg }

func (*Input) AvailableArchs() []string { return datakit.AllArch }

func (*Input) SampleMeasurement() []inputs.Measurement {
	var res = []inputs.Measurement{
		&kubernetesMetric{},
	}
	for _, resource := range resourceList {
		res = append(res, resource)
	}
	return res
}

func init() {
	inputs.Add(inputName, func() inputs.Input {
		return &Input{
			Tags:    make(map[string]string),
			chPause: make(chan bool, 1),
		}
	})
}
