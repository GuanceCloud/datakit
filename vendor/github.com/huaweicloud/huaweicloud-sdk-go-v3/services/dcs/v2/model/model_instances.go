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

type Instances struct {
	// 缓存实例ID。
	InstanceId *string `json:"instance_id,omitempty"`
	// 缓存实例名称。
	InstanceName *string `json:"instance_name,omitempty"`
}

func (o Instances) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "Instances struct{}"
	}

	return strings.Join([]string{"Instances", string(data)}, " ")
}
