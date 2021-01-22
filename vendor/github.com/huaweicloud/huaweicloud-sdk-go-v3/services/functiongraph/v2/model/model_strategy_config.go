/*
 * FunctionGraph
 *
 * API v2
 *
 */

package model

import (
	"encoding/json"

	"strings"
)

// 函数策略配置。
type StrategyConfig struct {
	// 0：函数被禁用;-1：函数被启用。
	Concurrency int32 `json:"concurrency"`
}

func (o StrategyConfig) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "StrategyConfig struct{}"
	}

	return strings.Join([]string{"StrategyConfig", string(data)}, " ")
}
