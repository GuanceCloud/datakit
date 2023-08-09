// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

// Package influxdb collects InfluxDB metrics.
package influxdb

import (
	"context"
	"crypto/tls"
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
	dkpt "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/io/point"
	dknet "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/net"
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
	Tagger  dkpt.GlobalTagger
}

var maxPauseCh = inputs.ElectionPauseChannelLength

func (i *Input) ElectionEnabled() bool {
	return i.Election
}

func (*Input) Catalog() string { return "influxdb" }

func (*Input) SampleConfig() string { return sampleConfig }

func (*Input) AvailableArchs() []string { return datakit.AllOSWithElection }

func (*Input) PipelineConfig() map[string]string { return nil }

func (i *Input) GetPipeline() []*tailer.Option {
	return []*tailer.Option{
		{
			Source:  inputName,
			Service: inputName,
			Pipeline: func() string {
				if i.Log != nil {
					return i.Log.Pipeline
				}
				return ""
			}(),
		},
	}
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

func (i *Input) RunPipeline() {
	if i.Log == nil || len(i.Log.Files) == 0 {
		return
	}

	opt := &tailer.Option{
		Source:            inputName,
		Service:           inputName,
		Pipeline:          i.Log.Pipeline,
		GlobalTags:        i.Tags,
		IgnoreStatus:      i.Log.IgnoreStatus,
		CharacterEncoding: i.Log.CharacterEncoding,
		MultilinePatterns: []string{i.Log.MultilineMatch},
		Done:              i.semStop.Wait(),
	}

	var err error
	i.tail, err = tailer.NewTailer(i.Log.Files, opt)
	if err != nil {
		l.Error(err)
		i.feeder.FeedLastError(err.Error(),
			dkio.WithLastErrorInput(inputName),
		)
		return
	}
	g := goroutine.NewGroup(goroutine.Option{Name: "inputs_influxdb"})
	g.Go(func(ctx context.Context) error {
		i.tail.Start()
		return nil
	})
}

func (i *Input) Run() {
	l = logger.SLogger(inputName)

	l.Infof("influxdb input started")

	i.Interval.Duration = config.ProtectedInterval(minInterval, maxInterval, i.Interval.Duration)
	var tlsCfg *tls.Config

	if i.TLSConf != nil {
		var err error
		tlsCfg, err = i.TLSConf.TLSConfig()
		if err != nil {
			l.Errorf("TLSConfig: %s", err)
			i.feeder.FeedLastError(err.Error(),
				dkio.WithLastErrorInput(inputName),
			)
			return
		}
	} else {
		tlsCfg = nil
	}

	i.client = &http.Client{
		Transport: &http.Transport{
			ResponseHeaderTimeout: i.Timeout.Duration,
			TLSClientConfig:       tlsCfg,
		},
		Timeout: i.Timeout.Duration,
	}

	tick := time.NewTicker(i.Interval.Duration)

	defer tick.Stop()
	for {
		if !i.pause {
			start := time.Now()
			if err := i.Collect(); err != nil {
				l.Errorf("Collect: %s", err)
				i.feeder.FeedLastError(err.Error(),
					dkio.WithLastErrorInput(inputName),
				)
			}

			if len(i.collectCache) > 0 {
				if err := i.feeder.Feed(
					inputName, point.Metric, i.collectCache,
					&dkio.Option{CollectCost: time.Since(start)},
				); err != nil {
					l.Errorf("Feed failed: %v", err)
					i.feeder.FeedLastError(err.Error(),
						dkio.WithLastErrorInput(inputName),
					)
				}
				i.collectCache = make([]*point.Point, 0)
			}
		} else {
			l.Debugf("not leader, skipped")
		}

		select {
		case <-datakit.Exit.Wait():
			i.exit()
			l.Infof("influxdb input exit")
			return

		case <-i.semStop.Wait():
			i.exit()
			l.Infof("influxdb input return")
			return

		case <-tick.C:

		case i.pause = <-i.pauseCh:
			// nil
		}
	}
}

func (i *Input) exit() {
	if i.tail != nil {
		i.tail.Close()
		l.Info("solr log exit")
	}
}

func (i *Input) Terminate() {
	if i.semStop != nil {
		i.semStop.Close()
	}
}

func (i *Input) Collect() error {
	ts := time.Now()

	req, err := http.NewRequest("GET", i.URL, nil)
	if err != nil {
		return err
	}
	if i.Username != "" || i.Password != "" {
		req.SetBasicAuth(i.Username, i.Password)
	}

	req.Header.Set("User-Agent", "Datakit/"+datakit.Version)
	resp, err := i.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close() //nolint:errcheck
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("influxdb: API responded with status-code %d, URL: %s, Resp: %s", resp.StatusCode, i.URL, resp.Body)
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
			for k, v := range i.Tags {
				pt.Tags[k] = v
			}

			if i.Election {
				pt.Tags = inputs.MergeTags(i.Tagger.ElectionTags(), pt.Tags, i.URL)
			} else {
				pt.Tags = inputs.MergeTags(i.Tagger.HostTags(), pt.Tags, i.URL)
			}

			metric := &measurement{
				name:   metricNamePrefix + pt.Name,
				tags:   pt.Tags,
				fields: pt.Values,
				ts:     ts,
			}
			i.collectCache = append(i.collectCache, metric.Point())
		}
	}
	return nil
}

func (i *Input) Pause() error {
	tick := time.NewTicker(inputs.ElectionPauseTimeout)
	defer tick.Stop()
	select {
	case i.pauseCh <- true:
		return nil
	case <-tick.C:
		return fmt.Errorf("pause %s failed", inputName)
	}
}

func (i *Input) Resume() error {
	tick := time.NewTicker(inputs.ElectionResumeTimeout)
	defer tick.Stop()
	select {
	case i.pauseCh <- false:
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
		Tagger:  dkpt.DefaultGlobalTagger(),
	}
}

func init() { //nolint:gochecknoinits
	inputs.Add(inputName, func() inputs.Input {
		return defaultInput()
	})
}
