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

// 修改配置参数请求体
type ModifyRedisConfigBody struct {
	// 实例配置项数组。
	RedisConfig *[]RedisConfig `json:"redis_config,omitempty"`
}

func (o ModifyRedisConfigBody) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "ModifyRedisConfigBody struct{}"
	}

	return strings.Join([]string{"ModifyRedisConfigBody", string(data)}, " ")
}
