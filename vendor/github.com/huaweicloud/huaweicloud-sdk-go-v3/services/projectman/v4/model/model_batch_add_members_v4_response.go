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
type BatchAddMembersV4Response struct {
	HttpStatusCode int `json:"-"`
}

func (o BatchAddMembersV4Response) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "BatchAddMembersV4Response struct{}"
	}

	return strings.Join([]string{"BatchAddMembersV4Response", string(data)}, " ")
}
