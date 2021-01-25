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

// bandwidth字段数据结构说明
type BandWidth struct {
	// 带宽名称
	Name *string `json:"name,omitempty"`
	// 带宽的共享类型。共享类型枚举：PER，表示独享；WHOLE，表示共享
	Sharetype BandWidthSharetype `json:"sharetype"`
	// 共享带宽ID。创建WHOLE类型带宽的弹性公网IP时可以指定之前的共享带宽创建。共享带宽的使用限制请参见“共享带宽简介”。 说明：当创建WHOLE类型的带宽时，该字段必选。
	Id *string `json:"id,omitempty"`
	// 取值范围：默认5Mbit/s~2000Mbit/s（具体范围以各Region配置为准，请参见管理控制台对应页面显示）。功能说明：带宽大小。共享带宽的大小有最小值限制，默认为5M。 说明：如果传入的参数为小数（如10.2）或者字符类型（如10），会自动强制转换为整数。带宽小于300Mbit/s时，步长支持1Mbit/s；带宽为300Mbit/s~1000Mbit/s时，步长支持50Mbit/s；带宽为1000Mbit/s~2000Mbit/s时，步长支持1000Mbit/s。如果sharetype是PER，该参数必选；如果sharetype是WHOLE并且id有值，该参数会忽略。
	Size int32 `json:"size"`
	// 带宽的计费类型。取值为：traffic（按流量计费）、bandwidth（按带宽计费）未传该字段，表示按带宽计费。字段值为空，表示按带宽计费。 说明：如果sharetype是WHOLE并且id有值，仅支持按带宽计费，该参数会忽略。
	Chargemode *BandWidthChargemode `json:"chargemode,omitempty"`
}

func (o BandWidth) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "BandWidth struct{}"
	}

	return strings.Join([]string{"BandWidth", string(data)}, " ")
}

type BandWidthSharetype struct {
	value string
}

type BandWidthSharetypeEnum struct {
	PER   BandWidthSharetype
	WHOLE BandWidthSharetype
}

func GetBandWidthSharetypeEnum() BandWidthSharetypeEnum {
	return BandWidthSharetypeEnum{
		PER: BandWidthSharetype{
			value: "PER",
		},
		WHOLE: BandWidthSharetype{
			value: "WHOLE",
		},
	}
}

func (c BandWidthSharetype) MarshalJSON() ([]byte, error) {
	return json.Marshal(c.value)
}

func (c *BandWidthSharetype) UnmarshalJSON(b []byte) error {
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

type BandWidthChargemode struct {
	value string
}

type BandWidthChargemodeEnum struct {
	TRAFFIC   BandWidthChargemode
	BANDWIDTH BandWidthChargemode
}

func GetBandWidthChargemodeEnum() BandWidthChargemodeEnum {
	return BandWidthChargemodeEnum{
		TRAFFIC: BandWidthChargemode{
			value: "traffic",
		},
		BANDWIDTH: BandWidthChargemode{
			value: "bandwidth",
		},
	}
}

func (c BandWidthChargemode) MarshalJSON() ([]byte, error) {
	return json.Marshal(c.value)
}

func (c *BandWidthChargemode) UnmarshalJSON(b []byte) error {
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
