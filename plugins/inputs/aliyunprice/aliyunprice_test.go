package aliyunprice

import (
	"fmt"
	"log"
	"testing"

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

//TestProductPriceModule 某个产品的对应付费模块信息 https://help.aliyun.com/document_detail/96469.html?spm=a2c4g.11186623.2.13.5a21634fRfjUAL
func TestProductPriceModule(t *testing.T) {

	//regionID := ""
	productCode := "ecs"
	subscriptionType := "Subscription"

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
	for _, img := range response.Images.Image {
		fmt.Printf("%s - %s\n", img.OSName, img.ImageId)
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
	req.ServicePeriodUnit = `Month`

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
			Config:     `InstanceType:ecs.g5.xlarge,IoOptimized:IoOptimized,ImageOs:linux,ImageID:coreos_2023_4_0_64_30G_alibase_20190319.vhd,InstanceTypeFamily:ecs.g5`,
			PriceType:  "Hour",
		},
		bssopenapi.GetPayAsYouGoPriceModuleList{
			ModuleCode: "SystemDisk",
			Config:     `SystemDisk.Category:cloud_efficiency,SystemDisk.Size:55`,
			PriceType:  "Hour",
		},
		bssopenapi.GetPayAsYouGoPriceModuleList{
			ModuleCode: "InternetMaxBandwidthOut",
			Config:     `InternetMaxBandwidthOut:1024`,
			PriceType:  "Hour",
		},
		bssopenapi.GetPayAsYouGoPriceModuleList{
			ModuleCode: "Region",
			Config:     `Region:ap-southeast-os30-a01`,
			PriceType:  "Hour",
		},
		bssopenapi.GetPayAsYouGoPriceModuleList{
			ModuleCode: "DataDisk",
			Config:     `DataDisk.Category:cloud_ssd,DataDisk.Size:130`,
			PriceType:  "Hour",
		},
		bssopenapi.GetPayAsYouGoPriceModuleList{
			ModuleCode: "DataDisk",
			Config:     `DataDisk.Category:cloud_ssd,DataDisk.Size:130`,
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
