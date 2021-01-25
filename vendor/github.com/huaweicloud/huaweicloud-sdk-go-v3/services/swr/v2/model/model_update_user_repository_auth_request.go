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
type UpdateUserRepositoryAuthRequest struct {
	ContentType UpdateUserRepositoryAuthRequestContentType `json:"Content-Type"`
	Namespace   string                                     `json:"namespace"`
	Repository  string                                     `json:"repository"`
	Body        *[]UserAuth                                `json:"body,omitempty"`
}

func (o UpdateUserRepositoryAuthRequest) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "UpdateUserRepositoryAuthRequest struct{}"
	}

	return strings.Join([]string{"UpdateUserRepositoryAuthRequest", string(data)}, " ")
}

type UpdateUserRepositoryAuthRequestContentType struct {
	value string
}

type UpdateUserRepositoryAuthRequestContentTypeEnum struct {
	APPLICATION_JSONCHARSETUTF_8 UpdateUserRepositoryAuthRequestContentType
	APPLICATION_JSON             UpdateUserRepositoryAuthRequestContentType
}

func GetUpdateUserRepositoryAuthRequestContentTypeEnum() UpdateUserRepositoryAuthRequestContentTypeEnum {
	return UpdateUserRepositoryAuthRequestContentTypeEnum{
		APPLICATION_JSONCHARSETUTF_8: UpdateUserRepositoryAuthRequestContentType{
			value: "application/json;charset=utf-8",
		},
		APPLICATION_JSON: UpdateUserRepositoryAuthRequestContentType{
			value: "application/json",
		},
	}
}

func (c UpdateUserRepositoryAuthRequestContentType) MarshalJSON() ([]byte, error) {
	return json.Marshal(c.value)
}

func (c *UpdateUserRepositoryAuthRequestContentType) UnmarshalJSON(b []byte) error {
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
