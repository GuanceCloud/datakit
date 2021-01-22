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
type CreateDatakeyWithoutPlaintextResponse struct {
	// 密钥ID。
	KeyId *string `json:"key_id,omitempty"`
	// DEK密文16进制，两位表示1byte。
	CipherText     *string `json:"cipher_text,omitempty"`
	HttpStatusCode int     `json:"-"`
}

func (o CreateDatakeyWithoutPlaintextResponse) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "CreateDatakeyWithoutPlaintextResponse struct{}"
	}

	return strings.Join([]string{"CreateDatakeyWithoutPlaintextResponse", string(data)}, " ")
}
