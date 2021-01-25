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

// IP白名单分组列表结构体
type Whitelist struct {
	// 白名单分组名称，每个实例支持创建4个分组。
	GroupName string `json:"group_name"`
	// 白名单分组下的IP列表,每个实例最多可以添加20个IP地址/地址段。如果有多个，可以用逗号分隔。不支持的IP和地址段：0.0.0.0和0.0.0.0/0
	IpList []string `json:"ip_list"`
}

func (o Whitelist) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "Whitelist struct{}"
	}

	return strings.Join([]string{"Whitelist", string(data)}, " ")
}
