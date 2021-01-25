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

// 创建弹性公网IP的extendparam字段数据结构说明
type ExtendParamEip struct {
	// 弹性公网IP的计费模式。若带宽计费类型为bandwidth，则支持prePaid和postPaid；若带宽计费类型为traffic，仅支持postPaid。取值范围：prePaid：预付费，即包年包月postPaid：后付费，即按需付费 说明：如果bandwidth对象中sharetype是WHOLE且id有值，弹性公网IP只能创建为按需付费的，故该参数传参“prePaid”无效。
	Chargingmode ExtendParamEipChargingmode `json:"chargingmode"`
}

func (o ExtendParamEip) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "ExtendParamEip struct{}"
	}

	return strings.Join([]string{"ExtendParamEip", string(data)}, " ")
}

type ExtendParamEipChargingmode struct {
	value string
}

type ExtendParamEipChargingmodeEnum struct {
	PRE_PAID  ExtendParamEipChargingmode
	POST_PAID ExtendParamEipChargingmode
}

func GetExtendParamEipChargingmodeEnum() ExtendParamEipChargingmodeEnum {
	return ExtendParamEipChargingmodeEnum{
		PRE_PAID: ExtendParamEipChargingmode{
			value: "prePaid",
		},
		POST_PAID: ExtendParamEipChargingmode{
			value: "postPaid",
		},
	}
}

func (c ExtendParamEipChargingmode) MarshalJSON() ([]byte, error) {
	return json.Marshal(c.value)
}

func (c *ExtendParamEipChargingmode) UnmarshalJSON(b []byte) error {
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
