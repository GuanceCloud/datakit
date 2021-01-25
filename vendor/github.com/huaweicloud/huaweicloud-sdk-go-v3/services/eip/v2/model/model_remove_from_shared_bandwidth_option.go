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

// 带宽对象
type RemoveFromSharedBandwidthOption struct {
	// 弹性公网IP从共享带宽移除后，会为此弹性公网IP创建独占带宽进行计费。  此参数表示弹性公网IP从共享带宽移除后，使用的独占带宽的计费类型。（bandwidth/traffic）
	ChargeMode RemoveFromSharedBandwidthOptionChargeMode `json:"charge_mode"`
	// 功能说明：要从共享带宽中移除的弹性公网IP或者IPv6端口信息  约束：WHOLE类型的带宽支持多个弹性公网IP或者IPv6端口，跟租户的配额相关，默认一个共享带宽的配额为20
	PublicipInfo []RemovePublicipInfo `json:"publicip_info"`
	// 弹性公网IP从共享带宽移除后，会为此弹性公网IP创建独占带宽进行计费。  此参数表示弹性公网IP从共享带宽移除后，使用的独占带宽的带宽大小。（M）取值范围：默认为1~2000Mbit/s. 可能因为局点配置不同而不同。也跟带宽的计费模式（bandwidth/traffic）相关。
	Size int32 `json:"size"`
}

func (o RemoveFromSharedBandwidthOption) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "RemoveFromSharedBandwidthOption struct{}"
	}

	return strings.Join([]string{"RemoveFromSharedBandwidthOption", string(data)}, " ")
}

type RemoveFromSharedBandwidthOptionChargeMode struct {
	value string
}

type RemoveFromSharedBandwidthOptionChargeModeEnum struct {
	BANDWIDTH RemoveFromSharedBandwidthOptionChargeMode
	TRAFFIC   RemoveFromSharedBandwidthOptionChargeMode
}

func GetRemoveFromSharedBandwidthOptionChargeModeEnum() RemoveFromSharedBandwidthOptionChargeModeEnum {
	return RemoveFromSharedBandwidthOptionChargeModeEnum{
		BANDWIDTH: RemoveFromSharedBandwidthOptionChargeMode{
			value: "bandwidth",
		},
		TRAFFIC: RemoveFromSharedBandwidthOptionChargeMode{
			value: "traffic",
		},
	}
}

func (c RemoveFromSharedBandwidthOptionChargeMode) MarshalJSON() ([]byte, error) {
	return json.Marshal(c.value)
}

func (c *RemoveFromSharedBandwidthOptionChargeMode) UnmarshalJSON(b []byte) error {
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
