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

// 工作项重要程度
type IssueItemSfv4Severity struct {
	// 重要程度id
	Id *int32 `json:"id,omitempty"`
	// 重要程度
	Name *string `json:"name,omitempty"`
}

func (o IssueItemSfv4Severity) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "IssueItemSfv4Severity struct{}"
	}

	return strings.Join([]string{"IssueItemSfv4Severity", string(data)}, " ")
}
