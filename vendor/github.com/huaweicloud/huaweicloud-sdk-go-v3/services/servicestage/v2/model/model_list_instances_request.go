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
type ListInstancesRequest struct {
	ApplicationId string                     `json:"application_id"`
	ComponentId   string                     `json:"component_id"`
	Limit         *int32                     `json:"limit,omitempty"`
	Offset        *int32                     `json:"offset,omitempty"`
	OrderBy       *string                    `json:"order_by,omitempty"`
	Order         *ListInstancesRequestOrder `json:"order,omitempty"`
}

func (o ListInstancesRequest) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "ListInstancesRequest struct{}"
	}

	return strings.Join([]string{"ListInstancesRequest", string(data)}, " ")
}

type ListInstancesRequestOrder struct {
	value string
}

type ListInstancesRequestOrderEnum struct {
	DESC ListInstancesRequestOrder
	ASC  ListInstancesRequestOrder
}

func GetListInstancesRequestOrderEnum() ListInstancesRequestOrderEnum {
	return ListInstancesRequestOrderEnum{
		DESC: ListInstancesRequestOrder{
			value: "desc",
		},
		ASC: ListInstancesRequestOrder{
			value: "asc",
		},
	}
}

func (c ListInstancesRequestOrder) MarshalJSON() ([]byte, error) {
	return json.Marshal(c.value)
}

func (c *ListInstancesRequestOrder) UnmarshalJSON(b []byte) error {
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
