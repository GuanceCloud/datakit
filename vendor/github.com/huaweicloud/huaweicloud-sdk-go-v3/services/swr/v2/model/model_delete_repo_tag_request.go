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
type DeleteRepoTagRequest struct {
	ContentType DeleteRepoTagRequestContentType `json:"Content-Type"`
	Namespace   string                          `json:"namespace"`
	Repository  string                          `json:"repository"`
	Tag         string                          `json:"tag"`
}

func (o DeleteRepoTagRequest) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "DeleteRepoTagRequest struct{}"
	}

	return strings.Join([]string{"DeleteRepoTagRequest", string(data)}, " ")
}

type DeleteRepoTagRequestContentType struct {
	value string
}

type DeleteRepoTagRequestContentTypeEnum struct {
	APPLICATION_JSONCHARSETUTF_8 DeleteRepoTagRequestContentType
	APPLICATION_JSON             DeleteRepoTagRequestContentType
}

func GetDeleteRepoTagRequestContentTypeEnum() DeleteRepoTagRequestContentTypeEnum {
	return DeleteRepoTagRequestContentTypeEnum{
		APPLICATION_JSONCHARSETUTF_8: DeleteRepoTagRequestContentType{
			value: "application/json;charset=utf-8",
		},
		APPLICATION_JSON: DeleteRepoTagRequestContentType{
			value: "application/json",
		},
	}
}

func (c DeleteRepoTagRequestContentType) MarshalJSON() ([]byte, error) {
	return json.Marshal(c.value)
}

func (c *DeleteRepoTagRequestContentType) UnmarshalJSON(b []byte) error {
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
