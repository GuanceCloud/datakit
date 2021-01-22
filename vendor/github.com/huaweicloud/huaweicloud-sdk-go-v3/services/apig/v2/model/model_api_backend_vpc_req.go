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

// VPC通道详情。vpc_channel_status = 1，则这个object类型为必填信息
type ApiBackendVpcReq struct {
	// 代理主机
	VpcChannelProxyHost *string `json:"vpc_channel_proxy_host,omitempty"`
	// VPC通道编号
	VpcChannelId string `json:"vpc_channel_id"`
}

func (o ApiBackendVpcReq) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "ApiBackendVpcReq struct{}"
	}

	return strings.Join([]string{"ApiBackendVpcReq", string(data)}, " ")
}
