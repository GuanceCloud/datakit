// Package rabbitmq collects rabbitmq metrics.
package rabbitmq

import (
	"fmt"
	"os"
	"time"

	"gitlab.jiagouyun.com/cloudcare-tools/cliutils"
	"gitlab.jiagouyun.com/cloudcare-tools/cliutils/logger"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/config"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/tailer"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"
)

var _ inputs.ElectionInput = (*Input)(nil)

func (*Input) SampleConfig() string { return sample }

func (*Input) Catalog() string { return inputName }

func (*Input) AvailableArchs() []string { return datakit.AllArch }

func (*Input) PipelineConfig() map[string]string { return map[string]string{"rabbitmq": pipelineCfg} }

func (n *Input) GetPipeline() []*tailer.Option {
	return []*tailer.Option{
		{
			Source:   inputName,
			Service:  inputName,
			Pipeline: n.Log.Pipeline,
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
		GlobalTags:        n.Tags,
		CharacterEncoding: n.Log.CharacterEncoding,
		MultilineMatch:    n.Log.MultilineMatch,
	}

	pl, err := config.GetPipelinePath(n.Log.Pipeline)
	if err != nil {
		io.FeedLastError(inputName, err.Error())
		l.Error(err)
		return
	}
	if _, err := os.Stat(pl); err != nil {
		l.Warn("%s missing: %s", pl, err.Error())
	} else {
		opt.Pipeline = pl
	}

	n.tail, err = tailer.NewTailer(n.Log.Files, opt, n.Log.IgnoreStatus)
	if err != nil {
		l.Errorf("NewTailer: %s", err)
		io.FeedLastError(inputName, err.Error())
		return
	}

	go n.tail.Start()
}

func (n *Input) Run() {
	l = logger.SLogger(inputName)
	l.Info("rabbitmq start")
	n.Interval.Duration = config.ProtectedInterval(minInterval, maxInterval, n.Interval.Duration)

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

			if n.semStopCompleted != nil {
				n.semStopCompleted.Close()
			}
			return

		case <-tick.C:

		case n.pause = <-n.pauseCh:
			// nil
		}
	}
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

		// wait stop completed
		if n.semStopCompleted != nil {
			for range n.semStopCompleted.Wait() {
				return
			}
		}
	}
}

type MetricFunc func(n *Input)

func (n *Input) getMetric() {
	n.start = time.Now()
	getFunc := []MetricFunc{getOverview, getNode, getQueues, getExchange}
	n.wg.Add(len(getFunc))
	for _, v := range getFunc {
		go func(gf MetricFunc) {
			defer n.wg.Done()
			gf(n)
		}(v)
	}
	n.wg.Wait()
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

			semStop:          cliutils.NewSem(),
			semStopCompleted: cliutils.NewSem(),
		}
		return s
	})
}
