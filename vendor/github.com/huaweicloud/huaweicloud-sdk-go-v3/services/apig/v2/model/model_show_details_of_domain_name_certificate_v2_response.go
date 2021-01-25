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
type ShowDetailsOfDomainNameCertificateV2Response struct {
	// 证书域名
	CommonName *string `json:"common_name,omitempty"`
	// SAN域名
	San *[]string `json:"san,omitempty"`
	// 证书版本
	Version *string `json:"version,omitempty"`
	// 公司、组织
	Organization *[]string `json:"organization,omitempty"`
	// 部门
	OrganizationalUnit *[]string `json:"organizational_unit,omitempty"`
	// 城市
	Locality *[]string `json:"locality,omitempty"`
	// 省份
	State *[]string `json:"state,omitempty"`
	// 国家
	Country *[]string `json:"country,omitempty"`
	// 证书有效期起始时间
	NotBefore *string `json:"not_before,omitempty"`
	// 证书有效期截止时间
	NotAfter *string `json:"not_after,omitempty"`
	// 序列号
	SerialNumber *string `json:"serial_number,omitempty"`
	// 颁发者
	Issuer *[]string `json:"issuer,omitempty"`
	// 签名算法
	SignatureAlgorithm *string `json:"signature_algorithm,omitempty"`
	HttpStatusCode     int     `json:"-"`
}

func (o ShowDetailsOfDomainNameCertificateV2Response) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "ShowDetailsOfDomainNameCertificateV2Response struct{}"
	}

	return strings.Join([]string{"ShowDetailsOfDomainNameCertificateV2Response", string(data)}, " ")
}
