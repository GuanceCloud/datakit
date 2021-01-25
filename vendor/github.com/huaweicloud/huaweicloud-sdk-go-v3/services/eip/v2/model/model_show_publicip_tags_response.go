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
type ShowPublicipTagsResponse struct {
	// 标签列表
	Tags           *[]ResourceTagResp `json:"tags,omitempty"`
	HttpStatusCode int                `json:"-"`
}

func (o ShowPublicipTagsResponse) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "ShowPublicipTagsResponse struct{}"
	}

	return strings.Join([]string{"ShowPublicipTagsResponse", string(data)}, " ")
}
