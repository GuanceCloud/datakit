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

type DiscountSimpleInfo struct {
	// |参数名称：折扣ID| |参数约束及描述：折扣ID|
	Id string `json:"id"`
	// |参数名称：折扣类型：取值为1：合同折扣（可以有多组）2：商务优惠（仅有一组）3：合作伙伴授予折扣（仅有一组）609：订单调价折扣| |参数的约束及描述：折扣类型：取值为1：合同折扣（可以有多组）2：商务优惠（仅有一组）3：合作伙伴授予折扣（仅有一组）609：订单调价折扣|
	Type int32 `json:"type"`
}

func (o DiscountSimpleInfo) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "DiscountSimpleInfo struct{}"
	}

	return strings.Join([]string{"DiscountSimpleInfo", string(data)}, " ")
}
