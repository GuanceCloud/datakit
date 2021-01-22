/*
 * CCE
 *
 * CCE开放API
 *
 */

package model

import (
	"encoding/json"
	"errors"
	"github.com/huaweicloud/huaweicloud-sdk-go-v3/core/converter"
	"strings"
)

// 插件状态信息
type AddonInstanceStatus struct {
	CurrentVersion *Versions `json:"currentVersion"`
	// 安装错误详情
	Message string `json:"message"`
	// 插件安装失败原因
	Reason string `json:"reason"`
	// 插件实例状态
	Status AddonInstanceStatusStatus `json:"status"`
	// 此插件版本，支持升级的集群版本
	TargetVersions *[]string `json:"targetVersions,omitempty"`
}

func (o AddonInstanceStatus) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "AddonInstanceStatus struct{}"
	}

	return strings.Join([]string{"AddonInstanceStatus", string(data)}, " ")
}

type AddonInstanceStatusStatus struct {
	value string
}

type AddonInstanceStatusStatusEnum struct {
	INSTALLING AddonInstanceStatusStatus
	UPGRADING  AddonInstanceStatusStatus
	FAILED     AddonInstanceStatusStatus
	RUNNING    AddonInstanceStatusStatus
}

func GetAddonInstanceStatusStatusEnum() AddonInstanceStatusStatusEnum {
	return AddonInstanceStatusStatusEnum{
		INSTALLING: AddonInstanceStatusStatus{
			value: "installing",
		},
		UPGRADING: AddonInstanceStatusStatus{
			value: "upgrading",
		},
		FAILED: AddonInstanceStatusStatus{
			value: "failed",
		},
		RUNNING: AddonInstanceStatusStatus{
			value: "running",
		},
	}
}

func (c AddonInstanceStatusStatus) MarshalJSON() ([]byte, error) {
	return json.Marshal(c.value)
}

func (c *AddonInstanceStatusStatus) UnmarshalJSON(b []byte) error {
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
