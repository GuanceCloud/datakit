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
type BatchAddMembersV4Request struct {
	ProjectId string                        `json:"project_id"`
	Body      *BatchAddMembersV4RequestBody `json:"body,omitempty"`
}

func (o BatchAddMembersV4Request) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "BatchAddMembersV4Request struct{}"
	}

	return strings.Join([]string{"BatchAddMembersV4Request", string(data)}, " ")
}
