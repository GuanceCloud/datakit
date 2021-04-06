package rabbitmq

import (
	"time"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
)

func (_ *Input) SampleConfig() string {
	return sample
}

func (_ *Input) Catalog() string {
	return inputName
}

func (n *Input) Run() {
	l.Info("rabbitmq start")
	client, err := n.createHttpClient()
	if err != nil {
		l.Errorf("[error] nginx init client err:%s", err.Error())
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
			if len(n.collectCache) > 0 {
				inputs.FeedMeasurement(inputName, io.Metric, n.collectCache, &io.Option{CollectCost: time.Since(n.start)})
				n.collectCache = n.collectCache[:]
			}
		case <-datakit.Exit.Wait():
			l.Info("nginx exit")
			return
		}
	}
}



func (n *Input) SampleMeasurement() []inputs.Measurement {
	return []inputs.Measurement{

	}
}





func init() {
	inputs.Add(inputName, func() inputs.Input {
		s := &Input{}
		return s
	})
}
