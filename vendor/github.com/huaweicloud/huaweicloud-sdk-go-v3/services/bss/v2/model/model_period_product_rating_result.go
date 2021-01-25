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

type PeriodProductRatingResult struct {
	// |参数名称：ID标识| |参数约束及描述：ID标识，来源于请求中的ID|
	Id *string `json:"id,omitempty"`
	// |参数名称：产品ID| |参数约束及描述：产品ID|
	ProductId *string `json:"product_id,omitempty"`
	// |参数名称：总额| |参数约束及描述：即最终优惠的金额|
	Amount float32 `json:"amount,omitempty"`
	// |参数名称：官网价| |参数约束及描述：官网价|
	OfficialWebsiteAmount float32 `json:"official_website_amount,omitempty"`
	// |参数名称：优惠额（官网价和总价的差）| |参数约束及描述：优惠额（官网价和总价的差）|
	DiscountAmount float32 `json:"discount_amount,omitempty"`
	// |参数名称：度量单位标识| |参数约束及描述：1：元|
	MeasureId *int32 `json:"measure_id,omitempty"`
}

func (o PeriodProductRatingResult) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "PeriodProductRatingResult struct{}"
	}

	return strings.Join([]string{"PeriodProductRatingResult", string(data)}, " ")
}
