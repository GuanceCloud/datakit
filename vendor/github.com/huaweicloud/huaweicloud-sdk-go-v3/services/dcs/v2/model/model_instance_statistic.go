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

// 实例的统计信息。
type InstanceStatistic struct {
	// 缓存实例网络入流量，单位：Kbps。
	InputKbps *string `json:"input_kbps,omitempty"`
	// 缓存实例网络出流量，单位：Kbps。
	OutputKbps *string `json:"output_kbps,omitempty"`
	// 实例ID。
	InstanceId *string `json:"instance_id,omitempty"`
	// 缓存存储的数据条数。
	Keys *int64 `json:"keys,omitempty"`
	// 缓存已经使用内存，单位：MB。
	UsedMemory *int64 `json:"used_memory,omitempty"`
	// 缓存的总内存，单位：MB。
	MaxMemory *int64 `json:"max_memory,omitempty"`
	// 缓存get命令被调用次数。
	CmdGetCount *int64 `json:"cmd_get_count,omitempty"`
	// 缓存set命令被调用次数。
	CmdSetCount *int64 `json:"cmd_set_count,omitempty"`
	// CPU使用率，单位：百分比。
	UsedCpu *string `json:"used_cpu,omitempty"`
}

func (o InstanceStatistic) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "InstanceStatistic struct{}"
	}

	return strings.Join([]string{"InstanceStatistic", string(data)}, " ")
}
