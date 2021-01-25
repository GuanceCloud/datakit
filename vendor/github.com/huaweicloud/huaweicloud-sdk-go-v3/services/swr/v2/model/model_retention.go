/*
 * SWR
 *
 * SWR API
 *
 */

package model

import (
	"encoding/json"

	"strings"
)

// 镜像老化规则
type Retention struct {
	// 回收规则匹配策略，or
	Algorithm string `json:"algorithm"`
	// ID
	Id int32 `json:"id"`
	// 镜像老化规则
	Rules []Rule `json:"rules"`
	// 保留字段
	Scope string `json:"scope"`
}

func (o Retention) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "Retention struct{}"
	}

	return strings.Join([]string{"Retention", string(data)}, " ")
}
