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
type UpdateKeyAliasRequest struct {
	VersionId string                     `json:"version_id"`
	Body      *UpdateKeyAliasRequestBody `json:"body,omitempty"`
}

func (o UpdateKeyAliasRequest) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "UpdateKeyAliasRequest struct{}"
	}

	return strings.Join([]string{"UpdateKeyAliasRequest", string(data)}, " ")
}
