package aliyunprice

import (
	"fmt"
	"time"

	"github.com/aliyun/alibaba-cloud-sdk-go/sdk/requests"
	"github.com/aliyun/alibaba-cloud-sdk-go/services/bssopenapi"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal"
)

const (
	eipSampleConfig = `
# ##弹性公网IP
#[[eip]]

# ##(optional) 自定义指标集名称，默认使用 aliyun_price
#metric_name = ''

# ##(optional) 采集间隔，默认一天
#interval = '1d'

# ##(required) 是否采集后付费价格, 默认false(即采集预付费价格)
#pay_as_you_go = false

# ##(required) 地域
#region = 'cn-hangzhou'

# ##(required) 带宽,单位MB
# ## 在pay_as_you_go下，如果 internet_charge_type=1，则忽略该值
#bandwidth = 1

# ##(required) 流量类型, 只在 pay_as_you_go=true 时需要设置
# ## 0:按固定带宽, 1:按使用流量,此时忽略 bandwidth
#internet_charge_type = 0

# ##(optional) 线路类型, 当 pay_as_you_go=true 时有效, 默认 BGP
#isp = 'BGP' 

# ##(optional)购买时长, 默认为1, 如果单位为Year, 则表示1年
#service_period_quantity = 1

# ##(optional)购买时长单位: Month，Year, 默认为 Year
#service_period_unit = "Year"

# ##(optional)购买份数, 默认1份
#quantity = 1
`
)

type Eip struct {
	MetricName string
	PayAsYouGo bool
	Interval   internal.Duration

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
