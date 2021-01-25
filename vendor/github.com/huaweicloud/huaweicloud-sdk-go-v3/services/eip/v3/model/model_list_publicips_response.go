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
type ListPublicipsResponse struct {
	// 本次请求的编号
	RequestId *string `json:"request_id,omitempty"`
	// 功能说明：弹性公网IP对象
	Publicips *[]PublicipShowResp `json:"publicips,omitempty"`
	PageInfo  *PageInfoOption     `json:"page_info,omitempty"`
	// 公网IP总条目数
	Total          *int32 `json:"total,omitempty"`
	HttpStatusCode int    `json:"-"`
}

func (o ListPublicipsResponse) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "ListPublicipsResponse struct{}"
	}

	return strings.Join([]string{"ListPublicipsResponse", string(data)}, " ")
}
