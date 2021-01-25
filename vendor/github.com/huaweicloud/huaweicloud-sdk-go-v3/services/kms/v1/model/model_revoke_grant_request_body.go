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

type RevokeGrantRequestBody struct {
	// 密钥ID，36字节，满足正则匹配“^[0-9a-z]{8}-[0-9a-z]{4}-[0-9a-z]{4}-[0-9a-z]{4}-[0-9a-z]{12}$”。 例如：0d0466b0-e727-4d9c-b35d-f84bb474a37f。
	KeyId *string `json:"key_id,omitempty"`
	// 授权ID，64字节，满足正则匹配“^[A-Fa-f0-9]{64}$”。 例如：7c9a3286af4fcca5f0a385ad13e1d21a50e27b6dbcab50f37f30f93b8939827d
	GrantId *string `json:"grant_id,omitempty"`
	// 请求消息序列号，36字节序列号。例如：919c82d4-8046-4722-9094-35c3c6524cff
	Sequence *string `json:"sequence,omitempty"`
}

func (o RevokeGrantRequestBody) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "RevokeGrantRequestBody struct{}"
	}

	return strings.Join([]string{"RevokeGrantRequestBody", string(data)}, " ")
}
