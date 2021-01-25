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

// id, 领域 14, '性能', 15, '功能', 16, '可靠性' 17, '网络安全' 18, '可维护性' 19, '其他DFX' 20, '可用性'
type IssueItemSfv4Domain struct {
	// 领域id
	Id *int32 `json:"id,omitempty"`
	// 领域
	Name *string `json:"name,omitempty"`
}

func (o IssueItemSfv4Domain) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "IssueItemSfv4Domain struct{}"
	}

	return strings.Join([]string{"IssueItemSfv4Domain", string(data)}, " ")
}
