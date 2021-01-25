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
type BatchBandwidthResp struct {
	// 功能说明：带宽类型，共享带宽默认为share。  取值范围：share，bgp，telcom，sbgp等。  share：共享带宽  bgp：动态bgp  telcom ：联通  sbgp：静态bgp
	BandwidthType *string `json:"bandwidth_type,omitempty"`
	// 功能说明：账单信息  如果billing_info不为空，说明是包周期的带宽
	BillingInfo *string `json:"billing_info,omitempty"`
	// 功能说明：按流量计费,按带宽计费还是按增强型95计费。  取值范围：bandwidth，traffic，95peak_plus(按增强型95计费)不返回或者为空时表示是bandwidth。  约束：只有共享带宽支持95peak_plus（按增强型95计费），按增强型95计费时需要指定保底百分比，默认是20%。
	ChargeMode *BatchBandwidthRespChargeMode `json:"charge_mode,omitempty"`
	// 功能说明：带宽唯一标识
	Id *string `json:"id,omitempty"`
	// 功能说明：带宽名称  取值范围：1-64个字符，支持数字、字母、中文、_(下划线)、-（中划线）、.（点）
	Name *string `json:"name,omitempty"`
	// 功能说明：带宽对应的弹性公网IP信息  约束：WHOLE类型的带宽支持多个弹性公网IP，PER类型的带宽只能对应一个弹性公网IP
	PublicipInfo *[]PublicipInfoResp `json:"publicip_info,omitempty"`
	// 功能说明：带宽类型，标识是否是共享带宽  取值范围：WHOLE，PER  WHOLE表示共享带宽；PER，表示独享带宽
	ShareType *BatchBandwidthRespShareType `json:"share_type,omitempty"`
	// 功能说明：带宽大小  取值范围：默认5Mbit/s~2000Mbit/s（具体范围以各区域配置为准，请参见控制台对应页面显示）。
	Size *int32 `json:"size,omitempty"`
	// 功能说明：用户所属租户ID
	TenantId *string `json:"tenant_id,omitempty"`
	// 功能说明：带宽的状态  取值范围：  FREEZED：冻结  NORMAL：正常
	Status *BatchBandwidthRespStatus `json:"status,omitempty"`
}

func (o BatchBandwidthResp) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "BatchBandwidthResp struct{}"
	}

	return strings.Join([]string{"BatchBandwidthResp", string(data)}, " ")
}

type BatchBandwidthRespChargeMode struct {
	value string
}

type BatchBandwidthRespChargeModeEnum struct {
	BANDWIDTH     BatchBandwidthRespChargeMode
	TRAFFIC       BatchBandwidthRespChargeMode
	E_95PEAK_PLUS BatchBandwidthRespChargeMode
}

func GetBatchBandwidthRespChargeModeEnum() BatchBandwidthRespChargeModeEnum {
	return BatchBandwidthRespChargeModeEnum{
		BANDWIDTH: BatchBandwidthRespChargeMode{
			value: "bandwidth",
		},
		TRAFFIC: BatchBandwidthRespChargeMode{
			value: "traffic",
		},
		E_95PEAK_PLUS: BatchBandwidthRespChargeMode{
			value: "95peak_plus",
		},
	}
}

func (c BatchBandwidthRespChargeMode) MarshalJSON() ([]byte, error) {
	return json.Marshal(c.value)
}

func (c *BatchBandwidthRespChargeMode) UnmarshalJSON(b []byte) error {
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

type BatchBandwidthRespShareType struct {
	value string
}

type BatchBandwidthRespShareTypeEnum struct {
	WHOLE BatchBandwidthRespShareType
	PER   BatchBandwidthRespShareType
}

func GetBatchBandwidthRespShareTypeEnum() BatchBandwidthRespShareTypeEnum {
	return BatchBandwidthRespShareTypeEnum{
		WHOLE: BatchBandwidthRespShareType{
			value: "WHOLE",
		},
		PER: BatchBandwidthRespShareType{
			value: "PER",
		},
	}
}

func (c BatchBandwidthRespShareType) MarshalJSON() ([]byte, error) {
	return json.Marshal(c.value)
}

func (c *BatchBandwidthRespShareType) UnmarshalJSON(b []byte) error {
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

type BatchBandwidthRespStatus struct {
	value string
}

type BatchBandwidthRespStatusEnum struct {
	FREEZED BatchBandwidthRespStatus
	NORMAL  BatchBandwidthRespStatus
}

func GetBatchBandwidthRespStatusEnum() BatchBandwidthRespStatusEnum {
	return BatchBandwidthRespStatusEnum{
		FREEZED: BatchBandwidthRespStatus{
			value: "FREEZED",
		},
		NORMAL: BatchBandwidthRespStatus{
			value: "NORMAL",
		},
	}
}

func (c BatchBandwidthRespStatus) MarshalJSON() ([]byte, error) {
	return json.Marshal(c.value)
}

func (c *BatchBandwidthRespStatus) UnmarshalJSON(b []byte) error {
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
