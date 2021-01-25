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

// 工作项操作的用户
type IssueRecordV4User struct {
	// 用户数字id
	UserNumId *int32 `json:"user_num_id,omitempty"`
	// 登录名
	UserName *string `json:"user_name,omitempty"`
	// 昵称
	NickName *string `json:"nick_name,omitempty"`
	// 用户32位的uuid
	UserId *string `json:"user_id,omitempty"`
}

func (o IssueRecordV4User) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "IssueRecordV4User struct{}"
	}

	return strings.Join([]string{"IssueRecordV4User", string(data)}, " ")
}
