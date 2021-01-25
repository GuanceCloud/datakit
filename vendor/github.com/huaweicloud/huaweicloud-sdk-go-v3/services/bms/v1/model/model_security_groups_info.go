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

// security_groups字段数据结构说明
type SecurityGroupsInfo struct {
	// 裸金属服务器对应的安全组ID，对创建裸金属服务器中配置的所有网卡生效。当该参数未指定时默认给裸金属服务器绑定default安全组。当该参数传值（UUID格式）时需要指定已有安全组的ID。获取已有安全组的方法请参见《虚拟私有云API参考》的“查询安全组列表”章节。
	Id *string `json:"id,omitempty"`
}

func (o SecurityGroupsInfo) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "SecurityGroupsInfo struct{}"
	}

	return strings.Join([]string{"SecurityGroupsInfo", string(data)}, " ")
}
