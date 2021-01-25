/*
 * APIG
 *
 * API网关（API Gateway）是为开发者、合作伙伴提供的高性能、高可用、高安全的API托管服务，帮助用户轻松构建、管理和发布任意规模的API。
 *
 */

package model

import (
	"encoding/json"

	"strings"
)

type VpcBaseInfo struct {
	// 云服务器ID
	EcsId *string `json:"ecs_id,omitempty"`
	// 云服务器名称
	EcsName *int32 `json:"ecs_name,omitempty"`
	// 是否使用级联方式  暂不支持
	CascadeFlag *bool `json:"cascade_flag,omitempty"`
}

func (o VpcBaseInfo) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "VpcBaseInfo struct{}"
	}

	return strings.Join([]string{"VpcBaseInfo", string(data)}, " ")
}
