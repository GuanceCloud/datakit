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

type ScriptInfo struct {
	Name *string `json:"name,omitempty"`
	// 脚本类型
	Type *ScriptInfoType `json:"type,omitempty"`
	// 脚本关联的目录
	Directory *string `json:"directory,omitempty"`
	// 脚本内容
	Content *string `json:"content,omitempty"`
	// 脚本关联的连接名称
	ConnectionName *string `json:"connectionName,omitempty"`
	// 脚本执行所在的数据库
	Database *string `json:"database,omitempty"`
	// 脚本关联的DLI队列名称
	QueueName *string `json:"queueName,omitempty"`
	// 脚本的配置项参数
	Configuration *string `json:"configuration,omitempty"`
}

func (o ScriptInfo) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "ScriptInfo struct{}"
	}

	return strings.Join([]string{"ScriptInfo", string(data)}, " ")
}

type ScriptInfoType struct {
	value string
}

type ScriptInfoTypeEnum struct {
	FLINK_SQL ScriptInfoType
	DLISQL    ScriptInfoType
	SPARK_SQL ScriptInfoType
	HIVE_SQL  ScriptInfoType
	DWSSQL    ScriptInfoType
	SHELL     ScriptInfoType
}

func GetScriptInfoTypeEnum() ScriptInfoTypeEnum {
	return ScriptInfoTypeEnum{
		FLINK_SQL: ScriptInfoType{
			value: "FlinkSQL",
		},
		DLISQL: ScriptInfoType{
			value: "DLISQL",
		},
		SPARK_SQL: ScriptInfoType{
			value: "SparkSQL",
		},
		HIVE_SQL: ScriptInfoType{
			value: "HiveSQL",
		},
		DWSSQL: ScriptInfoType{
			value: "DWSSQL",
		},
		SHELL: ScriptInfoType{
			value: "Shell",
		},
	}
}

func (c ScriptInfoType) MarshalJSON() ([]byte, error) {
	return json.Marshal(c.value)
}

func (c *ScriptInfoType) UnmarshalJSON(b []byte) error {
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
