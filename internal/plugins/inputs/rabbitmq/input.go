// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

// Package rabbitmq collects rabbitmq metrics.
package rabbitmq

import (
	"context"
	"fmt"
	"net"
	"net/url"
	"time"

	"github.com/GuanceCloud/cliutils"
	"github.com/GuanceCloud/cliutils/logger"
	"github.com/GuanceCloud/cliutils/point"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/config"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/goroutine"
	dkio "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/io"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/tailer"
)

var _ inputs.ElectionInput = (*Input)(nil)

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

func (ipt *Input) GetPipeline() []*tailer.Option {
	return []*tailer.Option{
		{
			Source:  inputName,
			Service: inputName,
			Pipeline: func() string {
				if ipt.Log != nil {
					return ipt.Log.Pipeline
				}
				return ""
			}(),
		},
	}
}

func (ipt *Input) RunPipeline() {
	if ipt.Log == nil || len(ipt.Log.Files) == 0 {
		return
	}

	opt := &tailer.Option{
		Source:            "rabbitmq",
		Service:           "rabbitmq",
		Pipeline:          ipt.Log.Pipeline,
		CharacterEncoding: ipt.Log.CharacterEncoding,
		MultilinePatterns: []string{ipt.Log.MultilineMatch},
		GlobalTags:        inputs.MergeTags(ipt.Tagger.HostTags(), ipt.Tags, ""),
		Done:              ipt.semStop.Wait(),
	}

	var err error
	ipt.tail, err = tailer.NewTailer(ipt.Log.Files, opt, ipt.Log.IgnoreStatus)
	if err != nil {
		l.Errorf("NewTailer: %s", err)
		ipt.feeder.FeedLastError(ipt.lastErr.Error(),
			dkio.WithLastErrorInput(inputName),
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
	if err := ipt.setHostIfNotLoopback(); err != nil {
		l.Errorf("failed to set host from url: %v", err)
	}
	client, err := ipt.createHTTPClient()
	if err != nil {
		l.Errorf("[error] rabbitmq init client err:%s", err.Error())
		return
	}
	ipt.client = client

	tick := time.NewTicker(ipt.Interval.Duration)
	defer tick.Stop()

	for {
		if !ipt.pause {
			ipt.getMetric()

			if ipt.lastErr != nil {
				ipt.feeder.FeedLastError(ipt.lastErr.Error(),
					dkio.WithLastErrorInput(inputName),
				)
				ipt.lastErr = nil
			}

			if len(ipt.collectCache) > 0 {
				if err := ipt.feeder.FeedV2(point.Metric, ipt.collectCache,
					dkio.WithCollectCost(time.Since(ipt.start)),
					dkio.WithElection(ipt.Election),
					dkio.WithInputName(inputName),
				); err != nil {
					l.Errorf("FeedMeasurement: %s", err.Error())
				}

				ipt.collectCache = ipt.collectCache[:0]
			}
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

		case <-tick.C:

		case ipt.pause = <-ipt.pauseCh:
			// nil
		}
	}
}

func (ipt *Input) setHostIfNotLoopback() error {
	uu, err := url.Parse(ipt.URL)
	if err != nil {
		return err
	}
	host, _, err := net.SplitHostPort(uu.Host)
	if err != nil {
		return err
	}
	if host != "localhost" && !net.ParseIP(host).IsLoopback() {
		ipt.host = host
	}
	return nil
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
	ipt.start = time.Now()
	getFunc := []MetricFunc{getOverview, getNode, getQueues, getExchange}
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
		&OverviewMeasurement{},
		&QueueMeasurement{},
		&ExchangeMeasurement{},
		&NodeMeasurement{},
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
