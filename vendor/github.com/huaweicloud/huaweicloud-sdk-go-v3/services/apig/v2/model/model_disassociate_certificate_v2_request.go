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

// Request Object
type DisassociateCertificateV2Request struct {
	ProjectId     string `json:"project_id"`
	InstanceId    string `json:"instance_id"`
	GroupId       string `json:"group_id"`
	CertificateId string `json:"certificate_id"`
}

func (o DisassociateCertificateV2Request) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "DisassociateCertificateV2Request struct{}"
	}

	return strings.Join([]string{"DisassociateCertificateV2Request", string(data)}, " ")
}
