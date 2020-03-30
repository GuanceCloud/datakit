package ucmon

import (
	"context"
	"sync"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/selfstat"

	"github.com/ucloud/ucloud-sdk-go/ucloud"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/models"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"
)

type (
	runningInstance struct {
		cfg   *ucInstance
		agent *ucMonitorAgent

		queryInfos []*queryListInfo

		timer *time.Timer

		ucCli *ucloud.Client

		logger *models.Logger
	}

	ucMonitorAgent struct {
		Instances []*ucInstance

		runningInstances []*runningInstance

		ctx         context.Context
		cancelFun   context.CancelFunc
		logger      *models.Logger
		accumulator telegraf.Accumulator

		once sync.Once
	}
)

func (ag *ucMonitorAgent) Init() error {

	ag.logger = &models.Logger{
		Errs: selfstat.Register("gather", "errors", nil),
		Name: `azurecms`,
	}

	return nil
}

func (_ *ucMonitorAgent) SampleConfig() string {
	return sampleConfig
}

func (_ *ucMonitorAgent) Description() string {
	return ""
}

func (_ *ucMonitorAgent) Gather(telegraf.Accumulator) error {
	return nil
}

func (ag *ucMonitorAgent) Start(acc telegraf.Accumulator) error {

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

func (ag *ucMonitorAgent) stop() {
	ag.cancelFun()
}

func (ag *ucMonitorAgent) Stop() {
	ag.once.Do(ag.stop)
}

func init() {
	inputs.Add("ucloud_monitor", func() telegraf.Input {
		ac := &ucMonitorAgent{}
		ac.ctx, ac.cancelFun = context.WithCancel(context.Background())
		return ac
	})
}
