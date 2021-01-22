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

// 集群证书有效期
type CertDuration struct {
	// 集群证书有效时间，单位为天，非管理员用户可申请 1-30天，管理员用户可申请 1-30天或无限限制（-1）
	Duration int32 `json:"duration"`
}

func (o CertDuration) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "CertDuration struct{}"
	}

	return strings.Join([]string{"CertDuration", string(data)}, " ")
}
