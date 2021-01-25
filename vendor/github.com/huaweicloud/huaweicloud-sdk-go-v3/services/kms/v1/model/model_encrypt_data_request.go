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
type EncryptDataRequest struct {
	VersionId string                  `json:"version_id"`
	Body      *EncryptDataRequestBody `json:"body,omitempty"`
}

func (o EncryptDataRequest) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "EncryptDataRequest struct{}"
	}

	return strings.Join([]string{"EncryptDataRequest", string(data)}, " ")
}
