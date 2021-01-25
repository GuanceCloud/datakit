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
type CancelKeyDeletionRequest struct {
	VersionId string                 `json:"version_id"`
	Body      *OperateKeyRequestBody `json:"body,omitempty"`
}

func (o CancelKeyDeletionRequest) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "CancelKeyDeletionRequest struct{}"
	}

	return strings.Join([]string{"CancelKeyDeletionRequest", string(data)}, " ")
}
