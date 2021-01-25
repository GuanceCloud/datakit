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
type ShowProjectSummaryV4Request struct {
	ProjectId string `json:"project_id"`
}

func (o ShowProjectSummaryV4Request) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "ShowProjectSummaryV4Request struct{}"
	}

	return strings.Join([]string{"ShowProjectSummaryV4Request", string(data)}, " ")
}
