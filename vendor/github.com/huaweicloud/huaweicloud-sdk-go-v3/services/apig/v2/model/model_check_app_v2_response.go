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
type CheckAppV2Response struct {
	// 名称
	Name *string `json:"name,omitempty"`
	// 描述
	Remark *string `json:"remark,omitempty"`
	// 编号
	Id             *string `json:"id,omitempty"`
	HttpStatusCode int     `json:"-"`
}

func (o CheckAppV2Response) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "CheckAppV2Response struct{}"
	}

	return strings.Join([]string{"CheckAppV2Response", string(data)}, " ")
}
