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
type ListProjectDemandStaticV4Request struct {
	ProjectId string `json:"project_id"`
}

func (o ListProjectDemandStaticV4Request) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "ListProjectDemandStaticV4Request struct{}"
	}

	return strings.Join([]string{"ListProjectDemandStaticV4Request", string(data)}, " ")
}
