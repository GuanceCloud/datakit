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
type ListAppQuantitiesV2Response struct {
	// 已进行API访问授权的APP个数
	AuthedNums *int32 `json:"authed_nums,omitempty"`
	// 未进行API访问授权的APP个数
	UnauthedNums   *int32 `json:"unauthed_nums,omitempty"`
	HttpStatusCode int    `json:"-"`
}

func (o ListAppQuantitiesV2Response) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "ListAppQuantitiesV2Response struct{}"
	}

	return strings.Join([]string{"ListAppQuantitiesV2Response", string(data)}, " ")
}
