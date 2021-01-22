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

type ShowRepoDomainsResponse struct {
	// 命名空间
	Namespace string `json:"namespace"`
	// 镜像仓库
	Repository string `json:"repository"`
	// 共享租户名
	AccessDomain string `json:"access_domain"`
	// 权限
	Permit string `json:"permit"`
	// 截止时间
	Deadline string `json:"deadline"`
	// 描述
	Description string `json:"description"`
	// 创建者ID
	CreatorId string `json:"creator_id"`
	// 创建者名称
	CreatorName string `json:"creator_name"`
	// 镜像创建时间，UTC时间格式，时间为UTC标准时间，用户需要根据本地时间计算偏移量；如东8区需要+8:00
	Created string `json:"created"`
	// 镜像更新时间，UTC时间格式，时间为UTC标准时间，用户需要根据本地时间计算偏移量；如东8区需要+8:00
	Updated string `json:"updated"`
	// 是否过期：true:有效；false:过期
	Status bool `json:"status"`
}

func (o ShowRepoDomainsResponse) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "ShowRepoDomainsResponse struct{}"
	}

	return strings.Join([]string{"ShowRepoDomainsResponse", string(data)}, " ")
}
