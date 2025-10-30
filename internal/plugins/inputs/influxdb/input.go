// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

// Package influxdb collects InfluxDB metrics.
package influxdb

import (
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"reflect"
	"time"

	"github.com/GuanceCloud/cliutils"
	"github.com/GuanceCloud/cliutils/logger"
	"github.com/GuanceCloud/cliutils/point"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/config"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/goroutine"
	dkio "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/io"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/metrics"
	dknet "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/net"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/ntp"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/tailer"
)

const (
	minInterval      = time.Second * 5
	maxInterval      = time.Minute * 10
	inputName        = "influxdb"
	metricNamePrefix = "influxdb_"
)

var (
	_ inputs.ElectionInput = (*Input)(nil)
	l                      = logger.DefaultSLogger("influxdb")
)

type Input struct {
	URLsDeprecated []string `toml:"urls,omitempty"`

	URL      string `toml:"url"`
	Username string `toml:"username"`
	Password string `toml:"password"`

	Timeout  datakit.Duration `toml:"timeout"`
	Interval datakit.Duration `toml:"interval"`

	Log *struct {
		Files             []string `toml:"files"`
		Pipeline          string   `toml:"pipeline"`
		IgnoreStatus      []string `toml:"ignore"`
		CharacterEncoding string   `toml:"character_encoding"`
		MultilineMatch    string   `toml:"multiline_match"`
	} `toml:"log"`

	TLSConf *dknet.TLSClientConfig `toml:"tlsconf"`
	Tags    map[string]string      `toml:"tags"`

	tail         *tailer.Tailer
	client       *http.Client
	collectCache []*point.Point

	Election bool `toml:"election"`
	pause    bool
	pauseCh  chan bool

	semStop *cliutils.Sem // start stop signal
	feeder  dkio.Feeder
	Tagger  datakit.GlobalTagger
}

var maxPauseCh = inputs.ElectionPauseChannelLength

func (ipt *Input) ElectionEnabled() bool {
	return ipt.Election
}

func (*Input) Catalog() string { return "influxdb" }

func (*Input) SampleConfig() string { return sampleConfig }

func (*Input) AvailableArchs() []string { return datakit.AllOSWithElection }

func (*Input) PipelineConfig() map[string]string { return nil }

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

func (*Input) SampleMeasurement() []inputs.Measurement {
	return []inputs.Measurement{
		&InfluxdbCqM{},
		&InfluxdbDatabaseM{},
		&InfluxdbHttpdM{},
		&InfluxdbMemstatsM{},
		&InfluxdbQueryExecutorM{},
		&InfluxdbRuntimeM{},
		&InfluxdbShardM{},
		&InfluxdbSubscriberM{},
		&InfluxdbTsm1CacheM{},
		&InfluxdbTsm1EngineM{},
		&InfluxdbTsm1FilestoreM{},
		&InfluxdbTsm1WalM{},
		&InfluxdbWriteM{},
	}
}

func (ipt *Input) RunPipeline() {
	if ipt.Log == nil || len(ipt.Log.Files) == 0 {
		return
	}

	opts := []tailer.Option{
		tailer.WithSource(inputName),
		tailer.WithService(inputName),
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
		l.Error(err)
		ipt.feeder.FeedLastError(err.Error(),
			metrics.WithLastErrorInput(inputName),
		)
		return
	}
	g := goroutine.NewGroup(goroutine.Option{Name: "inputs_influxdb"})
	g.Go(func(ctx context.Context) error {
		ipt.tail.Start()
		return nil
	})
}

func (ipt *Input) setup() {
	l = logger.SLogger(inputName)
	l.Infof("influxdb input started")
	ipt.Interval.Duration = config.ProtectedInterval(minInterval, maxInterval, ipt.Interval.Duration)
}

func (ipt *Input) tryInit() {
	tlsCfg, err := ipt.TLSConf.TLSConfigWithBase64()
	if err != nil {
		l.Errorf("TLSConfig: %s", err)
		ipt.feeder.FeedLastError(err.Error(),
			metrics.WithLastErrorInput(inputName),
		)
		return
	}

	ipt.client = &http.Client{
		Transport: &http.Transport{
			ResponseHeaderTimeout: ipt.Timeout.Duration,
			TLSClientConfig:       tlsCfg,
		},
		Timeout: ipt.Timeout.Duration,
	}
}

func (ipt *Input) Run() {
	ipt.setup()
	ipt.tryInit()

	tick := time.NewTicker(ipt.Interval.Duration)
	defer tick.Stop()
	start := ntp.Now()

	for {
		if !ipt.pause {
			collectStart := time.Now()
			if err := ipt.Collect(start.UnixNano()); err != nil {
				l.Errorf("Collect: %s", err)
				ipt.feeder.FeedLastError(err.Error(),
					metrics.WithLastErrorInput(inputName),
				)
			}

			if len(ipt.collectCache) > 0 {
				if err := ipt.feeder.Feed(point.Metric, ipt.collectCache,
					dkio.WithCollectCost(time.Since(collectStart)),
					dkio.WithElection(ipt.Election),
					dkio.WithSource(inputName),
				); err != nil {
					ipt.feeder.FeedLastError(err.Error(),
						metrics.WithLastErrorInput(inputName),
					)
					l.Errorf("feed measurement: %s", err)
				}
				ipt.collectCache = make([]*point.Point, 0)
			}
		} else {
			l.Debugf("not leader, skipped")
		}

		select {
		case <-datakit.Exit.Wait():
			ipt.exit()
			l.Infof("influxdb input exit")
			return

		case <-ipt.semStop.Wait():
			ipt.exit()
			l.Infof("influxdb input return")
			return

		case tt := <-tick.C:
			start = inputs.AlignTime(tt, start, ipt.Interval.Duration)

		case ipt.pause = <-ipt.pauseCh:
			// nil
		}
	}
}

func (ipt *Input) exit() {
	if ipt.tail != nil {
		ipt.tail.Close()
		l.Info("solr log exit")
	}
}

func (ipt *Input) Terminate() {
	if ipt.semStop != nil {
		ipt.semStop.Close()
	}
}

func (ipt *Input) Collect(ptTS int64) error {
	if ipt.client == nil {
		return fmt.Errorf("ipt.client is nil, un initialized")
	}

	req, err := http.NewRequest("GET", ipt.URL, nil)
	if err != nil {
		return err
	}
	if ipt.Username != "" || ipt.Password != "" {
		req.SetBasicAuth(ipt.Username, ipt.Password)
	}

	req.Header.Set("User-Agent", "Datakit/"+datakit.Version)
	resp, err := ipt.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close() //nolint:errcheck
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("influxdb: API responded with status-code %d, URL: %s, Resp: %s", resp.StatusCode, ipt.URL, resp.Body)
	}

	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	fc, err := DebugVarsDataParse2Point(data, MetricMap)
	if err != nil {
		return err
	}
	for {
		pt, err := fc()
		if err != nil {
			if reflect.TypeOf(err) == reflect.TypeOf(NoMoreDataError{}) || err.Error() == "no more data" {
				break
			} else {
				return err
			}
		}
		if pt != nil {
			if pt.Tags == nil {
				pt.Tags = make(map[string]string)
			}
			for k, v := range ipt.Tags {
				pt.Tags[k] = v
			}

			if ipt.Election {
				pt.Tags = inputs.MergeTags(ipt.Tagger.ElectionTags(), pt.Tags, ipt.URL)
			} else {
				pt.Tags = inputs.MergeTags(ipt.Tagger.HostTags(), pt.Tags, ipt.URL)
			}

			metric := &measurement{
				name:   metricNamePrefix + pt.Name,
				tags:   pt.Tags,
				fields: pt.Values,
				ts:     ptTS,
			}
			ipt.collectCache = append(ipt.collectCache, metric.Point())
		}
	}
	return nil
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

func defaultInput() *Input {
	return &Input{
		Interval: datakit.Duration{Duration: time.Second * 15},
		Timeout:  datakit.Duration{Duration: time.Second * 5},
		pauseCh:  make(chan bool, maxPauseCh),
		Election: true,

		semStop: cliutils.NewSem(),
		feeder:  dkio.DefaultFeeder(),
		Tagger:  datakit.DefaultGlobalTagger(),
	}
}

func init() { //nolint:gochecknoinits
	inputs.Add(inputName, func() inputs.Input {
		return defaultInput()
	})
}
