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

type CommentUserV4 struct {
	// 发表评论用户id
	UserNumId *int32 `json:"user_num_id,omitempty"`
	// 发表评论用户名称
	UserName *string `json:"user_name,omitempty"`
	// 发表评论用户昵称
	NickName *string `json:"nick_name,omitempty"`
}

func (o CommentUserV4) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "CommentUserV4 struct{}"
	}

	return strings.Join([]string{"CommentUserV4", string(data)}, " ")
}
