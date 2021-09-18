// +build windows

package wmi

import (
	"context"
	"os"
	"time"

	"gitlab.jiagouyun.com/cloudcare-tools/cliutils/logger"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"
)

var moduleLogger *logger.Logger

func (_ *Instance) SampleConfig() string {
	return sampleConfig
}

func (_ *Instance) Catalog() string {
	return `windows`
}

func (ag *Instance) Run() {
	moduleLogger = logger.SLogger(inputName)

	go func() {
		<-datakit.Exit.Wait()
		ag.cancelFun()
	}()

	if ag.MetricName == "" {
		ag.MetricName = "WMI"
	}

	if ag.Interval.Duration == 0 {
		ag.Interval.Duration = time.Minute * 5
	}

	ag.run(ag.ctx)
}

func (r *Instance) run(ctx context.Context) error {
	for {

		select {
		case <-ctx.Done():
			return context.Canceled
		default:
		}

		for _, query := range r.Queries {

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
				moduleLogger.Warnf("%s", err)
				continue
			}

			props := []string{}

			for _, ms := range query.Metrics {
				props = append(props, ms[0])
			}

			fieldsArr, err := DefaultClient.QueryEx(sql, props)
			if err != nil {
				moduleLogger.Errorf("query failed, %s", err)
				continue

			}

			hostname, _ := os.Hostname()

			tags := map[string]string{
				"host": hostname,
			}

			for k, v := range r.Tags {
				tags[k] = v
			}

			for _, fields := range fieldsArr {
				if r.isTest() {
					// pass
				} else {
					io.NamedFeedEx(inputName, datakit.Metric, r.MetricName, tags, fields)
				}
			}

			query.lastTime = time.Now()
		}

		if r.isTest() {
			return nil
		}
		datakit.SleepContext(ctx, time.Second)
	}
}

func NewAgent() *Instance {
	ac := &Instance{}
	ac.ctx, ac.cancelFun = context.WithCancel(context.Background())
	return ac
}

func init() {
	inputs.Add(inputName, func() inputs.Input {
		return NewAgent()
	})
}
