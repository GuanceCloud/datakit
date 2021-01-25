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

// Response Object
type ListOnDemandResourceRatingsResponse struct {
	// |参数名称：总额| |参数约束及描述：即最终优惠后的金额|
	Amount float32 `json:"amount,omitempty"`
	// |参数名称：优惠额| |参数约束及描述：（官网价和总价的差）|
	DiscountAmount float32 `json:"discount_amount,omitempty"`
	// |参数名称：官网价| |参数约束及描述：官网价|
	OfficialWebsiteAmount float32 `json:"official_website_amount,omitempty"`
	// |参数名称：度量单位标识| |参数约束及描述：1：元|
	MeasureId *int32 `json:"measure_id,omitempty"`
	// |参数名称：币种| |参数约束及描述：比如CNY|
	Currency *string `json:"currency,omitempty"`
	// |参数名称：产品询价结果| |参数的约束及描述：产品询价结果|
	ProductRatingResults *[]DemandProductRatingResult `json:"product_rating_results,omitempty"`
	HttpStatusCode       int                          `json:"-"`
}

func (o ListOnDemandResourceRatingsResponse) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "ListOnDemandResourceRatingsResponse struct{}"
	}

	return strings.Join([]string{"ListOnDemandResourceRatingsResponse", string(data)}, " ")
}
