/*
 * BSS
 *
 * Business Support System API
 *
 */

package model

import (
	"encoding/json"

	"strings"
)

type QuotaLimitInfo struct {
	// |参数名称：属性key值| |参数约束及描述：属性key值|
	LimitKey *string `json:"limit_key,omitempty"`
	// |参数名称：属性值| |参数约束以及描述：属性值|
	LimitValues *[]LimitValue `json:"limit_values,omitempty"`
}

func (o QuotaLimitInfo) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "QuotaLimitInfo struct{}"
	}

	return strings.Join([]string{"QuotaLimitInfo", string(data)}, " ")
}
