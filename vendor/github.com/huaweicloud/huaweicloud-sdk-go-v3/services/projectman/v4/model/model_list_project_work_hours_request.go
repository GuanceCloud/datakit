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
type ListProjectWorkHoursRequest struct {
	Body *ListProjectWorkHoursRequestBody `json:"body,omitempty"`
}

func (o ListProjectWorkHoursRequest) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "ListProjectWorkHoursRequest struct{}"
	}

	return strings.Join([]string{"ListProjectWorkHoursRequest", string(data)}, " ")
}
