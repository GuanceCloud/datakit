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

type AddMemberRequestV4 struct {
	// 租户id
	DomainId string `json:"domain_id"`
	// 用户在项目中的角色ID 3, 4, 5, 6, 7 , 8
	RoleId *int32 `json:"role_id,omitempty"`
	// 用户32位uuid
	UserId string `json:"user_id"`
}

func (o AddMemberRequestV4) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "AddMemberRequestV4 struct{}"
	}

	return strings.Join([]string{"AddMemberRequestV4", string(data)}, " ")
}
