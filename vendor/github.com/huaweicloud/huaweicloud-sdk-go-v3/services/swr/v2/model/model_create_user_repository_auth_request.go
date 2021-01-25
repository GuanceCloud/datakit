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
type CreateUserRepositoryAuthRequest struct {
	ContentType CreateUserRepositoryAuthRequestContentType `json:"Content-Type"`
	Namespace   string                                     `json:"namespace"`
	Repository  string                                     `json:"repository"`
	Body        *[]UserAuth                                `json:"body,omitempty"`
}

func (o CreateUserRepositoryAuthRequest) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "CreateUserRepositoryAuthRequest struct{}"
	}

	return strings.Join([]string{"CreateUserRepositoryAuthRequest", string(data)}, " ")
}

type CreateUserRepositoryAuthRequestContentType struct {
	value string
}

type CreateUserRepositoryAuthRequestContentTypeEnum struct {
	APPLICATION_JSONCHARSETUTF_8 CreateUserRepositoryAuthRequestContentType
	APPLICATION_JSON             CreateUserRepositoryAuthRequestContentType
}

func GetCreateUserRepositoryAuthRequestContentTypeEnum() CreateUserRepositoryAuthRequestContentTypeEnum {
	return CreateUserRepositoryAuthRequestContentTypeEnum{
		APPLICATION_JSONCHARSETUTF_8: CreateUserRepositoryAuthRequestContentType{
			value: "application/json;charset=utf-8",
		},
		APPLICATION_JSON: CreateUserRepositoryAuthRequestContentType{
			value: "application/json",
		},
	}
}

func (c CreateUserRepositoryAuthRequestContentType) MarshalJSON() ([]byte, error) {
	return json.Marshal(c.value)
}

func (c *CreateUserRepositoryAuthRequestContentType) UnmarshalJSON(b []byte) error {
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
