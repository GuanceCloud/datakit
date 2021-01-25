/*
 * DGC
 *
 * 数据湖治理中心DGC是具有数据全生命周期管理、智能数据管理能力的一站式治理运营平台，支持行业知识库智能化建设，支持大数据存储、大数据计算分析引擎等数据底座，帮助企业快速构建从数据接入到数据分析的端到端智能数据系统，消除数据孤岛，统一数据标准，加快数据变现，实现数字化转型
 *
 */

package model

import (
	"encoding/json"
	"errors"
	"github.com/huaweicloud/huaweicloud-sdk-go-v3/core/converter"
	"strings"
)

// Response Object
type ShowJobStatusResponse struct {
	Name      *string                      `json:"name,omitempty"`
	Status    *ShowJobStatusResponseStatus `json:"status,omitempty"`
	Starttime *string                      `json:"starttime,omitempty"`
	EndTime   *string                      `json:"endTime,omitempty"`
	// 状态最后更新时间
	LastUpdateTime *string               `json:"lastUpdateTime,omitempty"`
	Nodes          *[]RealTimeNodeStatus `json:"nodes,omitempty"`
	HttpStatusCode int                   `json:"-"`
}

func (o ShowJobStatusResponse) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "ShowJobStatusResponse struct{}"
	}

	return strings.Join([]string{"ShowJobStatusResponse", string(data)}, " ")
}

type ShowJobStatusResponseStatus struct {
	value string
}

type ShowJobStatusResponseStatusEnum struct {
	STARTTING ShowJobStatusResponseStatus
	NORMAL    ShowJobStatusResponseStatus
	EXCEPTION ShowJobStatusResponseStatus
	STOPPING  ShowJobStatusResponseStatus
	STOPPED   ShowJobStatusResponseStatus
}

func GetShowJobStatusResponseStatusEnum() ShowJobStatusResponseStatusEnum {
	return ShowJobStatusResponseStatusEnum{
		STARTTING: ShowJobStatusResponseStatus{
			value: "STARTTING",
		},
		NORMAL: ShowJobStatusResponseStatus{
			value: "NORMAL",
		},
		EXCEPTION: ShowJobStatusResponseStatus{
			value: "EXCEPTION",
		},
		STOPPING: ShowJobStatusResponseStatus{
			value: "STOPPING",
		},
		STOPPED: ShowJobStatusResponseStatus{
			value: "STOPPED",
		},
	}
}

func (c ShowJobStatusResponseStatus) MarshalJSON() ([]byte, error) {
	return json.Marshal(c.value)
}

func (c *ShowJobStatusResponseStatus) UnmarshalJSON(b []byte) error {
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
