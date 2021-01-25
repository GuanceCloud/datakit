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

type CreateCustomerV2Req struct {
	// |参数名称：客户的华为云账号名| |参数的约束及描述：该参数非必填，不能以“op_”或“shadow_”开头且不能全为数字。且只允许最大长度64的字符串,如果为空，随机生成。校验规则^[a-zA-Z0-9\\u00c0-\\u00ff-._ ]{0,64}$|
	DomainName *string `json:"domain_name,omitempty"`
	// |参数名称：手机号| |参数的约束及描述：如果接入的是华北站点，该字段必填，否则该字段忽略目前系统只支持中国手机，必须全部是数字。示例：13XXXXXXXXX|
	MobilePhone *string `json:"mobile_phone,omitempty"`
	// |参数名称：验证码| |参数的约束及描述：该参数必填，如果输入的是手机，就是手机验证码，如果输入的是邮箱，就是邮箱验证码|
	VerificationCode *string `json:"verification_code,omitempty"`
	// |参数名称：第3方系统的用户唯一标识| |参数的约束及描述：该参数必填，且只允许最大长度128的字符串|
	XaccountId string `json:"xaccount_id"`
	// |参数名称：华为分给合作伙伴的平台标识| |参数的约束及描述：该参数必填，且只允许最大长度30的字符串,该标识的具体值由华为分配|
	XaccountType string `json:"xaccount_type"`
	// |参数名称：密码| |参数的约束及描述：该参数选填，长度6~32位字符，至少包含以下四种字符中的两种： 大写字母、小写字母、数字、特殊字符，不能和账号名或倒序的账号名相同，不能包含手机号，不能包含邮箱|
	Password *string `json:"password,omitempty"`
	// |是否关闭营销消息| |参数的约束及描述：该参数选填。false：不关闭，True：关闭，默认不关闭|
	IsCloseMarketMs *string `json:"is_close_market_ms,omitempty"`
	// |合作类型| |参数的约束及描述：该参数选填。1：推荐。仅仅支持1|
	CooperationType *string `json:"cooperation_type,omitempty"`
	// |参数名称：二级渠道ID| |参数的约束及描述：该参数非必填，二级渠道ID，最大长度64|
	IndirectPartnerId *string `json:"indirect_partner_id,omitempty"`
	// |参数名称：是否返回关联结果| |参数的约束及描述：该参数非必填|
	IncludeAssociationResult *bool `json:"include_association_result,omitempty"`
}

func (o CreateCustomerV2Req) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "CreateCustomerV2Req struct{}"
	}

	return strings.Join([]string{"CreateCustomerV2Req", string(data)}, " ")
}
