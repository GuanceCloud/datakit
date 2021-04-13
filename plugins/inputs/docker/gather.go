package docker

import (
	"time"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io"
)

type gatherOption struct {
	IsObjectCategory bool
}

func (this *Input) gatherMetric(interval time.Duration) {
	tick := time.NewTicker(interval)
	defer tick.Stop()
	for {
		select {
		case <-datakit.Exit.Wait():
			return

		case <-tick.C:
			data, err := this.gather(&gatherOption{})
			if err != nil {
				l.Error(err)
			}
			if err := io.NamedFeed(data, io.Metric, inputName); err != nil {
				l.Error(err)
			}
		}
	}
}

func (this *Input) gatherObject(interval time.Duration) {
	tick := time.NewTicker(interval)
	defer tick.Stop()
	for {
		select {
		case <-datakit.Exit.Wait():
			return

		case <-tick.C:
			data, err := this.gather(&gatherOption{IsObjectCategory: true})
			if err != nil {
				l.Error(err)
			}
			if err := io.NamedFeed(data, io.Object, inputName); err != nil {
				l.Error(err)
			}
		}
	}
}

func (this *Input) gatherLoggoing(hitInterval time.Duration) {
	// 定期发现新容器，从而获取其日志数据
	tick := time.NewTicker(hitInterval)
	defer tick.Stop()
	for {
		select {
		case <-datakit.Exit.Wait():
			this.cancelTails()
			this.wg.Wait()
			return

		case <-tick.C:
			this.gatherLog()
		}
	}
}
