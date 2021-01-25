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

type CreateTopicRequestBody struct {
	// 创建topic的名字。Topic名称只能包含大写字母、小写字母、数字、-和_，且必须由大写字母、小写字母或数字开头，长度为1到255个字符。
	Name string `json:"name"`
	// Topic的显示名，推送邮件消息时，作为邮件发件人显示。显示名的长度为192byte或64个中文。默认值为空。
	DisplayName string `json:"display_name"`
	// 企业项目ID。非必选参数，当企业项目开关打开时需要传入该参数。
	EnterpriseProjectId *string `json:"enterprise_project_id,omitempty"`
}

func (o CreateTopicRequestBody) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "CreateTopicRequestBody struct{}"
	}

	return strings.Join([]string{"CreateTopicRequestBody", string(data)}, " ")
}
