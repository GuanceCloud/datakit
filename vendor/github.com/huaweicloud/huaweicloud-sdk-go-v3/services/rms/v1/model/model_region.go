/*
 * RMS
 *
 * Resource Manager Api
 *
 */

package model

import (
	"encoding/json"

	"strings"
)

// 区域
type Region struct {
	// 区域ID
	RegionId *string `json:"region_id,omitempty"`
	// 显示名称
	DisplayName *string `json:"display_name,omitempty"`
}

func (o Region) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "Region struct{}"
	}

	return strings.Join([]string{"Region", string(data)}, " ")
}
