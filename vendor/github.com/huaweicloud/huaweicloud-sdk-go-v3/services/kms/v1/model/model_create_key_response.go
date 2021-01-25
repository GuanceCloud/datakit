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
type CreateKeyResponse struct {
	KeyInfo        *KeKInfo `json:"key_info,omitempty"`
	HttpStatusCode int      `json:"-"`
}

func (o CreateKeyResponse) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "CreateKeyResponse struct{}"
	}

	return strings.Join([]string{"CreateKeyResponse", string(data)}, " ")
}
