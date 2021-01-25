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
type UpdateRetentionRequest struct {
	ContentType UpdateRetentionRequestContentType `json:"Content-Type"`
	Namespace   string                            `json:"namespace"`
	Repository  string                            `json:"repository"`
	RetentionId int32                             `json:"retention_id"`
	Body        *UpdateRetentionRequestBody       `json:"body,omitempty"`
}

func (o UpdateRetentionRequest) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "UpdateRetentionRequest struct{}"
	}

	return strings.Join([]string{"UpdateRetentionRequest", string(data)}, " ")
}

type UpdateRetentionRequestContentType struct {
	value string
}

type UpdateRetentionRequestContentTypeEnum struct {
	APPLICATION_JSONCHARSETUTF_8 UpdateRetentionRequestContentType
	APPLICATION_JSON             UpdateRetentionRequestContentType
}

func GetUpdateRetentionRequestContentTypeEnum() UpdateRetentionRequestContentTypeEnum {
	return UpdateRetentionRequestContentTypeEnum{
		APPLICATION_JSONCHARSETUTF_8: UpdateRetentionRequestContentType{
			value: "application/json;charset=utf-8",
		},
		APPLICATION_JSON: UpdateRetentionRequestContentType{
			value: "application/json",
		},
	}
}

func (c UpdateRetentionRequestContentType) MarshalJSON() ([]byte, error) {
	return json.Marshal(c.value)
}

func (c *UpdateRetentionRequestContentType) UnmarshalJSON(b []byte) error {
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
