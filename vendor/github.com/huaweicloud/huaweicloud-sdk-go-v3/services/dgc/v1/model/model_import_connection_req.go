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

type ImportConnectionReq struct {
	Path *string `json:"path,omitempty"`
	// 连接参数
	Params         *[]ConnectionParam                 `json:"params,omitempty"`
	SameNamePolicy *ImportConnectionReqSameNamePolicy `json:"sameNamePolicy,omitempty"`
}

func (o ImportConnectionReq) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "ImportConnectionReq struct{}"
	}

	return strings.Join([]string{"ImportConnectionReq", string(data)}, " ")
}

type ImportConnectionReqSameNamePolicy struct {
	value string
}

type ImportConnectionReqSameNamePolicyEnum struct {
	SKIP      ImportConnectionReqSameNamePolicy
	OVERWRITE ImportConnectionReqSameNamePolicy
}

func GetImportConnectionReqSameNamePolicyEnum() ImportConnectionReqSameNamePolicyEnum {
	return ImportConnectionReqSameNamePolicyEnum{
		SKIP: ImportConnectionReqSameNamePolicy{
			value: "SKIP",
		},
		OVERWRITE: ImportConnectionReqSameNamePolicy{
			value: "OVERWRITE",
		},
	}
}

func (c ImportConnectionReqSameNamePolicy) MarshalJSON() ([]byte, error) {
	return json.Marshal(c.value)
}

func (c *ImportConnectionReqSameNamePolicy) UnmarshalJSON(b []byte) error {
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
