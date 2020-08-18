package aliyunobject

import (
	"bytes"
	"context"
	"time"

	"gitlab.jiagouyun.com/cloudcare-tools/cliutils/logger"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"
)

var (
	inputName    = `aliyunobject`
	moduleLogger *logger.Logger
)

type subModule interface {
	run(*objectAgent)
}

func (_ *objectAgent) SampleConfig() string {
	var buf bytes.Buffer
	buf.WriteString(sampleConfig)
	buf.WriteString(ecsSampleConfig)
	buf.WriteString(slbSampleConfig)
	buf.WriteString(ossSampleConfig)
	buf.WriteString(rdsSampleConfig)
	buf.WriteString(redisSampleConfig)
	buf.WriteString(cdnSampleConfig)
	buf.WriteString(wafSampleConfig)
	buf.WriteString(elasticsearchSampleConfig)
	buf.WriteString(influxDBSampleConfig)
	buf.WriteString(onsSampleConfig)
	buf.WriteString(domainSampleConfig)
	buf.WriteString(ddsSampleConfig)
	return buf.String()
}

func (_ *objectAgent) Catalog() string {
	return `aliyun`
}

func (ag *objectAgent) Run() {

	moduleLogger = logger.SLogger(inputName)

	ag.ctx, ag.cancelFun = context.WithCancel(context.Background())

	go func() {
		<-datakit.Exit.Wait()
		ag.cancelFun()
	}()

	if ag.Interval.Duration == 0 {
		ag.Interval.Duration = time.Minute * 5
	}

	if ag.Ecs != nil {
		ag.addModule(ag.Ecs)
	}
	if ag.Slb != nil {
		ag.addModule(ag.Slb)
	}
	if ag.Oss != nil {
		ag.addModule(ag.Oss)
	}
	if ag.Rds != nil {
		ag.addModule(ag.Rds)
	}

	if ag.Ons != nil {
		ag.addModule(ag.Ons)
	}
	if ag.Dds != nil {
		ag.addModule(ag.Dds)
	}
	if ag.Domain != nil {
		ag.addModule(ag.Domain)
	}
	if ag.Redis != nil {
		ag.addModule(ag.Redis)
	}
	if ag.Cdn != nil {
		ag.addModule(ag.Cdn)
	}
	if ag.Waf != nil {
		ag.addModule(ag.Waf)
	}
	if ag.Es != nil {
		ag.addModule(ag.Es)
	}
	if ag.InfluxDB != nil {
		ag.addModule(ag.InfluxDB)

	}

	for _, s := range ag.subModules {
		ag.wg.Add(1)
		go func(s subModule) {
			defer ag.wg.Done()
			s.run(ag)
		}(s)
	}

	ag.wg.Wait()

	moduleLogger.Debugf("done")
}

func newAgent() *objectAgent {
	ag := &objectAgent{}
	return ag
}

func init() {
	inputs.Add(inputName, func() inputs.Input {
		return newAgent()
	})
}
