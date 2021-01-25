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

type SignBindingReq struct {
	// 签名密钥编号
	SignId string `json:"sign_id"`
	// API的发布记录编号
	PublishIds []string `json:"publish_ids"`
}

func (o SignBindingReq) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "SignBindingReq struct{}"
	}

	return strings.Join([]string{"SignBindingReq", string(data)}, " ")
}
