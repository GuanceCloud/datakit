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
type ShowNamespaceRequest struct {
	ContentType ShowNamespaceRequestContentType `json:"Content-Type"`
	Namespace   string                          `json:"namespace"`
}

func (o ShowNamespaceRequest) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "ShowNamespaceRequest struct{}"
	}

	return strings.Join([]string{"ShowNamespaceRequest", string(data)}, " ")
}

type ShowNamespaceRequestContentType struct {
	value string
}

type ShowNamespaceRequestContentTypeEnum struct {
	APPLICATION_JSONCHARSETUTF_8 ShowNamespaceRequestContentType
	APPLICATION_JSON             ShowNamespaceRequestContentType
}

func GetShowNamespaceRequestContentTypeEnum() ShowNamespaceRequestContentTypeEnum {
	return ShowNamespaceRequestContentTypeEnum{
		APPLICATION_JSONCHARSETUTF_8: ShowNamespaceRequestContentType{
			value: "application/json;charset=utf-8",
		},
		APPLICATION_JSON: ShowNamespaceRequestContentType{
			value: "application/json",
		},
	}
}

func (c ShowNamespaceRequestContentType) MarshalJSON() ([]byte, error) {
	return json.Marshal(c.value)
}

func (c *ShowNamespaceRequestContentType) UnmarshalJSON(b []byte) error {
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
