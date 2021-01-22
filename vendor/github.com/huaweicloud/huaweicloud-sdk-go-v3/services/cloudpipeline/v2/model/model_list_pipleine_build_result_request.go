/*
 * CloudPipeline
 *
 * devcloud pipeline api
 *
 */

package model

import (
	"encoding/json"

	"strings"
)

// Request Object
type ListPipleineBuildResultRequest struct {
	XLanguage *string `json:"X-Language,omitempty"`
	ProjectId string  `json:"project_id"`
	StartDate string  `json:"start_date"`
	EndDate   string  `json:"end_date"`
	Offset    int32   `json:"offset"`
	Limit     int32   `json:"limit"`
}

func (o ListPipleineBuildResultRequest) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "ListPipleineBuildResultRequest struct{}"
	}

	return strings.Join([]string{"ListPipleineBuildResultRequest", string(data)}, " ")
}
