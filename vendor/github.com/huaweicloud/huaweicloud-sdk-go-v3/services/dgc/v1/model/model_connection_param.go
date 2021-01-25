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

type ConnectionParam struct {
	Name           *string                        `json:"name,omitempty"`
	ConnectionType *ConnectionParamConnectionType `json:"connectionType,omitempty"`
	Params         *interface{}                   `json:"params,omitempty"`
}

func (o ConnectionParam) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "ConnectionParam struct{}"
	}

	return strings.Join([]string{"ConnectionParam", string(data)}, " ")
}

type ConnectionParamConnectionType struct {
	value string
}

type ConnectionParamConnectionTypeEnum struct {
	DWS         ConnectionParamConnectionType
	DLI         ConnectionParamConnectionType
	SPARK_SQL   ConnectionParamConnectionType
	HIVE        ConnectionParamConnectionType
	RDS         ConnectionParamConnectionType
	CLOUD_TABLE ConnectionParamConnectionType
}

func GetConnectionParamConnectionTypeEnum() ConnectionParamConnectionTypeEnum {
	return ConnectionParamConnectionTypeEnum{
		DWS: ConnectionParamConnectionType{
			value: "DWS",
		},
		DLI: ConnectionParamConnectionType{
			value: "DLI",
		},
		SPARK_SQL: ConnectionParamConnectionType{
			value: "SparkSQL",
		},
		HIVE: ConnectionParamConnectionType{
			value: "Hive",
		},
		RDS: ConnectionParamConnectionType{
			value: "RDS",
		},
		CLOUD_TABLE: ConnectionParamConnectionType{
			value: "CloudTable",
		},
	}
}

func (c ConnectionParamConnectionType) MarshalJSON() ([]byte, error) {
	return json.Marshal(c.value)
}

func (c *ConnectionParamConnectionType) UnmarshalJSON(b []byte) error {
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
