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

// root_volume字段数据结构说明
type RootVolume struct {
	// 裸金属服务器系统盘对应的磁盘类型，需要与系统所提供的磁盘类型相匹配。SATA：普通IO磁盘类型SAS：高IO磁盘类型SSD：超高IO磁盘类型
	Volumetype RootVolumeVolumetype `json:"volumetype"`
	// 系统盘大小，容量单位为GB，输入大小范围为[40-1024]。约束：系统盘大小取值应不小于镜像中系统盘的最小值（min_disk属性）。
	Size int32 `json:"size"`
	// 裸金属服务器系统盘对应的存储池的ID。 说明：使用专属分布式存储时需要该字段。存储池ID可以从管理控制台或者参考《专属分布式存储API参考》的“获取专属分布式存储池详情列表”章节获取。
	ClusterId *string `json:"cluster_id,omitempty"`
	// 裸金属服务器系统盘对应的磁盘存储类型。磁盘存储类型枚举值：DSS（专属分布式存储）。 说明：使用专属分布式存储时需要该字段。存储池类型可以从管理控制台或者参考《专属分布式存储API参考》的“获取专属分布式存储池详情列表”章节获取。
	ClusterType *RootVolumeClusterType `json:"cluster_type,omitempty"`
}

func (o RootVolume) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "RootVolume struct{}"
	}

	return strings.Join([]string{"RootVolume", string(data)}, " ")
}

type RootVolumeVolumetype struct {
	value string
}

type RootVolumeVolumetypeEnum struct {
	SATA RootVolumeVolumetype
	SAS  RootVolumeVolumetype
	SSD  RootVolumeVolumetype
}

func GetRootVolumeVolumetypeEnum() RootVolumeVolumetypeEnum {
	return RootVolumeVolumetypeEnum{
		SATA: RootVolumeVolumetype{
			value: "SATA",
		},
		SAS: RootVolumeVolumetype{
			value: "SAS",
		},
		SSD: RootVolumeVolumetype{
			value: "SSD",
		},
	}
}

func (c RootVolumeVolumetype) MarshalJSON() ([]byte, error) {
	return json.Marshal(c.value)
}

func (c *RootVolumeVolumetype) UnmarshalJSON(b []byte) error {
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

type RootVolumeClusterType struct {
	value string
}

type RootVolumeClusterTypeEnum struct {
	DSS RootVolumeClusterType
}

func GetRootVolumeClusterTypeEnum() RootVolumeClusterTypeEnum {
	return RootVolumeClusterTypeEnum{
		DSS: RootVolumeClusterType{
			value: "DSS",
		},
	}
}

func (c RootVolumeClusterType) MarshalJSON() ([]byte, error) {
	return json.Marshal(c.value)
}

func (c *RootVolumeClusterType) UnmarshalJSON(b []byte) error {
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
