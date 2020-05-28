package baiduIndex

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/selfstat"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/models"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"
)

type Baidu struct {
	BaiduIndex       []*BaiduIndexCfg
	runningInstances []*runningInstance
	ctx              context.Context
	cancelFun        context.CancelFunc
	accumulator      telegraf.Accumulator
	logger           *models.Logger
}

type runningInstance struct {
	cfg        *BaiduIndexCfg
	agent      *Baidu
	logger     *models.Logger
	metricName string
}

func (_ *Baidu) SampleConfig() string {
	return baiduIndexConfigSample
}

func (_ *Baidu) Description() string {
	return ""
}

func (_ *Baidu) Gather(telegraf.Accumulator) error {
	return nil
}

func (b *Baidu) Start(acc telegraf.Accumulator) error {
	if len(b.BaiduIndex) == 0 {
		log.Printf("W! [baiduIndex] no configuration found")
		return nil
	}

	b.logger = &models.Logger{
		Errs: selfstat.Register("gather", "errors", nil),
		Name: `baiduIndex`,
	}

	log.Printf("baiduIndex cdn start")

	b.accumulator = acc

	for _, instCfg := range b.BaiduIndex {
		r := &runningInstance{
			cfg:    instCfg,
			agent:  b,
			logger: b.logger,
		}

		r.metricName = instCfg.MetricName
		if r.metricName == "" {
			r.metricName = "baiduIndex"
		}

		if r.cfg.Interval.Duration == 0 {
			r.cfg.Interval.Duration = time.Minute * 10
		}

		b.runningInstances = append(b.runningInstances, r)

		go r.run(b.ctx)
	}

	return nil
}

func (b *Baidu) Stop() {
	b.cancelFun()
}

func (r *runningInstance) run(ctx context.Context) error {
	defer func() {
		if e := recover(); e != nil {

		}
	}()

	// go r.getHistory(ctx)
	for {
		select {
		case <-ctx.Done():
			return context.Canceled
		default:
		}

		fmt.Println("开始采集百度指数数据。。。")
		internal.SleepContext(ctx, r.cfg.Interval.Duration)
	}

	return nil
}

func init() {
	inputs.Add("baiduIndex", func() telegraf.Input {
		ac := &Baidu{}
		ac.ctx, ac.cancelFun = context.WithCancel(context.Background())
		return ac
	})
}
