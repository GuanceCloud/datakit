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
type CreateSpecialThrottlingConfigurationV2Request struct {
	ProjectId  string              `json:"project_id"`
	InstanceId string              `json:"instance_id"`
	ThrottleId string              `json:"throttle_id"`
	Body       *ThrottleSpecialReq `json:"body,omitempty"`
}

func (o CreateSpecialThrottlingConfigurationV2Request) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "CreateSpecialThrottlingConfigurationV2Request struct{}"
	}

	return strings.Join([]string{"CreateSpecialThrottlingConfigurationV2Request", string(data)}, " ")
}
