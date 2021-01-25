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

type UrlDomainsResp struct {
	// 域名编号
	Id *string `json:"id,omitempty"`
	// 访问域名
	Domain *string `json:"domain,omitempty"`
	// 域名cname状态： - 1：未解析 - 2：解析中 - 3：解析成功 - 4：解析失败
	CnameStatus *int32 `json:"cname_status,omitempty"`
	// SSL证书编号
	SslId *string `json:"ssl_id,omitempty"`
	// SSL证书名称
	SslName *string `json:"ssl_name,omitempty"`
}

func (o UrlDomainsResp) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "UrlDomainsResp struct{}"
	}

	return strings.Join([]string{"UrlDomainsResp", string(data)}, " ")
}
