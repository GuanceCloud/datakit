/*
 * kms
 *
 * KMS v1.0 API, open API
 *
 */

package model

import (
	"encoding/json"

	"strings"
)

type ScheduleKeyDeletionRequestBody struct {
	// 密钥ID，36字节，满足正则匹配“^[0-9a-z]{8}-[0-9a-z]{4}-[0-9a-z]{4}-[0-9a-z]{4}-[0-9a-z]{12}$”。 例如：0d0466b0-e727-4d9c-b35d-f84bb474a37f。
	KeyId *string `json:"key_id,omitempty"`
	// 计划多少天后删除密钥，取值为7到1096。
	PendingDays *string `json:"pending_days,omitempty"`
	// 请求消息序列号，36字节序列号。 例如：919c82d4-8046-4722-9094-35c3c6524cff
	Sequence *string `json:"sequence,omitempty"`
}

func (o ScheduleKeyDeletionRequestBody) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "ScheduleKeyDeletionRequestBody struct{}"
	}

	return strings.Join([]string{"ScheduleKeyDeletionRequestBody", string(data)}, " ")
}
