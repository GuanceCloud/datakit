/*
 * FunctionGraph
 *
 * API v2
 *
 */

package model

import (
	"encoding/json"

	"strings"
)

// 函数挂载结构体。
type FuncMount struct {
	// 挂载类型(sfs/sfsTurbo/ecs)，func_mounts非空时必选。
	MountType string `json:"mount_type"`
	// 挂载资源ID（对应云服务ID），func_mounts非空时必选。
	MountResource string `json:"mount_resource"`
	// 远端挂载路径（例如192.168.0.12:/data），如果mount_type为ecs，必选。
	MountSharePath *string `json:"mount_share_path,omitempty"`
	// 函数访问路径，func_mounts非空时必选。
	LocalMountPath string `json:"local_mount_path"`
}

func (o FuncMount) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "FuncMount struct{}"
	}

	return strings.Join([]string{"FuncMount", string(data)}, " ")
}
