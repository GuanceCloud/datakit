/*
 * BSS
 *
 * Business Support System API
 *
 */

package model

import (
	"encoding/json"

	"strings"
)

type RateOnDemandReq struct {
	// |参数名称：项目ID| |参数约束及描述：如果使用客户AK/SK或者Token，可以调用“通过assume_role方式获取用户token”接口获取“regionId”取值对应的project id。|
	ProjectId string `json:"project_id"`
	// |参数名称：精度模式| |参数约束及描述：精度模式：0：询价结果默认精度截取，按需最长保留到元后6位小数点，如0.000001元;1：询价结果保留10位精度，即最长保留到分后10位小数点，如：1.0000000001元. 说明：如果定价只到元后2位或者3位，那么价格也只到元后2位或者3位，不管传0或者穿1都一样，只有定价定到了小数点后面6位以上，传0和传1才有区别。|
	InquiryPrecision *int32 `json:"inquiry_precision,omitempty"`
	// |参数名称：产品信息列表| |参数的约束及描述：询价时要询价产品的信息的列表|
	ProductInfos []DemandProductInfo `json:"product_infos"`
}

func (o RateOnDemandReq) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "RateOnDemandReq struct{}"
	}

	return strings.Join([]string{"RateOnDemandReq", string(data)}, " ")
}
