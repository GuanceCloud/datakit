/*
 * BMS
 *
 * BMS Open API
 *
 */

package model

import (
	"encoding/json"

	"strings"
)

// 裸金属服务器挂载信息列表
type VolumeAttachments struct {
	// 挂载资源ID
	Id *string `json:"id,omitempty"`
	// 所属裸金属服务器ID
	ServerId *string `json:"serverId,omitempty"`
	// 挂载云磁盘ID
	VolumeId *string `json:"volumeId,omitempty"`
	// 挂载目录，例如“/dev/sdd”。
	Device *string `json:"device,omitempty"`
}

func (o VolumeAttachments) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "VolumeAttachments struct{}"
	}

	return strings.Join([]string{"VolumeAttachments", string(data)}, " ")
}
