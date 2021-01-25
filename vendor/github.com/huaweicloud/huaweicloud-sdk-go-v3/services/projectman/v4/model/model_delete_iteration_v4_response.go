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
type DeleteIterationV4Response struct {
	HttpStatusCode int `json:"-"`
}

func (o DeleteIterationV4Response) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "DeleteIterationV4Response struct{}"
	}

	return strings.Join([]string{"DeleteIterationV4Response", string(data)}, " ")
}
