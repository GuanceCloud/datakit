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
type BatchCreateKmsTagsResponse struct {
	HttpStatusCode int `json:"-"`
}

func (o BatchCreateKmsTagsResponse) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "BatchCreateKmsTagsResponse struct{}"
	}

	return strings.Join([]string{"BatchCreateKmsTagsResponse", string(data)}, " ")
}
