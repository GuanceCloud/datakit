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

type PersistentVolumeClaimStatus struct {
	// 显示volume实际具有的访问模式。
	AccessModes *[]string `json:"accessModes,omitempty"`
	// 底层卷的实际资源
	Capacity *string `json:"capacity,omitempty"`
	// PersistentVolumeClaim当前所处的状态
	Phase *string `json:"phase,omitempty"`
}

func (o PersistentVolumeClaimStatus) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "PersistentVolumeClaimStatus struct{}"
	}

	return strings.Join([]string{"PersistentVolumeClaimStatus", string(data)}, " ")
}
