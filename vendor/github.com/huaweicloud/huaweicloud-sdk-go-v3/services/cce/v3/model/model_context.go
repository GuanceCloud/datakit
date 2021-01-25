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

type Context struct {
	// 上下文cluster信息。
	Cluster *string `json:"cluster,omitempty"`
	// 上下文user信息。
	User *string `json:"user,omitempty"`
}

func (o Context) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "Context struct{}"
	}

	return strings.Join([]string{"Context", string(data)}, " ")
}
