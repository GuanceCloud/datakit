/*
 * RDS
 *
 * API v3
 *
 */

package model

import (
	"encoding/json"

	"strings"
)

// Response Object
type ListInstancesResponse struct {
	Instances *[]InstanceResponse `json:"instances,omitempty"`
	// 总实例数。
	TotalCount     *int32 `json:"total_count,omitempty"`
	HttpStatusCode int    `json:"-"`
}

func (o ListInstancesResponse) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "ListInstancesResponse struct{}"
	}

	return strings.Join([]string{"ListInstancesResponse", string(data)}, " ")
}
