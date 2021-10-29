// Package kubernetes collects k8s metrics/objects.
package kubernetes

import (
	"fmt"
	"time"

	"gitlab.jiagouyun.com/cloudcare-tools/cliutils/logger"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/net"
	timex "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/time"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"
)

var _ inputs.ElectionInput = (*Input)(nil)

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

	Tags         map[string]string `toml:"tags"`
	client       *client
	resourceList []resource
	exporterList []exporter

	chPause chan bool
	pause   bool
}

var l = logger.DefaultSLogger("kubernetes")

func (i *Input) Run() {
	l = logger.SLogger(inputName)

	if i.setup() {
		return
	}

	dur, _ := timex.ParseDuration(i.Interval)
	if dur < minMetricInterval {
		l.Debug("use default metric interval: 60s")
		dur = minMetricInterval
	}

	metricTick := time.NewTicker(dur)
	defer metricTick.Stop()

	objectTick := time.NewTicker(defaultObjectInterval)
	defer objectTick.Stop()

	if !i.pause {
		l.Info("first collection")
		i.gatherMetric()
		i.execExport()
		i.gatherObject()
	}

	for {
		select {
		case <-metricTick.C:
			if i.pause {
				l.Debugf("not leader, skipped (metrics)")
				continue
			}
			i.gatherMetric()

		case <-objectTick.C:
			if i.pause {
				l.Debugf("not leader, skipped (object)")
				continue
			}
			i.gatherObject()
			i.execExport()

		case <-datakit.Exit.Wait():
			i.Stop()
			l.Info("kubernetes exit")
			return

		case i.pause = <-i.chPause:
			// nil
		}
	}
}

func (i *Input) Stop() {
	for _, exporter := range i.exporterList {
		exporter.Stop()
	}
}

func (i *Input) setup() bool {
	for {
		select {
		case <-datakit.Exit.Wait():
			l.Info("exit")
			return true
		default:
			// nil
		}
		time.Sleep(time.Second)

		if err := i.buildClient(); err != nil {
			l.Error(err)
			continue
		}
		i.buildExporters()

		break
	}

	i.buildResources()

	return false
}

func (i *Input) buildClient() error {
	var cli *client
	var err error

	if i.URL == "" {
		return fmt.Errorf("invalid k8s url, cannot be empty")
	}

	if i.BearerToken != "" {
		cli, err = newClientFromBearerToken(i.URL, i.BearerToken)
		if err != nil {
			return err
		}
		l.Debug("use bearer_token")
		goto end
	}

	if i.BearerTokenString != "" {
		cli, err = newClientFromBearerTokenString(i.URL, i.BearerTokenString)
		if err != nil {
			return err
		}
		l.Debug("use bearer_token string")
		goto end
	}

	if i.TLSCA != "" {
		tlsconfig := net.TLSClientConfig{
			CaCerts:            []string{i.TLSCA},
			Cert:               i.TLSCert,
			CertKey:            i.TLSKey,
			InsecureSkipVerify: i.InsecureSkipVerify,
		}

		cli, err = newClientFromTLS(i.URL, &tlsconfig)
		if err != nil {
			return err
		}
		l.Debug("use tls config")
		goto end
	}

	l.Debug("not found https authority, token/tokenString/tls are empty")
end:
	if cli != nil {
		i.client = cli
		return nil
	}

	return fmt.Errorf("failed of build client")
}

func (i *Input) buildResources() {
	i.resourceList = []resource{
		// metric
		&kubernetesMetric{client: i.client, tags: i.Tags},
		// object
		&cluster{client: i.client, tags: i.Tags},
		&deployment{client: i.client, tags: i.Tags},
		&replicaSet{client: i.client, tags: i.Tags},
		&service{client: i.client, tags: i.Tags},
		&node{client: i.client, tags: i.Tags},
		&job{client: i.client, tags: i.Tags},
		&cronJob{client: i.client, tags: i.Tags},
		// &pod{client: i.client, tags: i.Tags},
	}
}

func (i *Input) buildExporters() {
	i.exporterList = []exporter{&pod{client: i.client, tags: i.Tags}}
}

func (i *Input) execExport() {
	for _, exporter := range i.exporterList {
		exporter.Export()
	}
}

func (i *Input) gatherObject() {
	if len(i.resourceList) < 2 {
		return
	}
	for _, resource := range i.resourceList[1:] {
		resource.Gather()
	}
}

func (i *Input) gatherMetric() {
	if len(i.resourceList) == 0 {
		return
	}
	i.resourceList[0].Gather()
}

func (i *Input) Pause() error {
	tick := time.NewTicker(inputs.ElectionPauseTimeout)
	select {
	case i.chPause <- true:
		return nil
	case <-tick.C:
		return fmt.Errorf("pause %s failed", inputName)
	}
}

func (i *Input) Resume() error {
	tick := time.NewTicker(inputs.ElectionResumeTimeout)
	select {
	case i.chPause <- false:
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

func init() { //nolint:gochecknoinits
	inputs.Add(inputName, func() inputs.Input {
		return &Input{
			Tags:    make(map[string]string),
			chPause: make(chan bool, inputs.ElectionPauseChannelLength),
		}
	})
}
