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

// Response Object
type ListKmsByTagsResponse struct {
	// 资源实例列表，详情请参见resource字段数据结构说明。
	Resources *[]ActionResources `json:"resources,omitempty"`
	// 总记录数。
	TotalCount     *int32 `json:"total_count,omitempty"`
	HttpStatusCode int    `json:"-"`
}

func (o ListKmsByTagsResponse) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "ListKmsByTagsResponse struct{}"
	}

	return strings.Join([]string{"ListKmsByTagsResponse", string(data)}, " ")
}
