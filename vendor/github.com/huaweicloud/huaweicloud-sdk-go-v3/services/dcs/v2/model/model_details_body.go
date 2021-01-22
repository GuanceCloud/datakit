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

// 详细信息结构体
type DetailsBody struct {
	// 变更前的容量，仅在变更规格时有值
	OldCapacity *string `json:"old_capacity,omitempty"`
	// 变更后的容量，仅在变更规格时有值
	NewCapacity *string `json:"new_capacity,omitempty"`
	// 是否开启公网访问，仅在开启公网访问时有值
	EnablePublicIp *bool `json:"enable_public_ip,omitempty"`
	// 公网IP的ID，仅在开启公网访问时有值
	PublicIpId *string `json:"public_ip_id,omitempty"`
	// 公网IP地址，仅在开启公网访问时有值
	PublicIpAddress *string `json:"public_ip_address,omitempty"`
	// 是否开启ssl，仅在开启ssl时有值
	EnableSsl *bool `json:"enable_ssl,omitempty"`
	// 变更前的缓存类型，仅在变更规格时有值
	OldCacheMode *string `json:"old_cache_mode,omitempty"`
	// 变更后的缓存类型，仅在变更规格时有值
	NewCacheMode *string `json:"new_cache_mode,omitempty"`
}

func (o DetailsBody) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "DetailsBody struct{}"
	}

	return strings.Join([]string{"DetailsBody", string(data)}, " ")
}
