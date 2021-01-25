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
type ShowVersionRequest struct {
	VersionId string `json:"version_id"`
}

func (o ShowVersionRequest) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "ShowVersionRequest struct{}"
	}

	return strings.Join([]string{"ShowVersionRequest", string(data)}, " ")
}
