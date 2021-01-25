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

// Response Object
type ShowBaremetalServerVolumeInfoResponse struct {
	// 裸金属服务器挂载信息列表，详情请参见表2 volumeAttachments字段数据结构说明。
	VolumeAttachments *[]VolumeAttachments `json:"volumeAttachments,omitempty"`
	HttpStatusCode    int                  `json:"-"`
}

func (o ShowBaremetalServerVolumeInfoResponse) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "ShowBaremetalServerVolumeInfoResponse struct{}"
	}

	return strings.Join([]string{"ShowBaremetalServerVolumeInfoResponse", string(data)}, " ")
}
