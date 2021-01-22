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
type CreateRetentionRequest struct {
	ContentType CreateRetentionRequestContentType `json:"Content-Type"`
	Namespace   string                            `json:"namespace"`
	Repository  string                            `json:"repository"`
	Body        *CreateRetentionRequestBody       `json:"body,omitempty"`
}

func (o CreateRetentionRequest) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "CreateRetentionRequest struct{}"
	}

	return strings.Join([]string{"CreateRetentionRequest", string(data)}, " ")
}

type CreateRetentionRequestContentType struct {
	value string
}

type CreateRetentionRequestContentTypeEnum struct {
	APPLICATION_JSONCHARSETUTF_8 CreateRetentionRequestContentType
	APPLICATION_JSON             CreateRetentionRequestContentType
}

func GetCreateRetentionRequestContentTypeEnum() CreateRetentionRequestContentTypeEnum {
	return CreateRetentionRequestContentTypeEnum{
		APPLICATION_JSONCHARSETUTF_8: CreateRetentionRequestContentType{
			value: "application/json;charset=utf-8",
		},
		APPLICATION_JSON: CreateRetentionRequestContentType{
			value: "application/json",
		},
	}
}

func (c CreateRetentionRequestContentType) MarshalJSON() ([]byte, error) {
	return json.Marshal(c.value)
}

func (c *CreateRetentionRequestContentType) UnmarshalJSON(b []byte) error {
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
