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
type RestartOrFlushInstancesResponse struct {
	// 删除/重启/清空实例的结果。
	Results        *[]BatchOpsResult `json:"results,omitempty"`
	HttpStatusCode int               `json:"-"`
}

func (o RestartOrFlushInstancesResponse) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "RestartOrFlushInstancesResponse struct{}"
	}

	return strings.Join([]string{"RestartOrFlushInstancesResponse", string(data)}, " ")
}
