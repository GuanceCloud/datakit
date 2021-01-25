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

type QuerySubCustomerListReq struct {
	// |参数名称：客户登录名称（如果客户创建了子用户，此处需要填写主账号登录名称。关于主账号和子用户的具体介绍请参见身份管理身份管理中“账号”和“IAM用户”的描述）。支持模糊查询。| |参数约束及描述：客户登录名称（如果客户创建了子用户，此处需要填写主账号登录名称。关于主账号和子用户的具体介绍请参见身份管理身份管理中“账号”和“IAM用户”的描述）。支持模糊查询。|
	AccountName *string `json:"account_name,omitempty"`
	// |参数名称：实名认证名称。支持模糊查询。| |参数约束及描述：实名认证名称。支持模糊查询。|
	Customer *string `json:"customer,omitempty"`
	// |参数名称：偏移量，从0开始| |参数约束及描述： 偏移量，从0开始|
	Offset *int32 `json:"offset,omitempty"`
	// |参数名称：每次查询的数量。默认10，最多100。| |参数约束及描述： 每次查询的数量。默认10，最多100。|
	Limit *int32 `json:"limit,omitempty"`
	// |参数名称：标签，支持模糊查找。| |参数约束及描述：非必填，最大长度64|
	Label *string `json:"label,omitempty"`
	// |参数名称：关联类型1.推荐，2.垫付，3.转售| |参数约束及描述：非必填，最大长度2|
	AssociationType *string `json:"association_type,omitempty"`
	// |参数名称：关联时间区间段开始，UTC时间。| |参数约束及描述：格式为：yyyy-MM-ddTHH:mm:ssZ，如“2019-05-06T08:05:01Z”。|
	AssociatedOnBegin *string `json:"associated_on_begin,omitempty"`
	// |参数名称：关联时间区间段结束，UTC时间| |参数约束及描述：格式为：yyyy-MM-ddTHH:mm:ssZ，如“2019-05-06T08:05:01Z”。|
	AssociatedOnEnd *string `json:"associated_on_end,omitempty"`
	// |参数名称：子客户ID| |参数约束及描述：非必填，最大长度64|
	CustomerId *string `json:"customer_id,omitempty"`
	// |参数名称：二级渠道商ID| |参数约束及描述：如果想查询二级渠道子客户的列表，该字段必须携带，最大长度64|
	IndirectPartnerId *string `json:"indirect_partner_id,omitempty"`
}

func (o QuerySubCustomerListReq) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "QuerySubCustomerListReq struct{}"
	}

	return strings.Join([]string{"QuerySubCustomerListReq", string(data)}, " ")
}
