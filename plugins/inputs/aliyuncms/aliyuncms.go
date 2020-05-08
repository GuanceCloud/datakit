package aliyuncms

import (
	"context"
	"sync"
	"time"

	"github.com/influxdata/telegraf"

	"golang.org/x/time/rate"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/models"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"

	"github.com/aliyun/alibaba-cloud-sdk-go/services/cms"
)

var (
	inputName = `aliyuncms`

	//理论上每个指标应该使用DescribeMetricMetaList接口返回的对应dimension，但有时该接口更新不及时，返回的并不是最新的，所以在这里将所有支持的dimension列出，并保持更新
	supportedDimensions = []string{
		"instanceId",
		"device",
		"state",
		"port",
		"vip",
		"nodeId",
		"queue",
		"region",
		"userId",
		"clusterId",
		"dbInstanceId",
		"tableSchema",
		"workerId",
		"role",
		"serviceName",
		"functionName",
		"groupId",
		"jobId",
		"taskId",
		"project",
		"logstore",
		"serviceId",
		"VpnConnectionId",
		"cenId",
		"geographicSpanId",
		"localRegionId",
		"oppositeRegionId",
		"src_region_id",
		"dst_region_id",
		"vbrInstanceId",
		"sourceRegion",
		"regionId",
		"appName",
		"appId",
		"domain_name",
		"isp",
		"loc",
		"productKey",
		"nodeIP",
		"projectName",
		"jobName",
		"database",
		"vhostQueue",
		"vhostName",
		"topic",
		"dspId",
		"sspId",
		"domain",
		"schema",
		"pipelineId",
		"groupName",
		"DedicatedHostId",
		"eniId",
		"gatewayId",
		"serverId",
		"host",
		"consumerGroup",
		"BucketName",
		"storageType",
		"Host",
		"Status",
		"apiUid",
		"ip",
		"protocol",
		"diskname",
		"mountpoint",
		"processName",
		"period",
		"gpuId",
		"appId",
		"direction",
		"appName",
		"podId",
		"subinstanceId",
		"alarm_type",
		"regionName",
		"SubscriptionName",
	}
)

type (
	runningInstance struct {
		cfg   *CMS
		agent *CmsAgent

		cmsClient *cms.Client

		reqs []*MetricsRequest

		logger *models.Logger

		limiter *rate.Limiter
	}

	CmsAgent struct {
		CMSs       []*CMS `toml:"cms"`
		ReportStat bool   `toml:"report_stat"`

		runningInstances []*runningInstance

		ctx       context.Context
		cancelFun context.CancelFunc

		logger *models.Logger

		accumulator telegraf.Accumulator

		wg sync.WaitGroup

		succedRequest int64
		faildRequest  int64
	}
)

func (_ *CmsAgent) SampleConfig() string {
	return aliyuncmsConfigSample
}

func (_ *CmsAgent) Description() string {
	return `Collect metrics from alibaba Cloud Monitor Service.`
}

func (_ *CmsAgent) Gather(telegraf.Accumulator) error {
	return nil
}

func (ac *CmsAgent) Start(acc telegraf.Accumulator) error {

	if len(ac.CMSs) == 0 {
		ac.logger.Warnf("no configuration found")
		return nil
	}

	ac.logger.Info("start")

	ac.accumulator = acc

	for _, cfg := range ac.CMSs {
		rc := &runningInstance{
			agent:  ac,
			cfg:    cfg,
			logger: ac.logger,
		}
		if cfg.Delay.Duration == 0 {
			cfg.Delay.Duration = time.Minute * 5
		}

		if rc.cfg.Interval.Duration == 0 {
			rc.cfg.Interval.Duration = time.Minute * 5
		}
		ac.runningInstances = append(ac.runningInstances, rc)

		ac.wg.Add(1)
		go func() {
			defer ac.wg.Done()
			rc.run(ac.ctx)
		}()
	}

	return nil
}

func (a *CmsAgent) Stop() {
	a.cancelFun()
	a.wg.Wait()
	a.logger.Info("done")
}

func NewAgent() *CmsAgent {
	ac := &CmsAgent{
		logger: &models.Logger{
			Name: inputName,
		},
	}
	ac.ctx, ac.cancelFun = context.WithCancel(context.Background())

	return ac
}

func init() {
	inputs.Add(inputName, func() telegraf.Input {
		return NewAgent()
	})
}
