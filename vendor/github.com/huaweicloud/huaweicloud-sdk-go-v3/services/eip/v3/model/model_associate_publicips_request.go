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

// Request Object
type AssociatePublicipsRequest struct {
	PublicipId string                         `json:"publicip_id"`
	Body       *AssociatePublicipsRequestBody `json:"body,omitempty"`
}

func (o AssociatePublicipsRequest) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "AssociatePublicipsRequest struct{}"
	}

	return strings.Join([]string{"AssociatePublicipsRequest", string(data)}, " ")
}
