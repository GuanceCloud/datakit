// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

// Package ingestioncanary implements the ingestion canary collector.
package ingestioncanary

import (
	"errors"
	"fmt"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/GuanceCloud/cliutils"
	"github.com/GuanceCloud/cliutils/logger"
	"github.com/GuanceCloud/cliutils/point"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/config"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/datakit"
	ingestioncanary "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/ingestion_canary"
	dkio "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/io"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/io/dataway"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/ntp"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs"
)

const (
	inputName = "ingestion_canary"
)

var l = logger.DefaultSLogger(inputName)

type Input struct {
	Interval        datakit.Duration  `toml:"interval"`
	QueryTimeout    datakit.Duration  `toml:"query_timeout"`
	PollInterval    datakit.Duration  `toml:"poll_interval"`
	ErrorRetries    int               `toml:"error_retries"`
	ResultWorkspace string            `toml:"result_workspace"`
	Categories      []string          `toml:"categories"`
	Logging         *LoggingConfig    `toml:"logging"`
	Tags            map[string]string `toml:"tags"`
	Election        bool              `toml:"election"`

	semStop *cliutils.Sem
	feeder  dkio.Feeder
	tagger  datakit.GlobalTagger

	pauseCh chan bool
	pause   bool

	round     int64
	canary    *ingestioncanary.Canary
	feedFuncs []func()

	resultDataway *dataway.Dataway // dataway for reporting results to another workspace
	writeURL      string
	resultPoints  []*point.Point // cached result points for batch reporting
	mu            sync.Mutex
}

type LoggingConfig struct {
	StorageIndex string `toml:"storage_index"`
}

func (*Input) Catalog() string { return inputName }

func (*Input) SampleConfig() string { return sampleConfig }

func (*Input) AvailableArchs() []string { return datakit.AllOSWithElection }

func (*Input) SampleMeasurement() []inputs.Measurement {
	return []inputs.Measurement{
		&IngestionCanaryResultMetric{},
		&IngestionCanaryMetric{},
		&IngestionCanaryLogging{},
		&IngestionCanaryTracing{},
	}
}

func (ipt *Input) initConfig() error {
	if config.Cfg.Dataway == nil {
		return errors.New("dataway not configured, DQL query will be skipped")
	}

	ipt.initFeedFuncs()

	// Initialize result sender if result_workspace is configured
	if err := ipt.initResultSender(); err != nil {
		return fmt.Errorf("init result sender failed: %w", err)
	}

	return nil
}

// initFeedFuncs initializes feed function array based on categories configuration.
func (ipt *Input) initFeedFuncs() {
	ipt.canary = ingestioncanary.New(
		ingestioncanary.WithName(inputName),
		ingestioncanary.WithTestType("collect"),
		ingestioncanary.WithEnableMetrics(true),
	)

	if len(ipt.Categories) == 0 {
		ipt.feedFuncs = []func(){
			ipt.feedMetric,
			ipt.feedLogging,
			ipt.feedTracing,
		}
		return
	}

	ipt.feedFuncs = []func(){}
	categoryMap := make(map[string]bool)
	for _, cat := range ipt.Categories {
		categoryMap[strings.ToLower(cat)] = true
	}

	if categoryMap["metric"] {
		ipt.feedFuncs = append(ipt.feedFuncs, ipt.feedMetric)
	}
	if categoryMap["logging"] {
		ipt.feedFuncs = append(ipt.feedFuncs, ipt.feedLogging)
	}
	if categoryMap["tracing"] {
		ipt.feedFuncs = append(ipt.feedFuncs, ipt.feedTracing)
	}
}

// initResultSender initializes the result dataway if result_workspace is configured.
func (ipt *Input) initResultSender() error {
	if ipt.ResultWorkspace == "" {
		return nil
	}
	writeURL, err := ipt.buildWriteURL()
	if err != nil {
		return fmt.Errorf("build write URL failed: %w", err)
	}
	ipt.writeURL = writeURL

	// Initialize dataway for reporting results
	dw := dataway.NewDefaultDataway()
	if err := dw.Init(); err != nil {
		return fmt.Errorf("init result dataway failed: %w", err)
	}

	ipt.resultDataway = dw

	return nil
}

func (ipt *Input) buildWriteURL() (string, error) {
	baseURL := ipt.ResultWorkspace

	// Parse the base URL
	u, err := url.Parse(baseURL)
	if err != nil {
		return "", fmt.Errorf("parse workspace URL failed: %w", err)
	}

	// Construct the metric write endpoint URL
	writeURL := fmt.Sprintf("%s://%s/v1/write/metric", u.Scheme, u.Host)
	if u.RawQuery != "" {
		writeURL += "?" + u.RawQuery
	}

	return writeURL, nil
}

func (ipt *Input) Run() {
	l = logger.SLogger(inputName)
	l.Info("ingestion_canary start.")

	if err := ipt.initConfig(); err != nil {
		l.Errorf("init config failed: %s", err.Error())
		return
	}

	duration := ipt.Interval.Duration

	tick := time.NewTicker(duration)
	defer tick.Stop()
	start := ntp.Now()

	for {
		if !ipt.pause {
			if err := ipt.Collect(start.UnixNano()); err != nil {
				l.Errorf("collect failed: %s", err)
			}
		} else {
			l.Infof("ingestion_canary paused")
		}

		select {
		case tt := <-tick.C:
			start = inputs.AlignTime(tt, start, duration)
		case <-datakit.Exit.Wait():
			l.Info("ingestion_canary exit")
			ipt.exit()
			return
		case <-ipt.semStop.Wait():
			l.Info("ingestion_canary return")
			ipt.exit()
			return
		case ipt.pause = <-ipt.pauseCh:
		}
	}
}

func (ipt *Input) Terminate() {
	if ipt.semStop != nil {
		ipt.semStop.Close()
	}
}

func (ipt *Input) exit() {
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

func defaultInput() *Input {
	ipt := &Input{
		Interval:     datakit.Duration{Duration: 10 * time.Minute},
		QueryTimeout: datakit.Duration{Duration: 5 * time.Minute},
		PollInterval: datakit.Duration{Duration: 500 * time.Millisecond},
		ErrorRetries: 10,
		semStop:      cliutils.NewSem(),
		Tags:         make(map[string]string),
		feeder:       dkio.DefaultFeeder(),
		tagger:       datakit.DefaultGlobalTagger(),
		pauseCh:      make(chan bool, inputs.ElectionPauseChannelLength),
		Election:     true,
	}

	return ipt
}

func init() { //nolint:gochecknoinits
	inputs.Add(inputName, func() inputs.Input {
		return defaultInput()
	})
}
