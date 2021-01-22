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
type BatchDeleteIssuesV4Request struct {
	ProjectId string                       `json:"project_id"`
	Body      *BatchDelelteIssuesRequestV4 `json:"body,omitempty"`
}

func (o BatchDeleteIssuesV4Request) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "BatchDeleteIssuesV4Request struct{}"
	}

	return strings.Join([]string{"BatchDeleteIssuesV4Request", string(data)}, " ")
}
