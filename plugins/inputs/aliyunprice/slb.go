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

# ##(optional) 自定义指标集名称，默认使用 aliyun_price
#metric_name = ''

# ##(optional) 采集间隔，默认一天
#interval = '1d'

# ##(required) 是否采集后付费价格, 默认false(即采集预付费价格)
#pay_as_you_go = false

# ##(required) 地域
#region = "cn-hangzhou"

# ##(required) 实例规格
#load_balancer_spec = 'slb.s1.small'

# ##(required) 带宽值, 单位MB
#bandwidth = 6

# ##(required) 计费类型
# ## 0:按固定带宽, 开通后即开始按固定带宽计费，和实例状态及使用流量无关
# ## 1:按使用流量, 开通后按照使用的流量进行计费，私网实例免流量费
#internet_traffic_out = 0

# ##(required) 是否公网
# ## false: 负载均衡实例仅提供公网IP，可以通过Internet访问的负载均衡服务
# ## true: 负载均衡实例仅提供阿里云私网IP地址（或VPC内的地址），只能通过阿里云内部网络访问该负载均衡服务
#private_net = false

# ##(optional)购买时长, 默认为1, 如果单位为Year, 则表示1年
#service_period_quantity = 1

# ##(optional)购买时长单位: Month，Year, 默认为 Year
#service_period_unit = "Year"

# ##(optional)购买份数, 默认1份
#quantity = 1
`
)

type Slb struct {
	MetricName string
	PayAsYouGo bool
	Interval   internal.Duration

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
		m:                                   e,
		payAsYouGo:                          e.PayAsYouGo,
		metricName:                          e.MetricName,
		region:                              e.Region,
		fetchModulePriceHistory:             make(map[string]time.Time),
		priceModuleInfos:                    make(map[string]*bssopenapi.ModuleList),
		productCodeForPriceModulesSubscript: "slb",
		productCodeForPriceModulesPayasugo:  "slb",
	}
	p.interval = e.Interval.Duration
	if p.interval == 0 {
		p.interval = defaultInterval
	}

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

	tags["Bandwidth"] = fmt.Sprintf("%d", e.Bandwidth)
	tags["LoadBalancerSpec"] = e.LoadBalancerSpec
	tags["InternetTrafficOut"] = fmt.Sprintf("%d", e.InternetTrafficOut)
	if e.PrivateNet {
		tags["PrivateNet"] = "1"
	} else {
		tags["PrivateNet"] = "0"
	}
	tags["Quantity"] = fmt.Sprintf("%d", e.Quantity)
	tags["ServicePeriodQuantity"] = fmt.Sprintf("%d", e.ServicePeriodQuantity)
	tags["ServicePeriodUnit"] = e.ServicePeriodUnit

	return tags
}

func (e *Slb) handleFields(fields map[string]interface{}) map[string]interface{} {
	return fields
}
