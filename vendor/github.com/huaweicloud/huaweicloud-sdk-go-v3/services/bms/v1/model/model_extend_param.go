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
	"strings"
)

// 创建裸金属服务器的extendparam字段数据结构说明
type ExtendParam struct {
	// 计费模式。取值范围：prePaid：预付费，即包年包月。默认值是prePaid。
	ChargingMode *ExtendParamChargingMode `json:"chargingMode,omitempty"`
	// 裸金属服务器所在区域ID。请参考地区和终端节点获取。
	RegionID *string `json:"regionID,omitempty"`
	// 订购周期类型。取值范围：month：月year：年 说明：chargingMode为prePaid时生效，且为必选值。
	PeriodType *ExtendParamPeriodType `json:"periodType,omitempty"`
	// 订购周期数。取值范围：periodType=month（周期类型为月）时，取值为[1-9]。periodType=year（周期类型为年）时，取值为1。 说明：chargingMode为prePaid时生效，且为必选值。
	PeriodNum *int32 `json:"periodNum,omitempty"`
	// 是否自动续订。true：自动续订false：不自动续订 说明：chargingMode为prePaid时生效，不指定该参数或者该参数值为空时默认为不自动续订。
	IsAutoRenew *string `json:"isAutoRenew,omitempty"`
	// 下单订购后，是否自动从客户的帐户中支付，而不需要客户手动去支付。true：是（自动支付）false：否（需要客户手动支付） 说明：chargingMode为prePaid时生效，不指定该参数或者该参数值为空时默认为客户手动支付。
	IsAutoPay *string `json:"isAutoPay,omitempty"`
	// 企业项目ID。该字段不传（或传为字符串“0”），则将资源绑定给默认企业项目。 说明：关于企业项目ID的获取及企业项目特性的详细信息，请参见《企业管理API参考》。
	EnterpriseProjectId *string `json:"enterprise_project_id,omitempty"`
}

func (o ExtendParam) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "ExtendParam struct{}"
	}

	return strings.Join([]string{"ExtendParam", string(data)}, " ")
}

type ExtendParamChargingMode struct {
	value string
}

type ExtendParamChargingModeEnum struct {
	PRE_PAID ExtendParamChargingMode
}

func GetExtendParamChargingModeEnum() ExtendParamChargingModeEnum {
	return ExtendParamChargingModeEnum{
		PRE_PAID: ExtendParamChargingMode{
			value: "prePaid",
		},
	}
}

func (c ExtendParamChargingMode) MarshalJSON() ([]byte, error) {
	return json.Marshal(c.value)
}

func (c *ExtendParamChargingMode) UnmarshalJSON(b []byte) error {
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

type ExtendParamPeriodType struct {
	value string
}

type ExtendParamPeriodTypeEnum struct {
	MONTH ExtendParamPeriodType
	YEAR  ExtendParamPeriodType
}

func GetExtendParamPeriodTypeEnum() ExtendParamPeriodTypeEnum {
	return ExtendParamPeriodTypeEnum{
		MONTH: ExtendParamPeriodType{
			value: "month",
		},
		YEAR: ExtendParamPeriodType{
			value: "year",
		},
	}
}

func (c ExtendParamPeriodType) MarshalJSON() ([]byte, error) {
	return json.Marshal(c.value)
}

func (c *ExtendParamPeriodType) UnmarshalJSON(b []byte) error {
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
