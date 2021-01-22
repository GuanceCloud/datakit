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
type UpdatePublicipResponse struct {
	Publicip       *PublicipShowResp `json:"publicip,omitempty"`
	HttpStatusCode int               `json:"-"`
}

func (o UpdatePublicipResponse) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "UpdatePublicipResponse struct{}"
	}

	return strings.Join([]string{"UpdatePublicipResponse", string(data)}, " ")
}
