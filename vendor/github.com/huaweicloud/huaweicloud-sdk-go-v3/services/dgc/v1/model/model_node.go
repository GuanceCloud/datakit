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

type Node struct {
	Name *string `json:"name,omitempty"`
	// 节点的类型
	NodeType *NodeNodeType `json:"nodeType,omitempty"`
	Location *Location     `json:"location,omitempty"`
	// 本节点依赖的前一个节点名称
	PreNodeNames *string `json:"preNodeNames,omitempty"`
	// 节点执行条件
	Condition *[]Condition `json:"condition,omitempty"`
	// 节点的属性
	NodeProperties *string `json:"nodeProperties,omitempty"`
	// 轮询节点执行结果时间间隔
	PollingInterval *int32 `json:"pollingInterval,omitempty"`
	// 节点最大执行时间
	MaxExecutionTime *int32 `json:"maxExecutionTime,omitempty"`
	// 节点失败重试次数
	RetryTimes *int32 `json:"retryTimes,omitempty"`
	// 失败重试时间间隔
	RetryInterval *int32 `json:"retryInterval,omitempty"`
	// 作业失败策略
	FailPolicy   *NodeFailPolicy `json:"failPolicy,omitempty"`
	EventTrigger *Event          `json:"eventTrigger,omitempty"`
	CronTrigger  *Cron           `json:"cronTrigger,omitempty"`
}

func (o Node) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "Node struct{}"
	}

	return strings.Join([]string{"Node", string(data)}, " ")
}

type NodeNodeType struct {
	value string
}

type NodeNodeTypeEnum struct {
	HIVE_SQL           NodeNodeType
	SPARK_SQL          NodeNodeType
	DWSSQL             NodeNodeType
	DLISQL             NodeNodeType
	SHELL              NodeNodeType
	CDM_JOB            NodeNodeType
	DIS_TRANSFER_TASK  NodeNodeType
	CS_JOB             NodeNodeType
	CLOUD_TABLE_MANAGE NodeNodeType
	OBS_MANAGER        NodeNodeType
	RESTAPI            NodeNodeType
	MACHINE_LEARNING   NodeNodeType
	SMN                NodeNodeType
	MRS_SPARK          NodeNodeType
	MAP_REDUCE         NodeNodeType
	DLI_SPARK          NodeNodeType
}

func GetNodeNodeTypeEnum() NodeNodeTypeEnum {
	return NodeNodeTypeEnum{
		HIVE_SQL: NodeNodeType{
			value: "HiveSQL",
		},
		SPARK_SQL: NodeNodeType{
			value: "SparkSQL",
		},
		DWSSQL: NodeNodeType{
			value: "DWSSQL",
		},
		DLISQL: NodeNodeType{
			value: "DLISQL",
		},
		SHELL: NodeNodeType{
			value: "Shell",
		},
		CDM_JOB: NodeNodeType{
			value: "CDMJob",
		},
		DIS_TRANSFER_TASK: NodeNodeType{
			value: "DISTransferTask",
		},
		CS_JOB: NodeNodeType{
			value: "CSJob",
		},
		CLOUD_TABLE_MANAGE: NodeNodeType{
			value: "CloudTableManage",
		},
		OBS_MANAGER: NodeNodeType{
			value: "OBSManager",
		},
		RESTAPI: NodeNodeType{
			value: "RESTAPI",
		},
		MACHINE_LEARNING: NodeNodeType{
			value: "MachineLearning",
		},
		SMN: NodeNodeType{
			value: "SMN",
		},
		MRS_SPARK: NodeNodeType{
			value: "MRSSpark",
		},
		MAP_REDUCE: NodeNodeType{
			value: "MapReduce",
		},
		DLI_SPARK: NodeNodeType{
			value: "DLISpark",
		},
	}
}

func (c NodeNodeType) MarshalJSON() ([]byte, error) {
	return json.Marshal(c.value)
}

func (c *NodeNodeType) UnmarshalJSON(b []byte) error {
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

type NodeFailPolicy struct {
	value string
}

type NodeFailPolicyEnum struct {
	FAIL    NodeFailPolicy
	IGNORE  NodeFailPolicy
	SUSPEND NodeFailPolicy
}

func GetNodeFailPolicyEnum() NodeFailPolicyEnum {
	return NodeFailPolicyEnum{
		FAIL: NodeFailPolicy{
			value: "FAIL",
		},
		IGNORE: NodeFailPolicy{
			value: "IGNORE",
		},
		SUSPEND: NodeFailPolicy{
			value: "SUSPEND",
		},
	}
}

func (c NodeFailPolicy) MarshalJSON() ([]byte, error) {
	return json.Marshal(c.value)
}

func (c *NodeFailPolicy) UnmarshalJSON(b []byte) error {
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
