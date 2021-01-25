/*
 * DDS
 *
 * API v3
 *
 */

package model

import (
	"encoding/json"

	"strings"
)

type InstanceItem struct {
	// 实例ID。
	InstanceId string `json:"instance_id"`
	// 实例名称
	InstanceName string `json:"instance_name"`
	// 标签列表。如果没有标签，默认为空数组。
	Tags []InstanceItemTagItem `json:"tags"`
}

func (o InstanceItem) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "InstanceItem struct{}"
	}

	return strings.Join([]string{"InstanceItem", string(data)}, " ")
}
