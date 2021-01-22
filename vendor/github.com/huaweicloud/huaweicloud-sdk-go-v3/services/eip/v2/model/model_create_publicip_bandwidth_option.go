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

// 创建弹性公网IP时，携带的待绑定带宽信息
type CreatePublicipBandwidthOption struct {
	// 功能说明：按流量计费还是按带宽计费。  其中IPv6国外默认是bandwidth，国内默认是traffic。取值为traffic，表示流量计费。
	ChargeMode *CreatePublicipBandwidthOptionChargeMode `json:"charge_mode,omitempty"`
	// 功能说明：带宽ID  创建WHOLE类型带宽的弹性公网IP时可以指定之前的共享带宽创建  取值范围：WHOLE类型的带宽ID
	Id *string `json:"id,omitempty"`
	// 功能说明：带宽名称  取值范围：1-64个字符，支持数字、字母、中文、_(下划线)、-（中划线）、.（点）  如果share_type是PER，该参数必须带,如果share_type是WHOLE并且id有值，该参数会忽略。
	Name *string `json:"name,omitempty"`
	// 功能说明：带宽类型 取值范围：PER，WHOLE。 约束：该字段为WHOLE时，必须指定带宽ID。
	ShareType CreatePublicipBandwidthOptionShareType `json:"share_type"`
	// 功能说明：带宽大小  取值范围：默认1Mbit/s~2000Mbit/s（具体范围以各区域配置为准，请参见控制台对应页面显示）。  约束：share_type是PER，该参数必须带，如果share_type是WHOLE并且id有值，该参数会忽略。  注意：调整带宽时的最小单位会根据带宽范围不同存在差异。  小于等于300Mbit/s：默认最小单位为1Mbit/s。  300Mbit/s~1000Mbit/s：默认最小单位为50Mbit/s。  大于1000Mbit/s：默认最小单位为500Mbit/s。
	Size *int32 `json:"size,omitempty"`
}

func (o CreatePublicipBandwidthOption) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "CreatePublicipBandwidthOption struct{}"
	}

	return strings.Join([]string{"CreatePublicipBandwidthOption", string(data)}, " ")
}

type CreatePublicipBandwidthOptionChargeMode struct {
	value string
}

type CreatePublicipBandwidthOptionChargeModeEnum struct {
	BANDWIDTH CreatePublicipBandwidthOptionChargeMode
	TRAFFIC   CreatePublicipBandwidthOptionChargeMode
}

func GetCreatePublicipBandwidthOptionChargeModeEnum() CreatePublicipBandwidthOptionChargeModeEnum {
	return CreatePublicipBandwidthOptionChargeModeEnum{
		BANDWIDTH: CreatePublicipBandwidthOptionChargeMode{
			value: "bandwidth",
		},
		TRAFFIC: CreatePublicipBandwidthOptionChargeMode{
			value: "traffic",
		},
	}
}

func (c CreatePublicipBandwidthOptionChargeMode) MarshalJSON() ([]byte, error) {
	return json.Marshal(c.value)
}

func (c *CreatePublicipBandwidthOptionChargeMode) UnmarshalJSON(b []byte) error {
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

type CreatePublicipBandwidthOptionShareType struct {
	value string
}

type CreatePublicipBandwidthOptionShareTypeEnum struct {
	WHOLE CreatePublicipBandwidthOptionShareType
	PER   CreatePublicipBandwidthOptionShareType
}

func GetCreatePublicipBandwidthOptionShareTypeEnum() CreatePublicipBandwidthOptionShareTypeEnum {
	return CreatePublicipBandwidthOptionShareTypeEnum{
		WHOLE: CreatePublicipBandwidthOptionShareType{
			value: "WHOLE",
		},
		PER: CreatePublicipBandwidthOptionShareType{
			value: "PER",
		},
	}
}

func (c CreatePublicipBandwidthOptionShareType) MarshalJSON() ([]byte, error) {
	return json.Marshal(c.value)
}

func (c *CreatePublicipBandwidthOptionShareType) UnmarshalJSON(b []byte) error {
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
