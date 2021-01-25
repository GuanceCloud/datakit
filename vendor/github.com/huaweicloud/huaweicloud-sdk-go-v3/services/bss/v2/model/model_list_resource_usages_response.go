/*
 * BSS
 *
 * Business Support System API
 *
 */

package model

import (
	"encoding/json"

	"strings"
)

// Response Object
type ListResourceUsagesResponse struct {
	// |参数名称：套餐包使用量信息| |参数的约束及描述：套餐包使用量信息|
	PackageUsageInfos *[]PackageUsageInfo `json:"package_usage_infos,omitempty"`
	HttpStatusCode    int                 `json:"-"`
}

func (o ListResourceUsagesResponse) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "ListResourceUsagesResponse struct{}"
	}

	return strings.Join([]string{"ListResourceUsagesResponse", string(data)}, " ")
}
