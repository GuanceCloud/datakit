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
type AssociateSignatureKeyV2Response struct {
	// API与签名密钥的绑定关系列表
	Bindings       *[]SignBindingApiResp `json:"bindings,omitempty"`
	HttpStatusCode int                   `json:"-"`
}

func (o AssociateSignatureKeyV2Response) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "AssociateSignatureKeyV2Response struct{}"
	}

	return strings.Join([]string{"AssociateSignatureKeyV2Response", string(data)}, " ")
}
