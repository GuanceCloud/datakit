package azurecms

import (
	"context"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/selfstat"

	"github.com/Azure/azure-sdk-for-go/services/preview/monitor/mgmt/2019-06-01/insights"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/models"
	//"github.com/Azure/go-autorest/tracing"
)

type (
	runningInstance struct {
		cfg   *azureInstance
		agent *azureMonitorAgent

		queryInfos []*queryListInfo

		timer *time.Timer

		metricDefinitionClient insights.MetricDefinitionsClient
		metricClient           insights.MetricsClient

		logger *models.Logger
	}

	azureMonitorAgent struct {
		Instances []*azureInstance

		runningInstances []*runningInstance

		ctx         context.Context
		cancelFun   context.CancelFunc
		logger      *models.Logger
		accumulator telegraf.Accumulator
	}
)

func (ag *azureMonitorAgent) Init() error {

	ag.logger = &models.Logger{
		Errs: selfstat.Register("gather", "errors", nil),
		Name: `azurecms`,
	}

	return nil
}

func (_ *azureMonitorAgent) SampleConfig() string {
	return sampleConfig
}

func (_ *azureMonitorAgent) Description() string {
	return ""
}

func (_ *azureMonitorAgent) Gather(telegraf.Accumulator) error {
	return nil
}

func (ag *azureMonitorAgent) Start(acc telegraf.Accumulator) error {

	if len(ag.Instances) == 0 {
		ag.logger.Warnf("no configuration found")
		return nil
	}

	ag.logger.Info("starting...")

	ag.accumulator = acc

	for _, c := range ag.Instances {
		rc := &runningInstance{
			agent:  ag,
			cfg:    c,
			logger: ag.logger,
		}
		ag.runningInstances = append(ag.runningInstances, rc)

		go rc.run(ag.ctx)
	}

	return nil
}

func (ag *azureMonitorAgent) Stop() {
	ag.cancelFun()
}
