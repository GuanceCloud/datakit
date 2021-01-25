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

type DemandProductRatingResult struct {
	// |参数名称：ID标识| |参数约束及描述：同一次询价中不能重复，用于标识返回询价结果和请求的映射关系|
	Id *string `json:"id,omitempty"`
	// |参数名称：寻到的产品ID| |参数约束及描述：寻到的产品ID|
	ProductId *string `json:"product_id,omitempty"`
	// |参数名称：总额| |参数约束及描述：即最终优惠的金额|
	Amount float32 `json:"amount,omitempty"`
	// |参数名称：优惠额（官网价和总价的差）| |参数约束及描述：优惠额（官网价和总价的差）|
	DiscountAmount float32 `json:"discount_amount,omitempty"`
	// |参数名称：官网价| |参数约束及描述：官网价|
	OfficialWebsiteAmount float32 `json:"official_website_amount,omitempty"`
	// |参数名称：度量单位标识| |参数约束及描述：1：元|
	MeasureId *int32 `json:"measure_id,omitempty"`
	// |参数名称：折扣优惠明细| |参数的约束及描述：包含产品本身的促销信息，同时包含商务或者伙伴折扣的优惠信息|
	DiscountRatingResults *[]DemandDiscountRatingResult `json:"discount_rating_results,omitempty"`
}

func (o DemandProductRatingResult) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "DemandProductRatingResult struct{}"
	}

	return strings.Join([]string{"DemandProductRatingResult", string(data)}, " ")
}
