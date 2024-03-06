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
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/export/doc"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/httpcli"
	dkio "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/io"
	dnet "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/net"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs/couchbase/collectors"
)

const (
	minInterval               = time.Second
	maxInterval               = time.Minute
	inputName                 = "couchbase"
	metricName                = inputName
	defaultIntervalDuration   = time.Second * 30
	defaultTimeout            = time.Second * 5
	minTimeout                = time.Second * 5
	maxTimeout                = time.Second * 30
	defaultInsecureSkipVerify = false
)

var (
	_ inputs.ReadEnv = (*Input)(nil)
	l                = logger.DefaultSLogger(inputName)

	_ inputs.ElectionInput = (*Input)(nil)
)

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

	semStop    *cliutils.Sem
	feeder     dkio.Feeder
	client     *collectors.Client
	mergedTags map[string]string
	tagger     datakit.GlobalTagger

	Election bool `toml:"election"`
	pause    bool
	pauseCh  chan bool
}

func (ipt *Input) Run() {
	if err := ipt.setup(); err != nil {
		l.Errorf("setup err: %v", err)
		return
	}

	tick := time.NewTicker(ipt.Interval)
	defer tick.Stop()

	for {
		if ipt.pause {
			l.Debug("%s election paused", inputName)
		} else {
			start := time.Now()

			if err := ipt.collect(); err != nil {
				ipt.feeder.FeedLastError(err.Error(),
					dkio.WithLastErrorInput(inputName),
				)
				l.Error(err)
			}

			if len(ipt.client.Pts) > 0 {
				if err := ipt.feeder.Feed(metricName, point.Metric, ipt.client.Pts,
					&dkio.Option{CollectCost: time.Since(start)}); err != nil {
					ipt.feeder.FeedLastError(err.Error(),
						dkio.WithLastErrorInput(inputName),
						dkio.WithLastErrorCategory(point.Metric),
					)
					l.Errorf("feed measurement: %s", err)
				}
			}
		}

		select {
		case <-tick.C:
		case <-datakit.Exit.Wait():
			l.Infof("%s input exit", inputName)
			return
		case <-ipt.semStop.Wait():
			l.Infof("%s input return", inputName)
			return
		case ipt.pause = <-ipt.pauseCh:
		}
	}
}

func (ipt *Input) setup() error {
	l = logger.SLogger(inputName)

	l.Infof("%s input started", inputName)
	ipt.Interval = config.ProtectedInterval(minInterval, maxInterval, ipt.Interval)
	if ipt.Election {
		ipt.mergedTags = inputs.MergeTags(ipt.tagger.ElectionTags(), ipt.Tags, ipt.Host)
	} else {
		ipt.mergedTags = inputs.MergeTags(ipt.tagger.HostTags(), ipt.Tags, ipt.Host)
	}
	l.Debugf("merged tags: %+#v", ipt.mergedTags)

	if err := ipt.parseURL(); err != nil {
		return err
	}

	cli, err := ipt.newClient()
	if err != nil {
		return err
	}
	ipt.client = cli

	return nil
}

func (ipt *Input) collect() error {
	if ipt.client == nil {
		return fmt.Errorf("i.client is nil")
	}

	if err := ipt.client.GetPts(); err != nil {
		return err
	}

	if len(ipt.client.Pts) == 0 {
		return fmt.Errorf("got nil pts")
	}

	return nil
}

func (ipt *Input) Terminate() {
	if ipt.semStop != nil {
		ipt.semStop.Close()
	}
}
func (*Input) Catalog() string          { return inputName }
func (*Input) SampleConfig() string     { return sampleCfg }
func (*Input) AvailableArchs() []string { return datakit.AllOSWithElection }
func (*Input) SampleMeasurement() []inputs.Measurement {
	return []inputs.Measurement{
		&nodeMeasurement{},
		&bucketInfoMeasurement{},
		&taskMeasurement{},
		&queryMeasurement{},
		&indexMeasurement{},
		&searchMeasurement{},
		&cbasMeasurement{},
		&eventingMeasurement{},
		&perNodeBucketStatsMeasurement{},
		&bucketStatsMeasurement{},
	}
}

func (ipt *Input) ElectionEnabled() bool {
	return ipt.Election
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

func (ipt *Input) GetENVDoc() []*inputs.ENVInfo {
	// nolint:lll
	infos := []*inputs.ENVInfo{
		{FieldName: "Interval", Default: `30s`},
		{FieldName: "Timeout", Default: `5s`},
		{FieldName: "Scheme", Type: doc.String, Example: `http or https`, Desc: "URL Scheme", DescZh: "网络协议"},
		{FieldName: "Host", Type: doc.String, Example: `127.0.0.1`, Desc: "server URL", DescZh: "服务器网址"},
		{FieldName: "Port", Type: doc.Int, Example: `8091 or 18091`, Desc: "Host port, If https will be 18091", DescZh: "端口号，https 用 18091"},
		{FieldName: "AdditionalPort", Type: doc.Int, Example: `9102 or 19102`, Desc: "Additional host port for index metric, If https will be 19102", DescZh: "附加的端口号，https 用 19102"},
		{FieldName: "User", Type: doc.String, Example: `Administrator`, Desc: "User name", DescZh: "登录名"},
		{FieldName: "Password", Type: doc.String, Example: `123456`, Desc: "Password", DescZh: "登录密码"},
		{FieldName: "TLSOpen", Type: doc.Boolean, Default: `false`, Desc: "TLS open"},
		{FieldName: "CacertFile", ENVName: "TLS_CA", Type: doc.String, Example: `/opt/ca.crt`, Desc: "TLS configuration"},
		{FieldName: "CertFile", ENVName: "TLS_CERT", Type: doc.String, Example: `/opt/peer.crt`, Desc: "TLS configuration"},
		{FieldName: "KeyFile", ENVName: "TLS_KEY", Type: doc.String, Example: `/opt/peer.key`, Desc: "TLS configuration"},
		{FieldName: "Election"},
		{FieldName: "Tags"},
	}

	return doc.SetENVDoc("ENV_INPUT_COUCHBASE_", infos)
}

// ReadEnv support envs：only for K8S.
func (ipt *Input) ReadEnv(envs map[string]string) {
	if str, ok := envs["ENV_INPUT_COUCHBASE_INTERVAL"]; ok {
		da, err := time.ParseDuration(str)
		if err != nil {
			l.Warnf("parse ENV_INPUT_COUCHBASE_INTERVAL to time.Duration: %s, ignore", err)
		} else {
			ipt.Interval = config.ProtectedInterval(minInterval, maxInterval, da)
		}
	}

	if str, ok := envs["ENV_INPUT_COUCHBASE_TIMEOUT"]; ok {
		da, err := time.ParseDuration(str)
		if err != nil {
			l.Warnf("parse ENV_INPUT_COUCHBASE_TIMEOUT to time.Duration: %s, ignore", err)
		} else {
			ipt.Timeout = config.ProtectedInterval(minTimeout, maxTimeout, da)
		}
	}

	if str, ok := envs["ENV_INPUT_COUCHBASE_SCHEME"]; ok {
		ipt.Scheme = str
	}

	if str, ok := envs["ENV_INPUT_COUCHBASE_HOST"]; ok {
		ipt.Host = str
	}

	if str, ok := envs["ENV_INPUT_COUCHBASE_PORT"]; ok {
		if port, err := strconv.Atoi(str); err != nil {
			l.Warnf("parse ENV_INPUT_COUCHBASE_PORT: %s, ignore", err)
		} else {
			ipt.Port = port
		}
	}

	if str, ok := envs["ENV_INPUT_COUCHBASE_ADDITIONAL_PORT"]; ok {
		if additionalPort, err := strconv.Atoi(str); err != nil {
			l.Warnf("parse ENV_INPUT_COUCHBASE_ADDITIONAL_PORT: %s, ignore", err)
		} else {
			ipt.AdditionalPort = additionalPort
		}
	}

	if str, ok := envs["ENV_INPUT_COUCHBASE_USER"]; ok {
		ipt.User = str
	}

	if str, ok := envs["ENV_INPUT_COUCHBASE_PASSWORD"]; ok {
		ipt.Password = str
	}

	if str, ok := envs["ENV_INPUT_COUCHBASE_TLS_OPEN"]; ok {
		open, err := strconv.ParseBool(str)
		if err != nil {
			l.Warnf("parse ENV_INPUT_COUCHBASE_TLS_OPEN: %s, ignore", err)
		} else {
			ipt.TLSOpen = open
		}
	}

	if str, ok := envs["ENV_INPUT_COUCHBASE_TLS_CA"]; ok {
		ipt.CacertFile = str
	}

	if str, ok := envs["ENV_INPUT_COUCHBASE_TLS_CERT"]; ok {
		ipt.CertFile = str
	}

	if str, ok := envs["ENV_INPUT_COUCHBASE_TLS_KEY"]; ok {
		ipt.KeyFile = str
	}

	if tagsStr, ok := envs["ENV_INPUT_COUCHBASE_TAGS"]; ok {
		tags := config.ParseGlobalTags(tagsStr)
		for k, v := range tags {
			ipt.Tags[k] = v
		}
	}

	if str, ok := envs["ENV_INPUT_COUCHBASE_ELECTION"]; ok {
		if election, err := strconv.ParseBool(str); err != nil {
			l.Warnf("parse ENV_INPUT_COUCHBASE_ELECTION: %s, ignore", err)
		} else {
			ipt.Election = election
		}
	}
}

func (ipt *Input) newClient() (*collectors.Client, error) {
	opt := &collectors.Option{
		TLSOpen:        ipt.TLSOpen,
		CacertFile:     ipt.CacertFile,
		CertFile:       ipt.CertFile,
		KeyFile:        ipt.KeyFile,
		Scheme:         ipt.Scheme,
		Host:           ipt.Host,
		Port:           ipt.Port,
		AdditionalPort: ipt.AdditionalPort,
		User:           ipt.User,
		Password:       ipt.Password,
	}

	client := &collectors.Client{
		Opt: opt,
		Ctx: &collectors.MetricContext{},
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

	client.MergedTags = ipt.mergedTags

	return client, nil
}

var maxPauseCh = inputs.ElectionPauseChannelLength

func NewInput() *Input {
	return &Input{
		pauseCh:  make(chan bool, maxPauseCh),
		Interval: defaultIntervalDuration,
		Timeout:  defaultTimeout,
		Election: true,
		Tags:     make(map[string]string),

		semStop: cliutils.NewSem(),
		feeder:  dkio.DefaultFeeder(),
		tagger:  datakit.DefaultGlobalTagger(),
	}
}

func init() { //nolint:gochecknoinits
	inputs.Add(inputName, func() inputs.Input {
		return NewInput()
	})
}
