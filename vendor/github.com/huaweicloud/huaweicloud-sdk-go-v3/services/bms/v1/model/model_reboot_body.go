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

// 重启裸金属服务器接口请求结构体
type RebootBody struct {
	Reboot *ServersInfoType `json:"reboot"`
}

func (o RebootBody) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "RebootBody struct{}"
	}

	return strings.Join([]string{"RebootBody", string(data)}, " ")
}
