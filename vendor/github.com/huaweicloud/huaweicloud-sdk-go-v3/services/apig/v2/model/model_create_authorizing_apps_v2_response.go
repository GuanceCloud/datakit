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
type CreateAuthorizingAppsV2Response struct {
	// API与APP的授权关系列表
	Auths          *[]AppAuthResp `json:"auths,omitempty"`
	HttpStatusCode int            `json:"-"`
}

func (o CreateAuthorizingAppsV2Response) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "CreateAuthorizingAppsV2Response struct{}"
	}

	return strings.Join([]string{"CreateAuthorizingAppsV2Response", string(data)}, " ")
}
