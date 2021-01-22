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
type DeletePublicipTagRequest struct {
	PublicipId string `json:"publicip_id"`
	Key        string `json:"key"`
}

func (o DeletePublicipTagRequest) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "DeletePublicipTagRequest struct{}"
	}

	return strings.Join([]string{"DeletePublicipTagRequest", string(data)}, " ")
}
