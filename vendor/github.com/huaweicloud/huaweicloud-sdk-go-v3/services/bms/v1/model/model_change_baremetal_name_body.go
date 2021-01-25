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

// 修改裸金属服务器名称接口请求结构体
type ChangeBaremetalNameBody struct {
	Server *ChangeBaremetalNameServer `json:"server"`
}

func (o ChangeBaremetalNameBody) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "ChangeBaremetalNameBody struct{}"
	}

	return strings.Join([]string{"ChangeBaremetalNameBody", string(data)}, " ")
}
