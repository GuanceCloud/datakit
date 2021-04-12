package rabbitmq

import (
	"path/filepath"
	"time"

	"gitlab.jiagouyun.com/cloudcare-tools/cliutils/logger"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
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

	if n.Log != nil {
		go func() {
			if err := n.Log.Init(); err != nil {
				l.Errorf("rabbitmq init tailf err:%s", err.Error())
				return
			}
			if n.Log.Option.Pipeline != "" {
				n.Log.Option.Pipeline = filepath.Join(datakit.PipelineDir, n.Log.Option.Pipeline)
			} else {
				n.Log.Option.Pipeline = filepath.Join(datakit.PipelineDir, "rabbitmq.p")
			}
			n.Log.Option.Source = inputName
			n.Log.Run()
		}()
	}

	client, err := n.createHttpClient()
	if err != nil {
		l.Errorf("[error] rabbitmq init client err:%s", err.Error())
		return
	}
	n.client = client
	if n.Interval.Duration == 0 {
		n.Interval.Duration = time.Second * 30
	}

	tick := time.NewTicker(n.Interval.Duration)
	defer tick.Stop()
	cleanCacheTick := time.NewTicker(time.Second * 5)
	defer cleanCacheTick.Stop()

	for {
		select {
		case <-tick.C:
			n.getMetric()
		case <-cleanCacheTick.C:
			if len(collectCache) > 0 {
				err := inputs.FeedMeasurement(inputName, io.Metric, collectCache, &io.Option{CollectCost: time.Since(n.start)})
				collectCache = collectCache[:]
				if err != nil {
					l.Errorf(err.Error())
					continue
				}
			}
		case <-datakit.Exit.Wait():
			if n.Log != nil {
				n.Log.Close()
				l.Info("rabbitmq log exit")
			}
			l.Info("rabbitmq exit")
			return
		}
	}
}

type MetricFunc func(n *Input)

func (n *Input) getMetric() {
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
		s := &Input{}
		return s
	})
}
