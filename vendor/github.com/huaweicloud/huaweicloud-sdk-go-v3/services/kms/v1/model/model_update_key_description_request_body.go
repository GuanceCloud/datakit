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

type UpdateKeyDescriptionRequestBody struct {
	// 密钥ID，36字节，满足正则匹配“^[0-9a-z]{8}-[0-9a-z]{4}-[0-9a-z]{4}-[0-9a-z]{4}-[0-9a-z]{12}$”。 例如：0d0466b0-e727-4d9c-b35d-f84bb474a37f。
	KeyId *string `json:"key_id,omitempty"`
	// 密钥描述，取值0到255字符。
	KeyDescription *string `json:"key_description,omitempty"`
	// 请求消息序列号，36字节序列号。 例如：919c82d4-8046-4722-9094-35c3c6524cff
	Sequence *string `json:"sequence,omitempty"`
}

func (o UpdateKeyDescriptionRequestBody) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "UpdateKeyDescriptionRequestBody struct{}"
	}

	return strings.Join([]string{"UpdateKeyDescriptionRequestBody", string(data)}, " ")
}
