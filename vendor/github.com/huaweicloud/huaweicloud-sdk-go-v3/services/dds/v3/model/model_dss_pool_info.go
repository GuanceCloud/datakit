/*
 * DDS
 *
 * API v3
 *
 */

package model

import (
	"encoding/json"

	"strings"
)

// 实例专属存储信息。
type DssPoolInfo struct {
	// 专属存储池所在az
	AzName string `json:"az_name"`
	// 专属存储池免费空间大小，单位GB
	FreeCapacityGb string `json:"free_capacity_gb"`
	// 专属存储池磁盘类型名称，可能取值如下：  - ULTRAHIGH，表示SSD。
	DssPoolVolumeType string `json:"dss_pool_volume_type"`
	// 专属存储池ID
	DssPoolId string `json:"dss_pool_id"`
	// 专属存储池当前状态，可能取值如下： - available，表示可用。 - deploying，表示正在部署。 - enlarging，表示正在扩容。 - frozen，表示冻结。 - sellout，表示售罄。
	DssPoolStatus string `json:"dss_pool_status"`
}

func (o DssPoolInfo) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "DssPoolInfo struct{}"
	}

	return strings.Join([]string{"DssPoolInfo", string(data)}, " ")
}
