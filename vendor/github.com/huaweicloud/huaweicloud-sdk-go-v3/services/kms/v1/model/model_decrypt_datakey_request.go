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
type DecryptDatakeyRequest struct {
	VersionId string                     `json:"version_id"`
	Body      *DecryptDatakeyRequestBody `json:"body,omitempty"`
}

func (o DecryptDatakeyRequest) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "DecryptDatakeyRequest struct{}"
	}

	return strings.Join([]string{"DecryptDatakeyRequest", string(data)}, " ")
}
