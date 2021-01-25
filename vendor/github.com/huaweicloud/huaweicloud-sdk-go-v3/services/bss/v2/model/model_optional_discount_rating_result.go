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

type OptionalDiscountRatingResult struct {
	// |参数名称：折扣优惠Id| |参数约束及描述：折扣优惠Id|
	DiscountId *string `json:"discount_id,omitempty"`
	// |参数名称：总额，即最终优惠后的金额，amount= official_website_amount - discountAmount|
	Amount float32 `json:"amount,omitempty"`
	// |参数名称：官网价| |参数约束及描述：官网价|
	OfficialWebsiteAmount float32 `json:"official_website_amount,omitempty"`
	// |参数名称：优惠额| |参数约束及描述：（官网价和总价的差）|
	DiscountAmount float32 `json:"discount_amount,omitempty"`
	// |参数名称：度量单位标识| |参数约束及描述：1：元|
	MeasureId *int32 `json:"measure_id,omitempty"`
	// |参数名称：折扣优惠类型：合同商务折扣：605(华为云BE场景下的合同商务折扣)、606(分销商BE场景下的合同商务折扣)伙伴授予折扣：607|
	DiscountType *int32 `json:"discount_type,omitempty"`
	// |参数名称：折扣名称| |参数约束及描述：折扣名称|
	DiscountName *string `json:"discount_name,omitempty"`
	// |参数名称：是否为最优折扣| |参数约束及描述：0：不是最优折扣；为缺省值。1：是最优折扣；最优折扣：在商务折扣、伙伴折扣中选择（优惠金额最大的折扣为最优，优惠金额相等则按此顺序排优先级），促销折扣，折扣券不参与最优折扣的计算|
	BestOffer *int32 `json:"best_offer,omitempty"`
	// |参数名称：产品询价结果| |参数的约束及描述：产品询价结果|
	ProductRatingResults *[]PeriodProductRatingResult `json:"product_rating_results,omitempty"`
}

func (o OptionalDiscountRatingResult) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "OptionalDiscountRatingResult struct{}"
	}

	return strings.Join([]string{"OptionalDiscountRatingResult", string(data)}, " ")
}
