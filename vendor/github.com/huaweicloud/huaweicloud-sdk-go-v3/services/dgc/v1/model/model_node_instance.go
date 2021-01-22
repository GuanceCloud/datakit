/*
 * DGC
 *
 * 数据湖治理中心DGC是具有数据全生命周期管理、智能数据管理能力的一站式治理运营平台，支持行业知识库智能化建设，支持大数据存储、大数据计算分析引擎等数据底座，帮助企业快速构建从数据接入到数据分析的端到端智能数据系统，消除数据孤岛，统一数据标准，加快数据变现，实现数字化转型
 *
 */

package model

import (
	"encoding/json"

	"strings"
)

type NodeInstance struct {
	NodeName       *string `json:"nodeName,omitempty"`
	Status         *string `json:"status,omitempty"`
	PlanTime       *int32  `json:"planTime,omitempty"`
	StartTime      *int32  `json:"startTime,omitempty"`
	EndTime        *int32  `json:"endTime,omitempty"`
	ExecuteTime    *int32  `json:"executeTime,omitempty"`
	NodeType       *string `json:"nodeType,omitempty"`
	RetryTimes     *int32  `json:"retryTimes,omitempty"`
	InstanceId     *int32  `json:"instanceId,omitempty"`
	InputRowCount  *int32  `json:"inputRowCount,omitempty"`
	OutputRowCount *int32  `json:"outputRowCount,omitempty"`
	LogPath        *string `json:"logPath,omitempty"`
}

func (o NodeInstance) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "NodeInstance struct{}"
	}

	return strings.Join([]string{"NodeInstance", string(data)}, " ")
}
