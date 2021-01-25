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

type IndirectPartnerInfo struct {
	// |参数名称：二级经销商ID| |参数约束及描述：二级经销商ID|
	IndirectPartnerId *string `json:"indirect_partner_id,omitempty"`
	// |参数名称：手机号码| |参数约束及描述：手机号码|
	MobilePhone *string `json:"mobile_phone,omitempty"`
	// |参数名称：邮箱| |参数约束及描述：邮箱|
	Email *string `json:"email,omitempty"`
	// |参数名称：二级经销商的账户名| |参数约束及描述：二级经销商的账户名|
	AccountName *string `json:"account_name,omitempty"`
	// |参数名称：二级经销商名称| |参数约束及描述：二级经销商名称|
	Name *string `json:"name,omitempty"`
	// |参数名称：关联时间，UTC时间（包括时区），比如2016-03-28T00:00:00Z| |参数约束及描述：关联时间，UTC时间（包括时区），比如2016-03-28T00:00:00Z|
	AssociatedOn *string `json:"associated_on,omitempty"`
}

func (o IndirectPartnerInfo) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "IndirectPartnerInfo struct{}"
	}

	return strings.Join([]string{"IndirectPartnerInfo", string(data)}, " ")
}
