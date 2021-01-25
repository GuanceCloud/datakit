/*
 * DevStar
 *
 * DevStar API
 *
 */

package model

import (
	"encoding/json"

	"strings"
)

type RepositoryInfo struct {
	// 代码仓的名称
	Name *string `json:"name,omitempty"`
	// 项目id
	ProjectId *string `json:"project_id,omitempty"`
	// 区域id
	RegionId *string `json:"region_id,omitempty"`
}

func (o RepositoryInfo) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "RepositoryInfo struct{}"
	}

	return strings.Join([]string{"RepositoryInfo", string(data)}, " ")
}
