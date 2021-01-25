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

type DownloadSlowlogRequestBody struct {
	// - 需要下载的文件的文件名列表。
	FileNameList *[]string `json:"file_name_list,omitempty"`
	// 节点ID列表，取空值，表示查询实例下所有允许查询的节点。使用请参考《DDS API参考》的“查询实例列表”响应消息表“nodes 数据结构说明”的“id”。允许查询的节点如下： - 集群下面的 shard节点 - 副本集、单节点下面的所有节点
	NodeIdList *[]string `json:"node_id_list,omitempty"`
}

func (o DownloadSlowlogRequestBody) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "DownloadSlowlogRequestBody struct{}"
	}

	return strings.Join([]string{"DownloadSlowlogRequestBody", string(data)}, " ")
}
