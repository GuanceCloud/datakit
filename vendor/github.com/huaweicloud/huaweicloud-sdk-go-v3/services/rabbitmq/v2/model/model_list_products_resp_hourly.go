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

type ListProductsRespHourly struct {
	// 消息引擎的名称，该字段显示为rabbitmq。
	Name *string `json:"name,omitempty"`
	// 消息引擎的版本，当前仅支持3.7.17。
	Version *string `json:"version,omitempty"`
	// 产品规格列表。
	Values *[]ListProductsRespValues `json:"values,omitempty"`
}

func (o ListProductsRespHourly) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "ListProductsRespHourly struct{}"
	}

	return strings.Join([]string{"ListProductsRespHourly", string(data)}, " ")
}
