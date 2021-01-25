/*
 * DMS
 *
 * DMS Document API
 *
 */

package model

import (
	"encoding/json"
	"errors"
	"github.com/huaweicloud/huaweicloud-sdk-go-v3/core/converter"
	"strings"
)

// Response Object
type ListConsumerGroupsResponse struct {
	// 队列ID。
	QueueId *string `json:"queue_id,omitempty"`
	// 队列的名称。
	QueueName *string `json:"queue_name,omitempty"`
	// 消费组列表。
	Groups *[]ListQueueGroupsRespGroups `json:"groups,omitempty"`
	// 该队列是否开启死信消息。仅当include_deadletter为true时，才有该响应参数。 - enable：表示开启。 - disable：表示不开启。
	RedrivePolicy  *ListConsumerGroupsResponseRedrivePolicy `json:"redrive_policy,omitempty"`
	HttpStatusCode int                                      `json:"-"`
}

func (o ListConsumerGroupsResponse) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "ListConsumerGroupsResponse struct{}"
	}

	return strings.Join([]string{"ListConsumerGroupsResponse", string(data)}, " ")
}

type ListConsumerGroupsResponseRedrivePolicy struct {
	value string
}

type ListConsumerGroupsResponseRedrivePolicyEnum struct {
	ENABLE  ListConsumerGroupsResponseRedrivePolicy
	DISABLE ListConsumerGroupsResponseRedrivePolicy
}

func GetListConsumerGroupsResponseRedrivePolicyEnum() ListConsumerGroupsResponseRedrivePolicyEnum {
	return ListConsumerGroupsResponseRedrivePolicyEnum{
		ENABLE: ListConsumerGroupsResponseRedrivePolicy{
			value: "enable",
		},
		DISABLE: ListConsumerGroupsResponseRedrivePolicy{
			value: "disable",
		},
	}
}

func (c ListConsumerGroupsResponseRedrivePolicy) MarshalJSON() ([]byte, error) {
	return json.Marshal(c.value)
}

func (c *ListConsumerGroupsResponseRedrivePolicy) UnmarshalJSON(b []byte) error {
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
