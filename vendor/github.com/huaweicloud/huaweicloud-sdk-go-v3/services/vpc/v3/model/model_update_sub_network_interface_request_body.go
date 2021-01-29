/*
 * VPC
 *
 * VPC Open API
 *
 */

package model

import (
	"encoding/json"

	"strings"
)

// This is a auto create Body Object
type UpdateSubNetworkInterfaceRequestBody struct {
	// 功能说明：是否只预检此次请求 取值范围： -true：发送检查请求，不会更新辅助弹性网卡。检查项包括是否填写了必需参数、请求格式、业务限制。如果检查不通过，则返回对应错误。如果检查通过，则返回响应码202。 -false（默认值）：发送正常请求，并直接更新辅助弹性网卡。
	DryRun              *bool                            `json:"dry_run,omitempty"`
	SubNetworkInterface *UpdateSubNetworkInterfaceOption `json:"sub_network_interface"`
}

func (o UpdateSubNetworkInterfaceRequestBody) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "UpdateSubNetworkInterfaceRequestBody struct{}"
	}

	return strings.Join([]string{"UpdateSubNetworkInterfaceRequestBody", string(data)}, " ")
}
