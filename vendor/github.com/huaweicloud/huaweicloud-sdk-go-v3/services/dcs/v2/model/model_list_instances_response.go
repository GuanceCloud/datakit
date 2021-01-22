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
type ListInstancesResponse struct {
	// 实例个数。
	InstanceNum *int32 `json:"instance_num,omitempty"`
	// 实例的详情数组。
	Instances      *[]InstanceListInfo `json:"instances,omitempty"`
	HttpStatusCode int                 `json:"-"`
}

func (o ListInstancesResponse) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "ListInstancesResponse struct{}"
	}

	return strings.Join([]string{"ListInstancesResponse", string(data)}, " ")
}
