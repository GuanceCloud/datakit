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

// 弹性公网IP对象
type PublicipCreateResp struct {
	// 带宽大小，单位为Mbit/s。
	BandwidthSize *int32 `json:"bandwidth_size,omitempty"`
	// 弹性公网IP申请时间（UTC时间）
	CreateTime *sdktime.SdkTime `json:"create_time,omitempty"`
	// 弹性公网IP唯一标识
	Id *string `json:"id,omitempty"`
	// IPv4时是申请到的弹性公网IP地址，IPv6时是IPv6地址对应的IPv4地址
	PublicIpAddress *string `json:"public_ip_address,omitempty"`
	// 功能说明：弹性公网IP的状态  取值范围：冻结FREEZED，绑定失败BIND_ERROR，绑定中BINDING，释放中PENDING_DELETE， 创建中PENDING_CREATE，创建中NOTIFYING，释放中NOTIFY_DELETE，更新中PENDING_UPDATE， 未绑定DOWN ，绑定ACTIVE，绑定ELB，绑定VPN，失败ERROR。
	Status *PublicipCreateRespStatus `json:"status,omitempty"`
	// 项目ID
	TenantId *string `json:"tenant_id,omitempty"`
	// 功能说明：弹性IP弹性公网IP的类型  取值范围：5_telcom（电信），5_union（联通），5_bgp（全动态BGP），5_sbgp（静态BGP），5_ipv6  东北-大连：5_telcom、5_union  华南-广州：5_bgp、5_sbgp  华东-上海二：5_bgp、5_sbgp  华北-北京一：5_bgp、5_sbgp、5_ipv6  亚太-香港：5_bgp  亚太-曼谷：5_bgp  亚太-新加坡：5_bgp  非洲-约翰内斯堡：5_bgp  西南-贵阳一：5_bgp、5_sbgp  华北-北京四：5_bgp、5_sbgp  约束：必须是系统具体支持的类型publicip_id为IPv4端口，所以\"publicip_type\"字段未给定时，默认为5_bgp。
	Type *string `json:"type,omitempty"`
	// IPv4时无此字段，IPv6时为申请到的弹性公网IP地址
	PublicIpv6Address *string `json:"public_ipv6_address,omitempty"`
	// IP版本信息，取值范围是4和6
	IpVersion *PublicipCreateRespIpVersion `json:"ip_version,omitempty"`
	// 企业项目ID。最大长度36字节，带“-”连字符的UUID格式，或者是字符串“0”。  创建弹性公网IP时，给弹性公网IP绑定企业项目ID。
	EnterpriseProjectId *string `json:"enterprise_project_id,omitempty"`
}

func (o PublicipCreateResp) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "PublicipCreateResp struct{}"
	}

	return strings.Join([]string{"PublicipCreateResp", string(data)}, " ")
}

type PublicipCreateRespStatus struct {
	value string
}

type PublicipCreateRespStatusEnum struct {
	FREEZED        PublicipCreateRespStatus
	BIND_ERROR     PublicipCreateRespStatus
	BINDING        PublicipCreateRespStatus
	PENDING_DELETE PublicipCreateRespStatus
	PENDING_CREATE PublicipCreateRespStatus
	NOTIFYING      PublicipCreateRespStatus
	NOTIFY_DELETE  PublicipCreateRespStatus
	PENDING_UPDATE PublicipCreateRespStatus
	DOWN           PublicipCreateRespStatus
	ACTIVE         PublicipCreateRespStatus
	ELB            PublicipCreateRespStatus
	ERROR          PublicipCreateRespStatus
	VPN            PublicipCreateRespStatus
}

func GetPublicipCreateRespStatusEnum() PublicipCreateRespStatusEnum {
	return PublicipCreateRespStatusEnum{
		FREEZED: PublicipCreateRespStatus{
			value: "FREEZED",
		},
		BIND_ERROR: PublicipCreateRespStatus{
			value: "BIND_ERROR",
		},
		BINDING: PublicipCreateRespStatus{
			value: "BINDING",
		},
		PENDING_DELETE: PublicipCreateRespStatus{
			value: "PENDING_DELETE",
		},
		PENDING_CREATE: PublicipCreateRespStatus{
			value: "PENDING_CREATE",
		},
		NOTIFYING: PublicipCreateRespStatus{
			value: "NOTIFYING",
		},
		NOTIFY_DELETE: PublicipCreateRespStatus{
			value: "NOTIFY_DELETE",
		},
		PENDING_UPDATE: PublicipCreateRespStatus{
			value: "PENDING_UPDATE",
		},
		DOWN: PublicipCreateRespStatus{
			value: "DOWN",
		},
		ACTIVE: PublicipCreateRespStatus{
			value: "ACTIVE",
		},
		ELB: PublicipCreateRespStatus{
			value: "ELB",
		},
		ERROR: PublicipCreateRespStatus{
			value: "ERROR",
		},
		VPN: PublicipCreateRespStatus{
			value: "VPN",
		},
	}
}

func (c PublicipCreateRespStatus) MarshalJSON() ([]byte, error) {
	return json.Marshal(c.value)
}

func (c *PublicipCreateRespStatus) UnmarshalJSON(b []byte) error {
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

type PublicipCreateRespIpVersion struct {
	value int32
}

type PublicipCreateRespIpVersionEnum struct {
	E_4 PublicipCreateRespIpVersion
	E_6 PublicipCreateRespIpVersion
}

func GetPublicipCreateRespIpVersionEnum() PublicipCreateRespIpVersionEnum {
	return PublicipCreateRespIpVersionEnum{
		E_4: PublicipCreateRespIpVersion{
			value: 4,
		}, E_6: PublicipCreateRespIpVersion{
			value: 6,
		},
	}
}

func (c PublicipCreateRespIpVersion) MarshalJSON() ([]byte, error) {
	return json.Marshal(c.value)
}

func (c *PublicipCreateRespIpVersion) UnmarshalJSON(b []byte) error {
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
