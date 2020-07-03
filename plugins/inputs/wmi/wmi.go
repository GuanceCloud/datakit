// +build windows

package wmi

import (
	"context"
	"sync"
	"time"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/models"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"
)

type (
	runningInstance struct {
		cfg   *Instance
		agent *WmiAgent

		logger *models.Logger
	}

	WmiAgent struct {
		Instances []*Instance `toml:"instances"`

		runningInstances []*runningInstance

		ctx       context.Context
		cancelFun context.CancelFunc

		logger *models.Logger

		wg sync.WaitGroup
	}
)

func (_ *WmiAgent) SampleConfig() string {
	return sampleConfig
}

// func (_ *WmiAgent) Description() string {
// 	return `Collect metrics from Windows WMI.`
// }

func (_ *WmiAgent) Catalog() string {
	return `wmi`
}

func (ag *WmiAgent) Run() {

	if len(ag.Instances) == 0 {
		ag.logger.Warnf("no configuration found")
		return
	}

	go func() {
		<-datakit.Exit.Wait()
		ag.cancelFun()
	}()

	for _, inst := range ag.Instances {
		if inst.MetricName == "" {
			inst.MetricName = "WMI"
		}

		rc := &runningInstance{
			agent:  ag,
			cfg:    inst,
			logger: ag.logger,
		}

		if rc.cfg.Interval.Duration == 0 {
			rc.cfg.Interval.Duration = time.Minute * 5
		}
		ag.runningInstances = append(ag.runningInstances, rc)

		ag.wg.Add(1)
		go func() {
			defer ag.wg.Done()
			rc.run(ag.ctx)
		}()
	}

	ag.wg.Wait()
}

func (r *runningInstance) run(ctx context.Context) error {

	for {

		select {
		case <-ctx.Done():
			return context.Canceled
		default:
		}

		for _, query := range r.cfg.Queries {

			select {
			case <-ctx.Done():
				return context.Canceled
			default:
			}

			if query.lastTime.IsZero() {
				query.lastTime = time.Now()
			} else {
				if time.Now().Sub(query.lastTime) < query.Interval.Duration {
					continue
				}
			}

			sql, err := query.ToSql()
			if err != nil {
				r.logger.Warnf("%s", err)
				continue
			}

			props := []string{}

			for _, ms := range query.Metrics {
				props = append(props, ms[0])
			}

			fieldsArr, err := DefaultClient.QueryEx(sql, props)
			if err != nil {
				r.logger.Errorf("query failed, %s", err)
				continue

			}

			for _, fields := range fieldsArr {
				io.FeedEx(io.Metric, r.cfg.MetricName, nil, fields)
			}

			query.lastTime = time.Now()
		}

		internal.SleepContext(ctx, time.Second)
	}
}

func NewAgent() *WmiAgent {
	ac := &WmiAgent{
		logger: &models.Logger{
			Name: inputName,
		},
	}
	ac.ctx, ac.cancelFun = context.WithCancel(context.Background())

	return ac
}

func init() {
	inputs.Add(inputName, func() inputs.Input {
		return NewAgent()
	})
}
