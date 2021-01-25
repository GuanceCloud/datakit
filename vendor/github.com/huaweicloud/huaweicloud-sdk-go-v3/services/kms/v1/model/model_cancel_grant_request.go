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
type CancelGrantRequest struct {
	VersionId string                  `json:"version_id"`
	Body      *RevokeGrantRequestBody `json:"body,omitempty"`
}

func (o CancelGrantRequest) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "CancelGrantRequest struct{}"
	}

	return strings.Join([]string{"CancelGrantRequest", string(data)}, " ")
}
