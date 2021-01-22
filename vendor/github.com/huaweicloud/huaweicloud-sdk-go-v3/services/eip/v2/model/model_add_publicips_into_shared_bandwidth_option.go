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

// 带宽对象
type AddPublicipsIntoSharedBandwidthOption struct {
	// 功能说明：要插入共享带宽的弹性公网IP或者IPv6端口信息  约束：WHOLE类型的带宽支持多个弹性公网IP或者IPv6端口，跟租户的配额相关，默认一个共享带宽的配额为20
	PublicipInfo []InsertPublicipInfo `json:"publicip_info"`
}

func (o AddPublicipsIntoSharedBandwidthOption) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "AddPublicipsIntoSharedBandwidthOption struct{}"
	}

	return strings.Join([]string{"AddPublicipsIntoSharedBandwidthOption", string(data)}, " ")
}
