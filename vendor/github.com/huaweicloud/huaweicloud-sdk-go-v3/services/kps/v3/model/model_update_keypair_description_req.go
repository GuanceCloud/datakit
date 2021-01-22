/*
 * kps
 *
 * kps v3 版本API
 *
 */

package model

import (
	"encoding/json"

	"strings"
)

// 更新SSH密钥对描述消息体
type UpdateKeypairDescriptionReq struct {
	// 描述信息
	Description string `json:"description"`
}

func (o UpdateKeypairDescriptionReq) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "UpdateKeypairDescriptionReq struct{}"
	}

	return strings.Join([]string{"UpdateKeypairDescriptionReq", string(data)}, " ")
}
