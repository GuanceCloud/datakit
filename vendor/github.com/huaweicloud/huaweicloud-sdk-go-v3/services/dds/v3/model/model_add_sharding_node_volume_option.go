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

type AddShardingNodeVolumeOption struct {
	// 指定新增的所有shard组的磁盘容量。取值范围：10GB~2000GB。
	Size int32 `json:"size"`
}

func (o AddShardingNodeVolumeOption) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "AddShardingNodeVolumeOption struct{}"
	}

	return strings.Join([]string{"AddShardingNodeVolumeOption", string(data)}, " ")
}
