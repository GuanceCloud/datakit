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
type UpdateIterationV4Response struct {
	HttpStatusCode int `json:"-"`
}

func (o UpdateIterationV4Response) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "UpdateIterationV4Response struct{}"
	}

	return strings.Join([]string{"UpdateIterationV4Response", string(data)}, " ")
}
