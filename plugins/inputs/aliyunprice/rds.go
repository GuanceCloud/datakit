package aliyunprice

import (
	"fmt"
	"time"

	"github.com/aliyun/alibaba-cloud-sdk-go/sdk/requests"
	"github.com/aliyun/alibaba-cloud-sdk-go/services/bssopenapi"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal"
)

const (
	rdsSampleConfig = `
# ##云数据库 RDS
#[[rds]]

# ##(optional) 自定义指标集名称，默认使用 aliyun_price
#metric_name = ''

# ##(optional) 采集间隔，默认一天
#interval = '1d'

# ##(required) 是否采集后付费价格, 默认false(即采集预付费价格)
#pay_as_you_go = false

# ##(required) 实例所在区域
#region = "cn-hangzhou"

# ##(required) 数据库类型, 支持 mysql, mssql, PostgreSQL, PPAS, MariaDB
#engine = 'mysql'

# ##(required) 数据库版本, 根据数据库类型设置
#engine_version = '8.0'

# ##(required) 数据库系列
# ## AlwaysOn #-集群版
# ## HighAvailability #-高可用版
# ## Finance #-三节点企业版
# ## Basic #-基础版
#series = 'Basic'

# ##(required) 存储类型
# ## local_ssd #-本地SSD盘
# ## cloud_essd #-ESSD云盘
# ## cloud_ssd #-SSD云盘
# ## cloud_essd2 #-ESSD PL2云盘
# ## cloud_essd3 #-ESSD PL3云盘
#db_instance_storage_type = 'cloud_ssd'

# ##(required) 存储大小, 单位GB
#db_instance_storage = 20

# ##(required) 实例规格, 可参考: https://help.aliyun.com/document_detail/26312.html?spm=a2c4g.11186623.2.14.37cc2c6crjKV5k
#db_instance_class = 'mysql.n2.medium.1'

# ##(optional) 网络类型, 0:经典网络, 1:专用网络
#db_network_type = 0

# ##(optional)购买时长, 默认为1, 如果单位为Year, 则表示1年
#service_period_quantity = 1

# ##(optional)购买时长单位: Month，Year, 默认为 Year
#service_period_unit = "Year"

# ##(optional)购买份数, 默认1份
#quantity = 1
`
)

type Rds struct {
	MetricName string
	PayAsYouGo bool
	Interval   internal.Duration

	Region string

	Engine                string
	EngineVersion         string
	Series                string
	DBInstanceStorageType string `toml:"db_instance_storage_type"`
	DBInstanceStorage     int    `toml:"db_instance_storage"`
	DBInstanceClass       string `toml:"db_instance_class"`
	DBNetworkType         int    `toml:"db_network_type"` //0：经典网络。1：专有网络。

	ServicePeriodQuantity int
	ServicePeriodUnit     string
	Quantity              int
}

func (e *Rds) toRequest() (*priceReq, error) {
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
		fetchModulePriceHistory:             make(map[string]time.Time),
		priceModuleInfos:                    make(map[string]*bssopenapi.ModuleList),
		productCodeForPriceModulesSubscript: "rds",
		productTypeForPriceModulesSubscript: "rds",
		productCodeForPriceModulesPayasugo:  "rds",
		productTypeForPriceModulesPayasugo:  "bards",
	}
	p.payAsYouGo = e.PayAsYouGo
	p.metricName = e.MetricName
	p.interval = e.Interval.Duration
	if p.interval == 0 {
		p.interval = 5 * time.Minute
	}
	p.region = e.Region

	//engineConfig := fmt.Sprintf("Engine:%s", e.Engine)
	//engineVersionConfig := fmt.Sprintf("EngineVersion:%s", e.EngineVersion)
	//seriesConfig := fmt.Sprintf("Series:%s", e.Series)
	//dbInstanceStorageTypeConfig := fmt.Sprintf("DBInstanceStorageType:%s", e.DBInstanceStorageType)
	//dbInstanceStorageConfig := fmt.Sprintf("DBInstanceStorage:%d", e.DBInstanceStorage)
	//dbInstanceStorageConfig := fmt.Sprintf("DBInstanceStorage:%d,Series:%s,DBInstanceStorageType:%s,Engine:%s,Region:%s", e.DBInstanceStorage, e.Series, e.DBInstanceStorageType, e.Engine, e.Region)
	//dbInstanceClassConfig := fmt.Sprintf("DBInstanceClass:%s,EngineVersion:5.7,Region:cn-hangzhou", e.DBInstanceClass)
	//dbInstanceClassConfig := fmt.Sprintf("DBInstanceClass:%s", e.DBInstanceClass)
	//dbNetworkTypeConfig := fmt.Sprintf("DBNetworkType:%d", e.DBNetworkType)

	if e.PayAsYouGo {
		p.payasyougoReq = bssopenapi.CreateGetPayAsYouGoPriceRequest()
		p.payasyougoReq.Scheme = "https"
		p.payasyougoReq.ProductCode = "rds"
		p.payasyougoReq.ProductType = "bards"
		p.payasyougoReq.SubscriptionType = `PayAsYouGo`
		p.payasyougoReq.Region = e.Region

		mods := []bssopenapi.GetPayAsYouGoPriceModuleList{
			{
				//按实际使用流量每日扣费
				ModuleCode: "DBFlowType",
				Config:     fmt.Sprintf("Region:%s,DBFlowType:1", e.Region),
				PriceType:  "Usage",
			},
			{
				ModuleCode: "DBInstanceStorage",
				Config:     fmt.Sprintf("DBInstanceStorage:%d,Series:%s,DBInstanceStorageType:%s,Engine:%s,EngineVersion:%s,Region:%s", e.DBInstanceStorage, e.Series, e.DBInstanceStorageType, e.Engine, e.EngineVersion, e.Region),
				PriceType:  "Hour",
			},
			{
				ModuleCode: "DBInstanceClass",
				Config:     fmt.Sprintf("DBInstanceClass:%s,EngineVersion:%s,Region:%s", e.DBInstanceClass, e.EngineVersion, e.Region),
				PriceType:  "Hour",
			},
		}

		p.payasyougoReq.ModuleList = &mods

	} else {
		p.subscriptionReq = bssopenapi.CreateGetSubscriptionPriceRequest()
		p.subscriptionReq.Scheme = `https`
		p.subscriptionReq.ProductCode = "rds"
		p.subscriptionReq.ProductType = "rds"
		p.subscriptionReq.SubscriptionType = `Subscription`
		p.subscriptionReq.OrderType = `NewOrder`
		p.subscriptionReq.Quantity = requests.NewInteger(e.Quantity)
		p.subscriptionReq.ServicePeriodQuantity = requests.NewInteger(e.ServicePeriodQuantity)
		p.subscriptionReq.ServicePeriodUnit = e.ServicePeriodUnit
		p.subscriptionReq.Region = e.Region

		mods := []bssopenapi.GetSubscriptionPriceModuleList{
			{
				ModuleCode: "DBInstanceStorage",
				Config:     fmt.Sprintf("DBInstanceStorage:%d,Series:%s,DBInstanceStorageType:%s,Engine:%s,EngineVersion:%s,Region:%s", e.DBInstanceStorage, e.Series, e.DBInstanceStorageType, e.Engine, e.EngineVersion, e.Region),
			},
			{
				ModuleCode: "DBInstanceClass",
				Config:     fmt.Sprintf("DBInstanceClass:%s,Engine:%s,EngineVersion:%s,Region:%s", e.DBInstanceClass, e.Engine, e.EngineVersion, e.Region),
			},
			{
				ModuleCode: "DBNetworkType",
				Config:     fmt.Sprintf("DBNetworkType:%d", e.DBNetworkType),
			},
		}

		p.subscriptionReq.ModuleList = &mods
	}

	return p, nil
}

func (e *Rds) handleTags(tags map[string]string) map[string]string {

	tags["Engine"] = e.Engine
	tags["EngineVersion"] = e.EngineVersion
	tags["Series"] = e.Series
	tags["DBInstanceStorageType"] = e.DBInstanceStorageType
	tags["DBInstanceStorage"] = fmt.Sprintf("%d", e.DBInstanceStorage)
	tags["DBInstanceClass"] = e.DBInstanceClass
	tags["Quantity"] = fmt.Sprintf("%d", e.Quantity)
	tags["ServicePeriodQuantity"] = fmt.Sprintf("%d", e.ServicePeriodQuantity)
	tags["ServicePeriodUnit"] = e.ServicePeriodUnit

	return tags
}

func (e *Rds) handleFields(fields map[string]interface{}) map[string]interface{} {
	return fields
}
