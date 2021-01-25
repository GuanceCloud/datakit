/*
 * ServiceStage
 *
 * ServiceStage的API,包括应用管理和仓库授权管理
 *
 */

package model

import (
	"encoding/json"

	"strings"
)

// 访问方式。
type ExternalAccesses struct {
	// ID。
	Id       *string                 `json:"id,omitempty"`
	Protocol *ExternalAccessProtocol `json:"protocol"`
	// 访问地址。
	Address string `json:"address"`
	// 应用组件进程监听端口
	ForwardPort int32                 `json:"forward_port"`
	Type        *ExternalAccessType   `json:"type,omitempty"`
	Status      *ExternalAccessStatus `json:"status,omitempty"`
	// 创建时间。
	CreateTime *int64 `json:"create_time,omitempty"`
	// 修改时间。
	UpdateTime *int64 `json:"update_time,omitempty"`
}

func (o ExternalAccesses) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "ExternalAccesses struct{}"
	}

	return strings.Join([]string{"ExternalAccesses", string(data)}, " ")
}
