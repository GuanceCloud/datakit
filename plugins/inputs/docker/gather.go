package docker

import (
	"fmt"
	"time"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io"
)

type gatherOption struct {
	IsObjectCategory bool
}

func (this *Input) gatherMetric() {
	tick := time.NewTicker(this.metricDuration)
	defer tick.Stop()
	for {
		select {
		case <-datakit.Exit.Wait():
			return

		case <-tick.C:
			startTime := time.Now()
			pts, err := this.gather()
			if err != nil {
				l.Error(err)
				io.FeedLastError(inputName, fmt.Sprintf("gather metric: %s", err.Error()))
				continue
			}
			cost := time.Since(startTime)
			if err := io.Feed(inputName, datakit.Metric, pts, &io.Option{CollectCost: cost}); err != nil {
				l.Error(err)
				io.FeedLastError(inputName, fmt.Sprintf("gather metric: %s", err.Error()))
			}
		}
	}
}

func (this *Input) gatherObject() {
	objectFunc := func() {
		startTime := time.Now()
		pts, err := this.gather(&gatherOption{IsObjectCategory: true})
		if err != nil {
			l.Error(err)
			io.FeedLastError(inputName, fmt.Sprintf("gather object: %s", err.Error()))
			return
		}
		cost := time.Since(startTime)
		if err := io.Feed(inputName, datakit.Object, pts, &io.Option{CollectCost: cost}); err != nil {
			l.Error(err)
			io.FeedLastError(inputName, fmt.Sprintf("gather object: %s", err.Error()))
		}
	}

	// 在即进入 for tick 之前，先执行一次，避免等待太久
	objectFunc()

	tick := time.NewTicker(this.objectDuration)
	defer tick.Stop()

	for {
		select {
		case <-datakit.Exit.Wait():
			return

		case <-tick.C:
			objectFunc()
		}
	}
}

func (this *Input) gatherLoggoing() {
	// 定期发现新容器，从而获取其日志数据
	tick := time.NewTicker(this.loggingHitDuration)
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
