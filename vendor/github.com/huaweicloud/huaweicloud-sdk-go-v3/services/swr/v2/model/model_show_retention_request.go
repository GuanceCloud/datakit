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
type ShowRetentionRequest struct {
	ContentType ShowRetentionRequestContentType `json:"Content-Type"`
	Namespace   string                          `json:"namespace"`
	Repository  string                          `json:"repository"`
	RetentionId int32                           `json:"retention_id"`
}

func (o ShowRetentionRequest) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "ShowRetentionRequest struct{}"
	}

	return strings.Join([]string{"ShowRetentionRequest", string(data)}, " ")
}

type ShowRetentionRequestContentType struct {
	value string
}

type ShowRetentionRequestContentTypeEnum struct {
	APPLICATION_JSONCHARSETUTF_8 ShowRetentionRequestContentType
	APPLICATION_JSON             ShowRetentionRequestContentType
}

func GetShowRetentionRequestContentTypeEnum() ShowRetentionRequestContentTypeEnum {
	return ShowRetentionRequestContentTypeEnum{
		APPLICATION_JSONCHARSETUTF_8: ShowRetentionRequestContentType{
			value: "application/json;charset=utf-8",
		},
		APPLICATION_JSON: ShowRetentionRequestContentType{
			value: "application/json",
		},
	}
}

func (c ShowRetentionRequestContentType) MarshalJSON() ([]byte, error) {
	return json.Marshal(c.value)
}

func (c *ShowRetentionRequestContentType) UnmarshalJSON(b []byte) error {
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
