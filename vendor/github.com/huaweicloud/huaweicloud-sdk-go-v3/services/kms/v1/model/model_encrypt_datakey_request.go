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
type EncryptDatakeyRequest struct {
	VersionId string                     `json:"version_id"`
	Body      *EncryptDatakeyRequestBody `json:"body,omitempty"`
}

func (o EncryptDatakeyRequest) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "EncryptDatakeyRequest struct{}"
	}

	return strings.Join([]string{"EncryptDatakeyRequest", string(data)}, " ")
}
