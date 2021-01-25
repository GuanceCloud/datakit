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

// 绑定密钥对描述消息体
type AssociateKeypairRequestBody struct {
	// SSH密钥对的名称
	KeypairName string         `json:"keypair_name"`
	Server      *EcsServerInfo `json:"server"`
}

func (o AssociateKeypairRequestBody) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "AssociateKeypairRequestBody struct{}"
	}

	return strings.Join([]string{"AssociateKeypairRequestBody", string(data)}, " ")
}
