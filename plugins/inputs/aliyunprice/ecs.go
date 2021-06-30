package aliyunprice

import (
	"fmt"
	"time"

	"github.com/aliyun/alibaba-cloud-sdk-go/sdk/requests"
	"github.com/aliyun/alibaba-cloud-sdk-go/services/bssopenapi"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
)

const (
	ecsSampleConfig = `
#[[inputs.aliyunprice.ecs]]

# ##(optional) custom metric name，default is 'aliyun_price'
#metric_name = ''

# ##(optional) collect interval，default is one day
#interval = '1d'

# ##(required) cllect PayAsYouGo price, default is false
#pay_as_you_go = false

# ##(required) ecs instance region
#region = "cn-hangzhou-dg-a01"

# ##(required) ecs instance
#instance_type = 'ecs.g5.xlarge'

# ##(required) ecs instance catalog
#instance_type_family = 'ecs.g5'

# ##(required) ecs instance os
#image_os = "linux"

# ##(required) system disk category
#system_disk_category = 'cloud_ssd'

# ##(required) system disk size, unit is GB, range:20-500
#system_disk_size = 20

# ##(required) whether bandwidth pay by traffic
#pay_by_traffic = false

# ##(required) for fixed bandwidth, is the vale; for pay by traffic, is the max value. unit is kbps
#internet_max_bandwidth_out = 1024

# ##(optional)Purchase duration, default is 1, so if unit is Year, then is one year
#service_period_quantity = 1

# ##(optional)unit of purchase duration: Month，Year, defalut is Year
#service_period_unit = "Year"

# ##(optional)Purchase quantity, default is 1
#quantity = 1

# ##(optional)data disk configuration
#[[inputs.aliyunprice.ecs.data_disks]]
# ##data disk category
#data_disk_category = 'cloud_ssd'

# ##data disk size, unit is GB
#data_disk_size = 40
`
)

type DataDisk struct {
	DataDiskCategory string //cloud, cloud_ssd...
	DataDiskSize     int    //GB
}

type Ecs struct {
	MetricName string
	PayAsYouGo bool
	Interval   datakit.Duration

	Region string

	//实例规格
	InstanceType       string //ecs.t1.small
	InstanceTypeFamily string //ecs.g5
	//IoOptimized        bool   //是否I/O 优化实例，默认false
	ImageOs string //windows or linux

	//系统盘
	SystemDiskCategory string //cloud, cloud_ssd, cloud_efficiency...
	SystemDiskSize     int    //GB

	//带宽/流量
	PayByTraffic            bool  //是否按流量计算
	InternetMaxBandwidthOut int64 //Kbps 使用固定带宽时的固定带宽值，默认1M

	ServicePeriodQuantity int    //购买时长
	ServicePeriodUnit     string //购买时长单位	Month，Year
	Quantity              int    //购买数量	整数

	//数据盘
	DataDisks []*DataDisk
}

func (e *Ecs) toRequest() (*priceReq, error) {
	if e.Quantity == 0 {
		e.Quantity = 1
	}

	if e.ServicePeriodQuantity == 0 {
		e.ServicePeriodQuantity = 1
	}

	if e.ServicePeriodUnit == "" {
		e.ServicePeriodUnit = "Year"
	}

	if e.ImageOs == "" {
		e.ImageOs = "linux"
	}

	p := &priceReq{
		m:                                   e,
		fetchModulePriceHistory:             make(map[string]time.Time),
		priceModuleInfos:                    make(map[string]*bssopenapi.ModuleList),
		productCodeForPriceModulesSubscript: "ecs",
		productCodeForPriceModulesPayasugo:  "ecs",
		//productTypeForPriceModulesPayasugo:  "bards",
	}
	p.payAsYouGo = e.PayAsYouGo
	p.metricName = e.MetricName
	p.interval = e.Interval.Duration
	if p.interval < 5*time.Minute {
		p.interval = 5 * time.Minute
	}
	p.region = e.Region

	instanceTypeConfig := fmt.Sprintf("InstanceType:%s,IoOptimized:IoOptimized,ImageOs:%s,InstanceTypeFamily:%s", e.InstanceType, e.ImageOs, e.InstanceTypeFamily)

	systemDiskConfig := fmt.Sprintf("SystemDisk.Category:%s,SystemDisk.Size:%d", e.SystemDiskCategory, e.SystemDiskSize)

	internetConfig := "NetworkType:1"
	if !e.PayByTraffic {
		if e.PayAsYouGo {
			internetConfig += fmt.Sprintf(`,InternetMaxBandwidthOut:%v,InternetMaxBandwidthOut.IsFlowType:0`, e.InternetMaxBandwidthOut)
		} else {
			internetConfig += fmt.Sprintf(`,InternetMaxBandwidthOut:%v,InternetMaxBandwidthOut.IsFlowType:5`, e.InternetMaxBandwidthOut)
		}
	} else {
		internetConfig += fmt.Sprintf(`,InternetMaxBandwidthOut:%v,InternetMaxBandwidthOut.IsFlowType:1`, e.InternetMaxBandwidthOut)
	}

	if e.PayAsYouGo {
		p.payasyougoReq = bssopenapi.CreateGetPayAsYouGoPriceRequest()
		p.payasyougoReq.Scheme = "https"
		p.payasyougoReq.ProductCode = "ecs"
		p.payasyougoReq.SubscriptionType = `PayAsYouGo`
		p.payasyougoReq.Region = e.Region

		mods := []bssopenapi.GetPayAsYouGoPriceModuleList{
			{
				ModuleCode: "InstanceType",
				Config:     instanceTypeConfig,
				PriceType:  "Hour",
			},

			{
				ModuleCode: "SystemDisk",
				Config:     systemDiskConfig,
				PriceType:  "Hour",
			},
			{
				ModuleCode: "InternetMaxBandwidthOut",
				Config:     internetConfig,
				PriceType:  "Hour",
			},
		}

		for _, dd := range e.DataDisks {
			mods = append(mods, bssopenapi.GetPayAsYouGoPriceModuleList{
				ModuleCode: "DataDisk",
				Config:     fmt.Sprintf(`DataDisk.Category:%s,DataDisk.Size:%d`, dd.DataDiskCategory, dd.DataDiskSize),
				PriceType:  "Hour",
			})
		}
		p.payasyougoReq.ModuleList = &mods

	} else {
		p.subscriptionReq = bssopenapi.CreateGetSubscriptionPriceRequest()
		p.subscriptionReq.Scheme = `https`
		p.subscriptionReq.ProductCode = "ecs"
		p.subscriptionReq.SubscriptionType = `Subscription`
		p.subscriptionReq.OrderType = `NewOrder`
		p.subscriptionReq.Quantity = requests.NewInteger(e.Quantity)
		p.subscriptionReq.ServicePeriodQuantity = requests.NewInteger(e.ServicePeriodQuantity)
		p.subscriptionReq.ServicePeriodUnit = e.ServicePeriodUnit
		p.subscriptionReq.Region = e.Region

		mods := []bssopenapi.GetSubscriptionPriceModuleList{
			{
				ModuleCode: "InstanceType",
				Config:     instanceTypeConfig,
			},
			// {
			// 	ModuleCode: "InternetTrafficOut",
			// 	Config:     fmt.Sprintf("Region:%s", e.Region),
			// },
			{
				ModuleCode: "SystemDisk",
				Config:     systemDiskConfig,
			},
			{
				ModuleCode: "InternetMaxBandwidthOut",
				Config:     internetConfig,
			},
		}
		for _, dd := range e.DataDisks {
			mods = append(mods, bssopenapi.GetSubscriptionPriceModuleList{
				ModuleCode: "DataDisk",
				Config:     fmt.Sprintf(`DataDisk.Category:%s,DataDisk.Size:%v`, dd.DataDiskCategory, dd.DataDiskSize),
			})
		}

		p.subscriptionReq.ModuleList = &mods
	}

	return p, nil
}

func (e *Ecs) handleTags(tags map[string]string) map[string]string {
	tags["InstanceType"] = e.InstanceType
	tags["InstanceTypeFamily"] = e.InstanceTypeFamily
	tags["ImageOs"] = e.ImageOs
	tags["SystemDiskCategory"] = e.SystemDiskCategory
	if e.PayByTraffic {
		tags["PayByTraffic"] = "1"
	} else {
		tags["PayByTraffic"] = "0"
	}
	tags["InternetMaxBandwidthOut"] = fmt.Sprintf("%d", e.InternetMaxBandwidthOut)
	tags["SystemDiskSize"] = fmt.Sprintf("%d", e.SystemDiskSize)
	tags["Quantity"] = fmt.Sprintf("%d", e.Quantity)
	tags["ServicePeriodQuantity"] = fmt.Sprintf("%d", e.ServicePeriodQuantity)
	tags["ServicePeriodUnit"] = e.ServicePeriodUnit

	return tags
}

func (e *Ecs) handleFields(fields map[string]interface{}) map[string]interface{} {
	return fields
}
