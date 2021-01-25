/*
 * RabbitMQ
 *
 * RabbitMQ Document API
 *
 */

package model

import (
	"encoding/json"
	"errors"
	"github.com/huaweicloud/huaweicloud-sdk-go-v3/core/converter"
	"strings"
)

// Request Object
type ListProductsRequest struct {
	Engine ListProductsRequestEngine `json:"engine"`
}

func (o ListProductsRequest) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "ListProductsRequest struct{}"
	}

	return strings.Join([]string{"ListProductsRequest", string(data)}, " ")
}

type ListProductsRequestEngine struct {
	value string
}

type ListProductsRequestEngineEnum struct {
	RABBITMQ ListProductsRequestEngine
}

func GetListProductsRequestEngineEnum() ListProductsRequestEngineEnum {
	return ListProductsRequestEngineEnum{
		RABBITMQ: ListProductsRequestEngine{
			value: "rabbitmq",
		},
	}
}

func (c ListProductsRequestEngine) MarshalJSON() ([]byte, error) {
	return json.Marshal(c.value)
}

func (c *ListProductsRequestEngine) UnmarshalJSON(b []byte) error {
	myConverter := converter.StringConverterFactory("string")
	if myConverter != nil {
		val, err := myConverter.CovertStringToInterface(strings.Trim(string(b[:]), "\""))
		if err == nil {
			c.value = val.(string)
			return nil
		}
		return err
	} else {
		return errors.New("convert enum data to string error")
	}
}
