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

// 创建带宽的请求体
type CreateSharedBandwidthOption struct {
	// 企业项目ID。最大长度36字节，带“-”连字符的UUID格式，或者是字符串“0”。  创建共享带宽时，给共享带宽绑定企业项目ID。
	EnterpriseProjectId *string `json:"enterprise_project_id,omitempty"`
	// 取值范围：1-64，支持数字、字母、中文、_(下划线)、-（中划线）、.（点）  功能说明：带宽名称
	Name string `json:"name"`
	// 功能说明：带宽大小。共享带宽的大小有最小值限制，默认为5M，可能因局点不同而不同。  取值范围：默认5Mbit/s~2000Mbit/s（具体范围以各区域配置为准，请参见控制台对应页面显示）。  如果传入的参数为小数（如 10.2）或者字符类型（如“10”），会自动强制转换为整数。  调整带宽时的最小单位会根据带宽范围不同存在差异。  小于等于300Mbit/s：默认最小单位为1Mbit/s。  300Mbit/s~1000Mbit/s：默认最小单位为50Mbit/s。  大于1000Mbit/s：默认最小单位为500Mbit/s。
	Size int32 `json:"size"`
	// 功能说明：按带宽计费还是按增强型95计费。  取值范围：bandwidth，95peak_plus(按增强型95计费)不返回或者为空时表示是bandwidth。  约束：只有共享带宽支持95peak_plus（按增强型95计费），按增强型95计费时需要指定保底百分比，默认是20%。
	ChargeMode *CreateSharedBandwidthOptionChargeMode `json:"charge_mode,omitempty"`
}

func (o CreateSharedBandwidthOption) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "CreateSharedBandwidthOption struct{}"
	}

	return strings.Join([]string{"CreateSharedBandwidthOption", string(data)}, " ")
}

type CreateSharedBandwidthOptionChargeMode struct {
	value string
}

type CreateSharedBandwidthOptionChargeModeEnum struct {
	BANDWIDTH     CreateSharedBandwidthOptionChargeMode
	E_95PEAK_PLUS CreateSharedBandwidthOptionChargeMode
}

func GetCreateSharedBandwidthOptionChargeModeEnum() CreateSharedBandwidthOptionChargeModeEnum {
	return CreateSharedBandwidthOptionChargeModeEnum{
		BANDWIDTH: CreateSharedBandwidthOptionChargeMode{
			value: "bandwidth",
		},
		E_95PEAK_PLUS: CreateSharedBandwidthOptionChargeMode{
			value: "95peak_plus",
		},
	}
}

func (c CreateSharedBandwidthOptionChargeMode) MarshalJSON() ([]byte, error) {
	return json.Marshal(c.value)
}

func (c *CreateSharedBandwidthOptionChargeMode) UnmarshalJSON(b []byte) error {
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
