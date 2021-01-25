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

// 需要绑定密钥对的虚拟机信息。
type DisassociateEcsServerInfo struct {
	// 需要绑定(替换或重置)SSH密钥对的虚拟机id
	Id   string `json:"id"`
	Auth *Auth  `json:"auth,omitempty"`
}

func (o DisassociateEcsServerInfo) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "DisassociateEcsServerInfo struct{}"
	}

	return strings.Join([]string{"DisassociateEcsServerInfo", string(data)}, " ")
}
