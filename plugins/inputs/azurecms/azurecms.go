package azurecms

import (
	"context"
	"sync"

	"github.com/influxdata/telegraf/selfstat"

	"github.com/Azure/azure-sdk-for-go/services/preview/monitor/mgmt/2019-06-01/insights"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/config"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/models"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"
	//"github.com/Azure/go-autorest/tracing"
)

type (
	runningInstance struct {
		cfg   *azureInstance
		agent *azureMonitorAgent

		queryInfos []*queryListInfo

		metricDefinitionClient insights.MetricDefinitionsClient
		metricClient           insights.MetricsClient

		logger *models.Logger
	}

	azureMonitorAgent struct {
		Instances []*azureInstance

		wg sync.WaitGroup

		ctx       context.Context
		cancelFun context.CancelFunc
		logger    *models.Logger
	}
)

func (ag *azureMonitorAgent) Init() error {

	ag.logger = &models.Logger{
		Errs: selfstat.Register("gather", "errors", nil),
		Name: `azurecms`,
	}

	return nil
}

func (_ *azureMonitorAgent) Catalog() string {
	return "azure"
}

func (_ *azureMonitorAgent) SampleConfig() string {
	return sampleConfig
}

// func (_ *azureMonitorAgent) Description() string {
// 	return ""
// }

func (ag *azureMonitorAgent) Run() {

	ag.logger = &models.Logger{
		Name: `azurecms`,
	}

	if len(ag.Instances) == 0 {
		ag.logger.Warnf("no configuration found")
		return
	}

	go func() {
		<-config.Exit.Wait()
		ag.cancelFun()
	}()

	for _, c := range ag.Instances {
		ag.wg.Add(1)

		go func(c *azureInstance) {
			defer ag.wg.Done()

			rc := &runningInstance{
				agent:  ag,
				cfg:    c,
				logger: ag.logger,
			}

			rc.run(ag.ctx)
		}(c)

	}

	ag.wg.Wait()
}

func init() {
	inputs.Add("azure_monitor", func() inputs.Input {
		ac := &azureMonitorAgent{}
		ac.ctx, ac.cancelFun = context.WithCancel(context.Background())
		return ac
	})
}
