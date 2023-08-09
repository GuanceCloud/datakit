// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

// Package couchbase collects couchbase metrics.
package couchbase

import (
	"fmt"
	"strconv"
	"time"

	"github.com/GuanceCloud/cliutils"
	"github.com/GuanceCloud/cliutils/logger"
	"github.com/GuanceCloud/cliutils/point"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/config"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/httpcli"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/io"
	dkpt "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/io/point"
	dnet "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/net"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs/couchbase/collectors"
)

var _ inputs.ElectionInput = (*Input)(nil)

const (
	inputName                 = "couchbase"
	catalog                   = "couchbase"
	defaultIntervalDuration   = time.Second * 30
	minInterval               = time.Second * 10
	maxInterval               = time.Second * 60
	defaultTimeout            = time.Second * 5
	minTimeout                = time.Second * 5
	maxTimeout                = time.Second * 30
	defaultInsecureSkipVerify = false

	sampleCfg = `
[[inputs.couchbase]]
  ## Collect interval, default is 30 seconds. (optional)
  # interval = "30s"
  
  ## Timeout: (defaults to "5s"). (optional)
  # timeout = "5s"

  ## Scheme, "http" or "https".
  scheme = "http"

  ## Host url or ip.
  host = "127.0.0.1"

  ## Host port. If "https" will be 18091.
  port = 8091

  ## Additional host port for index metric. If "https" will be 19102.
  additional_port = 9102

  ## Host user name.
  user = "Administrator"

  ## Host password.
  password = "123456"

  ## TLS configuration.
  tls_open = false
  # tls_ca = ""
  # tls_cert = "/var/cb/clientcertfiles/travel-sample.pem"
  # tls_key = "/var/cb/clientcertfiles/travel-sample.key"

  ## Disable setting host tag for this input
  disable_host_tag = false

  ## Disable setting instance tag for this input
  disable_instance_tag = false

  ## Set to 'true' to enable election.
  election = true

# [inputs.couchbase.tags]
  # some_tag = "some_value"
  # more_tag = "some_other_value"
`
)

var l = logger.DefaultSLogger(inputName)

type Input struct {
	Interval       time.Duration     `toml:"interval"`
	Timeout        time.Duration     `toml:"timeout"`
	Scheme         string            `toml:"scheme"`
	Host           string            `toml:"host"`
	Port           int               `toml:"port"`
	AdditionalPort int               `toml:"additional_port"`
	User           string            `toml:"user"`
	Password       string            `toml:"password"`
	TLSOpen        bool              `toml:"tls_open"`
	CacertFile     string            `toml:"tls_ca"`
	CertFile       string            `toml:"tls_cert"`
	KeyFile        string            `toml:"tls_key"`
	Tags           map[string]string `toml:"tags"`
	Election       bool              `toml:"election"`

	feeder        io.Feeder
	client        *collectors.Client
	chPause       chan bool
	pause         bool
	Tagger        dkpt.GlobalTagger
	semStop       *cliutils.Sem
	isInitialized bool
}

// Catalog catalog name.
func (*Input) Catalog() string {
	return catalog
}

// SampleConfig conf File samples, reflected in the document.
func (*Input) SampleConfig() string {
	return sampleCfg
}

// AvailableArchs OS support, reflected in the document.
func (*Input) AvailableArchs() []string { return datakit.AllOSWithElection }

func (*Input) AvailableArchsDCGM() []string {
	return []string{datakit.OSLabelLinux, datakit.LabelK8s}
}

// SampleMeasurement sample measurement results, reflected in the document.
func (*Input) SampleMeasurement() []inputs.Measurement {
	return []inputs.Measurement{
		&NodeMeasurement{},
		&BucketInfoMeasurement{},
		&TaskMeasurement{},
		&QueryMeasurement{},
		&IndexMeasurement{},
		&SearchMeasurement{},
		&CbasMeasurement{},
		&EventingMeasurement{},
		&PerNodeBucketStatsMeasurement{},
		&BucketStatsMeasurement{},
	}
}

func (i *Input) Run() {
	i.Interval = config.ProtectedInterval(minInterval, maxInterval, i.Interval)
	i.Timeout = config.ProtectedInterval(minInterval, maxInterval, i.Timeout)

	tick := time.NewTicker(i.Interval)
	defer tick.Stop()

	l.Info("couchbase start")

	for {
		if i.pause {
			l.Debug("couchbase paused")
		} else {
			l.Debugf("is leader, couchbase gathering...")
			start := time.Now()

			if err := i.Collect(); err != nil {
				i.feeder.FeedLastError(err.Error(),
					io.WithLastErrorInput(inputName),
				)
				l.Error(err)
			} else {
				if errFeed := i.feeder.Feed(inputName, point.Metric, i.client.Pts,
					&io.Option{CollectCost: time.Since(start)}); errFeed != nil {
					i.feeder.FeedLastError(err.Error(),
						io.WithLastErrorInput(inputName),
					)
					l.Error(errFeed)
				}
			}
		}

		select {
		case <-datakit.Exit.Wait():
			l.Info("couchbase exit")
			return

		case <-i.semStop.Wait():
			l.Info("couchbase return")
			return

		case <-tick.C:

		case i.pause = <-i.chPause:
			// nil
		}
	}
}

// Collect collect metrics from all URLs.
func (i *Input) Collect() error {
	if !i.isInitialized {
		if err := i.Init(); err != nil {
			return err
		}
	}

	if i.client == nil {
		return fmt.Errorf("i.client is nil")
	}

	if err := i.client.GetPts(); err != nil {
		return err
	}

	if len(i.client.Pts) == 0 {
		return fmt.Errorf("got nil pts")
	}

	return nil
}

func (i *Input) Terminate() {
	if i.semStop != nil {
		i.semStop.Close()
	}
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

// ElectionEnabled election.
func (i *Input) ElectionEnabled() bool {
	return i.Election
}

// ReadEnv support envs：only for K8S.
func (i *Input) ReadEnv(envs map[string]string) {
	if str, ok := envs["ENV_INPUT_COUCHBASE_INTERVAL"]; ok {
		da, err := time.ParseDuration(str)
		if err != nil {
			l.Warnf("parse ENV_INPUT_COUCHBASE_INTERVAL to time.Duration: %s, ignore", err)
		} else {
			i.Interval = config.ProtectedInterval(minInterval, maxInterval, da)
		}
	}

	if str, ok := envs["ENV_INPUT_COUCHBASE_TIMEOUT"]; ok {
		da, err := time.ParseDuration(str)
		if err != nil {
			l.Warnf("parse ENV_INPUT_COUCHBASE_TIMEOUT to time.Duration: %s, ignore", err)
		} else {
			i.Timeout = config.ProtectedInterval(minTimeout, maxTimeout, da)
		}
	}

	if str, ok := envs["ENV_INPUT_COUCHBASE_SCHEME"]; ok {
		i.Scheme = str
	}

	if str, ok := envs["ENV_INPUT_COUCHBASE_HOST"]; ok {
		i.Host = str
	}

	if str, ok := envs["ENV_INPUT_COUCHBASE_PORT"]; ok {
		if port, err := strconv.Atoi(str); err != nil {
			l.Warnf("parse ENV_INPUT_COUCHBASE_PORT: %s, ignore", err)
		} else {
			i.Port = port
		}
	}

	if str, ok := envs["ENV_INPUT_COUCHBASE_ADDITIONAL_PORT"]; ok {
		if additionalPort, err := strconv.Atoi(str); err != nil {
			l.Warnf("parse ENV_INPUT_COUCHBASE_ADDITIONAL_PORT: %s, ignore", err)
		} else {
			i.AdditionalPort = additionalPort
		}
	}

	if str, ok := envs["ENV_INPUT_COUCHBASE_USER"]; ok {
		i.User = str
	}

	if str, ok := envs["ENV_INPUT_COUCHBASE_PASSWORD"]; ok {
		i.Password = str
	}

	if str, ok := envs["ENV_INPUT_COUCHBASE_TLS_OPEN"]; ok {
		open, err := strconv.ParseBool(str)
		if err != nil {
			l.Warnf("parse ENV_INPUT_COUCHBASE_TLS_OPEN: %s, ignore", err)
		} else {
			i.TLSOpen = open
		}
	}

	if str, ok := envs["ENV_INPUT_COUCHBASE_TLS_CA"]; ok {
		i.CacertFile = str
	}

	if str, ok := envs["ENV_INPUT_COUCHBASE_TLS_CERT"]; ok {
		i.CertFile = str
	}

	if str, ok := envs["ENV_INPUT_COUCHBASE_TLS_KEY"]; ok {
		i.KeyFile = str
	}

	if tagsStr, ok := envs["ENV_INPUT_COUCHBASE_TAGS"]; ok {
		tags := config.ParseGlobalTags(tagsStr)
		for k, v := range tags {
			i.Tags[k] = v
		}
	}

	if str, ok := envs["ENV_INPUT_COUCHBASE_ELECTION"]; ok {
		if election, err := strconv.ParseBool(str); err != nil {
			l.Warnf("parse ENV_INPUT_COUCHBASE_ELECTION: %s, ignore", err)
		} else {
			i.Election = election
		}
	}
}

func (i *Input) Init() error {
	if err := i.parseURL(); err != nil {
		l.Errorf("parse url failed: %w", err)
		return err
	}

	cli, err := i.newClient()
	if err != nil {
		return err
	}
	i.client = cli
	i.isInitialized = true

	return nil
}

func (i *Input) newClient() (*collectors.Client, error) {
	opt := &collectors.Option{
		TLSOpen:        i.TLSOpen,
		CacertFile:     i.CacertFile,
		CertFile:       i.CertFile,
		KeyFile:        i.KeyFile,
		Scheme:         i.Scheme,
		Host:           i.Host,
		Port:           i.Port,
		AdditionalPort: i.AdditionalPort,
		User:           i.User,
		Password:       i.Password,
	}

	client := &collectors.Client{
		Opt:  opt,
		Ctx:  &collectors.MetricContext{},
		Tags: make(map[string]string),
	}

	cliopts := httpcli.NewOptions()

	if opt.TLSOpen {
		caCerts := []string{}
		insecureSkipVerify := defaultInsecureSkipVerify
		if len(opt.CacertFile) != 0 {
			caCerts = append(caCerts, opt.CacertFile)
		} else {
			insecureSkipVerify = true
		}
		tc := &dnet.TLSClientConfig{
			CaCerts:            caCerts,
			Cert:               opt.CertFile,
			CertKey:            opt.KeyFile,
			InsecureSkipVerify: insecureSkipVerify,
		}

		tlsConfig, err := tc.TLSConfig()
		if err != nil {
			return client, err
		}
		cliopts.TLSClientConfig = tlsConfig
	}

	client.SetClient(httpcli.Cli(cliopts))

	for k, v := range i.Tags {
		client.Tags[k] = v
	}
	if i.Election {
		client.URLTags = inputs.MergeTags(i.Tagger.ElectionTags(), i.Tags, i.Host)
	} else {
		client.URLTags = inputs.MergeTags(i.Tagger.HostTags(), i.Tags, i.Host)
	}

	return client, nil
}

var maxPauseCh = inputs.ElectionPauseChannelLength

func NewInput() *Input {
	return &Input{
		chPause:  make(chan bool, maxPauseCh),
		Interval: defaultIntervalDuration,
		Timeout:  defaultTimeout,
		Election: true,
		Tags:     make(map[string]string),

		semStop: cliutils.NewSem(),
		feeder:  io.DefaultFeeder(),
		Tagger:  dkpt.DefaultGlobalTagger(),
	}
}

func init() { //nolint:gochecknoinits
	inputs.Add(inputName, func() inputs.Input {
		return NewInput()
	})
}
