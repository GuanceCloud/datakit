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

type AppAuthReq struct {
	// 需要授权的环境编号
	EnvId string `json:"env_id"`
	// APP的编号列表
	AppIds []string `json:"app_ids"`
	// API的编号列表，可以选择租户自己的API，也可以选择从云市场上购买的API。
	ApiIds []string `json:"api_ids"`
}

func (o AppAuthReq) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "AppAuthReq struct{}"
	}

	return strings.Join([]string{"AppAuthReq", string(data)}, " ")
}
