package aliyunobject

import (
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
	return sampleConfig + ecsSampleConfig + slbSampleConfig + ossSampleConfig + rdsSampleConfig
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

	ag.addModule(ag.Ecs)
	ag.addModule(ag.Slb)
	ag.addModule(ag.Oss)
	ag.addModule(ag.Rds)

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
