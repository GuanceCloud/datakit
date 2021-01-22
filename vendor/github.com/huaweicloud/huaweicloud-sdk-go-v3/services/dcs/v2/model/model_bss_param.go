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

type BssParam struct {
	// 当选择包年包月时，该字段为必选，表示是否自动续订资源。 取值范围： - false：不自动续订； - true：自动续订； 默认值为：false 约束： 如果设置为自动续订，到期后，会自动续订一个月（自动续订时间后续可能会变化），详情可联系客服咨询。
	IsAutoRenew *BssParamIsAutoRenew `json:"is_auto_renew,omitempty"`
	// 功能说明：付费方式（预付费、按需付费；预付费，即包周期付费）。 取值范围： - prePaid：预付费，即包年包月； - postPaid：后付费，即按需付费； 默认值是postPaid。 后付费的场景下，bss_param参数的其他字段都会被忽略。
	ChargingMode BssParamChargingMode `json:"charging_mode"`
	// 功能说明：下单订购后，是否自动从客户的账户中支付；默认是“不自动支付” 。  取值范围： - true：是（自动支付，从账户余额自动扣费） - false：否（默认值，只提交订单不支付，需要客户手动去支付）  约束： 自动支付时，只能使用账户的现金支付；如果要使用代金券，请选择不自动支付，然后在用户费用中心，选择代金券支付。  **如果没有设置成自动支付，即设置为false时，在创建实例之后，实例状态为“支付中”，用户必须在“费用中心 > 我的订单”，完成订单支付，否则订单一直在支付中，实例没有创建成功**。
	IsAutoPay *BssParamIsAutoPay `json:"is_auto_pay,omitempty"`
	// 当选择包年包月时，该字段为必选，表示订购资源的周期类型。  取值范围如下： - month：表示包月 - year：表示包年
	PeriodType *BssParamPeriodType `json:"period_type,omitempty"`
	// 功能说明：订购周期数 取值范围：(后续会随运营策略变化) - period_type为month时，为[1,9]， - period_type为year时，为[1,3]  约束：同period_type约束。
	PeriodNum *int32 `json:"period_num,omitempty"`
}

func (o BssParam) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "BssParam struct{}"
	}

	return strings.Join([]string{"BssParam", string(data)}, " ")
}

type BssParamIsAutoRenew struct {
	value string
}

type BssParamIsAutoRenewEnum struct {
	TRUE  BssParamIsAutoRenew
	FALSE BssParamIsAutoRenew
}

func GetBssParamIsAutoRenewEnum() BssParamIsAutoRenewEnum {
	return BssParamIsAutoRenewEnum{
		TRUE: BssParamIsAutoRenew{
			value: "true",
		},
		FALSE: BssParamIsAutoRenew{
			value: "false",
		},
	}
}

func (c BssParamIsAutoRenew) MarshalJSON() ([]byte, error) {
	return json.Marshal(c.value)
}

func (c *BssParamIsAutoRenew) UnmarshalJSON(b []byte) error {
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

type BssParamChargingMode struct {
	value string
}

type BssParamChargingModeEnum struct {
	PRE_PAID  BssParamChargingMode
	POST_PAID BssParamChargingMode
}

func GetBssParamChargingModeEnum() BssParamChargingModeEnum {
	return BssParamChargingModeEnum{
		PRE_PAID: BssParamChargingMode{
			value: "prePaid",
		},
		POST_PAID: BssParamChargingMode{
			value: "postPaid",
		},
	}
}

func (c BssParamChargingMode) MarshalJSON() ([]byte, error) {
	return json.Marshal(c.value)
}

func (c *BssParamChargingMode) UnmarshalJSON(b []byte) error {
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

type BssParamIsAutoPay struct {
	value string
}

type BssParamIsAutoPayEnum struct {
	TRUE  BssParamIsAutoPay
	FALSE BssParamIsAutoPay
}

func GetBssParamIsAutoPayEnum() BssParamIsAutoPayEnum {
	return BssParamIsAutoPayEnum{
		TRUE: BssParamIsAutoPay{
			value: "true",
		},
		FALSE: BssParamIsAutoPay{
			value: "false",
		},
	}
}

func (c BssParamIsAutoPay) MarshalJSON() ([]byte, error) {
	return json.Marshal(c.value)
}

func (c *BssParamIsAutoPay) UnmarshalJSON(b []byte) error {
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

type BssParamPeriodType struct {
	value string
}

type BssParamPeriodTypeEnum struct {
	MONTH BssParamPeriodType
	YEAR  BssParamPeriodType
}

func GetBssParamPeriodTypeEnum() BssParamPeriodTypeEnum {
	return BssParamPeriodTypeEnum{
		MONTH: BssParamPeriodType{
			value: "month",
		},
		YEAR: BssParamPeriodType{
			value: "year",
		},
	}
}

func (c BssParamPeriodType) MarshalJSON() ([]byte, error) {
	return json.Marshal(c.value)
}

func (c *BssParamPeriodType) UnmarshalJSON(b []byte) error {
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
