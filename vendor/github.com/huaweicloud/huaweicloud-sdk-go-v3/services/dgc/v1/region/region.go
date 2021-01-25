package region

import (
	"fmt"
	"github.com/huaweicloud/huaweicloud-sdk-go-v3/core/region"
)

var CN_NORTH_1 = region.NewRegion("cn-north-1", "https://dayu-dlf.cn-north-1.myhuaweicloud.com")
var CN_NORTH_4 = region.NewRegion("cn-north-4", "https://dayu-dlf.cn-north-4.myhuaweicloud.com")
var CN_EAST_3 = region.NewRegion("cn-east-3", "https://dayu-dlf.cn-east-3.myhuaweicloud.com")
var CN_EAST_2 = region.NewRegion("cn-east-2", "https://dayu-dlf.cn-east-2.myhuaweicloud.com")
var CN_SOUTH_1 = region.NewRegion("cn-south-1", "https://dayu-dlf.cn-south-1.myhuaweicloud.com")
var AP_SOUTHEAST_3 = region.NewRegion("ap-southeast-3", "https://dayu-dlf.ap-southeast-3.myhuaweicloud.com")
var AP_SOUTHEAST_1 = region.NewRegion("ap-southeast-1", "https://dayu-dlf.ap-southeast-3.myhuaweicloud.com")
var RU_NORTHWEST_2 = region.NewRegion("ru-northwest-2", "https://dayu-dlf.ru-northwest-2.myhuaweicloud.com")

var staticFields = map[string]*region.Region{
	"cn-north-1":     CN_NORTH_1,
	"cn-north-4":     CN_NORTH_4,
	"cn-east-3":      CN_EAST_3,
	"cn-east-2":      CN_EAST_2,
	"cn-south-1":     CN_SOUTH_1,
	"ap-southeast-3": AP_SOUTHEAST_3,
	"ap-southeast-1": AP_SOUTHEAST_1,
	"ru-northwest-2": RU_NORTHWEST_2,
}

func ValueOf(regionId string) *region.Region {
	if regionId == "" {
		panic("unexpected empty parameter: regionId")
	}
	if _, ok := staticFields[regionId]; ok {
		return staticFields[regionId]
	}
	panic(fmt.Sprintf("unexpected regionId: %s", regionId))
}
