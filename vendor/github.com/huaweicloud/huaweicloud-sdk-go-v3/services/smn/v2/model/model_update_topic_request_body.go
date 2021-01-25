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

type UpdateTopicRequestBody struct {
	// Topic的显示名，推送邮件消息时，作为邮件发件人显示。显示名的长度为192byte或64个中文。
	DisplayName string `json:"display_name"`
}

func (o UpdateTopicRequestBody) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "UpdateTopicRequestBody struct{}"
	}

	return strings.Join([]string{"UpdateTopicRequestBody", string(data)}, " ")
}
