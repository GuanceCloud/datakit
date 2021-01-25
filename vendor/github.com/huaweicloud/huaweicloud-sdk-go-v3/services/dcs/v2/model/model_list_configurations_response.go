/*
 * DCS
 *
 * DCS V2版本API
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
type ListConfigurationsResponse struct {
	// 实例操作时间。格式为：2017-03-31T12:24:46.297Z
	ConfigTime *string `json:"config_time,omitempty"`
	// 实例ID。
	InstanceId *string `json:"instance_id,omitempty"`
	// 实例配置项数组。
	RedisConfig *[]QueryRedisConfig `json:"redis_config,omitempty"`
	// 实例修改状态 - UPDATING - FAILURE - SUCCESS
	ConfigStatus *ListConfigurationsResponseConfigStatus `json:"config_status,omitempty"`
	// 实例运行状态。
	Status         *string `json:"status,omitempty"`
	HttpStatusCode int     `json:"-"`
}

func (o ListConfigurationsResponse) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "ListConfigurationsResponse struct{}"
	}

	return strings.Join([]string{"ListConfigurationsResponse", string(data)}, " ")
}

type ListConfigurationsResponseConfigStatus struct {
	value string
}

type ListConfigurationsResponseConfigStatusEnum struct {
	UPDATING ListConfigurationsResponseConfigStatus
	FAILURE  ListConfigurationsResponseConfigStatus
	SUCCESS  ListConfigurationsResponseConfigStatus
}

func GetListConfigurationsResponseConfigStatusEnum() ListConfigurationsResponseConfigStatusEnum {
	return ListConfigurationsResponseConfigStatusEnum{
		UPDATING: ListConfigurationsResponseConfigStatus{
			value: "UPDATING",
		},
		FAILURE: ListConfigurationsResponseConfigStatus{
			value: "FAILURE",
		},
		SUCCESS: ListConfigurationsResponseConfigStatus{
			value: "SUCCESS",
		},
	}
}

func (c ListConfigurationsResponseConfigStatus) MarshalJSON() ([]byte, error) {
	return json.Marshal(c.value)
}

func (c *ListConfigurationsResponseConfigStatus) UnmarshalJSON(b []byte) error {
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
