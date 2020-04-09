package aliyunprice

import (
	"fmt"
	"time"

	"github.com/aliyun/alibaba-cloud-sdk-go/sdk/requests"
	"github.com/aliyun/alibaba-cloud-sdk-go/services/bssopenapi"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal"
)

const (
	sampleConfig = `
#region_id = ''
#access_id = ''
#access_key = ''

#[[ecs]]
#  ## optional, if empty, use 'aliyun_price'
#metric_name = ''
#interval = '1h'
#pay_as_you_go = false
#region = "cn-hangzhou-dg-a01"
#instance_type = 'ecs.g5.xlarge'
#instance_type_family = 'ecs.g5'
#  ## windows or linux
#image_os = "linux"
#  ## cloud_ssd=SSD 云盘; cloud_efficiency=高效云盘; cloud=普通云盘; ephemeral_ssd=本地SSD盘
#  ## see: https://help.aliyun.com/document_detail/25382.html?spm=5176.ecsbuyv3.storage.2.4d0e3675s1tQjx
#system_disk_category = "cloud_ssd"
#  ##unit:GB, range:20-500
#system_disk_size = 20
#  ## Bandwidth is billed based on the amount of data used in GB by hour
#pay_by_traffic = false
#  ##unit: kbps
#internet_max_bandwidth_out = 1024

# ##购买时长
#service_period_quantity = 1
# ##购买时长单位: Month，Year
#service_period_unit = "Year"
# ##购买数量
#quantity = 1

#[[ecs.data_disks]]
#data_disk_category = 'cloud_ssd'
# ##unit:GB
#data_disk_size = 40
`
)

type (
	DataDisk struct {
		DataDiskCategory string //cloud, cloud_ssd...
		DataDiskSize     int    //GB
	}

	Ecs struct {
		MetricName string
		PayAsYouGo bool
		Interval   internal.Duration

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

	Rds struct {
		MetricName string
		PayAsYouGo bool
		Interval   internal.Duration

		Engine                string
		EngineVersion         string
		Series                string
		DBInstanceStorageType string
		DBInstanceStorage     int
		DBInstanceClass       string
		DBNetworkType         string

		ServicePeriodQuantity int
		ServicePeriodUnit     string
		Quantity              int
	}
)

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

	p := &priceReq{}
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
		internetConfig += `,InternetMaxBandwidthOut.IsFlowType:1`
	}

	if e.PayAsYouGo {
		p.payasyougoReq = bssopenapi.CreateGetPayAsYouGoPriceRequest()
		p.payasyougoReq.Scheme = "https"
		p.payasyougoReq.ProductCode = "ecs"
		p.payasyougoReq.SubscriptionType = `PayAsYouGo`

		mods := []bssopenapi.GetPayAsYouGoPriceModuleList{
			bssopenapi.GetPayAsYouGoPriceModuleList{
				ModuleCode: "InstanceType",
				Config:     instanceTypeConfig,
				PriceType:  "Hour",
			},

			bssopenapi.GetPayAsYouGoPriceModuleList{
				ModuleCode: "SystemDisk",
				Config:     systemDiskConfig,
				PriceType:  "Hour",
			},
			bssopenapi.GetPayAsYouGoPriceModuleList{
				ModuleCode: "InternetMaxBandwidthOut",
				Config:     internetConfig,
				PriceType:  "Hour",
			},
			bssopenapi.GetPayAsYouGoPriceModuleList{
				ModuleCode: "Region",
				Config:     fmt.Sprintf(`Region:%s`, e.Region),
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

		mods := []bssopenapi.GetSubscriptionPriceModuleList{
			bssopenapi.GetSubscriptionPriceModuleList{
				ModuleCode: "InstanceType",
				Config:     instanceTypeConfig,
			},
			bssopenapi.GetSubscriptionPriceModuleList{
				ModuleCode: "SystemDisk",
				Config:     systemDiskConfig,
			},
			bssopenapi.GetSubscriptionPriceModuleList{
				ModuleCode: "InternetMaxBandwidthOut",
				Config:     internetConfig,
			},
			bssopenapi.GetSubscriptionPriceModuleList{
				ModuleCode: "Region",
				Config:     fmt.Sprintf(`Region:%s`, e.Region),
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
