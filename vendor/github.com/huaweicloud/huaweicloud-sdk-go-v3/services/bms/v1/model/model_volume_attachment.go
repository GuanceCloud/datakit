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

// volumeAttachment数据结构说明
type VolumeAttachment struct {
	// 要挂卷的卷ID。可以从云硬盘控制台查询，或者通过调用“查询云硬盘列表”API获取。
	VolumeId string `json:"volumeId"`
	// 磁盘挂载点，如/dev/sda、/dev/sdb。新增加的磁盘挂载点不能和已有的磁盘挂载点相同。需要根据已有设备名称顺序指定，否则不写device或device的值为空时，由系统自动生成。
	Device *string `json:"device,omitempty"`
}

func (o VolumeAttachment) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "VolumeAttachment struct{}"
	}

	return strings.Join([]string{"VolumeAttachment", string(data)}, " ")
}
