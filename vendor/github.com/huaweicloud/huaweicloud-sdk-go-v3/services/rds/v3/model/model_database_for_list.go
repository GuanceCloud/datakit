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

// 数据库信息。
type DatabaseForList struct {
	// 数据库名称。 数据库名称长度可在1～64个字符之间，由字母、数字、中划线、下划线或$组成，$累计总长度小于等于10个字符，（MySQL 8.0不可包含$）。
	Name string `json:"name"`
	// 数据库使用的字符集，例如utf8、gbk、ascii等MySQL支持的字符集。
	CharacterSet string `json:"character_set"`
}

func (o DatabaseForList) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "DatabaseForList struct{}"
	}

	return strings.Join([]string{"DatabaseForList", string(data)}, " ")
}
