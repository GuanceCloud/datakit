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

// 慢日志单个条目
type SlowlogItem struct {
	// 慢日志的唯一标识
	Id *int32 `json:"id,omitempty"`
	// 慢命令
	Command *string `json:"command,omitempty"`
	// 执行开始时间,格式为“2020-06-19T07:06:07Z”
	StartTime *string `json:"start_time,omitempty"`
	// 持续时间，单位是ms
	Duration *string `json:"duration,omitempty"`
	// 慢命令所在的分片名称，仅在实例类型为集群时支持
	ShardName *string `json:"shard_name,omitempty"`
}

func (o SlowlogItem) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "SlowlogItem struct{}"
	}

	return strings.Join([]string{"SlowlogItem", string(data)}, " ")
}
