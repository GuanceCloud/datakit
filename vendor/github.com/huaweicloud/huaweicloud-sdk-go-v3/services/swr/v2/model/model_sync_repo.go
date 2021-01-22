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

type SyncRepo struct {
	// 创建时间，UTC日期格式，时间为UTC标准时间，用户需要根据本地时间计算偏移量；如东8区需要+8:00
	CreatedAt string `json:"createdAt"`
	// 租户ID
	DomainID string `json:"domainID"`
	// 租户名
	DomainName string `json:"domainName"`
	// ID
	Id int32 `json:"id"`
	// 组织名
	Namespace string `json:"namespace"`
	// 是否覆盖
	Override bool `json:"override"`
	// 目的组织
	RemoteNamespace string `json:"remoteNamespace"`
	// 目的region
	RemoteRegionId string `json:"remoteRegionId"`
	// 仓库名
	RepoName string `json:"repoName"`
	// 自动同步
	SynAuto bool `json:"synAuto"`
	// 更新时间，UTC日期格式，时间为UTC标准时间，用户需要根据本地时间计算偏移量；如东8区需要+8:00
	UpdatedAt string `json:"updatedAt"`
}

func (o SyncRepo) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "SyncRepo struct{}"
	}

	return strings.Join([]string{"SyncRepo", string(data)}, " ")
}
