package aliyunprice

import (
	"fmt"
	"time"

	"github.com/aliyun/alibaba-cloud-sdk-go/sdk/requests"
	"github.com/aliyun/alibaba-cloud-sdk-go/services/bssopenapi"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
)

const (
	eipSampleConfig = `
#[[inputs.aliyunprice.eip]]

# ##(optional) custom metric name, default is aliyun_price
#metric_name = ''

# ##(optional) collect interval, default is one day
#interval = '1d'

# ##(required) cllect PayAsYouGo price, default is false
#pay_as_you_go = false

# ##(required) region
#region = 'cn-hangzhou'

# ##(required) bandwidth, unit is MB
# ## for pay_as_you_go, ignored when internet_charge_type=1
#bandwidth = 1

# ##(required) traffic type, only for pay_as_you_go=true
# ## 0:fixed bandwidth, 1:by used traffix, if true ignore bandwidth
#internet_charge_type = 0

# ##(optional) only for pay_as_you_go=true, default is BGP
#isp = 'BGP'

# ##(optional)Purchase duration, default is 1, so if unit is Year, then is one year
#service_period_quantity = 1

# ##(optional)unit of purchase duration: Month，Year, defalut is Year
#service_period_unit = "Year"

# ##(optional)Purchase quantity, default is 1
#quantity = 1
`
)

type Eip struct {
	MetricName string
	PayAsYouGo bool
	Interval   datakit.Duration

	Region string

	Bandwidth          int64
	InternetChargeType int    //only for PayAsYouGo，0:按固定带宽计费，1:按使用流量计费,此时忽略Bandwidth。
	ISP                string `toml:"isp"` //only for PayAsYouGo

	ServicePeriodQuantity int
	ServicePeriodUnit     string
	Quantity              int
}

func (e *Eip) toRequest() (*priceReq, error) {
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
		productCodeForPriceModulesSubscript: "EIP",
		productCodeForPriceModulesPayasugo:  "EIP",
	}
	p.interval = e.Interval.Duration
	if p.interval == 0 {
		p.interval = defaultInterval
	}

	bw := e.Bandwidth
	if p.payAsYouGo {
		//payAsYouGo下要传kbps
		bw = bw * 1024
	}

	bandwidthConfig := fmt.Sprintf("Bandwidth:%d", bw)
	if p.payAsYouGo {
		bandwidthConfig += fmt.Sprintf(",ISP:%s", e.ISP)
	}
	internetChargeTypeConfig := fmt.Sprintf("InternetChargeType:%d,ISP:%s", e.InternetChargeType, e.ISP)
	//eip服务
	instanceRentConfig := `InstanceRent:1,IsPortable:true` //是否可解绑

	if e.PayAsYouGo {
		p.payasyougoReq = bssopenapi.CreateGetPayAsYouGoPriceRequest()
		p.payasyougoReq.Scheme = "https"
		p.payasyougoReq.ProductCode = "eip"
		p.payasyougoReq.SubscriptionType = `PayAsYouGo`
		p.payasyougoReq.Region = e.Region

		mods := []bssopenapi.GetPayAsYouGoPriceModuleList{
			{
				ModuleCode: "InstanceRent",
				Config:     instanceRentConfig,
				PriceType:  "Hour",
			},
		}

		if e.InternetChargeType == 0 {
			mods = append(mods,
				bssopenapi.GetPayAsYouGoPriceModuleList{
					ModuleCode: "Bandwidth",
					Config:     bandwidthConfig,
					PriceType:  "Day",
				})
		} else {
			mods = append(mods,
				bssopenapi.GetPayAsYouGoPriceModuleList{
					ModuleCode: "InternetChargeType",
					Config:     internetChargeTypeConfig,
					PriceType:  "Usage",
				})
		}

		p.payasyougoReq.ModuleList = &mods

	} else {
		p.subscriptionReq = bssopenapi.CreateGetSubscriptionPriceRequest()
		p.subscriptionReq.Scheme = `https`
		p.subscriptionReq.ProductCode = "eip"
		p.subscriptionReq.SubscriptionType = `Subscription`
		p.subscriptionReq.OrderType = `NewOrder`
		p.subscriptionReq.Quantity = requests.NewInteger(e.Quantity)
		p.subscriptionReq.ServicePeriodQuantity = requests.NewInteger(e.ServicePeriodQuantity)
		p.subscriptionReq.ServicePeriodUnit = e.ServicePeriodUnit
		p.subscriptionReq.Region = e.Region

		mods := []bssopenapi.GetSubscriptionPriceModuleList{
			{
				ModuleCode: "Bandwidth",
				Config:     bandwidthConfig,
			},
		}

		p.subscriptionReq.ModuleList = &mods
	}

	return p, nil
}

func (e *Eip) handleTags(tags map[string]string) map[string]string {

	tags["Bandwidth"] = fmt.Sprintf("%d", e.Bandwidth)
	tags["InternetChargeType"] = fmt.Sprintf("%d", e.InternetChargeType)
	tags["ISP"] = e.ISP
	tags["Quantity"] = fmt.Sprintf("%d", e.Quantity)
	tags["ServicePeriodQuantity"] = fmt.Sprintf("%d", e.ServicePeriodQuantity)
	tags["ServicePeriodUnit"] = e.ServicePeriodUnit

	return tags
}

func (e *Eip) handleFields(fields map[string]interface{}) map[string]interface{} {
	return fields
}
