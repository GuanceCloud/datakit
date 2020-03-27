package aliyuncms

import (
	"context"
	"sync"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/metric"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/limiter"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/models"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"

	"github.com/aliyun/alibaba-cloud-sdk-go/services/cms"
)

var (
	inputName = `aliyuncms`

	batchInterval = 5 * time.Minute
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
	runningCMS struct {
		cfg   *CMS
		agent *CmsAgent

		cmsClient *cms.Client

		reqs []*MetricsRequest

		logger *models.Logger

		limiter *limiter.RateLimiter
	}

	CmsAgent struct {
		CMSs       []*CMS `toml:"cms"`
		ReportStat bool   `toml:"report_stat"`

		runningCms []*runningCMS

		ctx       context.Context
		cancelFun context.CancelFunc

		logger *models.Logger

		accumulator telegraf.Accumulator

		wg sync.WaitGroup

		inputStat     internal.InputStat
		succedRequest int64
		faildRequest  int64
	}
)

func (s *CmsAgent) IsRunning() bool {
	return s.inputStat.Stat > 0
}

func (s *CmsAgent) StatMetric() telegraf.Metric {
	if !s.ReportStat {
		return nil
	}
	metricname := "datakit_input_" + inputName
	tags := map[string]string{}
	fields := s.inputStat.Fields()
	fields["failed_request"] = s.faildRequest
	fields["succeed_request"] = s.succedRequest
	s.inputStat.ClearErrorID()
	m, _ := metric.New(metricname, tags, fields, time.Now())
	return m
}

func (_ *CmsAgent) SampleConfig() string {
	return aliyuncmsConfigSample
}

func (_ *CmsAgent) Description() string {
	return ""
}

func (_ *CmsAgent) Gather(telegraf.Accumulator) error {
	return nil
}

func (ac *CmsAgent) Start(acc telegraf.Accumulator) error {

	if len(ac.CMSs) == 0 {
		ac.logger.Warnf("no configuration found")
		return nil
	}

	ac.logger.Info("starting...")

	ac.accumulator = acc

	ac.inputStat.SetStat(len(ac.CMSs))

	for _, cfg := range ac.CMSs {
		rc := &runningCMS{
			agent:  ac,
			cfg:    cfg,
			logger: ac.logger,
		}
		ac.runningCms = append(ac.runningCms, rc)

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
		ac := NewAgent()
		return ac
	})
}
