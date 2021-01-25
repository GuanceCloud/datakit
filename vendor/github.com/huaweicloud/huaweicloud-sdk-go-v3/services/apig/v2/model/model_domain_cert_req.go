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

type DomainCertReq struct {
	// 证书内容
	CertContent string `json:"cert_content"`
	// 证书名称。长度为4 ~ 50位的字符串，字符串由中文、英文字母、数字、下划线组成，且只能以英文或中文开头。
	Name string `json:"name"`
	// 私钥内容
	PrivateKey string `json:"private_key"`
}

func (o DomainCertReq) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "DomainCertReq struct{}"
	}

	return strings.Join([]string{"DomainCertReq", string(data)}, " ")
}
