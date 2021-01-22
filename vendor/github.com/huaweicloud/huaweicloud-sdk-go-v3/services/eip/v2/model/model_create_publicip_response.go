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
type CreatePublicipResponse struct {
	Publicip       *PublicipCreateResp `json:"publicip,omitempty"`
	HttpStatusCode int                 `json:"-"`
}

func (o CreatePublicipResponse) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "CreatePublicipResponse struct{}"
	}

	return strings.Join([]string{"CreatePublicipResponse", string(data)}, " ")
}
