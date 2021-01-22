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
type DeleteProjectV4Response struct {
	HttpStatusCode int `json:"-"`
}

func (o DeleteProjectV4Response) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "DeleteProjectV4Response struct{}"
	}

	return strings.Join([]string{"DeleteProjectV4Response", string(data)}, " ")
}
