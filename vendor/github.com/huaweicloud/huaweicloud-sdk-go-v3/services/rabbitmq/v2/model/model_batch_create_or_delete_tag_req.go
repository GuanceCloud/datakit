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

type BatchCreateOrDeleteTagReq struct {
	// 操作标识（仅支持小写）: - create（创建） - delete（删除）
	Action *BatchCreateOrDeleteTagReqAction `json:"action,omitempty"`
	// 标签列表。
	Tags *[]CreateInstanceReqTags `json:"tags,omitempty"`
}

func (o BatchCreateOrDeleteTagReq) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "BatchCreateOrDeleteTagReq struct{}"
	}

	return strings.Join([]string{"BatchCreateOrDeleteTagReq", string(data)}, " ")
}

type BatchCreateOrDeleteTagReqAction struct {
	value string
}

type BatchCreateOrDeleteTagReqActionEnum struct {
	CREATE BatchCreateOrDeleteTagReqAction
	DELETE BatchCreateOrDeleteTagReqAction
}

func GetBatchCreateOrDeleteTagReqActionEnum() BatchCreateOrDeleteTagReqActionEnum {
	return BatchCreateOrDeleteTagReqActionEnum{
		CREATE: BatchCreateOrDeleteTagReqAction{
			value: "create",
		},
		DELETE: BatchCreateOrDeleteTagReqAction{
			value: "delete",
		},
	}
}

func (c BatchCreateOrDeleteTagReqAction) MarshalJSON() ([]byte, error) {
	return json.Marshal(c.value)
}

func (c *BatchCreateOrDeleteTagReqAction) UnmarshalJSON(b []byte) error {
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
