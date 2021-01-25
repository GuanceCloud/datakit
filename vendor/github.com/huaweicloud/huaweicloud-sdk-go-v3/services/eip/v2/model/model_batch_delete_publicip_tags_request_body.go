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

// 批量操作资源标签的请求体
type BatchDeletePublicipTagsRequestBody struct {
	// 标签列表
	Tags []ResourceTagOption `json:"tags"`
	// 操作标识  delete：删除  action为delete时，value可选
	Action BatchDeletePublicipTagsRequestBodyAction `json:"action"`
}

func (o BatchDeletePublicipTagsRequestBody) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "BatchDeletePublicipTagsRequestBody struct{}"
	}

	return strings.Join([]string{"BatchDeletePublicipTagsRequestBody", string(data)}, " ")
}

type BatchDeletePublicipTagsRequestBodyAction struct {
	value string
}

type BatchDeletePublicipTagsRequestBodyActionEnum struct {
	DELETE BatchDeletePublicipTagsRequestBodyAction
}

func GetBatchDeletePublicipTagsRequestBodyActionEnum() BatchDeletePublicipTagsRequestBodyActionEnum {
	return BatchDeletePublicipTagsRequestBodyActionEnum{
		DELETE: BatchDeletePublicipTagsRequestBodyAction{
			value: "delete",
		},
	}
}

func (c BatchDeletePublicipTagsRequestBodyAction) MarshalJSON() ([]byte, error) {
	return json.Marshal(c.value)
}

func (c *BatchDeletePublicipTagsRequestBodyAction) UnmarshalJSON(b []byte) error {
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
