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

type BackupDatabase struct {
	// SqlServer引擎指定备份的数据库。
	Name string `json:"name"`
}

func (o BackupDatabase) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "BackupDatabase struct{}"
	}

	return strings.Join([]string{"BackupDatabase", string(data)}, " ")
}
