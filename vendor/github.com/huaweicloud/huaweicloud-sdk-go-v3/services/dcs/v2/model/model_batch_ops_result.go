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

// 批量删除实例接口返回结构体
type BatchOpsResult struct {
	// 操作结果，取值有success或failed。
	Result *string `json:"result,omitempty"`
	// 缓存实例ID。
	Instance *string `json:"instance,omitempty"`
}

func (o BatchOpsResult) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "BatchOpsResult struct{}"
	}

	return strings.Join([]string{"BatchOpsResult", string(data)}, " ")
}
