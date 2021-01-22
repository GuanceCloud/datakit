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
type ListKeyDetailRequest struct {
	VersionId string                 `json:"version_id"`
	Body      *OperateKeyRequestBody `json:"body,omitempty"`
}

func (o ListKeyDetailRequest) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "ListKeyDetailRequest struct{}"
	}

	return strings.Join([]string{"ListKeyDetailRequest", string(data)}, " ")
}
