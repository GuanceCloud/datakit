// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

// Package rabbitmq collects rabbitmq metrics.
package rabbitmq

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/GuanceCloud/cliutils"
	"github.com/GuanceCloud/cliutils/logger"
	"github.com/influxdata/telegraf/plugins/common/tls"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/config"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/goroutine"
	dkio "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/io"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/metrics"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/ntp"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/tailer"
)

var _ inputs.ElectionInput = (*Input)(nil)

type rabbitmqlog struct {
	Files             []string `toml:"files"`
	Pipeline          string   `toml:"pipeline"`
	IgnoreStatus      []string `toml:"ignore"`
	CharacterEncoding string   `toml:"character_encoding"`
	MultilineMatch    string   `toml:"multiline_match"`
}

type Input struct {
	URL      string           `toml:"url"`
	Username string           `toml:"username"`
	Password string           `toml:"password"`
	Interval datakit.Duration `toml:"interval"`
	Log      *rabbitmqlog     `toml:"log"`

	Tags       map[string]string `toml:"tags"`
	mergedTags map[string]string

	Version            string
	Uptime             int
	CollectCoStatus    string
	CollectCoErrMsg    string
	LastCustomerObject *customerObjectMeasurement

	QueueNameIncludeDeprecated []string `toml:"queue_name_include,omitempty"`
	QueueNameExcludeDeprecated []string `toml:"queue_name_exclude,omitempty"`

	tls.ClientConfig

	// HTTP client
	client *http.Client

	tail    *tailer.Tailer
	lastErr error

	start time.Time

	Election bool `toml:"election"`
	pause    bool
	pauseCh  chan bool

	semStop *cliutils.Sem // start stop signal
	feeder  dkio.Feeder
	Tagger  datakit.GlobalTagger

	UpState int
}

func (*Input) SampleConfig() string { return sample }

func (*Input) Catalog() string { return inputName }

func (*Input) AvailableArchs() []string { return datakit.AllOSWithElection }

func (*Input) PipelineConfig() map[string]string { return map[string]string{"rabbitmq": pipelineCfg} }

func (ipt *Input) ElectionEnabled() bool {
	return ipt.Election
}

//nolint:lll
func (*Input) LogExamples() map[string]map[string]string {
	return map[string]map[string]string{
		"rabbitmq": {
			"RabbitMQ log": `2021-05-26 14:20:06.105 [warning] <0.12897.46> rabbitmqctl node_health_check and its HTTP API counterpart are DEPRECATED. See https://www.rabbitmq.com/monitoring.html#health-checks for replacement options.`,
		},
	}
}

func (ipt *Input) GetPipeline() []tailer.Option {
	opts := []tailer.Option{
		tailer.WithSource(inputName),
		tailer.WithService(inputName),
	}
	if ipt.Log != nil {
		opts = append(opts, tailer.WithPipeline(ipt.Log.Pipeline))
	}
	return opts
}

func (ipt *Input) RunPipeline() {
	if ipt.Log == nil || len(ipt.Log.Files) == 0 {
		return
	}

	opts := []tailer.Option{
		tailer.WithSource("rabbitmq"),
		tailer.WithService("rabbitmq"),
		tailer.WithPipeline(ipt.Log.Pipeline),
		tailer.WithIgnoredStatuses(ipt.Log.IgnoreStatus),
		tailer.WithCharacterEncoding(ipt.Log.CharacterEncoding),
		tailer.EnableMultiline(true),
		tailer.WithMaxMultilineLength(int64(float64(config.Cfg.Dataway.MaxRawBodySize) * 0.8)),
		tailer.WithMultilinePatterns([]string{ipt.Log.MultilineMatch}),
		tailer.WithExtraTags(inputs.MergeTags(ipt.Tagger.HostTags(), ipt.Tags, "")),
		tailer.EnableDebugFields(config.Cfg.EnableDebugFields),
	}

	var err error
	ipt.tail, err = tailer.NewTailer(ipt.Log.Files, opts...)
	if err != nil {
		l.Errorf("NewTailer: %s", err)
		ipt.feeder.FeedLastError(ipt.lastErr.Error(),
			metrics.WithLastErrorInput(inputName),
		)
		return
	}

	g := goroutine.NewGroup(goroutine.Option{Name: "inputs_rabbitmq"})
	g.Go(func(ctx context.Context) error {
		ipt.tail.Start()
		return nil
	})
}

func (ipt *Input) Run() {
	l = logger.SLogger(inputName)
	l.Info("rabbitmq start")

	ipt.Interval.Duration = config.ProtectedInterval(minInterval, maxInterval, ipt.Interval.Duration)

	if ipt.Election {
		ipt.mergedTags = inputs.MergeTags(ipt.Tagger.ElectionTags(), ipt.Tags, ipt.URL)
	} else {
		ipt.mergedTags = inputs.MergeTags(ipt.Tagger.HostTags(), ipt.Tags, ipt.URL)
	}

	if x, err := ipt.newClient(); err != nil {
		ipt.FeedCoByErr(err)
		l.Errorf("[error] rabbitmq init client err:%s", err.Error())
		return
	} else {
		ipt.client = x
	}

	tick := time.NewTicker(ipt.Interval.Duration)
	defer tick.Stop()

	ipt.start = ntp.Now()
	for {
		if !ipt.pause {
			ipt.setUpState()
			ipt.getMetric()
			ipt.FeedCoPts()

			if ipt.lastErr != nil {
				ipt.feeder.FeedLastError(ipt.lastErr.Error(),
					metrics.WithLastErrorInput(inputName),
				)
				ipt.setErrUpState()
				ipt.lastErr = nil
			}

			ipt.FeedUpMetric()
		} else {
			l.Debugf("not leader, skipped")
		}

		select {
		case <-datakit.Exit.Wait():
			ipt.exit()
			l.Info("rabbitmq exit")
			return

		case <-ipt.semStop.Wait():
			ipt.exit()
			l.Info("rabbitmq return")
			return

		case tt := <-tick.C:
			ipt.start = inputs.AlignTime(tt, ipt.start, ipt.Interval.Duration)

		case ipt.pause = <-ipt.pauseCh:
		}
	}
}

func (ipt *Input) exit() {
	if ipt.tail != nil {
		ipt.tail.Close()
		l.Info("rabbitmq log exit")
	}
}

func (ipt *Input) Terminate() {
	if ipt.semStop != nil {
		ipt.semStop.Close()
	}
}

type MetricFunc func(n *Input)

func (ipt *Input) getMetric() {
	// get overview first, to get cluster name
	getOverview(ipt)

	getFunc := []MetricFunc{getNode, getQueues, getExchange}

	g := goroutine.NewGroup(goroutine.Option{Name: "inputs_rabbitmq"})
	for _, v := range getFunc {
		func(gf MetricFunc) {
			g.Go(func(ctx context.Context) error {
				gf(ipt)
				return nil
			})
		}(v)
	}

	if err := g.Wait(); err != nil {
		l.Errorf("g.Wait failed: %v", err)
	}
}

func (*Input) SampleMeasurement() []inputs.Measurement {
	return []inputs.Measurement{
		&overviewMeasurement{},
		&queueMeasurement{},
		&exchangeMeasurement{},
		&nodeMeasurement{},
		&customerObjectMeasurement{},
		&inputs.UpMeasurement{},
	}
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

func (ipt *Input) newClient() (*http.Client, error) {
	tlsCfg, err := ipt.ClientConfig.TLSConfig()
	if err != nil {
		return nil, err
	}

	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: tlsCfg,
		},
		Timeout: time.Second * 10,
	}

	return client, nil
}

func (ipt *Input) requestJSON(u string, target interface{}) error {
	u = fmt.Sprintf("%s%s", ipt.URL, u)
	req, err := http.NewRequest("GET", u, nil)
	if err != nil {
		return err
	}
	req.SetBasicAuth(ipt.Username, ipt.Password)
	resp, err := ipt.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close() //nolint:errcheck
	return json.NewDecoder(resp.Body).Decode(target)
}

func newCountFieldInfo(desc string) *inputs.FieldInfo {
	return &inputs.FieldInfo{
		DataType: inputs.Int,
		Type:     inputs.Count,
		Unit:     inputs.NCount,
		Desc:     desc,
	}
}

func defaultInput() *Input {
	return &Input{
		Interval: datakit.Duration{Duration: time.Second * 10},
		pauseCh:  make(chan bool, inputs.ElectionPauseChannelLength),
		Election: true,
		semStop:  cliutils.NewSem(),
		feeder:   dkio.DefaultFeeder(),
		Tagger:   datakit.DefaultGlobalTagger(),
	}
}

func init() { //nolint:gochecknoinits
	inputs.Add(inputName, func() inputs.Input {
		return defaultInput()
	})
}
