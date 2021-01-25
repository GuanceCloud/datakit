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
type ListRetirableGrantsRequest struct {
	VersionId string                          `json:"version_id"`
	Body      *ListRetirableGrantsRequestBody `json:"body,omitempty"`
}

func (o ListRetirableGrantsRequest) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "ListRetirableGrantsRequest struct{}"
	}

	return strings.Join([]string{"ListRetirableGrantsRequest", string(data)}, " ")
}
