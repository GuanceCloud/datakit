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
type CreateDatakeyResponse struct {
	// 密钥ID。
	KeyId *string `json:"key_id,omitempty"`
	// DEK明文16进制，两位表示1byte。
	PlainText *string `json:"plain_text,omitempty"`
	// DEK密文16进制，两位表示1byte。
	CipherText     *string `json:"cipher_text,omitempty"`
	HttpStatusCode int     `json:"-"`
}

func (o CreateDatakeyResponse) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "CreateDatakeyResponse struct{}"
	}

	return strings.Join([]string{"CreateDatakeyResponse", string(data)}, " ")
}
