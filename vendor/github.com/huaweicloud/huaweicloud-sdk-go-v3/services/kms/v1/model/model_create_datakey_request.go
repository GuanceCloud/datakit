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
type CreateDatakeyRequest struct {
	VersionId string                    `json:"version_id"`
	Body      *CreateDatakeyRequestBody `json:"body,omitempty"`
}

func (o CreateDatakeyRequest) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "CreateDatakeyRequest struct{}"
	}

	return strings.Join([]string{"CreateDatakeyRequest", string(data)}, " ")
}
