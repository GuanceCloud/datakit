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

type DomainReq struct {
	// 自定义域名。长度为0-255位的字符串，需要符合域名规范。
	UrlDomain string `json:"url_domain"`
}

func (o DomainReq) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "DomainReq struct{}"
	}

	return strings.Join([]string{"DomainReq", string(data)}, " ")
}
