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

// Response Object
type ListTopicDetailsResponse struct {
	// 更新时间。时间格式为UTC时间，YYYY-MM-DDTHH:MM:SSZ。
	UpdateTime *string `json:"update_time,omitempty"`
	// 消息推送的策略。0表示发送失败，保留到失败队列，1表示直接丢弃发送失败的消息。
	PushPolicy *int32 `json:"push_policy,omitempty"`
	// 创建时间。时间格式为UTC时间，YYYY-MM-DDTHH:MM:SSZ。
	CreateTime *string `json:"create_time,omitempty"`
	// 创建Topic的名字。
	Name *string `json:"name,omitempty"`
	// Topic的唯一的资源标识。可以通过[查看主题列表获](https://support.huaweicloud.com/api-smn/smn_api_51004.html)取该标识。
	TopicUrn *string `json:"topic_urn,omitempty"`
	// Topic的显示名，推送邮件消息时，作为邮件发件人显示。
	DisplayName *string `json:"display_name,omitempty"`
	// 请求的唯一标识ID。
	RequestId *string `json:"request_id,omitempty"`
	// 企业项目ID。
	EnterpriseProjectId *string `json:"enterprise_project_id,omitempty"`
	HttpStatusCode      int     `json:"-"`
}

func (o ListTopicDetailsResponse) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "ListTopicDetailsResponse struct{}"
	}

	return strings.Join([]string{"ListTopicDetailsResponse", string(data)}, " ")
}
