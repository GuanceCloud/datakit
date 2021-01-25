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
type ListApplicationsRequest struct {
	Limit   *int32                        `json:"limit,omitempty"`
	Offset  *int32                        `json:"offset,omitempty"`
	OrderBy *string                       `json:"order_by,omitempty"`
	Order   *ListApplicationsRequestOrder `json:"order,omitempty"`
}

func (o ListApplicationsRequest) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "ListApplicationsRequest struct{}"
	}

	return strings.Join([]string{"ListApplicationsRequest", string(data)}, " ")
}

type ListApplicationsRequestOrder struct {
	value string
}

type ListApplicationsRequestOrderEnum struct {
	DESC ListApplicationsRequestOrder
	ASC  ListApplicationsRequestOrder
}

func GetListApplicationsRequestOrderEnum() ListApplicationsRequestOrderEnum {
	return ListApplicationsRequestOrderEnum{
		DESC: ListApplicationsRequestOrder{
			value: "desc",
		},
		ASC: ListApplicationsRequestOrder{
			value: "asc",
		},
	}
}

func (c ListApplicationsRequestOrder) MarshalJSON() ([]byte, error) {
	return json.Marshal(c.value)
}

func (c *ListApplicationsRequestOrder) UnmarshalJSON(b []byte) error {
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
