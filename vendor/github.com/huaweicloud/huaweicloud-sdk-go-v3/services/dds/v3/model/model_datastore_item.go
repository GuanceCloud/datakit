/*
 * DDS
 *
 * API v3
 *
 */

package model

import (
	"encoding/json"

	"strings"
)

// 数据库信息。
type DatastoreItem struct {
	// 数据库引擎。
	Type string `json:"type"`
	// 数据库版本号。
	Version string `json:"version"`
}

func (o DatastoreItem) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "DatastoreItem struct{}"
	}

	return strings.Join([]string{"DatastoreItem", string(data)}, " ")
}
