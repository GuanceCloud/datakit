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
type UpdateMembesRoleV4Request struct {
	ProjectId string                         `json:"project_id"`
	Body      *UpdateMembesRoleV4RequestBody `json:"body,omitempty"`
}

func (o UpdateMembesRoleV4Request) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "UpdateMembesRoleV4Request struct{}"
	}

	return strings.Join([]string{"UpdateMembesRoleV4Request", string(data)}, " ")
}
