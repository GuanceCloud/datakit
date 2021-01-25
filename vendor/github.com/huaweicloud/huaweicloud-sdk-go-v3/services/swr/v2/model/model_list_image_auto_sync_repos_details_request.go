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
type ListImageAutoSyncReposDetailsRequest struct {
	ContentType ListImageAutoSyncReposDetailsRequestContentType `json:"Content-Type"`
	Namespace   string                                          `json:"namespace"`
	Repository  string                                          `json:"repository"`
}

func (o ListImageAutoSyncReposDetailsRequest) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "ListImageAutoSyncReposDetailsRequest struct{}"
	}

	return strings.Join([]string{"ListImageAutoSyncReposDetailsRequest", string(data)}, " ")
}

type ListImageAutoSyncReposDetailsRequestContentType struct {
	value string
}

type ListImageAutoSyncReposDetailsRequestContentTypeEnum struct {
	APPLICATION_JSONCHARSETUTF_8 ListImageAutoSyncReposDetailsRequestContentType
	APPLICATION_JSON             ListImageAutoSyncReposDetailsRequestContentType
}

func GetListImageAutoSyncReposDetailsRequestContentTypeEnum() ListImageAutoSyncReposDetailsRequestContentTypeEnum {
	return ListImageAutoSyncReposDetailsRequestContentTypeEnum{
		APPLICATION_JSONCHARSETUTF_8: ListImageAutoSyncReposDetailsRequestContentType{
			value: "application/json;charset=utf-8",
		},
		APPLICATION_JSON: ListImageAutoSyncReposDetailsRequestContentType{
			value: "application/json",
		},
	}
}

func (c ListImageAutoSyncReposDetailsRequestContentType) MarshalJSON() ([]byte, error) {
	return json.Marshal(c.value)
}

func (c *ListImageAutoSyncReposDetailsRequestContentType) UnmarshalJSON(b []byte) error {
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
