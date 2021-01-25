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
type EnableKeyResponse struct {
	KeyInfo        *KeyStatusInfo `json:"key_info,omitempty"`
	HttpStatusCode int            `json:"-"`
}

func (o EnableKeyResponse) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "EnableKeyResponse struct{}"
	}

	return strings.Join([]string{"EnableKeyResponse", string(data)}, " ")
}
