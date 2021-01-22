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
type CreateGrantRequest struct {
	VersionId string                  `json:"version_id"`
	Body      *CreateGrantRequestBody `json:"body,omitempty"`
}

func (o CreateGrantRequest) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "CreateGrantRequest struct{}"
	}

	return strings.Join([]string{"CreateGrantRequest", string(data)}, " ")
}
