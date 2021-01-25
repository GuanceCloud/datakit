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

// 解绑弹性公网IP的请求体
type DisassociatePublicipsRequestBody struct {
	Publicip *DisassociatePublicipsOption `json:"publicip"`
}

func (o DisassociatePublicipsRequestBody) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "DisassociatePublicipsRequestBody struct{}"
	}

	return strings.Join([]string{"DisassociatePublicipsRequestBody", string(data)}, " ")
}
