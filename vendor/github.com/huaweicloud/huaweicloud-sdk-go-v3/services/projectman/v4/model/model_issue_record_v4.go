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

// 历史记录
type IssueRecordV4 struct {
	User *IssueRecordV4User `json:"user,omitempty"`
	// 操作的记录
	Details *[]IssueRecordV4Details `json:"details,omitempty"`
}

func (o IssueRecordV4) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "IssueRecordV4 struct{}"
	}

	return strings.Join([]string{"IssueRecordV4", string(data)}, " ")
}
