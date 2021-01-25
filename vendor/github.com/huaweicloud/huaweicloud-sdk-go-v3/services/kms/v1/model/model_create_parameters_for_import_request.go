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
type CreateParametersForImportRequest struct {
	VersionId string                             `json:"version_id"`
	Body      *GetParametersForImportRequestBody `json:"body,omitempty"`
}

func (o CreateParametersForImportRequest) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "CreateParametersForImportRequest struct{}"
	}

	return strings.Join([]string{"CreateParametersForImportRequest", string(data)}, " ")
}
