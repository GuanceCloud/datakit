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
type ListKmsByTagsRequest struct {
	ResourceInstances string                    `json:"resource_instances"`
	VersionId         string                    `json:"version_id"`
	Body              *ListKmsByTagsRequestBody `json:"body,omitempty"`
}

func (o ListKmsByTagsRequest) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "ListKmsByTagsRequest struct{}"
	}

	return strings.Join([]string{"ListKmsByTagsRequest", string(data)}, " ")
}
