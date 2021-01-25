/*
 * SWR
 *
 * SWR API
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
type CreateNamespaceAuthRequest struct {
	ContentType CreateNamespaceAuthRequestContentType `json:"Content-Type"`
	Namespace   string                                `json:"namespace"`
	Body        *[]UserAuth                           `json:"body,omitempty"`
}

func (o CreateNamespaceAuthRequest) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "CreateNamespaceAuthRequest struct{}"
	}

	return strings.Join([]string{"CreateNamespaceAuthRequest", string(data)}, " ")
}

type CreateNamespaceAuthRequestContentType struct {
	value string
}

type CreateNamespaceAuthRequestContentTypeEnum struct {
	APPLICATION_JSONCHARSETUTF_8 CreateNamespaceAuthRequestContentType
	APPLICATION_JSON             CreateNamespaceAuthRequestContentType
}

func GetCreateNamespaceAuthRequestContentTypeEnum() CreateNamespaceAuthRequestContentTypeEnum {
	return CreateNamespaceAuthRequestContentTypeEnum{
		APPLICATION_JSONCHARSETUTF_8: CreateNamespaceAuthRequestContentType{
			value: "application/json;charset=utf-8",
		},
		APPLICATION_JSON: CreateNamespaceAuthRequestContentType{
			value: "application/json",
		},
	}
}

func (c CreateNamespaceAuthRequestContentType) MarshalJSON() ([]byte, error) {
	return json.Marshal(c.value)
}

func (c *CreateNamespaceAuthRequestContentType) UnmarshalJSON(b []byte) error {
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
