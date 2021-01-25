/*
 * ProjectMan
 *
 * devcloud projectman api
 *
 */

package model

import (
	"encoding/json"

	"strings"
)

// Response Object
type ListIssuesV4Response struct {
	// 工作项列表
	Issues *[]IssueResponseV4 `json:"issues,omitempty"`
	// 总数
	Total          *int32 `json:"total,omitempty"`
	HttpStatusCode int    `json:"-"`
}

func (o ListIssuesV4Response) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "ListIssuesV4Response struct{}"
	}

	return strings.Join([]string{"ListIssuesV4Response", string(data)}, " ")
}
