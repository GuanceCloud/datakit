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
	// 弹性公网IP对象
	Publicips      *[]PublicipShowResp `json:"publicips,omitempty"`
	HttpStatusCode int                 `json:"-"`
}

func (o ListPublicipsResponse) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "ListPublicipsResponse struct{}"
	}

	return strings.Join([]string{"ListPublicipsResponse", string(data)}, " ")
}
