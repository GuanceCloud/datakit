/*
 * DCS
 *
 * DCS V2版本API
 *
 */

package model

import (
	"encoding/json"

	"strings"
)

// 重启、清空实例数据的请求体
type ChangeInstanceStatusBody struct {
	// 实例的ID列表。
	Instances *[]string `json:"instances,omitempty"`
	// 对实例的操作：  restart: 强制重启  soft_restart: 软重启，只重启进程  flush: 清空数据 > 当前版本，只有Redis 4.0和Redis 5.0实例支持清空数据功能，即flush操作。
	Action *string `json:"action,omitempty"`
}

func (o ChangeInstanceStatusBody) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "ChangeInstanceStatusBody struct{}"
	}

	return strings.Join([]string{"ChangeInstanceStatusBody", string(data)}, " ")
}
