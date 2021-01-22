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
type UpdateKeyAliasResponse struct {
	KeyInfo        *KeyAliasInfo `json:"key_info,omitempty"`
	HttpStatusCode int           `json:"-"`
}

func (o UpdateKeyAliasResponse) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "UpdateKeyAliasResponse struct{}"
	}

	return strings.Join([]string{"UpdateKeyAliasResponse", string(data)}, " ")
}
