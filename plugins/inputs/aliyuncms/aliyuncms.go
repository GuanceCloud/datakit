package aliyuncms

import (
	"context"
	"time"

	"gitlab.jiagouyun.com/cloudcare-tools/cliutils/logger"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"
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

	valueKeys = []string{
		"Average",
		"Minimum",
		"Maximum",
		"Value",
		"Sum",
		"SumPerMinute",
	}
)

var moduleLogger *logger.Logger

func (_ *CMS) SampleConfig() string {
	return aliyuncmsConfigSample
}

func (_ *CMS) Catalog() string {
	return `aliyun`
}

func (c *CMS) Run() {

	moduleLogger = logger.SLogger(inputName)

	c.apiCallInfo = &CloudApiCallInfo{
		details: map[string][]uint64{},
	}

	go func() {
		<-datakit.Exit.Wait()
		c.cancelFun()
	}()

	if c.Delay.Duration == 0 {
		c.Delay.Duration = time.Minute * 5
	}

	if c.Interval.Duration == 0 {
		c.Interval.Duration = time.Minute * 5
	}

	c.run(c.ctx)
}

func NewAgent(mode string) *CMS {
	ac := &CMS{}
	ac.mode = mode
	ac.ctx, ac.cancelFun = context.WithCancel(context.Background())
	return ac
}

func init() {
	inputs.Add(inputName, func() inputs.Input {
		return NewAgent("")
	})
}
