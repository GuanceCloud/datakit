/*
 * RabbitMQ
 *
 * RabbitMQ Document API
 *
 */

package model

import (
	"encoding/json"

	"strings"
)

// Response Object
type ListPluginsResponse struct {
	// 插件信息列表。
	Plugins        *[]ListPluginsRespPlugins `json:"plugins,omitempty"`
	HttpStatusCode int                       `json:"-"`
}

func (o ListPluginsResponse) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "ListPluginsResponse struct{}"
	}

	return strings.Join([]string{"ListPluginsResponse", string(data)}, " ")
}
