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
type ShowPublicipTagsRequest struct {
	PublicipId string `json:"publicip_id"`
}

func (o ShowPublicipTagsRequest) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "ShowPublicipTagsRequest struct{}"
	}

	return strings.Join([]string{"ShowPublicipTagsRequest", string(data)}, " ")
}
