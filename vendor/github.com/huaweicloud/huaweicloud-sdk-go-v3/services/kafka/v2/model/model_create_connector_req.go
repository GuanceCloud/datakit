/*
 * Kafka
 *
 * Kafka Document API
 *
 */

package model

import (
	"encoding/json"
	"errors"
	"github.com/huaweicloud/huaweicloud-sdk-go-v3/core/converter"
	"strings"
)

type CreateConnectorReq struct {
	// 部署connector的规格，基准带宽，表示单位时间内传送的最大数据量，单位Byte/秒。  取值范围：   - 100MB   - 300MB   - 600MB   - 1200MB  可以不填，则默认跟当前实例的规格是一致。  第一阶段实现先不填，保持和当前实例规格一致，后面再扩展可以选择不同的规格。
	Specification *CreateConnectorReqSpecification `json:"specification,omitempty"`
	// 转储节点数量。不能小于2个。 默认是2个。
	NodeCnt *string `json:"node_cnt,omitempty"`
	// 转储节点规格编码。
	SpecCode string `json:"spec_code"`
}

func (o CreateConnectorReq) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "CreateConnectorReq struct{}"
	}

	return strings.Join([]string{"CreateConnectorReq", string(data)}, " ")
}

type CreateConnectorReqSpecification struct {
	value string
}

type CreateConnectorReqSpecificationEnum struct {
	E_100_MB  CreateConnectorReqSpecification
	E_300_MB  CreateConnectorReqSpecification
	E_600_MB  CreateConnectorReqSpecification
	E_1200_MB CreateConnectorReqSpecification
}

func GetCreateConnectorReqSpecificationEnum() CreateConnectorReqSpecificationEnum {
	return CreateConnectorReqSpecificationEnum{
		E_100_MB: CreateConnectorReqSpecification{
			value: "100MB",
		},
		E_300_MB: CreateConnectorReqSpecification{
			value: "300MB",
		},
		E_600_MB: CreateConnectorReqSpecification{
			value: "600MB",
		},
		E_1200_MB: CreateConnectorReqSpecification{
			value: "1200MB",
		},
	}
}

func (c CreateConnectorReqSpecification) MarshalJSON() ([]byte, error) {
	return json.Marshal(c.value)
}

func (c *CreateConnectorReqSpecification) UnmarshalJSON(b []byte) error {
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
