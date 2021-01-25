/*
 * RDS
 *
 * API v3
 *
 */

package model

import (
	"encoding/json"

	"strings"
)

// 数据库及其权限。
type DatabaseWithPrivilege struct {
	// 数据库名称。
	Name string `json:"name"`
	// 是否为只读权限。
	Readonly bool `json:"readonly"`
}

func (o DatabaseWithPrivilege) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "DatabaseWithPrivilege struct{}"
	}

	return strings.Join([]string{"DatabaseWithPrivilege", string(data)}, " ")
}
