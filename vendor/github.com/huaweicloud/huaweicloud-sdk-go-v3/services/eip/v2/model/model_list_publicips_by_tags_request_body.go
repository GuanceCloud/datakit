/*
 * EIP
 *
 * 云服务接口
 *
 */

package model

import (
	"encoding/json"
	"errors"
	"github.com/huaweicloud/huaweicloud-sdk-go-v3/core/converter"
	"strings"
)

// 通过标签过滤弹性公网IP的请求体
type ListPublicipsByTagsRequestBody struct {
	// 包含标签，最多包含10个key。  每个key下面的value最多10个，结构体不能缺失，key不能为空或者空字符串。  Key不能重复，同一个key中values不能重复。
	Tags *[]TagReq `json:"tags,omitempty"`
	// 查询记录数（action为count时无此参数）
	Limit *int32 `json:"limit,omitempty"`
	// 索引位置， 从offset指定的下一条数据开始查询。 查询第一页数据时，不需要传入此参数，查询后续页码数据时，将查询前一页数据时响应体中的值带入此参数（action为count时无此参数）
	Offset *int32 `json:"offset,omitempty"`
	// 操作标识：  filter分页查询  count查询总数
	Action ListPublicipsByTagsRequestBodyAction `json:"action"`
	// 搜索字段，key为要匹配的字段，当前仅支持resource_name。value为匹配的值。此字段为固定字典值。
	Matches *[]MatchReq `json:"matches,omitempty"`
}

func (o ListPublicipsByTagsRequestBody) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "ListPublicipsByTagsRequestBody struct{}"
	}

	return strings.Join([]string{"ListPublicipsByTagsRequestBody", string(data)}, " ")
}

type ListPublicipsByTagsRequestBodyAction struct {
	value string
}

type ListPublicipsByTagsRequestBodyActionEnum struct {
	FILTER ListPublicipsByTagsRequestBodyAction
	COUNT  ListPublicipsByTagsRequestBodyAction
}

func GetListPublicipsByTagsRequestBodyActionEnum() ListPublicipsByTagsRequestBodyActionEnum {
	return ListPublicipsByTagsRequestBodyActionEnum{
		FILTER: ListPublicipsByTagsRequestBodyAction{
			value: "filter",
		},
		COUNT: ListPublicipsByTagsRequestBodyAction{
			value: "count",
		},
	}
}

func (c ListPublicipsByTagsRequestBodyAction) MarshalJSON() ([]byte, error) {
	return json.Marshal(c.value)
}

func (c *ListPublicipsByTagsRequestBodyAction) UnmarshalJSON(b []byte) error {
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
