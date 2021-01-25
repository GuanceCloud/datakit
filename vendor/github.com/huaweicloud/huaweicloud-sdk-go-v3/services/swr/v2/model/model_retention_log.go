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

type RetentionLog struct {
	// 创建时间
	CreatedAt string `json:"created_at"`
	// ID
	Id int32 `json:"id"`
	// 组织名
	Namespace string `json:"namespace"`
	// 镜像仓库名
	Repo string `json:"repo"`
	// 老化规则ID
	RetentionId int32 `json:"retention_id"`
	// 规则
	RuleType string `json:"rule_type"`
	// 镜像版本
	Tag string `json:"tag"`
}

func (o RetentionLog) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "RetentionLog struct{}"
	}

	return strings.Join([]string{"RetentionLog", string(data)}, " ")
}
