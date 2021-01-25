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
type AssociatePublicipsResponse struct {
	// 本次请求的编号
	RequestId      *string           `json:"request_id,omitempty"`
	Publicip       *PublicipShowResp `json:"publicip,omitempty"`
	HttpStatusCode int               `json:"-"`
}

func (o AssociatePublicipsResponse) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "AssociatePublicipsResponse struct{}"
	}

	return strings.Join([]string{"AssociatePublicipsResponse", string(data)}, " ")
}
