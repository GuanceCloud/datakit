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
type ShowRepositoryRequest struct {
	ContentType ShowRepositoryRequestContentType `json:"Content-Type"`
	Namespace   string                           `json:"namespace"`
	Repository  string                           `json:"repository"`
}

func (o ShowRepositoryRequest) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "ShowRepositoryRequest struct{}"
	}

	return strings.Join([]string{"ShowRepositoryRequest", string(data)}, " ")
}

type ShowRepositoryRequestContentType struct {
	value string
}

type ShowRepositoryRequestContentTypeEnum struct {
	APPLICATION_JSONCHARSETUTF_8 ShowRepositoryRequestContentType
	APPLICATION_JSON             ShowRepositoryRequestContentType
}

func GetShowRepositoryRequestContentTypeEnum() ShowRepositoryRequestContentTypeEnum {
	return ShowRepositoryRequestContentTypeEnum{
		APPLICATION_JSONCHARSETUTF_8: ShowRepositoryRequestContentType{
			value: "application/json;charset=utf-8",
		},
		APPLICATION_JSON: ShowRepositoryRequestContentType{
			value: "application/json",
		},
	}
}

func (c ShowRepositoryRequestContentType) MarshalJSON() ([]byte, error) {
	return json.Marshal(c.value)
}

func (c *ShowRepositoryRequestContentType) UnmarshalJSON(b []byte) error {
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
