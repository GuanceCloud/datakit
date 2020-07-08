package ucmon

import (
	"context"

	"go.uber.org/zap"

	"gitlab.jiagouyun.com/cloudcare-tools/cliutils/logger"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"
)

var (
	inputName    = `ucloud_monitor`
	moduleLogger *zap.SugaredLogger
)

func (_ *ucInstance) SampleConfig() string {
	return sampleConfig
}

func (_ *ucInstance) Catalog() string {
	return "ucloud"
}

// func (_ *ucMonitorAgent) Description() string {
// 	return ""
// }

func (ag *ucInstance) Run() {

	moduleLogger = logger.SLogger(inputName)

	go func() {
		<-datakit.Exit.Wait()
		ag.cancelFun()
	}()

	ag.run(ag.ctx)

}

func init() {
	inputs.Add(inputName, func() inputs.Input {
		ac := &ucInstance{}
		ac.ctx, ac.cancelFun = context.WithCancel(context.Background())
		return ac
	})
}
