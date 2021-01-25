/*
 * FunctionGraph
 *
 * API v2
 *
 */

package model

import (
	"encoding/json"
	"errors"
	"github.com/huaweicloud/huaweicloud-sdk-go-v3/core/converter"
	"strings"
)

type UpdateDependencyRequestBody struct {
	// depend_type为zip类型时必填，为文件流格式。
	DependFile *string `json:"depend_file,omitempty"`
	// depend_type为obs类型时，依赖包在obs的存储地址。
	DependLink *string `json:"depend_link,omitempty"`
	// 导入类型,目前支持obs和zip。
	DependType string `json:"depend_type"`
	// 运行时语言。
	Runtime UpdateDependencyRequestBodyRuntime `json:"runtime"`
	// 依赖包名称。必须以大、小写字母开头，以字母或数字结尾，只能由字母、数字、下划线、点和中划线组成，长度不超过96个字符。
	Name *string `json:"name,omitempty"`
	// 依赖包描述，不超过512个字符。
	Description *string `json:"description,omitempty"`
}

func (o UpdateDependencyRequestBody) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "UpdateDependencyRequestBody struct{}"
	}

	return strings.Join([]string{"UpdateDependencyRequestBody", string(data)}, " ")
}

type UpdateDependencyRequestBodyRuntime struct {
	value string
}

type UpdateDependencyRequestBodyRuntimeEnum struct {
	JAVA_8          UpdateDependencyRequestBodyRuntime
	NODE_JS_6_10    UpdateDependencyRequestBodyRuntime
	NODE_JS_8_10    UpdateDependencyRequestBodyRuntime
	NODE_JS_10_16   UpdateDependencyRequestBodyRuntime
	NODE_JS_12_13   UpdateDependencyRequestBodyRuntime
	PYTHON_2_7      UpdateDependencyRequestBodyRuntime
	_PYTHON_3_6     UpdateDependencyRequestBodyRuntime
	GO_1_8          UpdateDependencyRequestBodyRuntime
	C__NET_CORE_2_0 UpdateDependencyRequestBodyRuntime
	C__NET_CORE_2_1 UpdateDependencyRequestBodyRuntime
	C__NET_CORE_3_1 UpdateDependencyRequestBodyRuntime
	PHP_7_3         UpdateDependencyRequestBodyRuntime
}

func GetUpdateDependencyRequestBodyRuntimeEnum() UpdateDependencyRequestBodyRuntimeEnum {
	return UpdateDependencyRequestBodyRuntimeEnum{
		JAVA_8: UpdateDependencyRequestBodyRuntime{
			value: "Java 8",
		},
		NODE_JS_6_10: UpdateDependencyRequestBodyRuntime{
			value: "Node.js 6.10",
		},
		NODE_JS_8_10: UpdateDependencyRequestBodyRuntime{
			value: "Node.js 8.10",
		},
		NODE_JS_10_16: UpdateDependencyRequestBodyRuntime{
			value: "Node.js 10.16",
		},
		NODE_JS_12_13: UpdateDependencyRequestBodyRuntime{
			value: "Node.js 12.13",
		},
		PYTHON_2_7: UpdateDependencyRequestBodyRuntime{
			value: "Python 2.7",
		},
		_PYTHON_3_6: UpdateDependencyRequestBodyRuntime{
			value: "  Python 3.6",
		},
		GO_1_8: UpdateDependencyRequestBodyRuntime{
			value: "Go 1.8",
		},
		C__NET_CORE_2_0: UpdateDependencyRequestBodyRuntime{
			value: "C#(.NET Core 2.0)",
		},
		C__NET_CORE_2_1: UpdateDependencyRequestBodyRuntime{
			value: "C#(.NET Core 2.1)",
		},
		C__NET_CORE_3_1: UpdateDependencyRequestBodyRuntime{
			value: "C#(.NET Core 3.1)",
		},
		PHP_7_3: UpdateDependencyRequestBodyRuntime{
			value: "PHP 7.3",
		},
	}
}

func (c UpdateDependencyRequestBodyRuntime) MarshalJSON() ([]byte, error) {
	return json.Marshal(c.value)
}

func (c *UpdateDependencyRequestBodyRuntime) UnmarshalJSON(b []byte) error {
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
