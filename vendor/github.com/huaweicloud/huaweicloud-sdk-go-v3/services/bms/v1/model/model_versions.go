/*
 * BMS
 *
 * BMS Open API
 *
 */

package model

import (
	"encoding/json"
	"errors"
	"github.com/huaweicloud/huaweicloud-sdk-go-v3/core/converter"
	"github.com/huaweicloud/huaweicloud-sdk-go-v3/core/sdktime"
	"strings"
)

// 描述裸金属服务器API版本信息列表
type Versions struct {
	// API版本ID
	Id *VersionsId `json:"id,omitempty"`
	// API的url地址
	Links *[]VersionLinks `json:"links,omitempty"`
	// API支持的最小微版本号
	MinVersion *string `json:"min_version,omitempty"`
	// 这个是API版本的状态。可以是：CURRENT这是使用的API的首选版本；SUPPORTED：这是一个较老的，但仍然支持的API版本；DEPRECATED：一个被废弃的API版本，该版本将被删除
	Status *VersionsStatus `json:"status,omitempty"`
	// API支持的最大微版本号
	Version *string `json:"version,omitempty"`
	// API版本发布时间
	Updated *sdktime.SdkTime `json:"updated,omitempty"`
}

func (o Versions) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "Versions struct{}"
	}

	return strings.Join([]string{"Versions", string(data)}, " ")
}

type VersionsId struct {
	value string
}

type VersionsIdEnum struct {
	V1 VersionsId
}

func GetVersionsIdEnum() VersionsIdEnum {
	return VersionsIdEnum{
		V1: VersionsId{
			value: "v1",
		},
	}
}

func (c VersionsId) MarshalJSON() ([]byte, error) {
	return json.Marshal(c.value)
}

func (c *VersionsId) UnmarshalJSON(b []byte) error {
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

type VersionsStatus struct {
	value string
}

type VersionsStatusEnum struct {
	CURRENT    VersionsStatus
	SUPPORTED  VersionsStatus
	DEPRECATED VersionsStatus
}

func GetVersionsStatusEnum() VersionsStatusEnum {
	return VersionsStatusEnum{
		CURRENT: VersionsStatus{
			value: "CURRENT",
		},
		SUPPORTED: VersionsStatus{
			value: "SUPPORTED",
		},
		DEPRECATED: VersionsStatus{
			value: "DEPRECATED",
		},
	}
}

func (c VersionsStatus) MarshalJSON() ([]byte, error) {
	return json.Marshal(c.value)
}

func (c *VersionsStatus) UnmarshalJSON(b []byte) error {
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
