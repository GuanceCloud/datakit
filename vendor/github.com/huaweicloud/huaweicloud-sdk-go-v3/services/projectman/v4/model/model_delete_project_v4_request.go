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
type DeleteProjectV4Request struct {
	ProjectId string `json:"project_id"`
}

func (o DeleteProjectV4Request) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "DeleteProjectV4Request struct{}"
	}

	return strings.Join([]string{"DeleteProjectV4Request", string(data)}, " ")
}
