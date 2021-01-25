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

type NewCustomerV2 struct {
	// |参数名称：客户主账号登录名。| |参数约束及描述：客户主账号登录名。|
	CustomerName string `json:"customer_name"`
	// |参数名称：管理员手机号码。如果usePriMobilePhone为Y，则这个参数无效，否则必选。| |参数约束及描述：管理员手机号码。如果usePriMobilePhone为Y，则这个参数无效，否则必选。|
	MobilePhone *string `json:"mobile_phone,omitempty"`
	// |参数名称：是否使用企业主账号手机号码作为子账号手机号码：Y：是；N：否（默认值）。注：当为Y时，mobilePhone输入无效。| |参数约束及描述：是否使用企业主账号手机号码作为子账号手机号码：Y：是；N：否（默认值）。注：当为Y时，mobilePhone输入无效。|
	UsePriMobilePhone *string `json:"use_pri_mobile_phone,omitempty"`
	// |参数名称：客户登录密码。注：usePriMobilePhone为Y时才支持| |参数约束及描述：客户登录密码。注：usePriMobilePhone为Y时才支持|
	Password string `json:"password"`
	// |参数名称：验证码，只有输入企业子客户的手机号邮箱的情况下，才需要填写该字段| |参数约束及描述：验证码，只有输入企业子客户的手机号邮箱的情况下，才需要填写该字段|
	VerificationCode *string `json:"verification_code,omitempty"`
}

func (o NewCustomerV2) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "NewCustomerV2 struct{}"
	}

	return strings.Join([]string{"NewCustomerV2", string(data)}, " ")
}
