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

// Response Object
type BatchDeleteInstancesResponse struct {
	// 删除/重启/清空实例的结果。
	Results        *[]BatchOpsResult `json:"results,omitempty"`
	HttpStatusCode int               `json:"-"`
}

func (o BatchDeleteInstancesResponse) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "BatchDeleteInstancesResponse struct{}"
	}

	return strings.Join([]string{"BatchDeleteInstancesResponse", string(data)}, " ")
}
