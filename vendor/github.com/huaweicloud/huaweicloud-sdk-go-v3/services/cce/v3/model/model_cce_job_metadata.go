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

type CceJobMetadata struct {
	// 作业的创建时间。
	CreationTimestamp *string `json:"creationTimestamp,omitempty"`
	// 作业的ID。
	Uid *string `json:"uid,omitempty"`
	// 作业的更新时间。
	UpdateTimestamp *string `json:"updateTimestamp,omitempty"`
}

func (o CceJobMetadata) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "CceJobMetadata struct{}"
	}

	return strings.Join([]string{"CceJobMetadata", string(data)}, " ")
}
