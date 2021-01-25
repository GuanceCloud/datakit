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
type CreateParametersForImportResponse struct {
	// 密钥ID。
	KeyId *string `json:"key_id,omitempty"`
	// 密钥导入令牌。
	ImportToken *string `json:"import_token,omitempty"`
	// 导入参数到期时间，时间戳，即从1970年1月1日至该时间的总秒数。
	ExpirationTime *string `json:"expiration_time,omitempty"`
	// 加密密钥材料的公钥，base64格式。
	PublicKey      *string `json:"public_key,omitempty"`
	HttpStatusCode int     `json:"-"`
}

func (o CreateParametersForImportResponse) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "CreateParametersForImportResponse struct{}"
	}

	return strings.Join([]string{"CreateParametersForImportResponse", string(data)}, " ")
}
