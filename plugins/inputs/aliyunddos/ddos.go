package aliyunddos

import (
	"bytes"
	"fmt"
	"strings"
	"time"

	"github.com/aliyun/alibaba-cloud-sdk-go/sdk"
	"github.com/aliyun/alibaba-cloud-sdk-go/sdk/requests"
	influxdb "github.com/influxdata/influxdb1-client/v2"
	"gitlab.jiagouyun.com/cloudcare-tools/cliutils/logger"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io"

	"github.com/tidwall/gjson"
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

var (
	l         *logger.Logger
	inputName = "aliyunddos"
)

func (_ *DDoS) SampleConfig() string {
	return configSample
}

func (_ *DDoS) Catalog() string {
	return "aliyun"
}

func (_ *DDoS) Description() string {
	return ""
}

func (_ *DDoS) Gather() error {
	return nil
}

func (a *DDoS) Run() {
	l = logger.SLogger("aliyunddos")
	l.Info("aliyunddos input started...")

	a.checkCfg()

	cli, err := sdk.NewClientWithAccessKey(a.RegionID, a.AccessKeyID, a.AccessKeySecret)
	if err != nil {
		l.Errorf("create client failed, %s", err)
	}

	a.client = cli

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

func (a *DDoS) checkCfg() {
	// 采集频度
	a.IntervalDuration = 24 * time.Hour

	if a.Interval != "" {
		du, err := time.ParseDuration(a.Interval)
		if err != nil {
			l.Errorf("bad interval %s: %s, use default: 10m", a.Interval, err.Error())
		} else {
			a.IntervalDuration = du
		}
	}

	// 指标集名称
	if a.MetricName == "" {
		a.MetricName = inputName
	}
}

func (r *DDoS) getInstance(region string) error {
	var pageNumber = 1
	var pageSize = 10

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
			l.Error("getInstance failed", err)
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
			pt, err := io.MakeMetric(r.MetricName, tags, fields, time.Now())
			if err != nil {
				l.Errorf("make metric point error %v", err)
			}

			err = io.NamedFeed([]byte(pt), datakit.Metric, inputName)
			if err != nil {
				l.Errorf("push metric point error %v", err)
			}

			go r.describeInstanceDetails(item.Get("InstanceId").String(), region)
			go r.describeInstanceStatistics(item.Get("InstanceId").String(), region)
			go r.describeNetworkRules(item.Get("InstanceId").String(), region)
		}

		total := gjson.Parse(data).Get("TotalCount").Int()
		if int64(pageNumber*pageSize) >= total {
			break
		}

		pageNumber = pageNumber + 1
	}
	return nil
}

func (r *DDoS) command() {
	for _, region := range regions {
		go r.getInstance(region)
	}

	for _, region := range regions2 {
		go r.describeWebRules(region)
	}
}

func (r *DDoS) describeInstanceDetails(instanceID, region string) error {
	var lines [][]byte
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
		l.Error("describeInstanceDetails failed", err)
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

		pt, err := io.MakeMetric(r.MetricName, tags, fields, time.Now())
		if err != nil {
			l.Errorf("make metric point error %v", err)
		}

		lines = append(lines, pt)

		err = io.NamedFeed([]byte(pt), datakit.Metric, inputName)
		if err != nil {
			l.Errorf("push metric point error %v", err)
		}
	}

	r.resData = bytes.Join(lines, []byte("\n"))

	return nil
}

func (r *DDoS) describeInstanceStatistics(instanceID, region string) error {
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
		l.Error("describeInstanceStatistics failed", err)
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

		pt, err := io.MakeMetric(r.MetricName, tags, fields, time.Now())
		if err != nil {
			l.Errorf("make metric point error %v", err)
		}

		err = io.NamedFeed([]byte(pt), datakit.Metric, inputName)
		if err != nil {
			l.Errorf("push metric point error %v", err)
		}
	}

	return nil
}

func (r *DDoS) describeWebRules(region string) error {
	var pageNumber = 1
	var pageSize = 10

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
			l.Error("describeWebRules failed", err)
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

			pt, err := io.MakeMetric(r.MetricName, tags, fields, time.Now())
			if err != nil {
				l.Errorf("make metric point error %v", err)
			}

			err = io.NamedFeed([]byte(pt), datakit.Metric, inputName)
			if err != nil {
				l.Errorf("push metric point error %v", err)
			}
		}

		total := gjson.Parse(data).Get("TotalCount").Int()
		if int64(pageNumber*pageSize) >= total {
			break
		}

		pageNumber = pageNumber + 1
	}
	return nil
}

func (r *DDoS) describeNetworkRules(instanceID, region string) error {
	var pageNumber = 1
	var pageSize = 10

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
		request.QueryParams["InstanceId"] = instanceID

		response, err := r.client.ProcessCommonRequest(request)
		if err != nil {
			l.Error("describeNetworkRules failed", err)
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

			pt, err := influxdb.NewPoint(r.MetricName, tags, fields, time.Now())
			if err != nil {
				return err
			}

			err = io.NamedFeed([]byte(pt.String()), datakit.Metric, inputName)
		}

		total := gjson.Parse(data).Get("TotalCount").Int()
		if int64(pageNumber*pageSize) >= total {
			break
		}

		pageNumber = pageNumber + 1
	}

	return nil
}

func init() {
	inputs.Add(inputName, func() inputs.Input {
		return &DDoS{}
	})
}
