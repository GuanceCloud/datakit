package aliyunsecurity

import (
	"time"

	"github.com/aliyun/alibaba-cloud-sdk-go/services/aegis"
	"github.com/aliyun/alibaba-cloud-sdk-go/services/sas"
	influxdb "github.com/influxdata/influxdb1-client/v2"
	"gitlab.jiagouyun.com/cloudcare-tools/cliutils/logger"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"
)

var (
	l         *logger.Logger
	inputName = "aliyunsecurity"
)

var regions = []string{
	"cn-hangzhou",
	"ap-southeast-3",
}

func (_ *Security) Catalog() string {
	return "aliyun"
}

func (_ *Security) SampleConfig() string {
	return configSample
}

func (_ *Security) Description() string {
	return ""
}

func (_ *Security) Gather() error {
	return nil
}

func (a *Security) Run() {
	l = logger.SLogger("aliyunSecurity")

	l.Info("aliyunSecurity input started...")

	a.checkCfg()

	cli, err := sas.NewClientWithAccessKey(a.RegionID, a.AccessKeyID, a.AccessKeySecret)
	if err != nil {
		l.Errorf("create client failed, %s", err)
	}

	cli2, err := aegis.NewClientWithAccessKey(a.RegionID, a.AccessKeyID, a.AccessKeySecret)
	if err != nil {
		l.Errorf("create client failed, %s", err)
	}

	a.client = cli

	a.aclient = cli2

	tick := time.NewTicker(a.IntervalDuration)
	defer tick.Stop()

	for {
		select {
		case <-tick.C:
			// handle
			a.command()
		case <-datakit.Exit.Wait():
			l.Info("exit")
			return
		}
	}
}

func (r *Security) checkCfg() {
	// 采集频度
	r.IntervalDuration = 10 * time.Second

	if r.Interval != "" {
		du, err := time.ParseDuration(r.Interval)
		if err != nil {
			l.Errorf("bad interval %s: %s, use default: 10s", r.Interval, err.Error())
		} else {
			r.IntervalDuration = du
		}
	}

	// 指标集名称
	if r.MetricName == "" {
		r.MetricName = "aliyunsecurity"
	}
}

func (r *Security) command() {
	for _, region := range regions {
		go r.describeSummaryInfo(region)
		go r.describeSecurityStatInfo(region)
		go r.describeRiskCheckSummary(region)
	}
}

func (r *Security) describeSummaryInfo(region string) {
	request := aegis.CreateDescribeSummaryInfoRequest()
	request.Scheme = "https"
	request.RegionId = region

	response, err := r.aclient.DescribeSummaryInfo(request)
	if err != nil {
		l.Errorf("[sas] action DescribeSummaryInfo failed, %s", err.Error())
	}

	tags := map[string]string{}
	fields := map[string]interface{}{}

	tags["product"] = "sas"
	tags["type"] = "describeSummaryInfo"
	tags["region"] = region

	fields["aegis_client_online_count"] = response.AegisClientOnlineCount
	fields["aegis_client_offline_count"] = response.AegisClientOfflineCount
	fields["security_score"] = response.SecurityScore

	pt, err := influxdb.NewPoint(r.MetricName, tags, fields, time.Now())
	if err != nil {
		return
	}

	err = io.NamedFeed([]byte(pt.String()), io.Metric, inputName)
}

func (r *Security) describeSecurityStatInfo(region string) {
	request := aegis.CreateDescribeSecurityStatInfoRequest()
	request.Scheme = "https"
	request.RegionId = region

	response, err := r.aclient.DescribeSecurityStatInfo(request)
	if err != nil {
		l.Errorf("[sas] action DescribeSecurityStatInfo failed, %s", err.Error())
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

	pt, err := io.MakeMetric(r.MetricName, tags, fields, time.Now())
	if err != nil {
		l.Errorf("make metric point error %v", err)
	}

	err = io.NamedFeed([]byte(pt), io.Metric, inputName)
	if err != nil {
		l.Errorf("push metric point error %v", err)
	}
}

func (r *Security) describeRiskCheckSummary(region string) {
	// TrafficData
	request := sas.CreateDescribeRiskCheckSummaryRequest()
	request.Scheme = "https"

	response, err := r.client.DescribeRiskCheckSummary(request)
	if err != nil {
		l.Errorf("[sas] action DescribeRiskCheckSummary failed, %s", err.Error())
	}

	tags := map[string]string{}
	fields := map[string]interface{}{}

	tags["product"] = "sas"
	tags["type"] = "describeRiskCheckSummary"

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

	pt, err := io.MakeMetric(r.MetricName, tags, fields, time.Now())
	if err != nil {
		l.Errorf("make metric point error %v", err)
	}

	err = io.NamedFeed([]byte(pt), io.Metric, inputName)
	if err != nil {
		l.Errorf("push metric point error %v", err)
	}
}

func init() {
	inputs.Add(inputName, func() inputs.Input {
		return &Security{}
	})
}
