package aliyunprice

import (
	"fmt"
	"time"

	"github.com/aliyun/alibaba-cloud-sdk-go/sdk/requests"
	"github.com/aliyun/alibaba-cloud-sdk-go/services/bssopenapi"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
)

const (
	slbSampleConfig = `
#[[inputs.aliyunprice.slb]]

# ##(optional) custom metric name, default is aliyun_price
#metric_name = ''

# ##(optional) collect interval, default is one day
#interval = '1d'

# ##(required) cllect PayAsYouGo price, default is false
#pay_as_you_go = false

# ##(required) instance region
#region = "cn-hangzhou"

# ##(required) instance
#load_balancer_spec = 'slb.s1.small'

# ##(required) bandwidth, unit is MB
#bandwidth = 6

# ##(required) billing type
# ## 0:by fixed bandwidth
# ## 1:by traffic
#internet_traffic_out = 0

# ##(required) if private net, default is false
#private_net = false

# ##(optional)Purchase duration, default is 1, so if unit is Year, then is one year
#service_period_quantity = 1

# ##(optional)unit of purchase duration: Month，Year, defalut is Year
#service_period_unit = "Year"

# ##(optional)Purchase quantity, default is 1
#quantity = 1
`
)

type Slb struct {
	MetricName string
	PayAsYouGo bool
	Interval   datakit.Duration

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
