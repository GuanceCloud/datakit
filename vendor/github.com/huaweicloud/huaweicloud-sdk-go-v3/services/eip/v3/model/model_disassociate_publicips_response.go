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
type DisassociatePublicipsResponse struct {
	// 本次请求的编号
	RequestId      *string           `json:"request_id,omitempty"`
	Publicip       *PublicipShowResp `json:"publicip,omitempty"`
	HttpStatusCode int               `json:"-"`
}

func (o DisassociatePublicipsResponse) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "DisassociatePublicipsResponse struct{}"
	}

	return strings.Join([]string{"DisassociatePublicipsResponse", string(data)}, " ")
}
