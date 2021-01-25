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

// 配额信息
type ListQuotasResult struct {
	// 配额列表
	Resources []Resources `json:"resources"`
}

func (o ListQuotasResult) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "ListQuotasResult struct{}"
	}

	return strings.Join([]string{"ListQuotasResult", string(data)}, " ")
}
