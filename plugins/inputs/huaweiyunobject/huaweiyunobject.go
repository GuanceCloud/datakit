package huaweiyunobject

import (
	"bytes"
	"context"
	"time"

	"gitlab.jiagouyun.com/cloudcare-tools/cliutils/logger"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"
)

var (
	inputName    = `huaweiyunobject`
	moduleLogger *logger.Logger
)

type subModule interface {
	run(*objectAgent)
}

func (_ *objectAgent) SampleConfig() string {
	var buf bytes.Buffer
	buf.WriteString(sampleConfig)
	buf.WriteString(ecsSampleConfig)
	buf.WriteString(elbSampleConfig)
	buf.WriteString(obsSampleConfig)
	buf.WriteString(mysqlSampleConfig)
	return buf.String()
}

func (_ *objectAgent) Catalog() string {
	return `huaweiyun`
}

func (r *objectAgent) Test() (result *inputs.TestResult, err error) {
	return
}

func (ag *objectAgent) Run() {

	moduleLogger = logger.SLogger(inputName)

	ag.ctx, ag.cancelFun = context.WithCancel(context.Background())

	go func() {
		<-datakit.Exit.Wait()
		ag.cancelFun()
	}()

	if ag.Interval.Duration == 0 {
		ag.Interval.Duration = time.Hour * 6
	} else if ag.Interval.Duration < time.Hour*1 {
		ag.Interval.Duration = time.Hour * 1
	} else if ag.Interval.Duration > time.Hour*24 {
		ag.Interval.Duration = time.Hour * 24
	}

	if ag.Ecs != nil {
		ag.addModule(ag.Ecs)
	}
	if ag.Elb != nil {
		ag.addModule(ag.Elb)
	}
	if ag.Obs != nil {
		ag.addModule(ag.Obs)
	}
	if ag.Mysql != nil {
		ag.addModule(ag.Mysql)
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
