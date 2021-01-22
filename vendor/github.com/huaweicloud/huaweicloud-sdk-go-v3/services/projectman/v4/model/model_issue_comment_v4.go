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

type IssueCommentV4 struct {
	// 评论内容
	Comment *string `json:"comment,omitempty"`
	// 评论id
	Id *int32 `json:"id,omitempty"`
	// 评论时间
	CreatedTime *string        `json:"created_time,omitempty"`
	User        *CommentUserV4 `json:"user,omitempty"`
}

func (o IssueCommentV4) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "IssueCommentV4 struct{}"
	}

	return strings.Join([]string{"IssueCommentV4", string(data)}, " ")
}
