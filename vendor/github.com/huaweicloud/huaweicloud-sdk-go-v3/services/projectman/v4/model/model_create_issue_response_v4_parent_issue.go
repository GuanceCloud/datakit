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

// 父工作项
type CreateIssueResponseV4ParentIssue struct {
	// 父工作项id
	Id *int32 `json:"id,omitempty"`
	// 父工作项
	Name *string `json:"name,omitempty"`
}

func (o CreateIssueResponseV4ParentIssue) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "CreateIssueResponseV4ParentIssue struct{}"
	}

	return strings.Join([]string{"CreateIssueResponseV4ParentIssue", string(data)}, " ")
}
