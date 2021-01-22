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

// 迭代
type IssueItemSfv4Iteration struct {
	// 迭代id
	Id *int32 `json:"id,omitempty"`
	// 迭代名
	Name *string `json:"name,omitempty"`
}

func (o IssueItemSfv4Iteration) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "IssueItemSfv4Iteration struct{}"
	}

	return strings.Join([]string{"IssueItemSfv4Iteration", string(data)}, " ")
}
