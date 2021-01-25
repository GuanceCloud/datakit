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

type ListTopicsItem struct {
	// Topic的唯一的资源标识。
	TopicUrn string `json:"topic_urn"`
	// 创建topic的名字。
	Name string `json:"name"`
	// Topic的显示名，推送邮件消息时，作为邮件发件人显示。
	DisplayName string `json:"display_name"`
	// 消息推送的策略，该属性目前不支持修改，后续将支持修改。0表示发送失败，保留到失败队列，1表示直接丢弃发送失败的消息。
	PushPolicy int32 `json:"push_policy"`
	// 企业项目ID。
	EnterpriseProjectId string `json:"enterprise_project_id"`
}

func (o ListTopicsItem) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "ListTopicsItem struct{}"
	}

	return strings.Join([]string{"ListTopicsItem", string(data)}, " ")
}
