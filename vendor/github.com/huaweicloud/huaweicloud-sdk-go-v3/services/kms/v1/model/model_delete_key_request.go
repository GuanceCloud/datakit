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
type DeleteKeyRequest struct {
	VersionId string                          `json:"version_id"`
	Body      *ScheduleKeyDeletionRequestBody `json:"body,omitempty"`
}

func (o DeleteKeyRequest) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "DeleteKeyRequest struct{}"
	}

	return strings.Join([]string{"DeleteKeyRequest", string(data)}, " ")
}
