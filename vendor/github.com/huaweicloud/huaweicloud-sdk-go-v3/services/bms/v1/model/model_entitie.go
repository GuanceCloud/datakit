/*
 * BMS
 *
 * BMS Open API
 *
 */

package model

import (
	"encoding/json"

	"strings"
)

// Job操作的对象。根据不同Job类型，显示不同的内容。裸金属服务器相关操作显示server_id；网卡相关操作显示nic_id
type Entitie struct {
	// 裸金属服务器相关操作显示server_id
	ServerId *string `json:"server_id,omitempty"`
	// 网卡相关操作显示nic_id
	NicId *string `json:"nic_id,omitempty"`
}

func (o Entitie) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "Entitie struct{}"
	}

	return strings.Join([]string{"Entitie", string(data)}, " ")
}
