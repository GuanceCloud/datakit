package ucmon

import (
	"context"
	"sync"

	"github.com/ucloud/ucloud-sdk-go/ucloud"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/models"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"
)

type (
	runningInstance struct {
		cfg   *ucInstance
		agent *ucMonitorAgent

		queryInfos []*queryListInfo

		//timer *time.Timer

		ucCli *ucloud.Client

		logger *models.Logger
	}

	ucMonitorAgent struct {
		Instances []*ucInstance

		runningInstances []*runningInstance

		ctx       context.Context
		cancelFun context.CancelFunc
		logger    *models.Logger

		wg sync.WaitGroup
	}
)

func (ag *ucMonitorAgent) initialize() error {

	ag.logger = &models.Logger{
		Name: `ucmon`,
	}

	return nil
}

func (_ *ucMonitorAgent) SampleConfig() string {
	return sampleConfig
}

func (_ *ucMonitorAgent) Catalog() string {
	return "ucloud"
}

// func (_ *ucMonitorAgent) Description() string {
// 	return ""
// }

func (ag *ucMonitorAgent) Run() {

	ag.initialize()

	if len(ag.Instances) == 0 {
		ag.logger.Warnf("no configuration found")
		return
	}

	go func() {
		<-datakit.Exit.Wait()
		ag.cancelFun()
	}()

	for _, c := range ag.Instances {

		ag.wg.Add(1)
		go func(c *ucInstance) {
			defer ag.wg.Done()

			rc := &runningInstance{
				agent:  ag,
				cfg:    c,
				logger: ag.logger,
			}
			ag.runningInstances = append(ag.runningInstances, rc)

			rc.run(ag.ctx)
		}(c)

	}

	ag.wg.Done()

}

func init() {
	inputs.Add("ucloud_monitor", func() inputs.Input {
		ac := &ucMonitorAgent{}
		ac.ctx, ac.cancelFun = context.WithCancel(context.Background())
		return ac
	})
}
