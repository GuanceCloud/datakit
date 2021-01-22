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

// 关闭裸金属服务器接口请求结构体
type OsStopBody struct {
	OsStop *OsStopBodyType `json:"os-stop"`
}

func (o OsStopBody) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "OsStopBody struct{}"
	}

	return strings.Join([]string{"OsStopBody", string(data)}, " ")
}
