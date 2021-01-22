/*
 * CCE
 *
 * CCE开放API
 *
 */

package model

import (
	"encoding/json"

	"strings"
)

type Authentication struct {
	AuthenticatingProxy *AuthenticatingProxy `json:"authenticatingProxy,omitempty"`
	// 集群认证模式。  - kubernetes 1.11及之前版本的集群支持“x509”、“rbac”和“authenticating_proxy”，默认取值为“x509”。 - kubernetes 1.13及以上版本的集群支持“rbac”和“authenticating_proxy”，默认取值为“rbac”。
	Mode *string `json:"mode,omitempty"`
}

func (o Authentication) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "Authentication struct{}"
	}

	return strings.Join([]string{"Authentication", string(data)}, " ")
}
