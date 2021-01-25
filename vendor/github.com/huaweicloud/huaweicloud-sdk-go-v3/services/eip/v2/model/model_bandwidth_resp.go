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
type BandwidthResp struct {
	// 功能说明：带宽类型，共享带宽默认为share。  取值范围：share，bgp，telcom，sbgp等。  share：共享带宽  bgp：动态bgp  telcom ：联通  sbgp：静态bgp
	BandwidthType *string `json:"bandwidth_type,omitempty"`
	// 功能说明：账单信息  如果billinginfo不为空，说明是包周期的带宽
	BillingInfo *string `json:"billing_info,omitempty"`
	// 功能说明：按流量计费,按带宽计费还是按增强型95计费。  取值范围：bandwidth，traffic，95peak_plus(按增强型95计费)不返回或者为空时表示是bandwidth。  约束：只有共享带宽支持95peak_plus（按增强型95计费），按增强型95计费时需要指定保底百分比，默认是20%。
	ChargeMode *BandwidthRespChargeMode `json:"charge_mode,omitempty"`
	// 功能说明：带宽唯一标识
	Id *string `json:"id,omitempty"`
	// 功能说明：带宽名称  取值范围：1-64个字符，支持数字、字母、中文、_(下划线)、-（中划线）、.（点）
	Name *string `json:"name,omitempty"`
	// 功能说明：带宽对应的弹性公网IP信息  约束：WHOLE类型的带宽支持多个弹性公网IP，PER类型的带宽只能对应一个弹性公网IP
	PublicipInfo *[]PublicipInfoResp `json:"publicip_info,omitempty"`
	// 功能说明：带宽类型，标识是否是共享带宽  取值范围：WHOLE，PER  WHOLE表示共享带宽；PER，表示独享带宽
	ShareType *BandwidthRespShareType `json:"share_type,omitempty"`
	// 功能说明：带宽大小  取值范围：默认5Mbit/s~2000Mbit/s（具体范围以各区域配置为准，请参见控制台对应页面显示）。
	Size *int32 `json:"size,omitempty"`
	// 功能说明：用户所属租户ID
	TenantId *string `json:"tenant_id,omitempty"`
	// 企业项目ID。最大长度36字节，带“-”连字符的UUID格式，或者是字符串“0”。  创建带宽时，给带宽绑定企业项目ID。
	EnterpriseProjectId *string `json:"enterprise_project_id,omitempty"`
	// 功能说明：带宽的状态  取值范围：  FREEZED：冻结  NORMAL：正常
	Status *BandwidthRespStatus `json:"status,omitempty"`
	// 功能说明：是否开启企业级qos，仅共享带宽支持开启。（该字段仅在上海1局点返回）
	EnableBandwidthRules *bool `json:"enable_bandwidth_rules,omitempty"`
	// 功能说明：带宽支持的最大分组规则数。（该字段仅在上海1局点返回）
	RuleQuota *int32 `json:"rule_quota,omitempty"`
	// 功能说明：带宽规则对象（该字段仅在上海1局点返回）
	BandwidthRules *[]BandWidthRules `json:"bandwidth_rules,omitempty"`
	// 功能说明：资源创建时间，UTC时间  格式： yyyy-MM-ddTHH:mm:ss
	CreatedAt *string `json:"created_at,omitempty"`
	// 功能说明：资源更新时间，UTC时间  格式： yyyy-MM-ddTHH:mm:ss
	UpdatedAt *string `json:"updated_at,omitempty"`
}

func (o BandwidthResp) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "BandwidthResp struct{}"
	}

	return strings.Join([]string{"BandwidthResp", string(data)}, " ")
}

type BandwidthRespChargeMode struct {
	value string
}

type BandwidthRespChargeModeEnum struct {
	BANDWIDTH     BandwidthRespChargeMode
	TRAFFIC       BandwidthRespChargeMode
	E_95PEAK_PLUS BandwidthRespChargeMode
}

func GetBandwidthRespChargeModeEnum() BandwidthRespChargeModeEnum {
	return BandwidthRespChargeModeEnum{
		BANDWIDTH: BandwidthRespChargeMode{
			value: "bandwidth",
		},
		TRAFFIC: BandwidthRespChargeMode{
			value: "traffic",
		},
		E_95PEAK_PLUS: BandwidthRespChargeMode{
			value: "95peak_plus",
		},
	}
}

func (c BandwidthRespChargeMode) MarshalJSON() ([]byte, error) {
	return json.Marshal(c.value)
}

func (c *BandwidthRespChargeMode) UnmarshalJSON(b []byte) error {
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

type BandwidthRespShareType struct {
	value string
}

type BandwidthRespShareTypeEnum struct {
	WHOLE BandwidthRespShareType
	PER   BandwidthRespShareType
}

func GetBandwidthRespShareTypeEnum() BandwidthRespShareTypeEnum {
	return BandwidthRespShareTypeEnum{
		WHOLE: BandwidthRespShareType{
			value: "WHOLE",
		},
		PER: BandwidthRespShareType{
			value: "PER",
		},
	}
}

func (c BandwidthRespShareType) MarshalJSON() ([]byte, error) {
	return json.Marshal(c.value)
}

func (c *BandwidthRespShareType) UnmarshalJSON(b []byte) error {
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

type BandwidthRespStatus struct {
	value string
}

type BandwidthRespStatusEnum struct {
	FREEZED BandwidthRespStatus
	NORMAL  BandwidthRespStatus
}

func GetBandwidthRespStatusEnum() BandwidthRespStatusEnum {
	return BandwidthRespStatusEnum{
		FREEZED: BandwidthRespStatus{
			value: "FREEZED",
		},
		NORMAL: BandwidthRespStatus{
			value: "NORMAL",
		},
	}
}

func (c BandwidthRespStatus) MarshalJSON() ([]byte, error) {
	return json.Marshal(c.value)
}

func (c *BandwidthRespStatus) UnmarshalJSON(b []byte) error {
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
