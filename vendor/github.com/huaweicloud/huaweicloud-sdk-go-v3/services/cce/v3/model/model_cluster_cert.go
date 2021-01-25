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

type ClusterCert struct {
	// 证书授权数据。
	CertificateAuthorityData *string `json:"certificate-authority-data,omitempty"`
	// 不校验服务端证书，在 cluster 类型为 externalCluster 时，该值为 true。
	InsecureSkipTlsVerify *bool `json:"insecure-skip-tls-verify,omitempty"`
	// 服务器地址。
	Server *string `json:"server,omitempty"`
}

func (o ClusterCert) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "ClusterCert struct{}"
	}

	return strings.Join([]string{"ClusterCert", string(data)}, " ")
}
