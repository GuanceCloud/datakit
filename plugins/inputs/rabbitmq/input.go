package rabbitmq

import (
	"time"

	"gitlab.jiagouyun.com/cloudcare-tools/cliutils/logger"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/config"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"
)

func (_ *Input) SampleConfig() string {
	return sample
}

func (_ *Input) Catalog() string {
	return inputName
}

func (_ *Input) AvailableArchs() []string {
	return datakit.AllArch
}

func (_ *Input) PipelineConfig() map[string]string {
	pipelineMap := map[string]string{
		"rabbitmq": pipelineCfg,
	}
	return pipelineMap
}

func (n *Input) Run() {
	l = logger.SLogger(inputName)
	l.Info("rabbitmq start")
	n.Interval.Duration = config.ProtectedInterval(minInterval, maxInterval, n.Interval.Duration)

	if n.Log != nil {
		go func() {
			inputs.JoinPipelinePath(n.Log, "rabbitmq.p")
			n.Log.Source = inputName
			n.Log.Tags = map[string]string{}
			for k, v := range n.Tags {
				n.Log.Tags[k] = v
			}

			tail, err := inputs.NewTailer(n.Log)
			if err != nil {
				l.Errorf("init tailf err:%s", err.Error())
				return
			}
			n.tail = tail
			tail.Run()
		}()
	}

	client, err := n.createHttpClient()
	if err != nil {
		l.Errorf("[error] rabbitmq init client err:%s", err.Error())
		return
	}
	n.client = client

	tick := time.NewTicker(n.Interval.Duration)
	defer tick.Stop()

	for {
		select {
		case <-tick.C:
			n.getMetric()
			if len(collectCache) > 0 {
				err := inputs.FeedMeasurement(inputName, datakit.Metric, collectCache, &io.Option{CollectCost: time.Since(n.start)})
				collectCache = collectCache[:0]
				if err != nil {
					n.lastErr = err
					l.Errorf(err.Error())
					continue
				}
			}
			if n.lastErr != nil {
				io.FeedLastError(inputName, n.lastErr.Error())
				n.lastErr = nil
			}
		case <-datakit.Exit.Wait():
			if n.tail != nil {
				n.tail.Close()
				l.Info("rabbitmq log exit")
			}
			l.Info("rabbitmq exit")
			return
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

func init() {
	inputs.Add(inputName, func() inputs.Input {
		s := &Input{
			Interval: datakit.Duration{Duration: time.Second * 10},
		}
		return s
	})
}
