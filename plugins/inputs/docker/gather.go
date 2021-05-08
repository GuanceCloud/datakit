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

func (this *Input) gatherMetric(interval time.Duration) {
	tick := time.NewTicker(interval)
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

func (this *Input) gatherObject(interval time.Duration) {
	tick := time.NewTicker(interval)
	defer tick.Stop()
	for {
		select {
		case <-datakit.Exit.Wait():
			return

		case <-tick.C:
			startTime := time.Now()
			pts, err := this.gather(&gatherOption{IsObjectCategory: true})
			if err != nil {
				l.Error(err)
				io.FeedLastError(inputName, fmt.Sprintf("gather object: %s", err.Error()))
				continue
			}
			cost := time.Since(startTime)
			if err := io.Feed(inputName, datakit.Object, pts, &io.Option{CollectCost: cost}); err != nil {
				l.Error(err)
				io.FeedLastError(inputName, fmt.Sprintf("gather object: %s", err.Error()))
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
