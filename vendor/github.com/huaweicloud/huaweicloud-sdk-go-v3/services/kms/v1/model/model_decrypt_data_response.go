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
type DecryptDataResponse struct {
	// 密钥ID。
	KeyId *string `json:"key_id,omitempty"`
	// 明文。
	PlainText      *string `json:"plain_text,omitempty"`
	HttpStatusCode int     `json:"-"`
}

func (o DecryptDataResponse) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "DecryptDataResponse struct{}"
	}

	return strings.Join([]string{"DecryptDataResponse", string(data)}, " ")
}
