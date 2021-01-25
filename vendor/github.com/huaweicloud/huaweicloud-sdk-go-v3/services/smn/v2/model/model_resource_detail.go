/*
 * SMN
 *
 * SMN Open API
 *
 */

package model

import (
	"encoding/json"

	"strings"
)

type ResourceDetail struct {
	// 企业项目ID
	EnterpriseProjectId string `json:"enterprise_project_id"`
	// 详情ID
	DetailId string `json:"detailId"`
	// topic唯一标识
	TopicUrn string `json:"topic_urn"`
	// 显示名
	DisplayName string `json:"display_name"`
}

func (o ResourceDetail) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "ResourceDetail struct{}"
	}

	return strings.Join([]string{"ResourceDetail", string(data)}, " ")
}
