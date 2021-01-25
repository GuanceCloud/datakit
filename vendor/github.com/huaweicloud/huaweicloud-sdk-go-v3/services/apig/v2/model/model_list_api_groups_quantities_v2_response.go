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
type ListApiGroupsQuantitiesV2Response struct {
	// 未上架的API分组个数  暂不支持
	OffsellNums *int32 `json:"offsell_nums,omitempty"`
	// 已上架的API分组个数
	OnsellNums     *int32 `json:"onsell_nums,omitempty"`
	HttpStatusCode int    `json:"-"`
}

func (o ListApiGroupsQuantitiesV2Response) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "ListApiGroupsQuantitiesV2Response struct{}"
	}

	return strings.Join([]string{"ListApiGroupsQuantitiesV2Response", string(data)}, " ")
}
