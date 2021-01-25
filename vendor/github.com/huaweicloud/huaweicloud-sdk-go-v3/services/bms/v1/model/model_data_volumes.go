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

// data_volumes字段数据结构说明
type DataVolumes struct {
	// 裸金属服务器系统盘对应的磁盘类型，需要与系统所提供的磁盘类型相匹配。SATA：普通IO磁盘类型SAS：高IO磁盘类型SSD：超高IO磁盘类型约束：在专属云中申请裸金属服务器时，须使用专属企业存储，此时该字段前缀必须是DESS_ 。枚举值如下：DESS_SAS_ISCSI：普通I/O企业存储DESS_SAS_FC：普通I/O企业存储（低延时）DESS_MIX_ISCSI：高I/O企业存储DESS_MIX_FC：高I/O企业存储（低延时）DESS_SSD_ISCSI：超高I/O企业存储DESS_SSD_FC：超高I/O企业存储（低延时）所有用户，包年包月场景下，不能挂载DESS卷。 说明：企业存储支持的存储类型说明可以从管理控制台或参考《专属企业存储服务用户指南》的“申请专属企业存储”“申请专属企业存储”章节获取。
	Volumetype DataVolumesVolumetype `json:"volumetype"`
	// 数据盘大小，容量单位为GB，输入大小范围为[10-32768]。
	Size int32 `json:"size"`
	// 是否为共享磁盘。true为共享盘，false为普通云硬盘
	Shareable *bool `json:"shareable,omitempty"`
	// 裸金属服务器数据盘对应的存储池ID。 说明：使用专属分布式存储时需要该字段。存储池ID可以从管理控制台或者参考《专属分布式存储API参考》的“获取专属分布式存储池详情列表”章节获取。
	ClusterId *string `json:"cluster_id,omitempty"`
	// 裸金属服务器数据盘对应的磁盘存储类型。磁盘存储类型枚举值：DSS（专属分布式存储）。 说明：使用专属分布式存储时需要该字段。存储池类型可以从管理控制台或者参考《专属分布式存储API参考》的“获取专属分布式存储池详情列表”章节获取。
	ClusterType *DataVolumesClusterType `json:"cluster_type,omitempty"`
}

func (o DataVolumes) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "DataVolumes struct{}"
	}

	return strings.Join([]string{"DataVolumes", string(data)}, " ")
}

type DataVolumesVolumetype struct {
	value string
}

type DataVolumesVolumetypeEnum struct {
	SATA DataVolumesVolumetype
	SAS  DataVolumesVolumetype
	SSD  DataVolumesVolumetype
}

func GetDataVolumesVolumetypeEnum() DataVolumesVolumetypeEnum {
	return DataVolumesVolumetypeEnum{
		SATA: DataVolumesVolumetype{
			value: "SATA",
		},
		SAS: DataVolumesVolumetype{
			value: "SAS",
		},
		SSD: DataVolumesVolumetype{
			value: "SSD",
		},
	}
}

func (c DataVolumesVolumetype) MarshalJSON() ([]byte, error) {
	return json.Marshal(c.value)
}

func (c *DataVolumesVolumetype) UnmarshalJSON(b []byte) error {
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

type DataVolumesClusterType struct {
	value string
}

type DataVolumesClusterTypeEnum struct {
	DSS DataVolumesClusterType
}

func GetDataVolumesClusterTypeEnum() DataVolumesClusterTypeEnum {
	return DataVolumesClusterTypeEnum{
		DSS: DataVolumesClusterType{
			value: "DSS",
		},
	}
}

func (c DataVolumesClusterType) MarshalJSON() ([]byte, error) {
	return json.Marshal(c.value)
}

func (c *DataVolumesClusterType) UnmarshalJSON(b []byte) error {
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
