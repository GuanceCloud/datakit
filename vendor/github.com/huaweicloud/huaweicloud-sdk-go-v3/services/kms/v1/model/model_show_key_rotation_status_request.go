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
type ShowKeyRotationStatusRequest struct {
	VersionId string                 `json:"version_id"`
	Body      *OperateKeyRequestBody `json:"body,omitempty"`
}

func (o ShowKeyRotationStatusRequest) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "ShowKeyRotationStatusRequest struct{}"
	}

	return strings.Join([]string{"ShowKeyRotationStatusRequest", string(data)}, " ")
}
