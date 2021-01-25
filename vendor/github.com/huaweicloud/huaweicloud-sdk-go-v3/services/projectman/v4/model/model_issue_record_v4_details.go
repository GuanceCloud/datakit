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

type IssueRecordV4Details struct {
	// 操作属性
	Property *string `json:"property,omitempty"`
	// 上次的记录
	OldValue *string `json:"old_value,omitempty"`
	// 当前值
	NewValue *string `json:"new_value,omitempty"`
	// 操作
	Operation *string `json:"operation,omitempty"`
}

func (o IssueRecordV4Details) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "IssueRecordV4Details struct{}"
	}

	return strings.Join([]string{"IssueRecordV4Details", string(data)}, " ")
}
