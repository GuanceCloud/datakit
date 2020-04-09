package aliyunprice

import (
	"fmt"
	"log"
	"sort"
	"testing"

	"github.com/influxdata/toml"

	"github.com/aliyun/alibaba-cloud-sdk-go/sdk/requests"
	"github.com/aliyun/alibaba-cloud-sdk-go/services/bssopenapi"
	"github.com/aliyun/alibaba-cloud-sdk-go/services/ecs"
)

func apiClient() *bssopenapi.Client {
	client, err := bssopenapi.NewClientWithAccessKey("cn-hangzhou", "LTAI4FmCPgfKHVwDsPXEnVaF", "U5kt6Ce5Dmm5iqgJK1Gu2QSdfAyYrS")
	if err != nil {
		log.Fatalf("%s", err)
	}
	return client
}

func TestConfig(t *testing.T) {
	ecs1 := &Ecs{
		MetricName: "11",
		PayAsYouGo: false,
		//Interval:   "1h",
		Region:                  "cn-hangzhou-dg-a01",
		InstanceType:            "",
		InstanceTypeFamily:      "",
		ImageOs:                 "linux",
		SystemDiskCategory:      "cloud_ssd",
		SystemDiskSize:          20,
		PayByTraffic:            false,
		InternetMaxBandwidthOut: 1024,
		DataDisks: []*DataDisk{
			&DataDisk{
				DataDiskCategory: "cloud_ssd",
				DataDiskSize:     40,
			},
		},
		ServicePeriodQuantity: 1,
		ServicePeriodUnit:     "Year",
		Quantity:              1,
	}

	cfg := AliyunPriceAgent{
		EcsCfg: []*Ecs{ecs1},
	}

	if data, err := toml.Marshal(&cfg); err != nil {
		t.Errorf("%s", err)
	} else {
		log.Printf("%s", string(data))
	}
}

//TestProductPriceModule 某个产品的对应付费模块信息 https://help.aliyun.com/document_detail/96469.html?spm=a2c4g.11186623.2.13.5a21634fRfjUAL
func TestProductPriceModule(t *testing.T) {

	//regionID := ""
	productCode := "ecs"
	subscriptionType := "PayAsYouGo"

	req := bssopenapi.CreateDescribePricingModuleRequest()
	req.Scheme = `https`
	//req.RegionId = regionID
	req.ProductCode = productCode
	req.SubscriptionType = subscriptionType

	resp, err := apiClient().DescribePricingModule(req)
	if err != nil {
		log.Fatalf("DescribePricingModule failed: %s", err)
	}

	// for _, attr := range resp.Data.AttributeList.Attribute {
	// 	//attr.Values.AttributeValue //属性的值的取值范围
	// 	if attr.Code == "InstanceType" {
	// 		for _, v := range attr.Values.AttributeValue {
	// 			fmt.Printf("%s : %s - %s\n", v.Name, v.Value, v.Remark)
	// 		}
	// 	}
	// }

	for _, mod := range resp.Data.ModuleList.Module {
		log.Printf("%s", mod)
	}
}

func TestDescribeRegions(t *testing.T) {

}

func TestGetImages(t *testing.T) {
	cli, _ := ecs.NewClientWithAccessKey(`cn-hangzhou`, `LTAI4FmCPgfKHVwDsPXEnVaF`, `U5kt6Ce5Dmm5iqgJK1Gu2QSdfAyYrS`)

	req := ecs.CreateDescribeImagesRequest()
	req.Scheme = "https"
	//req.InstanceType = `ecs.g5.xlarge`
	//req.OSType = "linux"
	//req.Architecture = "x86_64"
	req.PageSize = requests.NewInteger(100)

	response, err := cli.DescribeImages(req)
	if err != nil {
		log.Fatalf("%s", err)
	}
	items := []string{}
	for _, img := range response.Images.Image {
		items = append(items, fmt.Sprintf("%s,%s", img.OSName, img.ImageId))
	}
	sort.Strings(items)
	for _, item := range items {
		fmt.Printf("%s\n", item)
	}
}

func TestGetEcsSubscriptionPrice(t *testing.T) {
	req := bssopenapi.CreateGetSubscriptionPriceRequest()
	req.Scheme = "https"
	req.ProductCode = "ecs"
	req.SubscriptionType = `Subscription`
	req.OrderType = `NewOrder`
	req.Quantity = requests.NewInteger(1)
	req.ServicePeriodQuantity = requests.NewInteger(1)
	req.ServicePeriodUnit = `Year`

	mods := []bssopenapi.GetSubscriptionPriceModuleList{
		bssopenapi.GetSubscriptionPriceModuleList{
			ModuleCode: "InstanceType",
			Config:     `InstanceType:ecs.g5.xlarge,IoOptimized:IoOptimized,ImageOs:linux,InstanceTypeFamily:ecs.g5`,
		},
		bssopenapi.GetSubscriptionPriceModuleList{
			ModuleCode: "SystemDisk",
			Config:     `SystemDisk.Category:cloud_efficiency,SystemDisk.Size:20`,
		},
		// bssopenapi.GetSubscriptionPriceModuleList{
		// 	ModuleCode: "ImageOs",
		// 	Config:     `Linux:linux`,
		// },
		bssopenapi.GetSubscriptionPriceModuleList{
			ModuleCode: "InternetMaxBandwidthOut",
			Config:     `InternetMaxBandwidthOut:1024,InternetMaxBandwidthOut.IsFlowType:5,NetworkType:1`,
		},
		bssopenapi.GetSubscriptionPriceModuleList{
			ModuleCode: "Region",
			Config:     `Region:ap-southeast-os30-a01`,
		},
		// bssopenapi.GetSubscriptionPriceModuleList{
		// 	ModuleCode: "DataDisk",
		// 	Config:     `DataDisk.Category:cloud_ssd,DataDisk.Size:130`,
		// },
		// bssopenapi.GetSubscriptionPriceModuleList{
		// 	ModuleCode: "DataDisk",
		// 	Config:     `DataDisk.Category:cloud_ssd,DataDisk.Size:130`,
		// },
	}

	req.ModuleList = &mods

	resp, err := apiClient().GetSubscriptionPrice(req)
	if err != nil {
		log.Fatalf("%s", err)
	}
	log.Printf("resp: %s", resp.String())

}

func TestGetEcsPayAsYouGoPrice(t *testing.T) {
	req := bssopenapi.CreateGetPayAsYouGoPriceRequest()
	req.Scheme = "https"
	req.ProductCode = "ecs"
	req.SubscriptionType = `PayAsYouGo`

	mods := []bssopenapi.GetPayAsYouGoPriceModuleList{
		bssopenapi.GetPayAsYouGoPriceModuleList{
			ModuleCode: "InstanceType",
			Config:     `InstanceType:ecs.s6-c1m1.small,IoOptimized:IoOptimized,ImageOs:linux,InstanceTypeFamily:ecs.s6`,
			PriceType:  "Hour",
		},
		bssopenapi.GetPayAsYouGoPriceModuleList{
			ModuleCode: "SystemDisk",
			Config:     `SystemDisk.Category:cloud_efficiency,SystemDisk.Size:20`,
			PriceType:  "Hour",
		},
		bssopenapi.GetPayAsYouGoPriceModuleList{
			ModuleCode: "InternetMaxBandwidthOut",
			Config:     `InternetMaxBandwidthOut:1024,InternetMaxBandwidthOut.IsFlowType:1`,
			PriceType:  "Hour",
		},
		bssopenapi.GetPayAsYouGoPriceModuleList{
			ModuleCode: "Region",
			Config:     `Region:cn-hangzhou-dg-a01`,
			PriceType:  "Hour",
		},
		// bssopenapi.GetPayAsYouGoPriceModuleList{
		// 	ModuleCode: "DataDisk",
		// 	Config:     `DataDisk.Category:cloud_ssd,DataDisk.Size:130`,
		// 	PriceType:  "Hour",
		// },
		// bssopenapi.GetPayAsYouGoPriceModuleList{
		// 	ModuleCode: "DataDisk",
		// 	Config:     `DataDisk.Category:cloud_ssd,DataDisk.Size:130`,
		// 	PriceType:  "Hour",
		// },
	}

	req.ModuleList = &mods

	resp, err := apiClient().GetPayAsYouGoPrice(req)
	if err != nil {
		log.Fatalf("%s", err)
	}
	log.Printf("resp: %s", resp.String())

}

func TestEIPPrice(t *testing.T) {

	req := bssopenapi.CreateGetPayAsYouGoPriceRequest()
	req.Scheme = "https"
	req.ProductCode = "eip"
	req.SubscriptionType = `PayAsYouGo`

	mods := []bssopenapi.GetPayAsYouGoPriceModuleList{
		bssopenapi.GetPayAsYouGoPriceModuleList{
			ModuleCode: "Bandwidth",
			Config:     `Bandwidth:1024`,
			PriceType:  "Day",
		},
		bssopenapi.GetPayAsYouGoPriceModuleList{
			ModuleCode: "InternetChargeType",
			Config:     `InternetChargeType:1`,
			PriceType:  "Usage",
		},
		// bssopenapi.GetPayAsYouGoPriceModuleList{
		// 	ModuleCode: "ISP",
		// 	Config:     `ISP:BGP`,
		// 	PriceType:  "Hour",
		// },
		bssopenapi.GetPayAsYouGoPriceModuleList{
			ModuleCode: "Region",
			Config:     `Region:cn-hangzhou-dg-a01`,
			PriceType:  "Hour",
		},
	}
	req.ModuleList = &mods

	resp, err := apiClient().GetPayAsYouGoPrice(req)
	if err != nil {
		log.Fatalf("%s", err)
	}
	log.Printf("resp: %s", resp.String())
}
