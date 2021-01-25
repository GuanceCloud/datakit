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

type QueryIndirectPartnersReq struct {
	// |参数名称：登录名称| |参数约束及描述：登录名称|
	AccountName *string `json:"account_name,omitempty"`
	// |参数名称：关联开始时间，UTC时间（包括时区），比如2016-03-28T00:00:00Z| |参数约束及描述：关联开始时间，UTC时间（包括时区），比如2016-03-28T00:00:00Z|
	AssociatedOnBegin *string `json:"associated_on_begin,omitempty"`
	// |参数名称：关联结束时间，UTC时间（包括时区），比如2016-03-28T00:00:00Z| |参数约束及描述：关联结束时间，UTC时间（包括时区），比如2016-03-28T00:00:00Z|
	AssociatedOnEnd *string `json:"associated_on_end,omitempty"`
	// |参数名称：偏移量，从0开始，默认是0| |参数的约束及描述：偏移量，从0开始，默认是0|
	Offset *int32 `json:"offset,omitempty"`
	// |参数名称：最大100，默认为10| |参数的约束及描述：最大100，默认为10|
	Limit *int32 `json:"limit,omitempty"`
	// |参数名称：二级经销商ID| |参数约束及描述：二级经销商ID|
	IndirectPartnerId *string `json:"indirect_partner_id,omitempty"`
}

func (o QueryIndirectPartnersReq) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "QueryIndirectPartnersReq struct{}"
	}

	return strings.Join([]string{"QueryIndirectPartnersReq", string(data)}, " ")
}
