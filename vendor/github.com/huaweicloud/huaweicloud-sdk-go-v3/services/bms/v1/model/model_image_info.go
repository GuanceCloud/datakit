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

// image数据结构说明
type ImageInfo struct {
	// 镜像ID，格式为UUID。
	Id *string `json:"id,omitempty"`
	// 镜像的名称
	Name *string `json:"name,omitempty"`
	// 镜像的类型。取值为：Linux（包括SUSE/RedHat/CentOS/Oracle Linux/EulerOS/Ubuntu操作系统）Windows（Windows操作系统）Other（ESXi操作系统）
	OsType *ImageInfoOsType `json:"__os_type,omitempty"`
	// 镜像相关快捷链接地址。
	Links *[]Links `json:"links,omitempty"`
}

func (o ImageInfo) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "ImageInfo struct{}"
	}

	return strings.Join([]string{"ImageInfo", string(data)}, " ")
}

type ImageInfoOsType struct {
	value string
}

type ImageInfoOsTypeEnum struct {
	LINUX   ImageInfoOsType
	WINDOWS ImageInfoOsType
	OTHER   ImageInfoOsType
}

func GetImageInfoOsTypeEnum() ImageInfoOsTypeEnum {
	return ImageInfoOsTypeEnum{
		LINUX: ImageInfoOsType{
			value: "Linux",
		},
		WINDOWS: ImageInfoOsType{
			value: "Windows",
		},
		OTHER: ImageInfoOsType{
			value: "Other",
		},
	}
}

func (c ImageInfoOsType) MarshalJSON() ([]byte, error) {
	return json.Marshal(c.value)
}

func (c *ImageInfoOsType) UnmarshalJSON(b []byte) error {
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
