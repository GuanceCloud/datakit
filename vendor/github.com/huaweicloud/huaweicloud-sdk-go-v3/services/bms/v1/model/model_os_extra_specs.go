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

// os_extra_specs数据结构说明
type OsExtraSpecs struct {
	// 识该规格对应的资源类型，取值范围为“ironic”。
	ResourceType OsExtraSpecsResourceType `json:"resource_type"`
	// 裸金属服务器的CPU架构类型，取值为：x86_64（适用于x86机型）aarch64（适用于ARM机型）
	CapabilitiescpuArch OsExtraSpecsCapabilitiescpuArch `json:"capabilities:cpu_arch"`
	// 磁盘物理规格描述信息。
	BaremetaldiskDetail string `json:"baremetal:disk_detail"`
	// 标示ironic类型的规格。
	CapabilitieshypervisorType string `json:"capabilities:hypervisor_type"`
	// 标识当前的规格是否支持挂载EVS卷。truefalse
	BaremetalSupportEvs *string `json:"baremetal:__support_evs,omitempty"`
	// 裸金属服务器启动源。LocalDisk：本地盘Volume：云硬盘（快速发放）
	BaremetalextBootType *OsExtraSpecsBaremetalextBootType `json:"baremetal:extBootType,omitempty"`
	// 裸金属服务器的规格类型。格式为规格的缩写，例如规格名称为“physical.o2.medium”，则规格类型为“o2m”。
	CapabilitiesboardType string `json:"capabilities:board_type"`
	// 实际可挂载网络数量。
	BaremetalnetNum string `json:"baremetal:net_num"`
	// 网卡物理规格描述信息。
	BaremetalnetcardDetail string `json:"baremetal:netcard_detail"`
	// CPU物理规格描述信息。
	BaremetalcpuDetail string `json:"baremetal:cpu_detail"`
	// 内存物理规格描述信息
	BaremetalmemoryDetail string `json:"baremetal:memory_detail"`
	// 裸金属服务器规格状态。不配置时等同于normal。normal：正常商用abandon：下线（即不显示）sellout：售罄obt：公测promotion：推荐（等同normal，也是商用）
	Condoperationstatus *OsExtraSpecsCondoperationstatus `json:"cond:operation:status,omitempty"`
	// 在某个AZ的裸金属服务器规格状态。此参数是AZ级配置，某个AZ没有在此参数中配置时默认使用cond:operation:status参数的取值。格式：az(xx)。()内为某个AZ下的裸金属服务器规格状态，()内必须填写状态，不填为无效配置。例如：规格在某个区域的az0正常商用，az1售罄，az2公测，az3正常商用，其他az显示下线，可配置为：“cond:operation:status”设置为“abandon”“cond:operation:az”设置为“az0(normal), az1(sellout), az2(obt), az3(promotion)” 说明：如果规格在某个AZ下的状态与cond:operation:status配置状态不同，必须配置该参数。
	Condoperationaz *string `json:"cond:operation:az,omitempty"`
}

func (o OsExtraSpecs) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "OsExtraSpecs struct{}"
	}

	return strings.Join([]string{"OsExtraSpecs", string(data)}, " ")
}

type OsExtraSpecsResourceType struct {
	value string
}

type OsExtraSpecsResourceTypeEnum struct {
	IRONIC OsExtraSpecsResourceType
}

func GetOsExtraSpecsResourceTypeEnum() OsExtraSpecsResourceTypeEnum {
	return OsExtraSpecsResourceTypeEnum{
		IRONIC: OsExtraSpecsResourceType{
			value: "ironic",
		},
	}
}

func (c OsExtraSpecsResourceType) MarshalJSON() ([]byte, error) {
	return json.Marshal(c.value)
}

func (c *OsExtraSpecsResourceType) UnmarshalJSON(b []byte) error {
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

type OsExtraSpecsCapabilitiescpuArch struct {
	value string
}

type OsExtraSpecsCapabilitiescpuArchEnum struct {
	X86_64  OsExtraSpecsCapabilitiescpuArch
	AARCH64 OsExtraSpecsCapabilitiescpuArch
}

func GetOsExtraSpecsCapabilitiescpuArchEnum() OsExtraSpecsCapabilitiescpuArchEnum {
	return OsExtraSpecsCapabilitiescpuArchEnum{
		X86_64: OsExtraSpecsCapabilitiescpuArch{
			value: "x86_64",
		},
		AARCH64: OsExtraSpecsCapabilitiescpuArch{
			value: "aarch64",
		},
	}
}

func (c OsExtraSpecsCapabilitiescpuArch) MarshalJSON() ([]byte, error) {
	return json.Marshal(c.value)
}

func (c *OsExtraSpecsCapabilitiescpuArch) UnmarshalJSON(b []byte) error {
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

type OsExtraSpecsBaremetalextBootType struct {
	value string
}

type OsExtraSpecsBaremetalextBootTypeEnum struct {
	LOCAL_DISK OsExtraSpecsBaremetalextBootType
	VOLUME     OsExtraSpecsBaremetalextBootType
}

func GetOsExtraSpecsBaremetalextBootTypeEnum() OsExtraSpecsBaremetalextBootTypeEnum {
	return OsExtraSpecsBaremetalextBootTypeEnum{
		LOCAL_DISK: OsExtraSpecsBaremetalextBootType{
			value: "LocalDisk",
		},
		VOLUME: OsExtraSpecsBaremetalextBootType{
			value: "Volume",
		},
	}
}

func (c OsExtraSpecsBaremetalextBootType) MarshalJSON() ([]byte, error) {
	return json.Marshal(c.value)
}

func (c *OsExtraSpecsBaremetalextBootType) UnmarshalJSON(b []byte) error {
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

type OsExtraSpecsCondoperationstatus struct {
	value string
}

type OsExtraSpecsCondoperationstatusEnum struct {
	NORMAL    OsExtraSpecsCondoperationstatus
	ABANDON   OsExtraSpecsCondoperationstatus
	SELLOUT   OsExtraSpecsCondoperationstatus
	OBT       OsExtraSpecsCondoperationstatus
	PROMOTION OsExtraSpecsCondoperationstatus
}

func GetOsExtraSpecsCondoperationstatusEnum() OsExtraSpecsCondoperationstatusEnum {
	return OsExtraSpecsCondoperationstatusEnum{
		NORMAL: OsExtraSpecsCondoperationstatus{
			value: "normal",
		},
		ABANDON: OsExtraSpecsCondoperationstatus{
			value: "abandon",
		},
		SELLOUT: OsExtraSpecsCondoperationstatus{
			value: "sellout",
		},
		OBT: OsExtraSpecsCondoperationstatus{
			value: "obt",
		},
		PROMOTION: OsExtraSpecsCondoperationstatus{
			value: "promotion",
		},
	}
}

func (c OsExtraSpecsCondoperationstatus) MarshalJSON() ([]byte, error) {
	return json.Marshal(c.value)
}

func (c *OsExtraSpecsCondoperationstatus) UnmarshalJSON(b []byte) error {
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
