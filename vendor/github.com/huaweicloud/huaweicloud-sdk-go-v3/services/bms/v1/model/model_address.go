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

// 虚拟私有云ID字段数据结构说明
type Address struct {
	// IP地址版本。4：代表IPv4。6：代表IPv6。
	Version *AddressVersion `json:"version,omitempty"`
	// IP地址
	Addr *string `json:"addr,omitempty"`
	// IP地址类型。fixed：代表私有IP地址。floating：代表浮动IP地址。
	OSEXTIPStype *AddressOSEXTIPStype `json:"OS-EXT-IPS:type,omitempty"`
	// MAC地址
	OSEXTIPSMACmacAddr *string `json:"OS-EXT-IPS-MAC:mac_addr,omitempty"`
	// IP地址对应的端口ID
	OSEXTIPSportId *string `json:"OS-EXT-IPS:port_id,omitempty"`
}

func (o Address) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "Address struct{}"
	}

	return strings.Join([]string{"Address", string(data)}, " ")
}

type AddressVersion struct {
	value int32
}

type AddressVersionEnum struct {
	E_4 AddressVersion
	E_6 AddressVersion
}

func GetAddressVersionEnum() AddressVersionEnum {
	return AddressVersionEnum{
		E_4: AddressVersion{
			value: 4,
		}, E_6: AddressVersion{
			value: 6,
		},
	}
}

func (c AddressVersion) MarshalJSON() ([]byte, error) {
	return json.Marshal(c.value)
}

func (c *AddressVersion) UnmarshalJSON(b []byte) error {
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

type AddressOSEXTIPStype struct {
	value string
}

type AddressOSEXTIPStypeEnum struct {
	FIXED    AddressOSEXTIPStype
	FLOATING AddressOSEXTIPStype
}

func GetAddressOSEXTIPStypeEnum() AddressOSEXTIPStypeEnum {
	return AddressOSEXTIPStypeEnum{
		FIXED: AddressOSEXTIPStype{
			value: "fixed",
		},
		FLOATING: AddressOSEXTIPStype{
			value: "floating",
		},
	}
}

func (c AddressOSEXTIPStype) MarshalJSON() ([]byte, error) {
	return json.Marshal(c.value)
}

func (c *AddressOSEXTIPStype) UnmarshalJSON(b []byte) error {
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
