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
type ShowtIssueCompletionRateRequest struct {
	ProjectId string `json:"project_id"`
}

func (o ShowtIssueCompletionRateRequest) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "ShowtIssueCompletionRateRequest struct{}"
	}

	return strings.Join([]string{"ShowtIssueCompletionRateRequest", string(data)}, " ")
}
