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

type UpdateFunctionReservedInstancesRequestBody struct {
	// 预留实例个数
	Count int32 `json:"count"`
}

func (o UpdateFunctionReservedInstancesRequestBody) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "UpdateFunctionReservedInstancesRequestBody struct{}"
	}

	return strings.Join([]string{"UpdateFunctionReservedInstancesRequestBody", string(data)}, " ")
}
