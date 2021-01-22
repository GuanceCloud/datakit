/*
 * DDS
 *
 * API v3
 *
 */

package model

import (
	"encoding/json"
	"errors"
	"github.com/huaweicloud/huaweicloud-sdk-go-v3/core/converter"
	"strings"
)

type ListInstancesByTagsRequestBody struct {
	// 索引位置偏移量，表示从第一条数据偏移offset条数据后开始查询。 - “action”值为“count”时，不传该参数。 - “action”值为“filter”时，取值必须为数字，不能为负数。默认取0值，表示从第一条数据开始查询。'
	Offset *string `json:"offset,omitempty"`
	// 查询记录数。   - “action”值为“count”时，不传该参数。   - “action”值为“filter”时，取值范围：1~100。不传该参数时，默认查询前100条实例信息。
	Limit *string `json:"limit,omitempty"`
	// 操作标识。   - 取值为“filter”，表示根据标签过滤条件查询实例。   - 取值为“count”，表示仅返回总记录数，禁止返回其他字段。
	Action ListInstancesByTagsRequestBodyAction `json:"action"`
	// 搜索字段。   - 该字段值为空，表示不按照实例名称或实例ID查询。   - 该字段值不为空
	Matches *[]QueryMatchItem `json:"matches,omitempty"`
	// 包含标签，最多包含10个key。
	Tags *[]QueryTagItem `json:"tags,omitempty"`
}

func (o ListInstancesByTagsRequestBody) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "ListInstancesByTagsRequestBody struct{}"
	}

	return strings.Join([]string{"ListInstancesByTagsRequestBody", string(data)}, " ")
}

type ListInstancesByTagsRequestBodyAction struct {
	value string
}

type ListInstancesByTagsRequestBodyActionEnum struct {
	FILTER ListInstancesByTagsRequestBodyAction
	COUNT  ListInstancesByTagsRequestBodyAction
}

func GetListInstancesByTagsRequestBodyActionEnum() ListInstancesByTagsRequestBodyActionEnum {
	return ListInstancesByTagsRequestBodyActionEnum{
		FILTER: ListInstancesByTagsRequestBodyAction{
			value: "filter",
		},
		COUNT: ListInstancesByTagsRequestBodyAction{
			value: "count",
		},
	}
}

func (c ListInstancesByTagsRequestBodyAction) MarshalJSON() ([]byte, error) {
	return json.Marshal(c.value)
}

func (c *ListInstancesByTagsRequestBodyAction) UnmarshalJSON(b []byte) error {
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
