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
type EnableKeyRotationRequest struct {
	VersionId string                 `json:"version_id"`
	Body      *OperateKeyRequestBody `json:"body,omitempty"`
}

func (o EnableKeyRotationRequest) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "EnableKeyRotationRequest struct{}"
	}

	return strings.Join([]string{"EnableKeyRotationRequest", string(data)}, " ")
}
