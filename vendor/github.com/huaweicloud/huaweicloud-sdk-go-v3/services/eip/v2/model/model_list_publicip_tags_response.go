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
type ListPublicipTagsResponse struct {
	// 标签列表
	Tags           *[]TagResp `json:"tags,omitempty"`
	HttpStatusCode int        `json:"-"`
}

func (o ListPublicipTagsResponse) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "ListPublicipTagsResponse struct{}"
	}

	return strings.Join([]string{"ListPublicipTagsResponse", string(data)}, " ")
}
