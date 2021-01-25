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
type DeleteUserRepositoryAuthRequest struct {
	ContentType DeleteUserRepositoryAuthRequestContentType `json:"Content-Type"`
	Namespace   string                                     `json:"namespace"`
	Repository  string                                     `json:"repository"`
	Body        *[]string                                  `json:"body,omitempty"`
}

func (o DeleteUserRepositoryAuthRequest) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "DeleteUserRepositoryAuthRequest struct{}"
	}

	return strings.Join([]string{"DeleteUserRepositoryAuthRequest", string(data)}, " ")
}

type DeleteUserRepositoryAuthRequestContentType struct {
	value string
}

type DeleteUserRepositoryAuthRequestContentTypeEnum struct {
	APPLICATION_JSONCHARSETUTF_8 DeleteUserRepositoryAuthRequestContentType
	APPLICATION_JSON             DeleteUserRepositoryAuthRequestContentType
}

func GetDeleteUserRepositoryAuthRequestContentTypeEnum() DeleteUserRepositoryAuthRequestContentTypeEnum {
	return DeleteUserRepositoryAuthRequestContentTypeEnum{
		APPLICATION_JSONCHARSETUTF_8: DeleteUserRepositoryAuthRequestContentType{
			value: "application/json;charset=utf-8",
		},
		APPLICATION_JSON: DeleteUserRepositoryAuthRequestContentType{
			value: "application/json",
		},
	}
}

func (c DeleteUserRepositoryAuthRequestContentType) MarshalJSON() ([]byte, error) {
	return json.Marshal(c.value)
}

func (c *DeleteUserRepositoryAuthRequestContentType) UnmarshalJSON(b []byte) error {
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
