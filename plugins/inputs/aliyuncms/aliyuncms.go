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
	batchInterval = time.Duration(5) * time.Minute
	metricPeriod  = time.Duration(5 * time.Minute)
	rateLimit     = 20

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
	RunningCMS struct {
		cfg *CMS

		client *cms.Client

		timer *time.Timer

		wg sync.WaitGroup

		aliCMS *AliCMS

		ctx context.Context

		logger *models.Logger
	}

	AliCMS struct {
		CMSs []*CMS `toml:"cms"`

		runningCms []*RunningCMS

		tags map[string]string

		ctx       context.Context
		cancelFun context.CancelFunc

		accumulator telegraf.Accumulator

		logger *models.Logger
	}
)

func (c *AliCMS) Init() error {

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
				req.Period = getPeriod(p.Name, m)
				req.MetricName = m
				req.Namespace = p.Name
				if ds, err := p.GenDimension(m); err == nil {
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

func (_ *AliCMS) SampleConfig() string {
	return aliyuncmsConfigSample
}

func (_ *AliCMS) Description() string {
	return ""
}

func (_ *AliCMS) Gather(telegraf.Accumulator) error {
	return nil
}

func (ac *AliCMS) Start(acc telegraf.Accumulator) error {

	if len(ac.CMSs) == 0 {
		ac.logger.Warnf("no configuration found")
		return nil
	}

	ac.logger.Info("starting...")

	ac.accumulator = acc

	for _, c := range ac.CMSs {
		rc := &RunningCMS{
			aliCMS: ac,
			cfg:    c,
			ctx:    ac.ctx,
			logger: ac.logger,
		}
		ac.runningCms = append(ac.runningCms, rc)
	}

	for _, rc := range ac.runningCms {

		go func(r *RunningCMS) {

			if err := r.run(); err != nil && err != context.Canceled {
				ac.logger.Errorf("%s", err)
			}

			r.logger.Infof("%s done", r.cfg.AccessKeyID)

		}(rc)

	}

	return nil
}

func (ac *AliCMS) Stop() {
	ac.cancelFun()
}

func init() {
	inputs.Add("aliyuncms", func() telegraf.Input {
		ac := &AliCMS{}
		ac.ctx, ac.cancelFun = context.WithCancel(context.Background())
		return ac
	})
}
