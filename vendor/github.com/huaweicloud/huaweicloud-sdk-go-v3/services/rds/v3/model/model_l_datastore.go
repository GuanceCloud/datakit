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

// 数据库版本信息。
type LDatastore struct {
	// 数据库版本ID。
	Id string `json:"id"`
	// 数据库版本号，只返回两位数的大版本号，例如MySQL 5.6.X版本，仅返回5.6。
	Name string `json:"name"`
}

func (o LDatastore) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "LDatastore struct{}"
	}

	return strings.Join([]string{"LDatastore", string(data)}, " ")
}
