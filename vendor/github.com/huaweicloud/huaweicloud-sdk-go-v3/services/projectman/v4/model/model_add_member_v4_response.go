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
type AddMemberV4Response struct {
	HttpStatusCode int `json:"-"`
}

func (o AddMemberV4Response) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "AddMemberV4Response struct{}"
	}

	return strings.Join([]string{"AddMemberV4Response", string(data)}, " ")
}
