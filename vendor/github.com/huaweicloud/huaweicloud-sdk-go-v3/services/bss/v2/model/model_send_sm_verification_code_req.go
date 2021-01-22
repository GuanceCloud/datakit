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

type SendSmVerificationCodeReq struct {
	// |参数名称：手机号| |参数约束及描述：手机号|
	MobilePhone string `json:"mobile_phone"`
	// |参数名称：超时时间，单位是分钟| |参数的约束及描述：超时时间，单位是分钟，短信传递10，邮箱传递60|
	Timeout *int32 `json:"timeout,omitempty"`
	// |参数名称：发送的短信的语言zh-cn: 中文en-us: 英语| |参数约束及描述：发送的短信的语言zh-cn: 中文en-us: 英语|
	Language *string `json:"language,omitempty"`
	// |参数名称：短信模板参数| |参数约束以及描述：短信模板参数|
	SmTemplateArgs *[]TemplateArgs `json:"sm_template_args,omitempty"`
}

func (o SendSmVerificationCodeReq) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "SendSmVerificationCodeReq struct{}"
	}

	return strings.Join([]string{"SendSmVerificationCodeReq", string(data)}, " ")
}
