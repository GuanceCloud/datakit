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

type UserTag struct {
	// 云服务器标签的键。不得以\"CCE-\"或\"__type_baremetal\"开头
	Key *string `json:"key,omitempty"`
	// 云服务器标签的值
	Value *string `json:"value,omitempty"`
}

func (o UserTag) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "UserTag struct{}"
	}

	return strings.Join([]string{"UserTag", string(data)}, " ")
}
