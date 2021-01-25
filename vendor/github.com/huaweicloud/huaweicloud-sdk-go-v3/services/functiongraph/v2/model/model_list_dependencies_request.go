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

// Request Object
type ListDependenciesRequest struct {
	DependencyType *ListDependenciesRequestDependencyType `json:"dependency_type,omitempty"`
	Runtime        *ListDependenciesRequestRuntime        `json:"runtime,omitempty"`
	Name           *string                                `json:"name,omitempty"`
	Marker         *string                                `json:"marker,omitempty"`
	Limit          *string                                `json:"limit,omitempty"`
}

func (o ListDependenciesRequest) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "ListDependenciesRequest struct{}"
	}

	return strings.Join([]string{"ListDependenciesRequest", string(data)}, " ")
}

type ListDependenciesRequestDependencyType struct {
	value string
}

type ListDependenciesRequestDependencyTypeEnum struct {
	PUBLIC  ListDependenciesRequestDependencyType
	PRIVATE ListDependenciesRequestDependencyType
	ALL     ListDependenciesRequestDependencyType
}

func GetListDependenciesRequestDependencyTypeEnum() ListDependenciesRequestDependencyTypeEnum {
	return ListDependenciesRequestDependencyTypeEnum{
		PUBLIC: ListDependenciesRequestDependencyType{
			value: "public",
		},
		PRIVATE: ListDependenciesRequestDependencyType{
			value: "private",
		},
		ALL: ListDependenciesRequestDependencyType{
			value: "all",
		},
	}
}

func (c ListDependenciesRequestDependencyType) MarshalJSON() ([]byte, error) {
	return json.Marshal(c.value)
}

func (c *ListDependenciesRequestDependencyType) UnmarshalJSON(b []byte) error {
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

type ListDependenciesRequestRuntime struct {
	value string
}

type ListDependenciesRequestRuntimeEnum struct {
	JAVA_8          ListDependenciesRequestRuntime
	NODE_JS_6_10    ListDependenciesRequestRuntime
	NODE_JS_8_10    ListDependenciesRequestRuntime
	NODE_JS_10_16   ListDependenciesRequestRuntime
	NODE_JS_12_13   ListDependenciesRequestRuntime
	PYTHON_2_7      ListDependenciesRequestRuntime
	_PYTHON_3_6     ListDependenciesRequestRuntime
	GO_1_8          ListDependenciesRequestRuntime
	C__NET_CORE_2_0 ListDependenciesRequestRuntime
	C__NET_CORE_2_1 ListDependenciesRequestRuntime
	C__NET_CORE_3_1 ListDependenciesRequestRuntime
	PHP_7_3         ListDependenciesRequestRuntime
}

func GetListDependenciesRequestRuntimeEnum() ListDependenciesRequestRuntimeEnum {
	return ListDependenciesRequestRuntimeEnum{
		JAVA_8: ListDependenciesRequestRuntime{
			value: "Java 8",
		},
		NODE_JS_6_10: ListDependenciesRequestRuntime{
			value: "Node.js 6.10",
		},
		NODE_JS_8_10: ListDependenciesRequestRuntime{
			value: "Node.js 8.10",
		},
		NODE_JS_10_16: ListDependenciesRequestRuntime{
			value: "Node.js 10.16",
		},
		NODE_JS_12_13: ListDependenciesRequestRuntime{
			value: "Node.js 12.13",
		},
		PYTHON_2_7: ListDependenciesRequestRuntime{
			value: "Python 2.7",
		},
		_PYTHON_3_6: ListDependenciesRequestRuntime{
			value: "  Python 3.6",
		},
		GO_1_8: ListDependenciesRequestRuntime{
			value: "Go 1.8",
		},
		C__NET_CORE_2_0: ListDependenciesRequestRuntime{
			value: "C#(.NET Core 2.0)",
		},
		C__NET_CORE_2_1: ListDependenciesRequestRuntime{
			value: "C#(.NET Core 2.1)",
		},
		C__NET_CORE_3_1: ListDependenciesRequestRuntime{
			value: "C#(.NET Core 3.1)",
		},
		PHP_7_3: ListDependenciesRequestRuntime{
			value: "PHP 7.3",
		},
	}
}

func (c ListDependenciesRequestRuntime) MarshalJSON() ([]byte, error) {
	return json.Marshal(c.value)
}

func (c *ListDependenciesRequestRuntime) UnmarshalJSON(b []byte) error {
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
