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
type ShowCurUserRoleRequest struct {
	ProjectId string `json:"project_id"`
}

func (o ShowCurUserRoleRequest) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "ShowCurUserRoleRequest struct{}"
	}

	return strings.Join([]string{"ShowCurUserRoleRequest", string(data)}, " ")
}
