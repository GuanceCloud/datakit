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

// 分片列表
type InstanceGroupListInfo struct {
	// 分片id
	GroupId *string `json:"group_id,omitempty"`
	// 分片名称
	GroupName *string `json:"group_name,omitempty"`
	// 每个分片包含的副本列表。
	ReplicationList *[]InstanceReplicationListInfo `json:"replication_list,omitempty"`
}

func (o InstanceGroupListInfo) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "InstanceGroupListInfo struct{}"
	}

	return strings.Join([]string{"InstanceGroupListInfo", string(data)}, " ")
}
