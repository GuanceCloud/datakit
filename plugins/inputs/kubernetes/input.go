// Package kubernetes collects k8s metrics/objects.
package kubernetes

import (
	"fmt"
	"time"

	"gitlab.jiagouyun.com/cloudcare-tools/cliutils"
	"gitlab.jiagouyun.com/cloudcare-tools/cliutils/logger"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/config"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/net"
	timex "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/time"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"
)

var (
	_ inputs.ElectionInput = (*Input)(nil)
	_ inputs.ReadEnv       = (*Input)(nil)
)

const (
	inputName = "kubernetes"
	catalog   = "kubernetes"

	minMetricInterval     = time.Minute * 1
	defaultObjectInterval = time.Minute * 5
)

type Input struct {
	URL      string `toml:"url"`
	Interval string `toml:"interval"`

	BearerToken       string `toml:"bearer_token"`
	BearerTokenString string `toml:"bearer_token_string"`
	TLSCA             string `toml:"tls_ca"`
	TLSCert           string `toml:"tls_cert"`
	TLSKey            string `toml:"tls_key"`

	DepercatedTimeout string `toml:"timeout"`

	InsecureSkipVerify bool `toml:"insecure_skip_verify"`

	Tags map[string]string `toml:"tags"`

	KubeAPIServerURLDeprecated string `toml:"kube_apiserver_url,omitempty"`

	client       *client
	resourceList []resource
	exporterList []exporter

	chPause chan bool
	pause   bool

	semStop *cliutils.Sem // start stop signal
}

var l = logger.DefaultSLogger("kubernetes")

func (ipt *Input) Run() {
	l = logger.SLogger(inputName)

	if ipt.setup() {
		return
	}

	dur, _ := timex.ParseDuration(ipt.Interval)
	if dur < minMetricInterval {
		l.Debug("use default metric interval: 60s")
		dur = minMetricInterval
	}

	metricTick := time.NewTicker(dur)
	defer metricTick.Stop()

	objectTick := time.NewTicker(defaultObjectInterval)
	defer objectTick.Stop()

	if !ipt.pause {
		l.Info("first collection")
		ipt.gatherMetric()
		ipt.execExport()
		ipt.gatherObject()
	}

	for {
		select {
		case <-metricTick.C:
			if ipt.pause {
				l.Debugf("not leader, skipped (metrics)")
				continue
			}
			ipt.gatherMetric()

		case <-objectTick.C:
			if ipt.pause {
				l.Debugf("not leader, skipped (object)")
				continue
			}
			ipt.gatherObject()
			ipt.execExport()

		case <-datakit.Exit.Wait():
			ipt.exit()
			l.Info("kubernetes exit")
			return

		case <-ipt.semStop.Wait():
			ipt.exit()
			l.Info("kubernetes return")
			return

		case ipt.pause = <-ipt.chPause:
			// nil
		}
	}
}

func (ipt *Input) exit() {
	ipt.Stop()
}

func (ipt *Input) Terminate() {
	if ipt.semStop != nil {
		ipt.semStop.Close()
	}
}

func (ipt *Input) Stop() {
	for _, exporter := range ipt.exporterList {
		exporter.Stop()
	}
}

func (ipt *Input) setup() bool {
	for {
		select {
		case <-datakit.Exit.Wait():
			l.Info("exit")
			return true
		default:
			// nil
		}
		time.Sleep(time.Second)

		if err := ipt.buildClient(); err != nil {
			l.Error(err)
			continue
		}
		ipt.buildExporters()

		break
	}

	ipt.buildResources()

	return false
}

func (ipt *Input) buildClient() error {
	var cli *client
	var err error

	if ipt.URL == "" {
		return fmt.Errorf("invalid k8s url, cannot be empty")
	}

	if ipt.BearerToken != "" {
		cli, err = newClientFromBearerToken(ipt.URL, ipt.BearerToken)
		if err != nil {
			return err
		}
		l.Debug("use bearer_token")
		goto end
	}

	if ipt.BearerTokenString != "" {
		cli, err = newClientFromBearerTokenString(ipt.URL, ipt.BearerTokenString)
		if err != nil {
			return err
		}
		l.Debug("use bearer_token string")
		goto end
	}

	if ipt.TLSCA != "" {
		tlsconfig := net.TLSClientConfig{
			CaCerts:            []string{ipt.TLSCA},
			Cert:               ipt.TLSCert,
			CertKey:            ipt.TLSKey,
			InsecureSkipVerify: ipt.InsecureSkipVerify,
		}

		cli, err = newClientFromTLS(ipt.URL, &tlsconfig)
		if err != nil {
			return err
		}
		l.Debug("use tls config")
		goto end
	}

	l.Debug("not found https authority, token/tokenString/tls are empty")
end:
	if cli != nil {
		ipt.client = cli
		return nil
	}

	return fmt.Errorf("failed of build client")
}

func (ipt *Input) buildResources() {
	ipt.resourceList = []resource{
		// metric
		&kubernetesMetric{client: ipt.client, tags: ipt.Tags},
		// object
		&cluster{client: ipt.client, tags: ipt.Tags},
		&deployment{client: ipt.client, tags: ipt.Tags},
		&replicaSet{client: ipt.client, tags: ipt.Tags},
		&service{client: ipt.client, tags: ipt.Tags},
		&node{client: ipt.client, tags: ipt.Tags},
		&job{client: ipt.client, tags: ipt.Tags},
		&cronJob{client: ipt.client, tags: ipt.Tags},
		// &pod{client: ipt.client, tags: ipt.Tags},
	}
}

func (ipt *Input) buildExporters() {
	ipt.exporterList = []exporter{&pod{client: ipt.client, tags: ipt.Tags}}
}

func (ipt *Input) execExport() {
	for _, exporter := range ipt.exporterList {
		exporter.Export()
	}
}

func (ipt *Input) gatherObject() {
	if len(ipt.resourceList) < 2 {
		return
	}
	for _, resource := range ipt.resourceList[1:] {
		resource.Gather()
	}
}

func (ipt *Input) gatherMetric() {
	if len(ipt.resourceList) == 0 {
		return
	}
	ipt.resourceList[0].Gather()
}

func (ipt *Input) Pause() error {
	tick := time.NewTicker(inputs.ElectionPauseTimeout)
	select {
	case ipt.chPause <- true:
		return nil
	case <-tick.C:
		return fmt.Errorf("pause %s failed", inputName)
	}
}

func (ipt *Input) Resume() error {
	tick := time.NewTicker(inputs.ElectionResumeTimeout)
	select {
	case ipt.chPause <- false:
		return nil
	case <-tick.C:
		return fmt.Errorf("resume %s failed", inputName)
	}
}

func (*Input) Catalog() string { return catalog }

func (*Input) SampleConfig() string { return sampleCfg }

func (*Input) AvailableArchs() []string { return datakit.AllArch }

func (*Input) SampleMeasurement() []inputs.Measurement {
	res := []inputs.Measurement{
		&kubernetesMetric{},
	}
	for _, resource := range resourceList {
		res = append(res, resource)
	}

	return res
}

// ReadEnv support envs：
//   ENV_INPUT_K8S_TAGS : "a=b,c=d"
func (ipt *Input) ReadEnv(envs map[string]string) {
	if tagsStr, ok := envs["ENV_INPUT_K8S_TAGS"]; ok {
		tags := config.ParseGlobalTags(tagsStr)
		for k, v := range tags {
			ipt.Tags[k] = v
		}
	}
}

func init() { //nolint:gochecknoinits
	inputs.Add(inputName, func() inputs.Input {
		return &Input{
			Tags:    make(map[string]string),
			chPause: make(chan bool, inputs.ElectionPauseChannelLength),

			semStop: cliutils.NewSem(),
		}
	})
}
