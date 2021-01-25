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

// servers字段数据结构说明
type ServersList struct {
	// 裸金属服务器ID。可以从裸金属服务器控制台查询，或者通过调用7.3.4-查询裸金属服务器列表（OpenStack原生）API获取。
	Id string `json:"id"`
}

func (o ServersList) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "ServersList struct{}"
	}

	return strings.Join([]string{"ServersList", string(data)}, " ")
}
