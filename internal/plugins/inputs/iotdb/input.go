// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package iotdb

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/GuanceCloud/cliutils"
	"github.com/GuanceCloud/cliutils/logger"
	"github.com/GuanceCloud/cliutils/point"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/datakit"
	dkio "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/io"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/ntp"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs"
)

const defaultInterval = 30 * time.Second

type Input struct {
	Endpoint string            `toml:"endpoint"`
	Username string            `toml:"username"`
	Password string            `toml:"password"`
	Interval time.Duration     `toml:"interval"`
	Timeout  time.Duration     `toml:"timeout"`
	Tags     map[string]string `toml:"tags"`
	Election bool              `toml:"election"`

	pauseCh chan bool
	semStop *cliutils.Sem
	feeder  dkio.Feeder
	Tagger  datakit.GlobalTagger
	client  *http.Client
}

func (*Input) Catalog() string { return "db" }

func (*Input) AvailableArchs() []string { return datakit.AllOSWithElection }

func (*Input) SampleConfig() string { return sampleConfig }

func (*Input) SampleMeasurement() []inputs.Measurement {
	return []inputs.Measurement{&Measurement{}}
}

func (ipt *Input) ElectionEnabled() bool { return ipt.Election }

func (ipt *Input) Terminate() {
	if ipt.semStop != nil {
		ipt.semStop.Close()
	}
}

func (ipt *Input) Pause() error { return ipt.setPaused(true) }

func (ipt *Input) Resume() error { return ipt.setPaused(false) }

func (ipt *Input) setPaused(paused bool) error {
	if ipt.pauseCh == nil {
		return fmt.Errorf("%s pause channel is not initialized", inputName)
	}
	timer := time.NewTimer(3 * time.Second)
	defer timer.Stop()
	select {
	case ipt.pauseCh <- paused:
		return nil
	case <-timer.C:
		return fmt.Errorf("%s pause state change timed out", inputName)
	}
}

func (ipt *Input) Run() {
	l = logger.SLogger(inputName)
	ipt.setDefaults()
	ipt.client = &http.Client{Timeout: ipt.Timeout}
	defer ipt.client.CloseIdleConnections()

	if err := ipt.collect(context.Background()); err != nil {
		l.Warnf("initial scrape failed: %v", err)
	}

	ticker := time.NewTicker(ipt.Interval)
	defer ticker.Stop()
	paused := false
	for {
		select {
		case <-datakit.Exit.Wait():
			return
		case <-ipt.semStop.Wait():
			return
		case paused = <-ipt.pauseCh:
			if paused {
				l.Infof("%s input paused", inputName)
			} else {
				l.Infof("%s input resumed", inputName)
			}
		case <-ticker.C:
			if paused {
				continue
			}
			if err := ipt.collect(context.Background()); err != nil {
				l.Warnf("scrape failed: %v", err)
			}
		}
	}
}

func (ipt *Input) setDefaults() {
	if ipt.Endpoint == "" {
		ipt.Endpoint = defaultEndpoint
	}
	if ipt.Interval <= 0 {
		ipt.Interval = defaultInterval
	}
	if ipt.Timeout <= 0 {
		ipt.Timeout = defaultTimeout
	}
	if ipt.Tags == nil {
		ipt.Tags = map[string]string{}
	}
	if ipt.pauseCh == nil {
		ipt.pauseCh = make(chan bool, 8)
	}
	if ipt.semStop == nil {
		ipt.semStop = cliutils.NewSem()
	}
	if ipt.feeder == nil {
		ipt.feeder = dkio.DefaultFeeder()
	}
	if ipt.Tagger == nil {
		ipt.Tagger = datakit.DefaultGlobalTagger()
	}
}

func (ipt *Input) collect(ctx context.Context) error {
	scrapedAt := ntp.Now()
	samples, err := scrapeMetrics(ctx, ipt.client, ipt.Endpoint, ipt.Username, ipt.Password, scrapedAt)
	if err != nil {
		return err
	}
	if len(samples) == 0 {
		l.Warnf("IoTDB metrics endpoint returned no samples")
		return nil
	}

	pts := make([]*point.Point, 0, len(samples))
	for _, sample := range samples {
		tags := map[string]string{
			"metric_name": sample.name,
			"metric_type": sample.metricType,
		}
		for key, value := range sample.labels {
			if key == "metric_name" || key == "metric_type" {
				continue
			}
			tags[key] = value
		}
		if ipt.Election {
			tags = inputs.MergeTagsWrapper(tags, ipt.Tagger.ElectionTags(), ipt.Tags, ipt.Endpoint)
		} else {
			tags = inputs.MergeTagsWrapper(tags, ipt.Tagger.HostTags(), ipt.Tags, ipt.Endpoint)
		}

		opts := append(point.DefaultMetricOptions(), point.WithTimestamp(sample.timestamp.UnixNano()))
		pts = append(pts, point.NewPoint(inputName,
			append(point.NewTags(tags), point.NewKVs(map[string]interface{}{"value": sample.value})...),
			opts...))
	}

	if err := ipt.feeder.Feed(point.Metric, pts,
		dkio.WithElection(ipt.Election),
		dkio.WithSource(inputName),
		dkio.WithInput(inputName)); err != nil {
		return fmt.Errorf("feed IoTDB metrics: %w", err)
	}
	return nil
}
