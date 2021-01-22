/*
 * kps
 *
 * kps v3 版本API
 *
 */

package model

import (
	"encoding/json"

	"strings"
)

// Response Object
type AssociateKeypairResponse struct {
	// 任务下发成功返回的ID
	TaskId         *string `json:"task_id,omitempty"`
	HttpStatusCode int     `json:"-"`
}

func (o AssociateKeypairResponse) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "AssociateKeypairResponse struct{}"
	}

	return strings.Join([]string{"AssociateKeypairResponse", string(data)}, " ")
}
