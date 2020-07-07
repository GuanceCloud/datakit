package azurecms

import (
	"context"

	"gitlab.jiagouyun.com/cloudcare-tools/cliutils/logger"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"
	"go.uber.org/zap"
	//"github.com/Azure/go-autorest/tracing"
)

var (
	inputName    = `azure_monitor`
	moduleLogger *zap.SugaredLogger
)

func (_ *azureInstance) Catalog() string {
	return "azure"
}

func (_ *azureInstance) SampleConfig() string {
	return sampleConfig
}

// func (_ *azureMonitorAgent) Description() string {
// 	return ""
// }

func (ag *azureInstance) Run() {

	moduleLogger = logger.SLogger(inputName)

	go func() {
		<-datakit.Exit.Wait()
		ag.cancelFun()
	}()

	ag.run(ag.ctx)
}

func init() {
	inputs.Add(inputName, func() inputs.Input {
		ac := &azureInstance{}
		ac.ctx, ac.cancelFun = context.WithCancel(context.Background())
		return ac
	})
}
