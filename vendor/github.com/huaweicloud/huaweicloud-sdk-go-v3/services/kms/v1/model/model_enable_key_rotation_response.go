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
type EnableKeyRotationResponse struct {
	HttpStatusCode int `json:"-"`
}

func (o EnableKeyRotationResponse) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "EnableKeyRotationResponse struct{}"
	}

	return strings.Join([]string{"EnableKeyRotationResponse", string(data)}, " ")
}
