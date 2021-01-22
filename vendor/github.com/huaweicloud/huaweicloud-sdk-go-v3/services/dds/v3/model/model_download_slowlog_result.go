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

type DownloadSlowlogResult struct {
	// 文件下载链接
	FileLink string `json:"file_link"`
	// 文件名称
	FileName string `json:"file_name"`
	// 文件大小
	FileSize string `json:"file_size"`
	// 文件所属节点名称
	NodeName string `json:"node_name"`
	// 文件查询状态
	Status string `json:"status"`
	// 文件最后更新时间
	UpdateAt int32 `json:"update_at"`
}

func (o DownloadSlowlogResult) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "DownloadSlowlogResult struct{}"
	}

	return strings.Join([]string{"DownloadSlowlogResult", string(data)}, " ")
}
