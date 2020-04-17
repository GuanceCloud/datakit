package aliyunprice

import (
	"fmt"
	"time"

	"github.com/aliyun/alibaba-cloud-sdk-go/sdk/requests"
	"github.com/aliyun/alibaba-cloud-sdk-go/services/bssopenapi"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal"
)

const (
	slbSampleConfig = `
# ##负载均衡
#[[slb]]
##metric_name = ''
##description = ''
##interval = '1d'
#pay_as_you_go = true
#region = "cn-hangzhou"
#load_balancer_spec = "slb.s1.small"
#bandwidth = 6 #unit is MB
#internet_traffic_out = 0 #0:按固定带宽，1:按使用流量
#private_net = false #是否公网
##service_period_quantity = 1
##service_period_unit = "Year"
##quantity = 1
`
)

/*"Value": "0",
"Name": "按固定带宽计费",
"Remark": "开通后即开始按固定带宽计费，和实例状态及使用流量无关"

"Value": "1",
"Name": "按使用流量计费",
"Remark": "开通后按照使用的流量进行计费，私网实例免流量费"
*/

/*
	{
		"Value": "internet",
		"Name": "公网",
		"Remark": " 负载均衡实例仅提供公网IP，可以通过Internet访问的负载均衡服务"
	},
	{
		"Value": "intranet",
		"Name": "私网",
		"Remark": "负载均衡实例仅提供阿里云私网IP地址（或VPC内的地址），只能通过阿里云内部网络访问该负载均衡服务"
	}
*/

type Slb struct {
	MetricName  string
	Description string
	PayAsYouGo  bool
	Interval    internal.Duration

	Region string

	LoadBalancerSpec   string
	Bandwidth          int64 //kbps
	InternetTrafficOut int
	PrivateNet         bool

	ServicePeriodQuantity int
	ServicePeriodUnit     string
	Quantity              int
}

func (e *Slb) toRequest() (*priceReq, error) {
	if e.Quantity == 0 {
		e.Quantity = 1
	}

	if e.ServicePeriodQuantity == 0 {
		e.ServicePeriodQuantity = 1
	}

	if e.ServicePeriodUnit == "" {
		e.ServicePeriodUnit = "Year"
	}

	p := &priceReq{
		fetchModulePriceHistory:             make(map[string]time.Time),
		priceModuleInfos:                    make(map[string]*bssopenapi.ModuleList),
		productCodeForPriceModulesSubscript: "slb",
		productCodeForPriceModulesPayasugo:  "slb",
	}
	p.m = e
	p.payAsYouGo = e.PayAsYouGo
	p.metricName = e.MetricName
	p.interval = e.Interval.Duration
	if p.interval == 0 {
		p.interval = defaultInterval
	}
	p.region = e.Region

	bw := e.Bandwidth * 1024

	addressType := "internet"
	if e.PrivateNet {
		addressType = "intranet"
	}

	// bandwidthConfig := fmt.Sprintf("Bandwidth:%d", bw)
	// if p.payAsYouGo {
	// 	bandwidthConfig += fmt.Sprintf(",ISP:%s", e.ISP)
	// }
	//internetChargeTypeConfig := fmt.Sprintf("InternetChargeType:%d,ISP:%s", e.InternetChargeType, e.ISP)

	//slb服务
	instanceRentConfig := fmt.Sprintf(`InstanceRent:1,AddressType:%s`, addressType)

	if e.PayAsYouGo {
		p.payasyougoReq = bssopenapi.CreateGetPayAsYouGoPriceRequest()
		p.payasyougoReq.Scheme = "https"
		p.payasyougoReq.ProductCode = "slb"
		p.payasyougoReq.SubscriptionType = `PayAsYouGo`
		p.payasyougoReq.Region = e.Region

		mods := []bssopenapi.GetPayAsYouGoPriceModuleList{
			{
				ModuleCode: "LoadBalancerSpec",
				Config:     fmt.Sprintf("LoadBalancerSpec:%s", e.LoadBalancerSpec),
				PriceType:  "Hour",
			},
			{
				ModuleCode: "InternetTrafficOut",
				Config:     fmt.Sprintf("InternetTrafficOut:%d", e.InternetTrafficOut),
				PriceType:  "Usage",
			},
			{
				ModuleCode: "InstanceRent",
				Config:     instanceRentConfig,
				PriceType:  "Hour",
			},
		}

		if e.InternetTrafficOut == 0 {
			mods = append(mods,
				bssopenapi.GetPayAsYouGoPriceModuleList{
					ModuleCode: "Bandwidth",
					Config:     fmt.Sprintf("Bandwidth:%d", bw),
					PriceType:  "Hour",
				})
		}

		p.payasyougoReq.ModuleList = &mods

	} else {
		p.subscriptionReq = bssopenapi.CreateGetSubscriptionPriceRequest()
		p.subscriptionReq.Scheme = `https`
		p.subscriptionReq.ProductCode = "slb"
		p.subscriptionReq.SubscriptionType = `Subscription`
		p.subscriptionReq.OrderType = `NewOrder`
		p.subscriptionReq.Quantity = requests.NewInteger(e.Quantity)
		p.subscriptionReq.ServicePeriodQuantity = requests.NewInteger(e.ServicePeriodQuantity)
		p.subscriptionReq.ServicePeriodUnit = e.ServicePeriodUnit
		p.subscriptionReq.Region = e.Region

		mods := []bssopenapi.GetSubscriptionPriceModuleList{
			{
				ModuleCode: "LoadBalancerSpec",
				Config:     fmt.Sprintf("LoadBalancerSpec:%s", e.LoadBalancerSpec),
			},
			{
				ModuleCode: "InternetTrafficOut",
				Config:     fmt.Sprintf("InternetTrafficOut:%d", e.InternetTrafficOut),
			},
			{
				ModuleCode: "InstanceRent",
				Config:     instanceRentConfig,
			},
		}

		if !e.PrivateNet {
			mods = append(mods,
				bssopenapi.GetSubscriptionPriceModuleList{
					ModuleCode: "Bandwidth",
					Config:     fmt.Sprintf("Bandwidth:%d", bw),
				})
		}

		p.subscriptionReq.ModuleList = &mods
	}

	return p, nil
}

func (e *Slb) handleTags(tags map[string]string) map[string]string {

	tags["Description"] = e.Description
	tags["Bandwidth"] = fmt.Sprintf("%d", e.Bandwidth)
	tags["LoadBalancerSpec"] = e.LoadBalancerSpec
	tags["InternetTrafficOut"] = fmt.Sprintf("%d", e.InternetTrafficOut)
	if e.PrivateNet {
		tags["PrivateNet"] = "1"
	} else {
		tags["PrivateNet"] = "0"
	}
	tags["Quantity"] = fmt.Sprintf("%d x %d%s", e.Quantity, e.ServicePeriodQuantity, e.ServicePeriodUnit)

	return tags
}

func (e *Slb) handleFields(fields map[string]interface{}) map[string]interface{} {
	return fields
}
