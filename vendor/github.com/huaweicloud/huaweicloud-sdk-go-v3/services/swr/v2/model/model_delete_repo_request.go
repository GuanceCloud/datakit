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
type DeleteRepoRequest struct {
	ContentType DeleteRepoRequestContentType `json:"Content-Type"`
	Namespace   string                       `json:"namespace"`
	Repository  string                       `json:"repository"`
}

func (o DeleteRepoRequest) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "DeleteRepoRequest struct{}"
	}

	return strings.Join([]string{"DeleteRepoRequest", string(data)}, " ")
}

type DeleteRepoRequestContentType struct {
	value string
}

type DeleteRepoRequestContentTypeEnum struct {
	APPLICATION_JSONCHARSETUTF_8 DeleteRepoRequestContentType
	APPLICATION_JSON             DeleteRepoRequestContentType
}

func GetDeleteRepoRequestContentTypeEnum() DeleteRepoRequestContentTypeEnum {
	return DeleteRepoRequestContentTypeEnum{
		APPLICATION_JSONCHARSETUTF_8: DeleteRepoRequestContentType{
			value: "application/json;charset=utf-8",
		},
		APPLICATION_JSON: DeleteRepoRequestContentType{
			value: "application/json",
		},
	}
}

func (c DeleteRepoRequestContentType) MarshalJSON() ([]byte, error) {
	return json.Marshal(c.value)
}

func (c *DeleteRepoRequestContentType) UnmarshalJSON(b []byte) error {
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
