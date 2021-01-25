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

// 工作项类型
type IssueItemSfv4Tracker struct {
	// 类型id
	Id *int32 `json:"id,omitempty"`
	// 类型名称
	Name *string `json:"name,omitempty"`
}

func (o IssueItemSfv4Tracker) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "IssueItemSfv4Tracker struct{}"
	}

	return strings.Join([]string{"IssueItemSfv4Tracker", string(data)}, " ")
}
