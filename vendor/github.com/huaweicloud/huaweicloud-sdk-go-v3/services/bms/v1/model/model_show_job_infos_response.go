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
	"github.com/huaweicloud/huaweicloud-sdk-go-v3/core/sdktime"
	"strings"
)

// Response Object
type ShowJobInfosResponse struct {
	// Job的状态。SUCCESS：成功RUNNING：运行中FAIL：失败INIT：正在初始化
	Status   *ShowJobInfosResponseStatus `json:"status,omitempty"`
	Entities *Entities                   `json:"entities,omitempty"`
	// Job ID
	JobId *string `json:"job_id,omitempty"`
	// Job的类型，包含以下类型：baremetalBatchCreate：批量创建裸金属服务器baremetalBatchOperate：批量修改裸金属服务器电源状态baremetalBatchCreate：批量创建裸金属服务器baremetalChangeOsVolumeBoot：切换快速发放裸金属服务器操作系统baremetalChangeOsLocalDisk：切换本地盘裸金属服务器操作系统baremetalVolumeBootReinstallOs：重装快速发放裸金属服务器操作系统baremetalReinstallOs：重装本地盘裸金属服务器操作系统baremetalAttachVolume：挂载单个磁盘baremetalDetachVolume：卸载单个磁盘baremetalBatchAttachVolume：裸金属服务器批量挂载共享磁盘
	JobType *string `json:"job_type,omitempty"`
	// 开始时间。时间戳格式为ISO 8601，例如：2019-04-25T20:04:47.591Z
	BeginTime *sdktime.SdkTime `json:"begin_time,omitempty"`
	// 结束时间。时间戳格式为ISO 8601，例如：2019-04-26T20:04:47.591Z
	EndTime *sdktime.SdkTime `json:"end_time,omitempty"`
	// Job执行失败时的错误码
	ErrorCode *string `json:"error_code,omitempty"`
	// Job执行失败时的错误原因
	FailReason *string `json:"fail_reason,omitempty"`
	// 出现错误时，返回的错误消息
	Message *string `json:"message,omitempty"`
	// 出现错误时，返回的错误码。错误码和其对应的含义请参考8.1-状态码。
	Code           *string `json:"code,omitempty"`
	HttpStatusCode int     `json:"-"`
}

func (o ShowJobInfosResponse) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "ShowJobInfosResponse struct{}"
	}

	return strings.Join([]string{"ShowJobInfosResponse", string(data)}, " ")
}

type ShowJobInfosResponseStatus struct {
	value string
}

type ShowJobInfosResponseStatusEnum struct {
	SUCCESS ShowJobInfosResponseStatus
	RUNNING ShowJobInfosResponseStatus
	FAIL    ShowJobInfosResponseStatus
	INIT    ShowJobInfosResponseStatus
}

func GetShowJobInfosResponseStatusEnum() ShowJobInfosResponseStatusEnum {
	return ShowJobInfosResponseStatusEnum{
		SUCCESS: ShowJobInfosResponseStatus{
			value: "SUCCESS",
		},
		RUNNING: ShowJobInfosResponseStatus{
			value: "RUNNING",
		},
		FAIL: ShowJobInfosResponseStatus{
			value: "FAIL",
		},
		INIT: ShowJobInfosResponseStatus{
			value: "INIT",
		},
	}
}

func (c ShowJobInfosResponseStatus) MarshalJSON() ([]byte, error) {
	return json.Marshal(c.value)
}

func (c *ShowJobInfosResponseStatus) UnmarshalJSON(b []byte) error {
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
