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
type BatchDeleteMembersV4Request struct {
	ProjectId string                           `json:"project_id"`
	Body      *BatchDeleteMembersV4RequestBody `json:"body,omitempty"`
}

func (o BatchDeleteMembersV4Request) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "BatchDeleteMembersV4Request struct{}"
	}

	return strings.Join([]string{"BatchDeleteMembersV4Request", string(data)}, " ")
}
