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
type ListSignatureKeysBindedToApiV2Request struct {
	ProjectId  string  `json:"project_id"`
	InstanceId string  `json:"instance_id"`
	ApiId      string  `json:"api_id"`
	SignId     *string `json:"sign_id,omitempty"`
	SignName   *string `json:"sign_name,omitempty"`
	EnvId      *string `json:"env_id,omitempty"`
	Offset     *int64  `json:"offset,omitempty"`
	Limit      *int32  `json:"limit,omitempty"`
}

func (o ListSignatureKeysBindedToApiV2Request) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "ListSignatureKeysBindedToApiV2Request struct{}"
	}

	return strings.Join([]string{"ListSignatureKeysBindedToApiV2Request", string(data)}, " ")
}
