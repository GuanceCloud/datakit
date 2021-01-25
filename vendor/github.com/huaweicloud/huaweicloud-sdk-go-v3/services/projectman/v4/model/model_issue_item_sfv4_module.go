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

// 模块
type IssueItemSfv4Module struct {
	// 模块id
	Id *int32 `json:"id,omitempty"`
	// 模块
	Name *string `json:"name,omitempty"`
}

func (o IssueItemSfv4Module) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "IssueItemSfv4Module struct{}"
	}

	return strings.Join([]string{"IssueItemSfv4Module", string(data)}, " ")
}
