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
type ShowAccessDomainRequest struct {
	ContentType  ShowAccessDomainRequestContentType `json:"Content-Type"`
	Namespace    string                             `json:"namespace"`
	Repository   string                             `json:"repository"`
	AccessDomain string                             `json:"access_domain"`
}

func (o ShowAccessDomainRequest) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "ShowAccessDomainRequest struct{}"
	}

	return strings.Join([]string{"ShowAccessDomainRequest", string(data)}, " ")
}

type ShowAccessDomainRequestContentType struct {
	value string
}

type ShowAccessDomainRequestContentTypeEnum struct {
	APPLICATION_JSONCHARSETUTF_8 ShowAccessDomainRequestContentType
	APPLICATION_JSON             ShowAccessDomainRequestContentType
}

func GetShowAccessDomainRequestContentTypeEnum() ShowAccessDomainRequestContentTypeEnum {
	return ShowAccessDomainRequestContentTypeEnum{
		APPLICATION_JSONCHARSETUTF_8: ShowAccessDomainRequestContentType{
			value: "application/json;charset=utf-8",
		},
		APPLICATION_JSON: ShowAccessDomainRequestContentType{
			value: "application/json",
		},
	}
}

func (c ShowAccessDomainRequestContentType) MarshalJSON() ([]byte, error) {
	return json.Marshal(c.value)
}

func (c *ShowAccessDomainRequestContentType) UnmarshalJSON(b []byte) error {
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
