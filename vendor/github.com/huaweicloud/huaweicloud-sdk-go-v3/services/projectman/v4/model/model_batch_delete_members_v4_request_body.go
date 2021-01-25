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

type BatchDeleteMembersV4RequestBody struct {
	// 用户id
	UserIds []string `json:"user_ids"`
}

func (o BatchDeleteMembersV4RequestBody) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "BatchDeleteMembersV4RequestBody struct{}"
	}

	return strings.Join([]string{"BatchDeleteMembersV4RequestBody", string(data)}, " ")
}
