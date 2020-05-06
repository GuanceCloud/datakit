package aliyunddos

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/aliyun/alibaba-cloud-sdk-go/sdk"
	"github.com/aliyun/alibaba-cloud-sdk-go/sdk/requests"
	"github.com/aliyun/alibaba-cloud-sdk-go/services/aegis"
	"github.com/tidwall/gjson"

	"github.com/influxdata/telegraf"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/models"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"
)

var regions = []string{
	"cn-hangzhou",
	"ap-southeast-3",
}

var regions2 = []string{
	"cn-hangzhou",
	"ap-southeast-1",
}

var regions3 = []string{
	"cn-hangzhou",
	"cn-hongkong",
}

type AliyunDDoS struct {
	DDoS        []*DDoS
	ctx         context.Context
	cancelFun   context.CancelFunc
	accumulator telegraf.Accumulator
	logger      *models.Logger

	runningInstances []*runningInstance
}

type runningInstance struct {
	cfg        *DDoS
	agent      *AliyunDDoS
	logger     *models.Logger
	client     *sdk.Client
	aclient    *aegis.Client
	metricName string
}

func (_ *AliyunDDoS) SampleConfig() string {
	return configSample
}

func (_ *AliyunDDoS) Description() string {
	return ""
}

func (_ *AliyunDDoS) Gather(telegraf.Accumulator) error {
	return nil
}

func (_ *AliyunDDoS) Init() error {
	return nil
}

func (a *AliyunDDoS) Start(acc telegraf.Accumulator) error {
	fmt.Println("======ddos采集 start")
	a.logger = &models.Logger{
		Name: `aliyunddos`,
	}

	if len(a.DDoS) == 0 {
		a.logger.Warnf("no configuration found")
		return nil
	}

	a.logger.Infof("starting...")

	a.accumulator = acc

	for _, instCfg := range a.DDoS {
		r := &runningInstance{
			cfg:    instCfg,
			agent:  a,
			logger: a.logger,
		}
		r.metricName = instCfg.MetricName
		if r.metricName == "" {
			r.metricName = "aliyun_ddos"
		}

		if r.cfg.Interval.Duration == 0 {
			r.cfg.Interval.Duration = time.Minute * 10
		}

		cli, err := sdk.NewClientWithAccessKey(instCfg.RegionID, instCfg.AccessKeyID, instCfg.AccessKeySecret)
		if err != nil {
			r.logger.Errorf("create client failed, %s", err)
			return err
		}

		r.client = cli
		a.runningInstances = append(a.runningInstances, r)

		go r.run(a.ctx)
	}
	return nil
}

func (a *AliyunDDoS) Stop() {
	a.cancelFun()
}

func (r *runningInstance) run(ctx context.Context) error {
	defer func() {
		if e := recover(); e != nil {

		}
	}()

	cli, err := sdk.NewClientWithAccessKey(r.cfg.RegionID, r.cfg.AccessKeyID, r.cfg.AccessKeySecret)
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

		fmt.Println("======ddos采集")

		r.command()

		internal.SleepContext(ctx, r.cfg.Interval.Duration)
	}
}

func (r *runningInstance) getInstance(region string) error {
	var pageNumber = 1
	var pageSize = 50

	for {
		request := requests.NewCommonRequest()
		request.Method = "POST"
		request.Scheme = "https" // https | http
		request.Domain = "ddoscoo.cn-hangzhou.aliyuncs.com"
		request.Version = "2020-01-01"
		request.ApiName = "DescribeInstances"

		request.QueryParams["RegionId"] = region
		request.QueryParams["PageSize"] = fmt.Sprintf("%d", pageSize)
		request.QueryParams["PageNumber"] = fmt.Sprintf("%d", pageNumber)

		response, err := r.client.ProcessCommonRequest(request)
		if err != nil {
			r.logger.Error("instance failed")
			return err
		}

		data := response.GetHttpContentString()
		instanceArr := gjson.Parse(data).Get("InstanceIds").Array()

		for _, item := range instanceArr {
			tags := map[string]string{}
			fields := map[string]interface{}{}

			tags["edition"] = item.Get("Edition").String()
			tags["product"] = "ddos"
			tags["action"] = "getInstance"
			tags["region"] = region

			fields["instanceId"] = item.Get("InstanceId").String()
			fields["remark"] = item.Get("Remark").String()

			go r.describeInstanceDetails(item.Get("InstanceId").String(), region)
			go r.describeInstanceStatistics(item.Get("InstanceId").String(), region)
		}

		total := gjson.Parse(data).Get("TotalCount").Int()
		if int64(pageNumber*pageSize) >= total {
			break
		}

		pageNumber = pageNumber + 1
	}
	return nil
}

func (r *runningInstance) command() {
	for _, region := range regions {
		go r.getInstance(region)
	}

	for _, region := range regions2 {
		go r.describeWebRules(region)
		go r.describeNetworkRules(region)
	}

	for _, region := range regions3 {
		go r.describePayInfo(region)
	}
}

func (r *runningInstance) describeInstanceDetails(instanceID, region string) error {
	request := requests.NewCommonRequest()
	request.Method = "POST"
	request.Scheme = "https"
	request.Domain = "ddoscoo.cn-hangzhou.aliyuncs.com"
	request.Version = "2020-01-01"
	request.ApiName = "DescribeInstanceDetails"

	request.QueryParams["RegionId"] = region
	request.QueryParams["InstanceIds.1"] = instanceID

	response, err := r.client.ProcessCommonRequest(request)
	if err != nil {
		r.logger.Error("instance detail failed")
		return err
	}

	data := response.GetHttpContentString()
	instanceArr := gjson.Parse(data).Get("InstanceDetails").Array()

	for _, item := range instanceArr {
		tags := map[string]string{}
		fields := map[string]interface{}{}

		tags["edition"] = item.Get("Edition").String()
		tags["product"] = "ddos"
		tags["action"] = "describeInstanceDetails"
		tags["region"] = region

		eipStatus := []string{}
		eip := []string{}

		for _, obj := range item.Get("EipInfos").Array() {
			eipStatus = append(eipStatus, obj.Get("Status").String())
			eip = append(eip, obj.Get("Eip").String())
		}
		fields["eipStatus"] = strings.Join(eipStatus, "\\")
		fields["eip"] = strings.Join(eip, "\\")
		fields["line"] = item.Get("line").String()
		fields["instanceId"] = item.Get("InstanceId").String()
		r.agent.accumulator.AddFields(r.metricName, fields, tags)
	}

	return nil
}

func (r *runningInstance) describeInstanceStatistics(instanceID, region string) error {
	request := requests.NewCommonRequest()
	request.Method = "POST"
	request.Scheme = "https" // https | http
	request.Domain = "ddoscoo.cn-hangzhou.aliyuncs.com"
	request.Version = "2020-01-01"
	request.ApiName = "DescribeInstanceStatistics"

	request.QueryParams["RegionId"] = "cn-hangzhou"
	request.QueryParams["InstanceIds.1"] = instanceID

	response, err := r.client.ProcessCommonRequest(request)
	if err != nil {
		r.logger.Error("instance detail failed")
		return err
	}

	data := response.GetHttpContentString()
	instanceArr := gjson.Parse(data).Get("InstanceStatistics").Array()

	for _, item := range instanceArr {
		tags := map[string]string{}
		fields := map[string]interface{}{}

		tags["product"] = "ddos"
		tags["action"] = "describeInstanceStatistics"
		tags["region"] = region

		fields["defenseCountUsage"] = item.Get("DefenseCountUsage").Int()
		fields["domainUsage"] = item.Get("DomainUsage").Int()
		fields["instanceId"] = item.Get("InstanceId").String()
		fields["portUsage"] = item.Get("PortUsage").Int()
		fields["siteUsage"] = item.Get("SiteUsage").Int()

		r.agent.accumulator.AddFields(r.metricName, fields, tags)
	}

	return nil
}

func (r *runningInstance) describeWebRules(region string) error {
	var pageNumber = 1
	var pageSize = 50

	for {
		request := requests.NewCommonRequest()
		request.Method = "POST"
		request.Scheme = "https" // https | http
		request.Domain = "ddoscoo.cn-hangzhou.aliyuncs.com"
		request.Version = "2020-01-01"
		request.ApiName = "DescribeWebRules"

		request.QueryParams["RegionId"] = region
		request.QueryParams["PageSize"] = fmt.Sprintf("%d", pageSize)
		request.QueryParams["PageNumber"] = fmt.Sprintf("%d", pageNumber)

		response, err := r.client.ProcessCommonRequest(request)
		if err != nil {
			r.logger.Error("instance failed")
			return err
		}

		data := response.GetHttpContentString()
		instanceArr := gjson.Parse(data).Get("WebRules").Array()

		for _, item := range instanceArr {
			tags := map[string]string{}
			fields := map[string]interface{}{}

			tags["product"] = "ddos"
			tags["action"] = "describeWebRules"
			tags["region"] = region

			fields["CcEnabled"] = item.Get("CcEnabled").Bool()
			fields["SslProtocols"] = item.Get("SslProtocols").String()
			fields["CcRuleEnabled"] = item.Get("CcRuleEnabled").Bool()
			fields["SslCiphers"] = item.Get("SslCiphers").String()
			fields["CertName"] = item.Get("CertName").String()
			fields["Domain"] = item.Get("Domain").String()
			fields["Http2Enable"] = item.Get("Http2Enable").Bool()
			fields["Cname"] = item.Get("Cname").String()
			fields["CcTemplate"] = item.Get("CcTemplate").String()

			for _, obj := range item.Get("ProxyTypes").Array() {
				proxyKey := obj.Get("ProxyType").String()
				fields[proxyKey] = obj.Get("ProxyPorts").Array()[0].Int()
			}

			for _, obj := range item.Get("RealServers").Array() {
				key := fmt.Sprintf("real_server_%s", obj.Get("RsType").String())
				fields[key] = obj.Get("RealServer").String()
			}

			r.agent.accumulator.AddFields(r.metricName, fields, tags)
		}

		total := gjson.Parse(data).Get("TotalCount").Int()
		if int64(pageNumber*pageSize) >= total {
			break
		}

		pageNumber = pageNumber + 1
	}
	return nil
}

func (r *runningInstance) describeNetworkRules(region string) error {
	var pageNumber = 1
	var pageSize = 50

	for {
		request := requests.NewCommonRequest()
		request.Method = "POST"
		request.Scheme = "https" // https | http
		request.Domain = "ddoscoo.cn-hangzhou.aliyuncs.com"
		request.Version = "2020-01-01"
		request.ApiName = "DescribeNetworkRules"

		request.QueryParams["RegionId"] = region
		request.QueryParams["PageSize"] = fmt.Sprintf("%d", pageSize)
		request.QueryParams["PageNumber"] = fmt.Sprintf("%d", pageNumber)

		response, err := r.client.ProcessCommonRequest(request)
		if err != nil {
			r.logger.Error("instance failed")
			return err
		}

		data := response.GetHttpContentString()

		instanceArr := gjson.Parse(data).Get("NetworkRules").Array()

		for _, item := range instanceArr {
			tags := map[string]string{}
			fields := map[string]interface{}{}

			tags["product"] = "ddos"
			tags["action"] = "describeNetworkRules"
			tags["region"] = region
			tags["Protocol"] = item.Get("Protocol").String()

			fields["backendPort"] = item.Get("BackendPort").Int()
			fields["isAutoCreate"] = item.Get("IsAutoCreate").Bool()
			fields["instanceId"] = item.Get("InstanceId").String()
			fields["frontendPort"] = item.Get("FrontendPort").Int()

			realServer := []string{}

			for _, obj := range item.Get("RealServers").Array() {
				realServer = append(realServer, obj.String())
			}
			fields["realServer"] = strings.Join(realServer, "\\")

			r.agent.accumulator.AddFields(r.metricName, fields, tags)
		}

		total := gjson.Parse(data).Get("TotalCount").Int()
		if int64(pageNumber*pageSize) >= total {
			break
		}

		pageNumber = pageNumber + 1
	}

	return nil
}

func (r *runningInstance) describePayInfo(region string) error {
	request := requests.NewCommonRequest()
	request.Method = "POST"
	request.Scheme = "https" // https | http
	request.Domain = "wafopenapi.cn-hangzhou.aliyuncs.com"
	request.Version = "2018-01-17"
	request.ApiName = "DescribePayInfo"

	request.QueryParams["RegionId"] = "cn-hangzhou"

	response, err := r.client.ProcessCommonRequest(request)
	if err != nil {
		r.logger.Error("instance detail failed")
		return err
	}

	data := response.GetHttpContentString()

	instanceArr := gjson.Parse(data).Get("Result").Array()

	for _, item := range instanceArr {
		tags := map[string]string{}
		fields := map[string]interface{}{}

		tags["product"] = "waf"
		tags["action"] = "describePayInfo"
		tags["region"] = item.Get("Region").String()

		fields["endDate"] = item.Get("EndDate").Int()
		fields["inDebt"] = item.Get("InDebt").Int()
		fields["instanceId"] = item.Get("InstanceId").String()
		fields["PayType"] = item.Get("PayType").Int()
		fields["trial"] = item.Get("Trial").Int()
		fields["region"] = item.Get("Region").String()
		fields["remainDay"] = item.Get("RemainDay").Int()
		fields["status"] = item.Get("Status").Int()

		r.agent.accumulator.AddFields(r.metricName, fields, tags)
	}

	return nil
}

func init() {
	inputs.Add("aliyunddos", func() telegraf.Input {
		ac := &AliyunDDoS{}
		ac.ctx, ac.cancelFun = context.WithCancel(context.Background())
		return ac
	})
}
