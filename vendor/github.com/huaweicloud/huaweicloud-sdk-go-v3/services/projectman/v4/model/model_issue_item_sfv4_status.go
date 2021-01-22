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

// 工作项状态
type IssueItemSfv4Status struct {
	// 状态id
	Id *int32 `json:"id,omitempty"`
	// 状态名称
	Name *string `json:"name,omitempty"`
}

func (o IssueItemSfv4Status) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "IssueItemSfv4Status struct{}"
	}

	return strings.Join([]string{"IssueItemSfv4Status", string(data)}, " ")
}
