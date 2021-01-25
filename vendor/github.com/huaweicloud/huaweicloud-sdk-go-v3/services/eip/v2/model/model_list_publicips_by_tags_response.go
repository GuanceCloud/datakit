/*
 * EIP
 *
 * 云服务接口
 *
 */

package model

import (
	"encoding/json"

	"strings"
)

// Response Object
type ListPublicipsByTagsResponse struct {
	// resource对象列表
	Resources *[]ListResourceResp `json:"resources,omitempty"`
	// 总记录数
	TotalCount     *int32 `json:"total_count,omitempty"`
	HttpStatusCode int    `json:"-"`
}

func (o ListPublicipsByTagsResponse) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "ListPublicipsByTagsResponse struct{}"
	}

	return strings.Join([]string{"ListPublicipsByTagsResponse", string(data)}, " ")
}
