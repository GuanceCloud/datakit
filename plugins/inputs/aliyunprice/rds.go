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
# ## hints for configuration of rds
#
# ## 'engine': mysql, mssql, PostgreSQL, PPAS, MariaDB
# ## 'engint-version' is set according to 'engine', e.g., 5.7 for mysql
#
# ## 'series':
# ##   AlwaysOn #-集群版
# ##   HighAvailability #-高可用版
# ##   Finance #-三节点企业版
# ##   Basic #-基础版
#
# ## 'db_instance_class': according to engine, e.g., "rds.mysql.s2.large". 
# ##   See https://help.aliyun.com/document_detail/26312.html?spm=a2c4g.11186623.2.14.37cc2c6crjKV5k
#
# ## 'db_instance_storage_type':
# ##   local_ssd #-本地SSD盘
# ##   cloud_essd #-ESSD云盘
# ##   cloud_ssd #-SSD云盘
# ##   cloud_essd2 #-ESSD PL2云盘
# ##   cloud_essd3 #-ESSD PL3云盘
#
# ## 'db_instance_storage': unit is GB
#
# ## 'db_network_type': 0:经典网络, 1:专用网络

#[[rds]]
##metric_name = 'rds_price'
##description = ''
##interval = '1d'
##pay_as_you_go = false
#region = "cn-hangzhou"
#engine = ''
#engine_version = ''
#series = ''
#db_instance_storage_type = ''
#db_instance_storage = 20
#db_instance_class = ''
#db_network_type = 0
##service_period_quantity = 1
##service_period_unit = "Year"
##quantity = 1
`
)

type Rds struct {
	MetricName  string
	Description string
	PayAsYouGo  bool
	Interval    internal.Duration

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

	tags["Description"] = e.Description
	tags["Engine"] = e.Engine
	tags["EngineVersion"] = e.EngineVersion
	tags["Series"] = e.Series
	tags["DBInstanceStorageType"] = e.DBInstanceStorageType
	tags["DBInstanceStorage"] = fmt.Sprintf("%d", e.DBInstanceStorage)
	tags["DBInstanceClass"] = e.DBInstanceClass
	tags["Quantity"] = fmt.Sprintf("%d x %d%s", e.Quantity, e.ServicePeriodQuantity, e.ServicePeriodUnit)

	return tags
}

func (e *Rds) handleFields(fields map[string]interface{}) map[string]interface{} {
	return fields
}
