/*
 * ECS
 *
 * ECS Open API
 *
 */

package model

import (
	"encoding/json"

	"strings"
)

//
type BatchDeleteServerNicOption struct {
	// 网卡Port ID。
	Id string `json:"id"`
}

func (o BatchDeleteServerNicOption) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "BatchDeleteServerNicOption struct{}"
	}

	return strings.Join([]string{"BatchDeleteServerNicOption", string(data)}, " ")
}
