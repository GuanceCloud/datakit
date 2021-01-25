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
type DeleteRetentionRequest struct {
	ContentType DeleteRetentionRequestContentType `json:"Content-Type"`
	Namespace   string                            `json:"namespace"`
	Repository  string                            `json:"repository"`
	RetentionId int32                             `json:"retention_id"`
}

func (o DeleteRetentionRequest) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "DeleteRetentionRequest struct{}"
	}

	return strings.Join([]string{"DeleteRetentionRequest", string(data)}, " ")
}

type DeleteRetentionRequestContentType struct {
	value string
}

type DeleteRetentionRequestContentTypeEnum struct {
	APPLICATION_JSONCHARSETUTF_8 DeleteRetentionRequestContentType
	APPLICATION_JSON             DeleteRetentionRequestContentType
}

func GetDeleteRetentionRequestContentTypeEnum() DeleteRetentionRequestContentTypeEnum {
	return DeleteRetentionRequestContentTypeEnum{
		APPLICATION_JSONCHARSETUTF_8: DeleteRetentionRequestContentType{
			value: "application/json;charset=utf-8",
		},
		APPLICATION_JSON: DeleteRetentionRequestContentType{
			value: "application/json",
		},
	}
}

func (c DeleteRetentionRequestContentType) MarshalJSON() ([]byte, error) {
	return json.Marshal(c.value)
}

func (c *DeleteRetentionRequestContentType) UnmarshalJSON(b []byte) error {
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
