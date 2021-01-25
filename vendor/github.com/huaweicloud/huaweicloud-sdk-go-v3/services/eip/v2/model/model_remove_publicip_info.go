/*
 * EIP
 *
 * 云服务接口
 *
 */

package model

import (
	"encoding/json"

	"strings"
)

// 共享带宽插入/移除弹性公网IP的publicip_info字段
type RemovePublicipInfo struct {
	// 功能说明：若publicip_id为弹性公网IP的id，则该字段可自动忽略。若publicip_id为IPv6端口PORT的id，则该字段必填：5_dualStack(目前仅北京4局点支持)
	PublicipType *string `json:"publicip_type,omitempty"`
	// 功能说明：带宽对应的弹性公网IP或IPv6端口PORT的唯一标识
	PublicipId string `json:"publicip_id"`
}

func (o RemovePublicipInfo) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "RemovePublicipInfo struct{}"
	}

	return strings.Join([]string{"RemovePublicipInfo", string(data)}, " ")
}
