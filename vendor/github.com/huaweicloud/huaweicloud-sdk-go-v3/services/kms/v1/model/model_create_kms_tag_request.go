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

// Request Object
type CreateKmsTagRequest struct {
	VersionId string                   `json:"version_id"`
	KeyId     string                   `json:"key_id"`
	Body      *CreateKmsTagRequestBody `json:"body,omitempty"`
}

func (o CreateKmsTagRequest) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "CreateKmsTagRequest struct{}"
	}

	return strings.Join([]string{"CreateKmsTagRequest", string(data)}, " ")
}
