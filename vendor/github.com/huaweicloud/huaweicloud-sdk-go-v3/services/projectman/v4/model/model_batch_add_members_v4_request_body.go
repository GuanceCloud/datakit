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

type BatchAddMembersV4RequestBody struct {
	// 添加的用户信息
	Users []BatchAddMemberRequestV4 `json:"users"`
}

func (o BatchAddMembersV4RequestBody) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "BatchAddMembersV4RequestBody struct{}"
	}

	return strings.Join([]string{"BatchAddMembersV4RequestBody", string(data)}, " ")
}
