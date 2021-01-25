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

// Response Object
type DeleteKeyResponse struct {
	// 密钥ID
	KeyId *string `json:"key_id,omitempty"`
	// 密钥状态： - 2为启用状态 - 3为禁用状态 - 4为计划删除状态 - 5为等待导入状态 - 7为冻结状态
	KeyState       *string `json:"key_state,omitempty"`
	HttpStatusCode int     `json:"-"`
}

func (o DeleteKeyResponse) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "DeleteKeyResponse struct{}"
	}

	return strings.Join([]string{"DeleteKeyResponse", string(data)}, " ")
}
