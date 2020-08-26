package tencentobject

import (
	"bytes"
	"context"
	"time"

	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common"

	"gitlab.jiagouyun.com/cloudcare-tools/cliutils/logger"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"
)

var (
	inputName    = `tencentobject`
	moduleLogger *logger.Logger
)

type subModule interface {
	run(*objectAgent)
}

func (_ *objectAgent) SampleConfig() string {
	var buf bytes.Buffer
	buf.WriteString(sampleConfig)
	buf.WriteString(cvmSampleConfig)
	buf.WriteString(cosSampleConfig)
	buf.WriteString(clbSampleConfig)
	buf.WriteString(cdbSampleConfig)
	// buf.WriteString(redisSampleConfig)
	// buf.WriteString(cdnSampleConfig)
	// buf.WriteString(wafSampleConfig)
	// buf.WriteString(elasticsearchSampleConfig)
	// buf.WriteString(influxDBSampleConfig)
	return buf.String()
}

func (_ *objectAgent) Catalog() string {
	return `tencentcloud`
}

func (ag *objectAgent) getCredential() *common.Credential {
	credential := common.NewCredential(
		ag.AccessKeyID,
		ag.AccessKeySecret,
	)
	return credential
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

	if ag.Cvm != nil {
		ag.addModule(ag.Cvm)
	}
	if ag.Cos != nil {
		ag.addModule(ag.Cos)
	}
	if ag.Clb != nil {
		ag.addModule(ag.Clb)
	}
	if ag.Cdb != nil {
		ag.addModule(ag.Cdb)
	}
	// if ag.Redis != nil {
	// 	ag.addModule(ag.Redis)
	// }
	// if ag.Cdn != nil {
	// 	ag.addModule(ag.Cdn)
	// }
	// if ag.Waf != nil {
	// 	ag.addModule(ag.Waf)
	// }
	// if ag.Es != nil {
	// 	ag.addModule(ag.Es)
	// }
	// if ag.InfluxDB != nil {
	// 	ag.addModule(ag.InfluxDB)
	// }

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
