/*
 * RDS
 *
 * API v3
 *
 */

package model

import (
	"encoding/json"
	"errors"
	"github.com/huaweicloud/huaweicloud-sdk-go-v3/core/converter"
	"strings"
)

type Volume struct {
	// 磁盘类型。 取值范围如下，区分大小写： - COMMON，表示SATA。 - HIGH，表示SAS。 - ULTRAHIGH，表示SSD。 - ULTRAHIGHPRO，表示SSD尊享版，仅支持超高性能型尊享版（需申请权限）。
	Type VolumeType `json:"type"`
	// 磁盘大小，单位为GB。 取值范围：40GB~4000GB，必须为10的整数倍。  部分用户支持40GB~6000GB，如果您想创建存储空间最大为6000GB的数据库实例，或提高扩容上限到10000GB，请联系客服开通。  说明：对于只读实例，该参数无效，磁盘大小，默认和主实例相同。
	Size int32 `json:"size"`
}

func (o Volume) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "Volume struct{}"
	}

	return strings.Join([]string{"Volume", string(data)}, " ")
}

type VolumeType struct {
	value string
}

type VolumeTypeEnum struct {
	ULTRAHIGH    VolumeType
	HIGH         VolumeType
	COMMON       VolumeType
	NVMESSD      VolumeType
	ULTRAHIGHPRO VolumeType
}

func GetVolumeTypeEnum() VolumeTypeEnum {
	return VolumeTypeEnum{
		ULTRAHIGH: VolumeType{
			value: "ULTRAHIGH",
		},
		HIGH: VolumeType{
			value: "HIGH",
		},
		COMMON: VolumeType{
			value: "COMMON",
		},
		NVMESSD: VolumeType{
			value: "NVMESSD",
		},
		ULTRAHIGHPRO: VolumeType{
			value: "ULTRAHIGHPRO",
		},
	}
}

func (c VolumeType) MarshalJSON() ([]byte, error) {
	return json.Marshal(c.value)
}

func (c *VolumeType) UnmarshalJSON(b []byte) error {
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
