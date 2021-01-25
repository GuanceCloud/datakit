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

// metadata数据结构说明
type MetadataList struct {
	// 裸金属服务器的计费类型。1：按包年包月计费（即prePaid：预付费方式）。
	ChargingMode *MetadataListChargingMode `json:"chargingMode,omitempty"`
	// 按“包年/包月”计费的裸金属服务器对应的订单ID。
	MeteringOrderId *string `json:"metering.order_id,omitempty"`
	// 按“包年/包月”计费的裸金属服务器对应的产品ID
	MeteringProductId *string `json:"metering.product_id,omitempty"`
	// 裸金属服务器所属的虚拟私有云ID
	VpcId *string `json:"vpc_id,omitempty"`
	// 裸金属服务器操作系统对应的镜像ID
	MeteringImageId *string `json:"metering.image_id,omitempty"`
	// 镜像类型，目前支持：公共镜像（gold）私有镜像（private）共享镜像（shared）
	MeteringImagetype *MetadataListMeteringImagetype `json:"metering.imagetype,omitempty"`
	// 裸金属服务器的网卡列表。
	BaremetalPortIDList *string `json:"baremetalPortIDList,omitempty"`
	// 裸金属服务器对应的资源规格编码，格式为：{规格ID}.{os_type}，例如physical.o2.medium.linux。
	MeteringResourcespeccode *string `json:"metering.resourcespeccode,omitempty"`
	// 裸金属服务器对应的资源类型，取值为：hws.resource.type.pm
	MeteringResourcetype *string `json:"metering.resourcetype,omitempty"`
	// 裸金属服务器操作系统对应的镜像名称
	ImageName *string `json:"image_name,omitempty"`
	// 用户ID（登录管理控制台，进入我的凭证，即可看到“用户ID”）
	OpSvcUserid *string `json:"op_svc_userid,omitempty"`
	// 操作系统类型，取值为：Linux、Windows
	OsType *MetadataListOsType `json:"os_type,omitempty"`
	// 裸金属服务器是否支持EVS卷。
	BmsSupportEvs *string `json:"__bms_support_evs,omitempty"`
	// 操作系统位数，一般取值为“32”或者“64”。
	OsBit *MetadataListOsBit `json:"os_bit,omitempty"`
}

func (o MetadataList) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "MetadataList struct{}"
	}

	return strings.Join([]string{"MetadataList", string(data)}, " ")
}

type MetadataListChargingMode struct {
	value string
}

type MetadataListChargingModeEnum struct {
	E_1 MetadataListChargingMode
}

func GetMetadataListChargingModeEnum() MetadataListChargingModeEnum {
	return MetadataListChargingModeEnum{
		E_1: MetadataListChargingMode{
			value: "1",
		},
	}
}

func (c MetadataListChargingMode) MarshalJSON() ([]byte, error) {
	return json.Marshal(c.value)
}

func (c *MetadataListChargingMode) UnmarshalJSON(b []byte) error {
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

type MetadataListMeteringImagetype struct {
	value string
}

type MetadataListMeteringImagetypeEnum struct {
	GOLD    MetadataListMeteringImagetype
	PRIVATE MetadataListMeteringImagetype
	SHARED  MetadataListMeteringImagetype
}

func GetMetadataListMeteringImagetypeEnum() MetadataListMeteringImagetypeEnum {
	return MetadataListMeteringImagetypeEnum{
		GOLD: MetadataListMeteringImagetype{
			value: "gold",
		},
		PRIVATE: MetadataListMeteringImagetype{
			value: "private",
		},
		SHARED: MetadataListMeteringImagetype{
			value: "shared",
		},
	}
}

func (c MetadataListMeteringImagetype) MarshalJSON() ([]byte, error) {
	return json.Marshal(c.value)
}

func (c *MetadataListMeteringImagetype) UnmarshalJSON(b []byte) error {
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

type MetadataListOsType struct {
	value string
}

type MetadataListOsTypeEnum struct {
	LINUX   MetadataListOsType
	WINDOWS MetadataListOsType
}

func GetMetadataListOsTypeEnum() MetadataListOsTypeEnum {
	return MetadataListOsTypeEnum{
		LINUX: MetadataListOsType{
			value: "Linux",
		},
		WINDOWS: MetadataListOsType{
			value: "Windows",
		},
	}
}

func (c MetadataListOsType) MarshalJSON() ([]byte, error) {
	return json.Marshal(c.value)
}

func (c *MetadataListOsType) UnmarshalJSON(b []byte) error {
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

type MetadataListOsBit struct {
	value string
}

type MetadataListOsBitEnum struct {
	E_32 MetadataListOsBit
	E_64 MetadataListOsBit
}

func GetMetadataListOsBitEnum() MetadataListOsBitEnum {
	return MetadataListOsBitEnum{
		E_32: MetadataListOsBit{
			value: "32",
		},
		E_64: MetadataListOsBit{
			value: "64",
		},
	}
}

func (c MetadataListOsBit) MarshalJSON() ([]byte, error) {
	return json.Marshal(c.value)
}

func (c *MetadataListOsBit) UnmarshalJSON(b []byte) error {
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
