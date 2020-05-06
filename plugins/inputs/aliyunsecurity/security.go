package aliyunsecurity

import (
	"context"
	"time"

	"github.com/aliyun/alibaba-cloud-sdk-go/services/aegis"
	"github.com/aliyun/alibaba-cloud-sdk-go/services/sas"

	"github.com/influxdata/telegraf"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/models"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"
)

var regions = []string{
	"cn-hangzhou",
	"ap-southeast-3",
}

type AliyunSecurity struct {
	Security    []*Security
	ctx         context.Context
	cancelFun   context.CancelFunc
	accumulator telegraf.Accumulator
	logger      *models.Logger

	runningInstances []*runningInstance
}

type runningInstance struct {
	cfg        *Security
	agent      *AliyunSecurity
	logger     *models.Logger
	client     *sas.Client
	aclient    *aegis.Client
	metricName string
}

func (_ *AliyunSecurity) SampleConfig() string {
	return configSample
}

func (_ *AliyunSecurity) Description() string {
	return ""
}

func (_ *AliyunSecurity) Gather(telegraf.Accumulator) error {
	return nil
}

func (_ *AliyunSecurity) Init() error {
	return nil
}

func (a *AliyunSecurity) Start(acc telegraf.Accumulator) error {
	a.logger = &models.Logger{
		Name: `aliyunsecurity`,
	}

	if len(a.Security) == 0 {
		a.logger.Warnf("no configuration found")
		return nil
	}

	a.logger.Infof("starting...")

	a.accumulator = acc

	for _, instCfg := range a.Security {
		r := &runningInstance{
			cfg:    instCfg,
			agent:  a,
			logger: a.logger,
		}
		r.metricName = instCfg.MetricName
		if r.metricName == "" {
			r.metricName = "aliyun_security"
		}

		if r.cfg.Interval.Duration == 0 {
			r.cfg.Interval.Duration = time.Minute * 10
		}

		cli, err := sas.NewClientWithAccessKey(instCfg.RegionID, instCfg.AccessKeyID, instCfg.AccessKeySecret)
		if err != nil {
			r.logger.Errorf("create client failed, %s", err)
			return err
		}

		cli2, err := aegis.NewClientWithAccessKey(instCfg.RegionID, instCfg.AccessKeyID, instCfg.AccessKeySecret)
		if err != nil {
			r.logger.Errorf("create client failed, %s", err)
			return err
		}

		r.client = cli

		r.aclient = cli2

		a.runningInstances = append(a.runningInstances, r)

		go r.run(a.ctx)
	}
	return nil
}

func (a *AliyunSecurity) Stop() {
	a.cancelFun()
}

func (r *runningInstance) run(ctx context.Context) error {
	defer func() {
		if e := recover(); e != nil {

		}
	}()

	cli, err := sas.NewClientWithAccessKey(r.cfg.RegionID, r.cfg.AccessKeyID, r.cfg.AccessKeySecret)
	if err != nil {
		r.logger.Errorf("create client failed, %s", err)
		return err
	}
	r.client = cli

	for {
		select {
		case <-ctx.Done():
			return context.Canceled
		default:
		}

		r.command()

		internal.SleepContext(ctx, r.cfg.Interval.Duration)
	}
}

func (r *runningInstance) command() {
	for _, region := range regions {
		go r.describeSummaryInfo(region)
		go r.describeSecurityStatInfo(region)
		go r.describeRiskCheckSummary(region)
	}
}

func (r *runningInstance) describeSummaryInfo(region string) {
	request := aegis.CreateDescribeSummaryInfoRequest()
	request.Scheme = "https"
	request.RegionId = region

	response, err := r.aclient.DescribeSummaryInfo(request)
	if err != nil {
		r.logger.Errorf("[sas] action DescribeSummaryInfo failed, %s", err.Error())
	}

	tags := map[string]string{}
	fields := map[string]interface{}{}

	tags["product"] = "sas"
	tags["type"] = "describeSummaryInfo"
	tags["region"] = region

	fields["aegis_client_online_count"] = response.AegisClientOnlineCount
	fields["aegis_client_offline_count"] = response.AegisClientOfflineCount
	fields["security_score"] = response.SecurityScore

	r.agent.accumulator.AddFields(r.metricName, fields, tags)
}

func (r *runningInstance) describeSecurityStatInfo(region string) {
	request := aegis.CreateDescribeSecurityStatInfoRequest()
	request.Scheme = "https"
	request.RegionId = region

	response, err := r.aclient.DescribeSecurityStatInfo(request)
	if err != nil {
		r.logger.Errorf("[sas] action DescribeSecurityStatInfo failed, %s", err.Error())
	}

	tags := map[string]string{}
	fields := map[string]interface{}{}

	tags["product"] = "sas"
	tags["type"] = "describeSecurityStatInfo"
	tags["region"] = region

	fields["security_event_total_count"] = response.SecurityEvent.TotalCount
	fields["security_event_serious_count"] = response.SecurityEvent.SeriousCount
	fields["security_event_suspicious_count"] = response.SecurityEvent.SuspiciousCount
	fields["security_event_remind_count"] = response.SecurityEvent.RemindCount

	fields["attack_event_total_count"] = response.AttackEvent.TotalCount

	fields["health_check_total_count"] = response.HealthCheck.TotalCount
	fields["health_check_medium_count"] = response.HealthCheck.MediumCount
	fields["health_check_high_count"] = response.HealthCheck.HighCount
	fields["health_check_lowcount"] = response.HealthCheck.LowCount

	fields["vulnerability_nntf_count"] = response.Vulnerability.NntfCount
	fields["vulnerability_later_count"] = response.Vulnerability.LaterCount
	fields["vulnerability_asap_count"] = response.Vulnerability.AsapCount
	fields["vulnerability_total_count"] = response.Vulnerability.TotalCount

	r.agent.accumulator.AddFields(r.metricName, fields, tags)
}

func (r *runningInstance) describeRiskCheckSummary(region string) {
	// TrafficData
	request := sas.CreateDescribeRiskCheckSummaryRequest()
	request.Scheme = "https"
	request.RegionId = region

	response, err := r.client.DescribeRiskCheckSummary(request)
	if err != nil {
		r.logger.Errorf("[sas] action DescribeRiskCheckSummary failed, %s", err.Error())
	}

	tags := map[string]string{}
	fields := map[string]interface{}{}

	tags["product"] = "sas"
	tags["type"] = "describeRiskCheckSummary"
	tags["region"] = region

	fields["risk_check_summary_risk_count"] = response.RiskCheckSummary.RiskCount

	for _, item := range response.RiskCheckSummary.RiskLevelCount {
		if item.Key == "medium" {
			fields["risk_check_summary_risk_count_medium"] = item.Count
		} else if item.Key == "high" {
			fields["risk_check_summary_risk_count_high"] = item.Count
		} else if item.Key == "low" {
			fields["risk_check_summary_risk_count_low"] = item.Count
		}
	}

	r.agent.accumulator.AddFields(r.metricName, fields, tags)
}

func init() {
	inputs.Add("aliyunsecurity", func() telegraf.Input {
		ac := &AliyunSecurity{}
		ac.ctx, ac.cancelFun = context.WithCancel(context.Background())
		return ac
	})
}
