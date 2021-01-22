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

// flavors数据结构说明
type FlavorsResp struct {
	// 裸金属服务器规格的ID
	Id string `json:"id"`
	// 裸金属服务器规格的名称
	Name string `json:"name"`
	// 该裸金属服务器规格对应的CPU核数。
	Vcpus *string `json:"vcpus,omitempty"`
	// 该裸金属服务器规格对应的内存大小，单位为MB。
	Ram *int32 `json:"ram,omitempty"`
	// 该裸金属服务器规格对应要求系统盘大小，0为不限制。
	Disk *string `json:"disk,omitempty"`
	// 未使用
	Swap *string `json:"swap,omitempty"`
	// 未使用
	OSFLVEXTDATAephemeral *int32 `json:"OS-FLV-EXT-DATA:ephemeral,omitempty"`
	// 未使用
	OSFLVDISABLEDdisabled *bool `json:"OS-FLV-DISABLED:disabled,omitempty"`
	// 未使用
	RxtxFactor *float32 `json:"rxtx_factor,omitempty"`
	// 未使用
	RxtxQuota *string `json:"rxtx_quota,omitempty"`
	// 未使用
	RxtxCap *string `json:"rxtx_cap,omitempty"`
	// 是否是公共规格。false：私有规格；true：公共规格
	OsFlavorAccessisPublic *bool `json:"os-flavor-access:is_public,omitempty"`
	// 规格相关快捷链接地址，详情请参见表3 links字段数据结构说明。
	Links        *[]LinksInfo  `json:"links,omitempty"`
	OsExtraSpecs *OsExtraSpecs `json:"os_extra_specs"`
}

func (o FlavorsResp) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "FlavorsResp struct{}"
	}

	return strings.Join([]string{"FlavorsResp", string(data)}, " ")
}
