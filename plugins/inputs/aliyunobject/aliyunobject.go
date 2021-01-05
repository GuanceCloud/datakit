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
	sampleConf   = ""
)

type subModule interface {
	run(*objectAgent)
	disabled() bool
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

func (ag *objectAgent) Test() (*inputs.TestResult, error) {
	ag.mode = "test"
	ag.testResult = &inputs.TestResult{}
	ag.Run()
	return ag.testResult, ag.testError
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

	if ag.Ecs == nil {
		ag.Ecs = &Ecs{}
	}
	if ag.Slb != nil {
		ag.Slb = &Slb{}
	}
	if ag.Oss == nil {
		ag.Oss = &Oss{}
	}
	if ag.Rds == nil {
		ag.Rds = &Rds{}
	}
	if ag.Ons == nil {
		ag.Ons = &Ons{}
	}
	if ag.Dds == nil {
		ag.Dds = &Dds{}
	}
	if ag.Domain == nil {
		ag.Domain = &Domain{}
	}
	if ag.Redis == nil {
		ag.Redis = &Redis{}
	}
	if ag.Cdn == nil {
		ag.Cdn = &Cdn{}
	}
	if ag.Waf == nil {
		ag.Waf = &Waf{}
	}
	if ag.Es == nil {
		ag.Es = &Elasticsearch{}
	}
	if ag.InfluxDB != nil {
		ag.InfluxDB = &InfluxDB{}
	}

	ag.addModule(ag.Ecs)
	ag.addModule(ag.Slb)
	ag.addModule(ag.Oss)
	ag.addModule(ag.Rds)
	ag.addModule(ag.Ons)
	ag.addModule(ag.Dds)
	ag.addModule(ag.Domain)
	ag.addModule(ag.Redis)
	ag.addModule(ag.Cdn)
	ag.addModule(ag.Waf)
	ag.addModule(ag.Es)
	ag.addModule(ag.InfluxDB)

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
