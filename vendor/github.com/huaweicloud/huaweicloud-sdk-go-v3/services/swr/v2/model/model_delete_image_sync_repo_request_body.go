/*
 * SWR
 *
 * SWR API
 *
 */

package model

import (
	"encoding/json"

	"strings"
)

// 需要删除镜像自动同步任务的信息
type DeleteImageSyncRepoRequestBody struct {
	// 目标region ID。
	RemoteRegionId string `json:"remoteRegionId"`
	// 目标组织
	RemoteNamespace string `json:"remoteNamespace"`
}

func (o DeleteImageSyncRepoRequestBody) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "DeleteImageSyncRepoRequestBody struct{}"
	}

	return strings.Join([]string{"DeleteImageSyncRepoRequestBody", string(data)}, " ")
}
