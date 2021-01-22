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
type DecryptDatakeyResponse struct {
	// DEK明文的16进制字符串。
	DataKey *string `json:"data_key,omitempty"`
	// DEK明文字节长度。
	DatakeyLength *string `json:"datakey_length,omitempty"`
	// DEK明文的SHA256值对应的16进制字符串。
	DatakeyDgst    *string `json:"datakey_dgst,omitempty"`
	HttpStatusCode int     `json:"-"`
}

func (o DecryptDatakeyResponse) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "DecryptDatakeyResponse struct{}"
	}

	return strings.Join([]string{"DecryptDatakeyResponse", string(data)}, " ")
}
