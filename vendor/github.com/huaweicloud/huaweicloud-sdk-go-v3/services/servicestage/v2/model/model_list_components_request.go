/*
 * ServiceStage
 *
 * ServiceStage的API,包括应用管理和仓库授权管理
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
type ListComponentsRequest struct {
	ApplicationId string                      `json:"application_id"`
	Limit         *int32                      `json:"limit,omitempty"`
	Offset        *int32                      `json:"offset,omitempty"`
	OrderBy       *string                     `json:"order_by,omitempty"`
	Order         *ListComponentsRequestOrder `json:"order,omitempty"`
}

func (o ListComponentsRequest) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "ListComponentsRequest struct{}"
	}

	return strings.Join([]string{"ListComponentsRequest", string(data)}, " ")
}

type ListComponentsRequestOrder struct {
	value string
}

type ListComponentsRequestOrderEnum struct {
	DESC ListComponentsRequestOrder
	ASC  ListComponentsRequestOrder
}

func GetListComponentsRequestOrderEnum() ListComponentsRequestOrderEnum {
	return ListComponentsRequestOrderEnum{
		DESC: ListComponentsRequestOrder{
			value: "desc",
		},
		ASC: ListComponentsRequestOrder{
			value: "asc",
		},
	}
}

func (c ListComponentsRequestOrder) MarshalJSON() ([]byte, error) {
	return json.Marshal(c.value)
}

func (c *ListComponentsRequestOrder) UnmarshalJSON(b []byte) error {
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
