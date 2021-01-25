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

type MessageTemplate struct {
	// 模板ID。
	MessageTemplateId string `json:"message_template_id"`
	// 模板名称。
	MessageTemplateName string `json:"message_template_name"`
	// 模板协议类型。  目前支持的协议包括：  “email”：邮件传输协议。  “sms”：短信传输协议。  “functionstage”：FunctionGraph（函数）传输协议。  “functiongraph”：FunctionGraph（工作流）传输协议。  “dms”：DMS传输协议。  “http”、“https”：HTTP/HTTPS传输协议。
	Protocol string `json:"protocol"`
	// 模板tag列表
	TagNames []string `json:"tag_names"`
	// 模板创建时间 时间格式为UTC时间，YYYY-MM-DDTHH:MM:SSZ。
	CreateTime string `json:"create_time"`
	// 模板最后更新时间 时间格式为UTC时间，YYYY-MM-DDTHH:MM:SSZ。
	UpdateTime string `json:"update_time"`
}

func (o MessageTemplate) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "MessageTemplate struct{}"
	}

	return strings.Join([]string{"MessageTemplate", string(data)}, " ")
}
