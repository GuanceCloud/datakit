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

// 更新带宽对象(三个参数不能全为空)
type UpdateBandwidthOption struct {
	// 取值范围：1-64，支持数字、字母、中文、_(下划线)、-（中划线），为空表示不修改名称  功能说明：带宽名称  约束：name、size必须有一个参数有值
	Name *string `json:"name,omitempty"`
	// 取值范围：默认1Mbit/s~2000Mbit/s（具体范围以各区域配置为准，请参见控制台对应页面显示），不带此参数时表示不修改大小  功能说明：带宽大小，单位Mbit/s。  约束：name、size必须有一个参数有值  如果传入的参数为小数（如 10.2）或者字符类型（如“10”），会自动强制转换为整数。  约束：name、size必须要有一个参数有值。  调整带宽时的最小单位会根据带宽范围不同存在差异:  小于等于300Mbit/s：默认最小单位为1Mbit/s。  300Mbit/s~1000Mbit/s：默认最小单位为50Mbit/s。  大于1000Mbit/s：默认最小单位为500Mbit/s。
	Size *int32 `json:"size,omitempty"`
	// 功能说明：按流量计费,按带宽计费还是按增强型95计费。  取值范围：bandwidth，traffic，95peak_plus(按增强型95计费)不返回或者为空时表示是bandwidth。  约束：只有共享带宽支持95peak_plus（按增强型95计费），按增强型95计费时需要指定保底百分比，默认是20%。
	ChargeMode *UpdateBandwidthOptionChargeMode `json:"charge_mode,omitempty"`
}

func (o UpdateBandwidthOption) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "UpdateBandwidthOption struct{}"
	}

	return strings.Join([]string{"UpdateBandwidthOption", string(data)}, " ")
}

type UpdateBandwidthOptionChargeMode struct {
	value string
}

type UpdateBandwidthOptionChargeModeEnum struct {
	BANDWIDTH     UpdateBandwidthOptionChargeMode
	TRAFFIC       UpdateBandwidthOptionChargeMode
	E_95PEAK_PLUS UpdateBandwidthOptionChargeMode
}

func GetUpdateBandwidthOptionChargeModeEnum() UpdateBandwidthOptionChargeModeEnum {
	return UpdateBandwidthOptionChargeModeEnum{
		BANDWIDTH: UpdateBandwidthOptionChargeMode{
			value: "bandwidth",
		},
		TRAFFIC: UpdateBandwidthOptionChargeMode{
			value: "traffic",
		},
		E_95PEAK_PLUS: UpdateBandwidthOptionChargeMode{
			value: "95peak_plus",
		},
	}
}

func (c UpdateBandwidthOptionChargeMode) MarshalJSON() ([]byte, error) {
	return json.Marshal(c.value)
}

func (c *UpdateBandwidthOptionChargeMode) UnmarshalJSON(b []byte) error {
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
