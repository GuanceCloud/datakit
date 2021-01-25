/*
 * DMS
 *
 * DMS Document API
 *
 */

package model

import (
	"encoding/json"

	"strings"
)

type SendMessagesRespMessages struct {
	// 错误描述信息。
	Error *string `json:"error,omitempty"`
	// 错误码。
	ErrorCode *int32 `json:"error_code,omitempty"`
	// 发送消息的状态。 0：表示发送成功。 1：表示发送失败，失败原因参考对应的error和error_code。
	State *int32 `json:"state,omitempty"`
	// 消息ID。
	Id *string `json:"id,omitempty"`
}

func (o SendMessagesRespMessages) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "SendMessagesRespMessages struct{}"
	}

	return strings.Join([]string{"SendMessagesRespMessages", string(data)}, " ")
}
