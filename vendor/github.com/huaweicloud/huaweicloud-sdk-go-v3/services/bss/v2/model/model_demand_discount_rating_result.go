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

type DemandDiscountRatingResult struct {
	// |参数名称：优惠标识ID| |参数约束及描述：优惠标识ID|
	DiscountId *string `json:"discount_id,omitempty"`
	// |参数名称：合同商务优惠类型：基于官网价计算优惠605 华为云商务-折扣率，一口价，华为云用户606 渠道商务-折扣率，一口价，BP用户伙伴折扣优惠类型：基于官网价计算优惠607 合作伙伴授予折扣-折扣率|
	DiscountType *int32 `json:"discount_type,omitempty"`
	// 优惠金额
	Amount float32 `json:"amount,omitempty"`
	// |参数名称：度量单位标识| |参数约束及描述：1：元|
	MeasureId *int32 `json:"measure_id,omitempty"`
	// |参数名称：折扣名称| |参数约束及描述：折扣名称|
	DiscountName *string `json:"discount_name,omitempty"`
}

func (o DemandDiscountRatingResult) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "DemandDiscountRatingResult struct{}"
	}

	return strings.Join([]string{"DemandDiscountRatingResult", string(data)}, " ")
}
