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

type BatchAddMemberRequestV4 struct {
	// 用户在项目中的角色ID 3, 4, 5, 6, 7 , 8
	RoleId *int32 `json:"role_id,omitempty"`
	// 用户32位uuid
	UserId string `json:"user_id"`
}

func (o BatchAddMemberRequestV4) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "BatchAddMemberRequestV4 struct{}"
	}

	return strings.Join([]string{"BatchAddMemberRequestV4", string(data)}, " ")
}
