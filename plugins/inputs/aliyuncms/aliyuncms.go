package aliyuncms

import (
	"context"

	"sync"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/selfstat"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/models"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"

	"github.com/aliyun/alibaba-cloud-sdk-go/services/cms"
)

var (
	batchInterval = 5 * time.Minute
	//metricPeriod  = time.Duration(5 * time.Minute)
	rateLimit = 20

	dms = []string{
		"instanceId",
		"userId",
		"consumerGroup",
		"topic",
		"BucketName",
		"storageType",
		"Host",
		"tableSchema",
		"Status",
		"workerId",
		"apiUid",
		"projectName",
		"jobName",
		"ip",
		"port",
		"protocol",
		"vip",
		"groupId",
		"clusterId",
		"nodeIP",
		"vbrInstanceId",
		"cenId",
		"serviceId",
		"diskname",
		"mountpoint",
		"state",
		"processName",
		"period",
		"device",
		"gpuId",
		"role",
		"appId",
		"direction",
		"pipelineId",
		"domain",
		"appName",
		"serviceName",
		"functionName",
		"podId",
		"subinstanceId",
		"dspId",
		"sspId",
		"logstore",
		"project",
		"alarm_type",
		"queue",
		"regionName",
		"SubscriptionName",
	}
)

type (
	runningCMS struct {
		cfg   *CMS
		agent *cmsAgent

		client *cms.Client

		timer *time.Timer

		wg sync.WaitGroup

		ctx    context.Context
		logger *models.Logger
	}

	cmsAgent struct {
		CMSs []*CMS `toml:"cms"`

		runningCms []*runningCMS

		ctx       context.Context
		cancelFun context.CancelFunc

		logger *models.Logger

		accumulator telegraf.Accumulator
	}
)

func (c *cmsAgent) Init() error {

	c.logger = &models.Logger{
		Errs: selfstat.Register("gather", "errors", nil),
		Name: `aliyuncms`,
	}

	for _, item := range c.CMSs {
		for _, p := range item.Project {
			for _, m := range p.Metrics.MetricNames {
				req := cms.CreateDescribeMetricListRequest()
				req.Scheme = "https"
				req.RegionId = item.RegionID
				req.Period = p.getPeriod(m)
				req.MetricName = m
				req.Namespace = p.Name
				if ds, err := p.genDimension(m, c.logger); err == nil {
					req.Dimensions = ds
				}
				MetricsRequests = append(MetricsRequests, &MetricsRequest{
					q: req,
				})
			}
		}
	}

	return nil
}

func (_ *cmsAgent) SampleConfig() string {
	return aliyuncmsConfigSample
}

func (_ *cmsAgent) Description() string {
	return ""
}

func (_ *cmsAgent) Gather(telegraf.Accumulator) error {
	return nil
}

func (ac *cmsAgent) Start(acc telegraf.Accumulator) error {

	if len(ac.CMSs) == 0 {
		ac.logger.Warnf("no configuration found")
		return nil
	}

	ac.logger.Info("starting...")

	ac.accumulator = acc

	for _, c := range ac.CMSs {
		rc := &runningCMS{
			agent:  ac,
			cfg:    c,
			ctx:    ac.ctx,
			logger: ac.logger,
		}
		ac.runningCms = append(ac.runningCms, rc)

		go rc.run(ac.ctx)
	}

	return nil
}

func (a *cmsAgent) Stop() {
	a.cancelFun()
}

func init() {
	inputs.Add("aliyuncms", func() telegraf.Input {
		ac := &cmsAgent{}
		ac.ctx, ac.cancelFun = context.WithCancel(context.Background())
		return ac
	})
}
