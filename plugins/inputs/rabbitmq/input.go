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
	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/config"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/goroutine"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/tailer"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"
)

var _ inputs.ElectionInput = (*Input)(nil)

func (*Input) SampleConfig() string { return sample }

func (*Input) Catalog() string { return inputName }

func (*Input) AvailableArchs() []string { return datakit.AllOSWithElection }

func (*Input) PipelineConfig() map[string]string { return map[string]string{"rabbitmq": pipelineCfg} }

func (n *Input) ElectionEnabled() bool {
	return n.Election
}

//nolint:lll
func (n *Input) LogExamples() map[string]map[string]string {
	return map[string]map[string]string{
		"rabbitmq": {
			"RabbitMQ log": `2021-05-26 14:20:06.105 [warning] <0.12897.46> rabbitmqctl node_health_check and its HTTP API counterpart are DEPRECATED. See https://www.rabbitmq.com/monitoring.html#health-checks for replacement options.`,
		},
	}
}

func (n *Input) GetPipeline() []*tailer.Option {
	return []*tailer.Option{
		{
			Source:  inputName,
			Service: inputName,
			Pipeline: func() string {
				if n.Log != nil {
					return n.Log.Pipeline
				}
				return ""
			}(),
		},
	}
}

func (n *Input) RunPipeline() {
	if n.Log == nil || len(n.Log.Files) == 0 {
		return
	}

	if n.Log.Pipeline == "" {
		n.Log.Pipeline = "rabbitmq.p" // use default
	}

	opt := &tailer.Option{
		Source:            "rabbitmq",
		Service:           "rabbitmq",
		Pipeline:          n.Log.Pipeline,
		GlobalTags:        n.Tags,
		CharacterEncoding: n.Log.CharacterEncoding,
		MultilinePatterns: []string{n.Log.MultilineMatch},
		Done:              n.semStop.Wait(),
	}

	var err error
	n.tail, err = tailer.NewTailer(n.Log.Files, opt, n.Log.IgnoreStatus)
	if err != nil {
		l.Errorf("NewTailer: %s", err)
		io.FeedLastError(inputName, err.Error())
		return
	}

	g := goroutine.NewGroup(goroutine.Option{Name: "inputs_rabbitmq"})
	g.Go(func(ctx context.Context) error {
		n.tail.Start()
		return nil
	})
}

func (n *Input) Run() {
	l = logger.SLogger(inputName)
	l.Info("rabbitmq start")
	n.Interval.Duration = config.ProtectedInterval(minInterval, maxInterval, n.Interval.Duration)
	if err := n.setHostIfNotLoopback(); err != nil {
		l.Errorf("failed to set host from url: %v", err)
	}
	client, err := n.createHTTPClient()
	if err != nil {
		l.Errorf("[error] rabbitmq init client err:%s", err.Error())
		return
	}
	n.client = client

	tick := time.NewTicker(n.Interval.Duration)
	defer tick.Stop()

	for {
		if !n.pause {
			n.getMetric()

			if n.lastErr != nil {
				io.FeedLastError(inputName, n.lastErr.Error())
				n.lastErr = nil
			}

			if len(collectCache) > 0 {
				if err := inputs.FeedMeasurement(inputName,
					datakit.Metric,
					collectCache,
					&io.Option{CollectCost: time.Since(n.start)}); err != nil {
					l.Errorf("FeedMeasurement: %s", err.Error())
				}

				collectCache = collectCache[:0]
			}
		} else {
			l.Debugf("not leader, skipped")
		}

		select {
		case <-datakit.Exit.Wait():
			n.exit()
			l.Info("rabbitmq exit")
			return

		case <-n.semStop.Wait():
			n.exit()
			l.Info("rabbitmq return")
			return

		case <-tick.C:

		case n.pause = <-n.pauseCh:
			// nil
		}
	}
}

func (n *Input) setHostIfNotLoopback() error {
	uu, err := url.Parse(n.URL)
	if err != nil {
		return err
	}
	host, _, err := net.SplitHostPort(uu.Host)
	if err != nil {
		return err
	}
	if host != "localhost" && !net.ParseIP(host).IsLoopback() {
		n.host = host
	}
	return nil
}

func (n *Input) exit() {
	if n.tail != nil {
		n.tail.Close()
		l.Info("rabbitmq log exit")
	}
}

func (n *Input) Terminate() {
	if n.semStop != nil {
		n.semStop.Close()
	}
}

type MetricFunc func(n *Input)

func (n *Input) getMetric() {
	n.start = time.Now()
	getFunc := []MetricFunc{getOverview, getNode, getQueues, getExchange}
	g := goroutine.NewGroup(goroutine.Option{Name: "inputs_rabbitmq"})
	for _, v := range getFunc {
		func(gf MetricFunc) {
			g.Go(func(ctx context.Context) error {
				gf(n)
				return nil
			})
		}(v)
	}
	_ = g.Wait()
}

func (n *Input) SampleMeasurement() []inputs.Measurement {
	return []inputs.Measurement{
		&OverviewMeasurement{},
		&QueueMeasurement{},
		&ExchangeMeasurement{},
		&NodeMeasurement{},
	}
}

func (n *Input) Pause() error {
	tick := time.NewTicker(inputs.ElectionPauseTimeout)
	defer tick.Stop()
	select {
	case n.pauseCh <- true:
		return nil
	case <-tick.C:
		return fmt.Errorf("pause %s failed", inputName)
	}
}

func (n *Input) Resume() error {
	tick := time.NewTicker(inputs.ElectionResumeTimeout)
	defer tick.Stop()
	select {
	case n.pauseCh <- false:
		return nil
	case <-tick.C:
		return fmt.Errorf("resume %s failed", inputName)
	}
}

func init() { //nolint:gochecknoinits
	inputs.Add(inputName, func() inputs.Input {
		s := &Input{
			Interval: datakit.Duration{Duration: time.Second * 10},
			pauseCh:  make(chan bool, inputs.ElectionPauseChannelLength),
			Election: true,

			semStop: cliutils.NewSem(),
		}
		return s
	})
}
