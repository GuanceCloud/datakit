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
type DeleteNamespacesRequest struct {
	ContentType DeleteNamespacesRequestContentType `json:"Content-Type"`
	Namespace   string                             `json:"namespace"`
}

func (o DeleteNamespacesRequest) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "DeleteNamespacesRequest struct{}"
	}

	return strings.Join([]string{"DeleteNamespacesRequest", string(data)}, " ")
}

type DeleteNamespacesRequestContentType struct {
	value string
}

type DeleteNamespacesRequestContentTypeEnum struct {
	APPLICATION_JSONCHARSETUTF_8 DeleteNamespacesRequestContentType
	APPLICATION_JSON             DeleteNamespacesRequestContentType
}

func GetDeleteNamespacesRequestContentTypeEnum() DeleteNamespacesRequestContentTypeEnum {
	return DeleteNamespacesRequestContentTypeEnum{
		APPLICATION_JSONCHARSETUTF_8: DeleteNamespacesRequestContentType{
			value: "application/json;charset=utf-8",
		},
		APPLICATION_JSON: DeleteNamespacesRequestContentType{
			value: "application/json",
		},
	}
}

func (c DeleteNamespacesRequestContentType) MarshalJSON() ([]byte, error) {
	return json.Marshal(c.value)
}

func (c *DeleteNamespacesRequestContentType) UnmarshalJSON(b []byte) error {
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
