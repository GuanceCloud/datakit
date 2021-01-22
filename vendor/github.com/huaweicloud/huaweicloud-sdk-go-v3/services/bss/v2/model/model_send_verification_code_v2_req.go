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

type SendVerificationCodeV2Req struct {
	// |参数名称：发送类型：1：发送短信验证码。2：发送邮件验证码。| |参数的约束及描述：发送类型：1：发送短信验证码。2：发送邮件验证码。|
	ReceiverType int32 `json:"receiver_type"`
	// |参数名称：验证码超时时间。如果不填的话，采用系统默认超时时间5分钟。单位：分钟| |参数的约束及描述：验证码超时时间。如果不填的话，采用系统默认超时时间5分钟。单位：分钟|
	Timeout *int32 `json:"timeout,omitempty"`
	// |参数名称：手机号。目前系统只支持中国手机，必须全部是数字。示例：13XXXXXXXXX| |参数约束及描述：手机号。目前系统只支持中国手机，必须全部是数字。示例：13XXXXXXXXX|
	MobilePhone *string `json:"mobile_phone,omitempty"`
	// |参数名称：根据语言如果查询不到对应模板信息，就取系统默认语言对应的模板信息。zh-cn：中文；en-us：英文。| |参数约束及描述：根据语言如果查询不到对应模板信息，就取系统默认语言对应的模板信息。zh-cn：中文；en-us：英文。|
	Lang *string `json:"lang,omitempty"`
	// |参数名称：场景| |参数的约束及描述：该参数非必填，29：注册；18：实名认证个人银行卡认证；不填写默认为29|
	Scene *int32 `json:"scene,omitempty"`
	// |参数名称：客户ID，如果scene=18的时候必填。| |参数约束及描述：客户ID，如果scene=18的时候必填。|
	CustomerId *string `json:"customer_id,omitempty"`
}

func (o SendVerificationCodeV2Req) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "SendVerificationCodeV2Req struct{}"
	}

	return strings.Join([]string{"SendVerificationCodeV2Req", string(data)}, " ")
}
