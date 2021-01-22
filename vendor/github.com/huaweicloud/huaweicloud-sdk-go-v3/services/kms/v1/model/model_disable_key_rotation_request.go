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
type DisableKeyRotationRequest struct {
	VersionId string                 `json:"version_id"`
	Body      *OperateKeyRequestBody `json:"body,omitempty"`
}

func (o DisableKeyRotationRequest) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "DisableKeyRotationRequest struct{}"
	}

	return strings.Join([]string{"DisableKeyRotationRequest", string(data)}, " ")
}
