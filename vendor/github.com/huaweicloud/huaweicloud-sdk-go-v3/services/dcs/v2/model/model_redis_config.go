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

type RedisConfig struct {
	// 实例配置项的值。
	ParamValue string `json:"param_value"`
	// 实例配置项名。
	ParamName string `json:"param_name"`
	// 实例配置项ID。
	ParamId string `json:"param_id"`
}

func (o RedisConfig) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "RedisConfig struct{}"
	}

	return strings.Join([]string{"RedisConfig", string(data)}, " ")
}
