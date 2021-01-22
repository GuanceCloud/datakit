/*
 * SWR
 *
 * SWR API
 *
 */

package model

import (
	"encoding/json"

	"strings"
)

type Link struct {
	// 链接
	Href string `json:"href"`
	// 描述
	Rel string `json:"rel"`
}

func (o Link) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "Link struct{}"
	}

	return strings.Join([]string{"Link", string(data)}, " ")
}
