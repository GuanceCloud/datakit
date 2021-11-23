//go:build windows
// +build windows

package wmi

import (
	"context"
	"os"
	"time"

	"gitlab.jiagouyun.com/cloudcare-tools/cliutils"
	"gitlab.jiagouyun.com/cloudcare-tools/cliutils/logger"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"
)

var moduleLogger *logger.Logger

func (*Instance) SampleConfig() string {
	return sampleConfig
}

func (*Instance) Catalog() string {
	return `windows`
}

func (ag *Instance) Run() {
	moduleLogger = logger.SLogger(inputName)

	go func() {
		select {
		case <-datakit.Exit.Wait():
		case <-ag.semStop.Wait():
		}
		ag.exit()
		ag.cancelFun()
	}()

	if ag.MetricName == "" {
		ag.MetricName = "WMI"
	}

	if ag.Interval.Duration == 0 {
		ag.Interval.Duration = time.Minute * 5
	}

	_ = ag.run(ag.ctx)
}

func (ag *Instance) exit() {
	ag.cancelFun()
}

func (ag *Instance) Terminate() {
	if ag.semStop != nil {
		ag.semStop.Close()
	}
}

func (ag *Instance) run(ctx context.Context) error {
	for {
		select {
		case <-ctx.Done():
			return context.Canceled
		default:
		}

		for _, query := range ag.Queries {
			select {
			case <-ctx.Done():
				return context.Canceled
			default:
			}

			if query.lastTime.IsZero() {
				query.lastTime = time.Now()
			} else if time.Since(query.lastTime) < query.Interval.Duration {
				continue
			}

			sql, err := query.ToSQL()
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

			for k, v := range ag.Tags {
				tags[k] = v
			}

			for _, fields := range fieldsArr {
				if ag.isTest() {
					// pass
				} else {
					_ = io.NamedFeedEx(inputName, datakit.Metric, ag.MetricName, tags, fields)
				}
			}

			query.lastTime = time.Now()
		}

		if ag.isTest() {
			return nil
		}
		_ = datakit.SleepContext(ctx, time.Second)
	}
}

func NewAgent() *Instance {
	ac := &Instance{}
	ac.ctx, ac.cancelFun = context.WithCancel(context.Background())
	ac.semStop = cliutils.NewSem()
	return ac
}

func init() { //nolint:gochecknoinits
	inputs.Add(inputName, func() inputs.Input {
		return NewAgent()
	})
}
