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
type CreateKmsTagResponse struct {
	HttpStatusCode int `json:"-"`
}

func (o CreateKmsTagResponse) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "CreateKmsTagResponse struct{}"
	}

	return strings.Join([]string{"CreateKmsTagResponse", string(data)}, " ")
}
