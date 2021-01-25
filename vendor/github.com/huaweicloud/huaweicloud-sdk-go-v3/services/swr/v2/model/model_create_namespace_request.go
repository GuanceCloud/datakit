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
type CreateNamespaceRequest struct {
	ContentType CreateNamespaceRequestContentType `json:"Content-Type"`
	Body        *CreateNamespaceRequestBody       `json:"body,omitempty"`
}

func (o CreateNamespaceRequest) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "CreateNamespaceRequest struct{}"
	}

	return strings.Join([]string{"CreateNamespaceRequest", string(data)}, " ")
}

type CreateNamespaceRequestContentType struct {
	value string
}

type CreateNamespaceRequestContentTypeEnum struct {
	APPLICATION_JSONCHARSETUTF_8 CreateNamespaceRequestContentType
	APPLICATION_JSON             CreateNamespaceRequestContentType
}

func GetCreateNamespaceRequestContentTypeEnum() CreateNamespaceRequestContentTypeEnum {
	return CreateNamespaceRequestContentTypeEnum{
		APPLICATION_JSONCHARSETUTF_8: CreateNamespaceRequestContentType{
			value: "application/json;charset=utf-8",
		},
		APPLICATION_JSON: CreateNamespaceRequestContentType{
			value: "application/json",
		},
	}
}

func (c CreateNamespaceRequestContentType) MarshalJSON() ([]byte, error) {
	return json.Marshal(c.value)
}

func (c *CreateNamespaceRequestContentType) UnmarshalJSON(b []byte) error {
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
