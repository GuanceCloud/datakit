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

// interfaceAttachments字段数据结构说明
type InterfaceAttachments struct {
	// 网卡端口状态。取值为：ACTIVE、BUILD、DOWN
	PortState *InterfaceAttachmentsPortState `json:"port_state,omitempty"`
	// 网卡私网IP信息列表，详情请参见表3 fixed_ips字段数据结构说明。
	FixedIps *[]FixedIps `json:"fixed_ips,omitempty"`
	// 网卡端口所属子网的网络ID（network_id）。
	NetId *string `json:"net_id,omitempty"`
	// 网卡端口ID。
	PortId *string `json:"port_id,omitempty"`
	// 网卡Mac地址信息
	MacAddr *string `json:"mac_addr,omitempty"`
	// 从guest os中，网卡的驱动类型
	DriverMode *string `json:"driver_mode,omitempty"`
	// 弹性网卡在Linux GuestOS里的BDF号
	PciAddress *string `json:"pci_address,omitempty"`
}

func (o InterfaceAttachments) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "InterfaceAttachments struct{}"
	}

	return strings.Join([]string{"InterfaceAttachments", string(data)}, " ")
}

type InterfaceAttachmentsPortState struct {
	value string
}

type InterfaceAttachmentsPortStateEnum struct {
	ACTIVE InterfaceAttachmentsPortState
	BUILD  InterfaceAttachmentsPortState
	DOWN   InterfaceAttachmentsPortState
}

func GetInterfaceAttachmentsPortStateEnum() InterfaceAttachmentsPortStateEnum {
	return InterfaceAttachmentsPortStateEnum{
		ACTIVE: InterfaceAttachmentsPortState{
			value: "ACTIVE",
		},
		BUILD: InterfaceAttachmentsPortState{
			value: "BUILD",
		},
		DOWN: InterfaceAttachmentsPortState{
			value: "DOWN",
		},
	}
}

func (c InterfaceAttachmentsPortState) MarshalJSON() ([]byte, error) {
	return json.Marshal(c.value)
}

func (c *InterfaceAttachmentsPortState) UnmarshalJSON(b []byte) error {
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
