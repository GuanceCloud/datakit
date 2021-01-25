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

// 绑定弹性公网IP的请求体
type AssociatePublicipsRequestBody struct {
	Publicip *AssociatePublicipsOption `json:"publicip"`
}

func (o AssociatePublicipsRequestBody) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "AssociatePublicipsRequestBody struct{}"
	}

	return strings.Join([]string{"AssociatePublicipsRequestBody", string(data)}, " ")
}
