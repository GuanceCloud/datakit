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
type CreateSecretRequest struct {
	ContentType CreateSecretRequestContentType `json:"Content-Type"`
	Projectname *string                        `json:"projectname,omitempty"`
}

func (o CreateSecretRequest) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "CreateSecretRequest struct{}"
	}

	return strings.Join([]string{"CreateSecretRequest", string(data)}, " ")
}

type CreateSecretRequestContentType struct {
	value string
}

type CreateSecretRequestContentTypeEnum struct {
	APPLICATION_JSONCHARSETUTF_8 CreateSecretRequestContentType
	APPLICATION_JSON             CreateSecretRequestContentType
}

func GetCreateSecretRequestContentTypeEnum() CreateSecretRequestContentTypeEnum {
	return CreateSecretRequestContentTypeEnum{
		APPLICATION_JSONCHARSETUTF_8: CreateSecretRequestContentType{
			value: "application/json;charset=utf-8",
		},
		APPLICATION_JSON: CreateSecretRequestContentType{
			value: "application/json",
		},
	}
}

func (c CreateSecretRequestContentType) MarshalJSON() ([]byte, error) {
	return json.Marshal(c.value)
}

func (c *CreateSecretRequestContentType) UnmarshalJSON(b []byte) error {
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
