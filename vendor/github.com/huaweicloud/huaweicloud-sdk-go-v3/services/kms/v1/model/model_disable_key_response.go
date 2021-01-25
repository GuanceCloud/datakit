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
type DisableKeyResponse struct {
	KeyInfo        *KeyStatusInfo `json:"key_info,omitempty"`
	HttpStatusCode int            `json:"-"`
}

func (o DisableKeyResponse) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "DisableKeyResponse struct{}"
	}

	return strings.Join([]string{"DisableKeyResponse", string(data)}, " ")
}
