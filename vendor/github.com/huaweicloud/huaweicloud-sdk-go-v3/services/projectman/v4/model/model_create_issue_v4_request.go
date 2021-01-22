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
type CreateIssueV4Request struct {
	ProjectId string                `json:"project_id"`
	Body      *CreateIssueRequestV4 `json:"body,omitempty"`
}

func (o CreateIssueV4Request) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "CreateIssueV4Request struct{}"
	}

	return strings.Join([]string{"CreateIssueV4Request", string(data)}, " ")
}
