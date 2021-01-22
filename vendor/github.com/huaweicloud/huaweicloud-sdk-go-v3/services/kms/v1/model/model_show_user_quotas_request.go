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
type ShowUserQuotasRequest struct {
	VersionId string `json:"version_id"`
}

func (o ShowUserQuotasRequest) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "ShowUserQuotasRequest struct{}"
	}

	return strings.Join([]string{"ShowUserQuotasRequest", string(data)}, " ")
}
