/*
 * CCE
 *
 * CCE开放API
 *
 */

package model

import (
	"encoding/json"

	"strings"
)

type Users struct {
	// 当前为固定值“user“。
	Name *string `json:"name,omitempty"`
	User *User   `json:"user,omitempty"`
}

func (o Users) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "Users struct{}"
	}

	return strings.Join([]string{"Users", string(data)}, " ")
}
