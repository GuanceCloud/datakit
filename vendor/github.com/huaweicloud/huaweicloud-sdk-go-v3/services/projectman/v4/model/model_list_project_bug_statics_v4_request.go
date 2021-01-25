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
type ListProjectBugStaticsV4Request struct {
	ProjectId string `json:"project_id"`
}

func (o ListProjectBugStaticsV4Request) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "ListProjectBugStaticsV4Request struct{}"
	}

	return strings.Join([]string{"ListProjectBugStaticsV4Request", string(data)}, " ")
}
