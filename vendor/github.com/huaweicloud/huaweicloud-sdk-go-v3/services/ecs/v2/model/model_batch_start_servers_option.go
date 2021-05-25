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
type BatchStartServersOption struct {
	// 云服务器ID列表
	Servers []ServerId `json:"servers"`
}

func (o BatchStartServersOption) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "BatchStartServersOption struct{}"
	}

	return strings.Join([]string{"BatchStartServersOption", string(data)}, " ")
}
