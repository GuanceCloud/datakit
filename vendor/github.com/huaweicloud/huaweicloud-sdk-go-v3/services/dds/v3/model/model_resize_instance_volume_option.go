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

type ResizeInstanceVolumeOption struct {
	// 角色组ID。 - 对于集群实例，该参数为shard组ID。 - 对于副本集和单节点实例，不传该参数。
	GroupId *string `json:"group_id,omitempty"`
	// 待扩容到的磁盘容量。取值为10的整数倍，并且大于当前磁盘容量。 - 对于集群实例，表示扩容到的单个shard组的磁盘容量。取值范围：10GB~2000GB。 - 对于副本集实例，表示扩容到的实例的磁盘容量，取值范围：10GB~2000GB。 - 对于单节点实例，表示扩容到的实例的磁盘容量，取值范围：10GB~1000GB。
	Size int32 `json:"size"`
}

func (o ResizeInstanceVolumeOption) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "ResizeInstanceVolumeOption struct{}"
	}

	return strings.Join([]string{"ResizeInstanceVolumeOption", string(data)}, " ")
}
