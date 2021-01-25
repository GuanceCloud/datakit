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

type CreateKmsTagRequestBody struct {
	Tag *TagItem `json:"tag,omitempty"`
}

func (o CreateKmsTagRequestBody) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "CreateKmsTagRequestBody struct{}"
	}

	return strings.Join([]string{"CreateKmsTagRequestBody", string(data)}, " ")
}
