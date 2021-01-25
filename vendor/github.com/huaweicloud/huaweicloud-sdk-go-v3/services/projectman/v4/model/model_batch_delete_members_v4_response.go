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
type BatchDeleteMembersV4Response struct {
	HttpStatusCode int `json:"-"`
}

func (o BatchDeleteMembersV4Response) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "BatchDeleteMembersV4Response struct{}"
	}

	return strings.Join([]string{"BatchDeleteMembersV4Response", string(data)}, " ")
}
