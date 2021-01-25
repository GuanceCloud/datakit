/*
 * SMN
 *
 * SMN Open API
 *
 */

package model

import (
	"encoding/json"

	"strings"
)

type CreateMessageTemplateRequestBody struct {
	// 创建模板的名称。只能包含大写字母、小写字母、数字、-和_，且必须由大写字母、小写字母或数字开头，长度在1到64个字符之间。
	MessageTemplateName string `json:"message_template_name"`
	// 模板支持的协议类型。  目前支持的协议包括：  “email”：邮件传输协议。  “sms”：短信传输协议。  “functionstage”：FunctionGraph（函数）传输协议。  “functiongraph”：FunctionGraph（工作流）传输协议。  “dms”：DMS传输协议。  “http”、“https”：HTTP/HTTPS传输协议。
	Protocol string `json:"protocol"`
	// 模板内容，模板目前仅支持纯文本模式。模板内容不能空，最大支持256KB。
	Content string `json:"content"`
}

func (o CreateMessageTemplateRequestBody) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "CreateMessageTemplateRequestBody struct{}"
	}

	return strings.Join([]string{"CreateMessageTemplateRequestBody", string(data)}, " ")
}
