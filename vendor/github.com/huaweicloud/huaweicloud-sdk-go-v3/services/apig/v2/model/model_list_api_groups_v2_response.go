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
type ListApiGroupsV2Response struct {
	// 满足条件的分组总数
	Total *int32 `json:"total,omitempty"`
	// 本次返回的列表长度
	Size *int32 `json:"size,omitempty"`
	// 分组列表
	Groups         *[]ApiGroupDetailResp `json:"groups,omitempty"`
	HttpStatusCode int                   `json:"-"`
}

func (o ListApiGroupsV2Response) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "ListApiGroupsV2Response struct{}"
	}

	return strings.Join([]string{"ListApiGroupsV2Response", string(data)}, " ")
}
