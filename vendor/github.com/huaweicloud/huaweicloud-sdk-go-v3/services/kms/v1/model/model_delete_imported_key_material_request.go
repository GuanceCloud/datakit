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
type DeleteImportedKeyMaterialRequest struct {
	VersionId string                 `json:"version_id"`
	Body      *OperateKeyRequestBody `json:"body,omitempty"`
}

func (o DeleteImportedKeyMaterialRequest) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "DeleteImportedKeyMaterialRequest struct{}"
	}

	return strings.Join([]string{"DeleteImportedKeyMaterialRequest", string(data)}, " ")
}
