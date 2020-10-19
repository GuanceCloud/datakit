package aliyuncms

import (
	"context"
	"time"

	"golang.org/x/time/rate"

	"gitlab.jiagouyun.com/cloudcare-tools/cliutils/logger"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
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

var moduleLogger *logger.Logger

type (
	runningInstance struct {
		cfg *CMS

		cmsClient *cms.Client

		reqs []*MetricsRequest

		limiter *rate.Limiter
	}
)

func (_ *CMS) SampleConfig() string {
	return aliyuncmsConfigSample
}

func (ac *CMS) Test() (*inputs.TestResult, error) {
	ac.mode = "test"
	ac.testResult = &inputs.TestResult{}
	ac.Run()
	return ac.testResult, ac.testError
}

func (_ *CMS) Catalog() string {
	return `aliyun`
}

func (ac *CMS) Run() {

	moduleLogger = logger.SLogger(inputName)

	go func() {
		<-datakit.Exit.Wait()
		ac.cancelFun()
	}()

	if ac.Delay.Duration == 0 {
		ac.Delay.Duration = time.Minute * 5
	}

	if ac.Interval.Duration == 0 {
		ac.Interval.Duration = time.Minute * 5
	}

	rc := &runningInstance{
		cfg: ac,
	}

	rc.run(ac.ctx)
}

func NewAgent() *CMS {
	ac := &CMS{}
	ac.ctx, ac.cancelFun = context.WithCancel(context.Background())
	return ac
}

func init() {
	inputs.Add(inputName, func() inputs.Input {
		return NewAgent()
	})
}
