/*
 * APIG
 *
 * API网关（API Gateway）是为开发者、合作伙伴提供的高性能、高可用、高安全的API托管服务，帮助用户轻松构建、管理和发布任意规模的API。
 *
 */

package model

import (
	"encoding/json"

	"strings"
)

// Response Object
type ListApiQuantitiesV2Response struct {
	// API总个数
	InstanceNum *int32 `json:"instance_num,omitempty"`
	// 已发布到release环境的API个数
	NumsOnRelease *int32 `json:"nums_on_release,omitempty"`
	// 未发布到release环境的API个数
	NumsOffRelease *int32 `json:"nums_off_release,omitempty"`
	HttpStatusCode int    `json:"-"`
}

func (o ListApiQuantitiesV2Response) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "ListApiQuantitiesV2Response struct{}"
	}

	return strings.Join([]string{"ListApiQuantitiesV2Response", string(data)}, " ")
}
