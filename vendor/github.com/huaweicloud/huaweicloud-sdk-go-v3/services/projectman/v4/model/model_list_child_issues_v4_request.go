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

// Request Object
type ListChildIssuesV4Request struct {
	ProjectId string `json:"project_id"`
	IssueId   int32  `json:"issue_id"`
}

func (o ListChildIssuesV4Request) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "ListChildIssuesV4Request struct{}"
	}

	return strings.Join([]string{"ListChildIssuesV4Request", string(data)}, " ")
}
