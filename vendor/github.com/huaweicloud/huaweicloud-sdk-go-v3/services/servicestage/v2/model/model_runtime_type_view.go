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

// 运行时参数。
type RuntimeTypeView struct {
	// 类型名称。
	TypeName *string `json:"type_name,omitempty"`
	// 显示名称。
	DisplayName *string `json:"display_name,omitempty"`
	// 容器默认端口。
	ContainerDefaultPort *int32 `json:"container_default_port,omitempty"`
	// 类型描述。
	TypeDesc *string `json:"type_desc,omitempty"`
}

func (o RuntimeTypeView) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "RuntimeTypeView struct{}"
	}

	return strings.Join([]string{"RuntimeTypeView", string(data)}, " ")
}
