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

// Response Object
type UpdateProjectV4Response struct {
	HttpStatusCode int `json:"-"`
}

func (o UpdateProjectV4Response) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "UpdateProjectV4Response struct{}"
	}

	return strings.Join([]string{"UpdateProjectV4Response", string(data)}, " ")
}
