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

type CreateRepoDomainsRequestBody struct {
	// 共享租户名称
	AccessDomain string `json:"access_domain"`
	// 当前只支持read权限
	Permit string `json:"permit"`
	// 截止时间，UTC时间格式。永久有效为forever
	Deadline string `json:"deadline"`
	// 描述
	Description *string `json:"description,omitempty"`
}

func (o CreateRepoDomainsRequestBody) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "CreateRepoDomainsRequestBody struct{}"
	}

	return strings.Join([]string{"CreateRepoDomainsRequestBody", string(data)}, " ")
}
