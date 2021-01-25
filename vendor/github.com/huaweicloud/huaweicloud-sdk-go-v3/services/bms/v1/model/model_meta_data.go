/*
 * BMS
 *
 * BMS Open API
 *
 */

package model

import (
	"encoding/json"

	"strings"
)

// 更新裸金属服务器元数据
type MetaData struct {
	Metadata *KeyValue `json:"metadata"`
}

func (o MetaData) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "MetaData struct{}"
	}

	return strings.Join([]string{"MetaData", string(data)}, " ")
}
