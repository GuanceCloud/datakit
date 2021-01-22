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
type DeleteNamespaceAuthRequest struct {
	ContentType DeleteNamespaceAuthRequestContentType `json:"Content-Type"`
	Namespace   string                                `json:"namespace"`
	Body        *[]string                             `json:"body,omitempty"`
}

func (o DeleteNamespaceAuthRequest) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "DeleteNamespaceAuthRequest struct{}"
	}

	return strings.Join([]string{"DeleteNamespaceAuthRequest", string(data)}, " ")
}

type DeleteNamespaceAuthRequestContentType struct {
	value string
}

type DeleteNamespaceAuthRequestContentTypeEnum struct {
	APPLICATION_JSONCHARSETUTF_8 DeleteNamespaceAuthRequestContentType
	APPLICATION_JSON             DeleteNamespaceAuthRequestContentType
}

func GetDeleteNamespaceAuthRequestContentTypeEnum() DeleteNamespaceAuthRequestContentTypeEnum {
	return DeleteNamespaceAuthRequestContentTypeEnum{
		APPLICATION_JSONCHARSETUTF_8: DeleteNamespaceAuthRequestContentType{
			value: "application/json;charset=utf-8",
		},
		APPLICATION_JSON: DeleteNamespaceAuthRequestContentType{
			value: "application/json",
		},
	}
}

func (c DeleteNamespaceAuthRequestContentType) MarshalJSON() ([]byte, error) {
	return json.Marshal(c.value)
}

func (c *DeleteNamespaceAuthRequestContentType) UnmarshalJSON(b []byte) error {
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
