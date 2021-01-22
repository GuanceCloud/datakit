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
	"github.com/huaweicloud/huaweicloud-sdk-go-v3/core/sdktime"
	"strings"
)

// 公网IP字段信息
type PublicipShowResp struct {
	// 功能说明：弹性公网IP唯一标识
	Id *string `json:"id,omitempty"`
	// 功能说明：项目ID
	ProjectId *string `json:"project_id,omitempty"`
	// 功能说明：IP版本信息 取值范围：4表示公网IP地址为public_ip_address地址;6表示公网IP地址为public_ipv6_address地址\"
	IpVersion *PublicipShowRespIpVersion `json:"ip_version,omitempty"`
	// 功能说明：弹性公网IP或者IPv6端口的地址
	PublicIpAddress *string `json:"public_ip_address,omitempty"`
	// 功能说明：IPv4时无此字段，IPv6时为申请到的弹性公网IP地址
	PublicIpv6Address *string `json:"public_ipv6_address,omitempty"`
	// 废弃，功能由publicip_pool_name继承，默认不显示。功能说明：弹性公网IP的网络类型
	NetworkType *string `json:"network_type,omitempty"`
	// 功能说明：弹性公网IP的状态  取值范围：冻结FREEZED，绑定失败BIND_ERROR，绑定中BINDING，释放中PENDING_DELETE， 创建中PENDING_CREATE，创建中NOTIFYING，释放中NOTIFY_DELETE，更新中PENDING_UPDATE， 未绑定DOWN ，绑定ACTIVE，绑定ELB，绑定VPN，失败ERROR。
	Status *PublicipShowRespStatus `json:"status,omitempty"`
	// 功能说明：弹性公网IP描述信息 约束：用户以自定义方式标识资源，系统不感知
	Description *string `json:"description,omitempty"`
	// 公网EIP的出口名称,多出口特性开关打开展示
	GroupName *string `json:"group_name,omitempty"`
	// 功能说明：资源创建UTC时间 格式:yyyy-MM-ddTHH:mm:ssZ
	CreatedAt *sdktime.SdkTime `json:"created_at,omitempty"`
	// 功能说明：资源更新UTC时间 格式:yyyy-MM-ddTHH:mm:ssZ
	UpdatedAt *sdktime.SdkTime `json:"updated_at,omitempty"`
	// 功能说明：弹性公网IP类型
	Type      *PublicipShowRespType  `json:"type,omitempty"`
	Vnic      *VnicInfo              `json:"vnic,omitempty"`
	Bandwidth *PublicipBandwidthInfo `json:"bandwidth,omitempty"`
	// 功能说明：企业项目ID。最大长度36字节,带“-”连字符的UUID格式,或者是字符串“0”。创建弹性公网IP时,给弹性公网IP绑定企业项目ID。
	EnterpriseProjectId *string `json:"enterprise_project_id,omitempty"`
	// 功能说明：公网IP的订单信息 约束：包周期才会有订单信息，按需资源此字段为空
	BillingInfo *string `json:"billing_info,omitempty"`
	// 功能说明：记录公网IP当前的冻结状态 约束：metadata类型，标识欠费冻结、公安冻结 取值范围：police，locked
	LockStatus *string `json:"lock_status,omitempty"`
	// 功能说明：公网IP绑定的实例类型 取值范围：PORT、NATGW、ELB、ELBV1、VPN、null
	AssociateInstanceType *PublicipShowRespAssociateInstanceType `json:"associate_instance_type,omitempty"`
	// 功能说明：公网IP绑定的实例ID
	AssociateInstanceId *string `json:"associate_instance_id,omitempty"`
	// 功能说明：公网IP所属网络的ID。publicip_pool_name对应的网络ID
	PublicipPoolId *string `json:"publicip_pool_id,omitempty"`
	// 功能说明：弹性公网IP的网络类型, 包括公共池类型，如5_bgp/5_sbgp...，和用户购买的专属池。 专属池见publcip_pool相关接口
	PublicipPoolName *string `json:"publicip_pool_name,omitempty"`
	// 功能说明：弹性公网IP名称
	Alias   *string      `json:"alias,omitempty"`
	Profile *ProfileInfo `json:"profile,omitempty"`
	// 默认不显示。该字段仅仅用于表示eip的bgp类型是否是真实的静态sbgp * 1. 如果为true，则该eip可以切换bgp类型 * 2. 如果为false，则该eip不可以切换bgp类型
	FakeNetworkType *bool `json:"fake_network_type,omitempty"`
	// 默认不显示。用户标签
	Tags *[]TagsInfo `json:"tags,omitempty"`
	// 默认不显示。记录实例的更上一层归属。例如associate_instance_type为PORT，此字段记录PORT的device_id和device_owner信息。仅有限场景记录。
	AssociateInstanceMetadata *string `json:"associate_instance_metadata,omitempty"`
}

func (o PublicipShowResp) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "PublicipShowResp struct{}"
	}

	return strings.Join([]string{"PublicipShowResp", string(data)}, " ")
}

type PublicipShowRespIpVersion struct {
	value int32
}

type PublicipShowRespIpVersionEnum struct {
	E_4 PublicipShowRespIpVersion
	E_6 PublicipShowRespIpVersion
}

func GetPublicipShowRespIpVersionEnum() PublicipShowRespIpVersionEnum {
	return PublicipShowRespIpVersionEnum{
		E_4: PublicipShowRespIpVersion{
			value: 4,
		}, E_6: PublicipShowRespIpVersion{
			value: 6,
		},
	}
}

func (c PublicipShowRespIpVersion) MarshalJSON() ([]byte, error) {
	return json.Marshal(c.value)
}

func (c *PublicipShowRespIpVersion) UnmarshalJSON(b []byte) error {
	myConverter := converter.StringConverterFactory("int32")
	if myConverter != nil {
		val, err := myConverter.CovertStringToInterface(strings.Trim(string(b[:]), "\""))
		if err == nil {
			c.value = val.(int32)
			return nil
		}
		return err
	} else {
		return errors.New("convert enum data to int32 error")
	}
}

type PublicipShowRespStatus struct {
	value string
}

type PublicipShowRespStatusEnum struct {
	FREEZED        PublicipShowRespStatus
	BIND_ERROR     PublicipShowRespStatus
	BINDING        PublicipShowRespStatus
	PENDING_DELETE PublicipShowRespStatus
	PENDING_CREATE PublicipShowRespStatus
	NOTIFYING      PublicipShowRespStatus
	NOTIFY_DELETE  PublicipShowRespStatus
	PENDING_UPDATE PublicipShowRespStatus
	DOWN           PublicipShowRespStatus
	ACTIVE         PublicipShowRespStatus
	ELB            PublicipShowRespStatus
	ERROR          PublicipShowRespStatus
	VPN            PublicipShowRespStatus
}

func GetPublicipShowRespStatusEnum() PublicipShowRespStatusEnum {
	return PublicipShowRespStatusEnum{
		FREEZED: PublicipShowRespStatus{
			value: "FREEZED",
		},
		BIND_ERROR: PublicipShowRespStatus{
			value: "BIND_ERROR",
		},
		BINDING: PublicipShowRespStatus{
			value: "BINDING",
		},
		PENDING_DELETE: PublicipShowRespStatus{
			value: "PENDING_DELETE",
		},
		PENDING_CREATE: PublicipShowRespStatus{
			value: "PENDING_CREATE",
		},
		NOTIFYING: PublicipShowRespStatus{
			value: "NOTIFYING",
		},
		NOTIFY_DELETE: PublicipShowRespStatus{
			value: "NOTIFY_DELETE",
		},
		PENDING_UPDATE: PublicipShowRespStatus{
			value: "PENDING_UPDATE",
		},
		DOWN: PublicipShowRespStatus{
			value: "DOWN",
		},
		ACTIVE: PublicipShowRespStatus{
			value: "ACTIVE",
		},
		ELB: PublicipShowRespStatus{
			value: "ELB",
		},
		ERROR: PublicipShowRespStatus{
			value: "ERROR",
		},
		VPN: PublicipShowRespStatus{
			value: "VPN",
		},
	}
}

func (c PublicipShowRespStatus) MarshalJSON() ([]byte, error) {
	return json.Marshal(c.value)
}

func (c *PublicipShowRespStatus) UnmarshalJSON(b []byte) error {
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

type PublicipShowRespType struct {
	value string
}

type PublicipShowRespTypeEnum struct {
	EIP              PublicipShowRespType
	DUALSTACK        PublicipShowRespType
	DUALSTACK_SUBNET PublicipShowRespType
}

func GetPublicipShowRespTypeEnum() PublicipShowRespTypeEnum {
	return PublicipShowRespTypeEnum{
		EIP: PublicipShowRespType{
			value: "EIP",
		},
		DUALSTACK: PublicipShowRespType{
			value: "DUALSTACK",
		},
		DUALSTACK_SUBNET: PublicipShowRespType{
			value: "DUALSTACK_SUBNET",
		},
	}
}

func (c PublicipShowRespType) MarshalJSON() ([]byte, error) {
	return json.Marshal(c.value)
}

func (c *PublicipShowRespType) UnmarshalJSON(b []byte) error {
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

type PublicipShowRespAssociateInstanceType struct {
	value string
}

type PublicipShowRespAssociateInstanceTypeEnum struct {
	PORT  PublicipShowRespAssociateInstanceType
	NATGW PublicipShowRespAssociateInstanceType
	ELB   PublicipShowRespAssociateInstanceType
	ELBV1 PublicipShowRespAssociateInstanceType
	VPN   PublicipShowRespAssociateInstanceType
	NULL  PublicipShowRespAssociateInstanceType
}

func GetPublicipShowRespAssociateInstanceTypeEnum() PublicipShowRespAssociateInstanceTypeEnum {
	return PublicipShowRespAssociateInstanceTypeEnum{
		PORT: PublicipShowRespAssociateInstanceType{
			value: "PORT",
		},
		NATGW: PublicipShowRespAssociateInstanceType{
			value: "NATGW",
		},
		ELB: PublicipShowRespAssociateInstanceType{
			value: "ELB",
		},
		ELBV1: PublicipShowRespAssociateInstanceType{
			value: "ELBV1",
		},
		VPN: PublicipShowRespAssociateInstanceType{
			value: "VPN",
		},
		NULL: PublicipShowRespAssociateInstanceType{
			value: "null",
		},
	}
}

func (c PublicipShowRespAssociateInstanceType) MarshalJSON() ([]byte, error) {
	return json.Marshal(c.value)
}

func (c *PublicipShowRespAssociateInstanceType) UnmarshalJSON(b []byte) error {
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
