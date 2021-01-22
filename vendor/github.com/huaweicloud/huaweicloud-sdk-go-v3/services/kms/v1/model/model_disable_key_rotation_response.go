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
type DisableKeyRotationResponse struct {
	HttpStatusCode int `json:"-"`
}

func (o DisableKeyRotationResponse) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "DisableKeyRotationResponse struct{}"
	}

	return strings.Join([]string{"DisableKeyRotationResponse", string(data)}, " ")
}
