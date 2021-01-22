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
type CreateImageSyncRepoRequest struct {
	ContentType CreateImageSyncRepoRequestContentType `json:"Content-Type"`
	Namespace   string                                `json:"namespace"`
	Repository  string                                `json:"repository"`
	Body        *CreateImageSyncRepoRequestBody       `json:"body,omitempty"`
}

func (o CreateImageSyncRepoRequest) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "CreateImageSyncRepoRequest struct{}"
	}

	return strings.Join([]string{"CreateImageSyncRepoRequest", string(data)}, " ")
}

type CreateImageSyncRepoRequestContentType struct {
	value string
}

type CreateImageSyncRepoRequestContentTypeEnum struct {
	APPLICATION_JSONCHARSETUTF_8 CreateImageSyncRepoRequestContentType
	APPLICATION_JSON             CreateImageSyncRepoRequestContentType
}

func GetCreateImageSyncRepoRequestContentTypeEnum() CreateImageSyncRepoRequestContentTypeEnum {
	return CreateImageSyncRepoRequestContentTypeEnum{
		APPLICATION_JSONCHARSETUTF_8: CreateImageSyncRepoRequestContentType{
			value: "application/json;charset=utf-8",
		},
		APPLICATION_JSON: CreateImageSyncRepoRequestContentType{
			value: "application/json",
		},
	}
}

func (c CreateImageSyncRepoRequestContentType) MarshalJSON() ([]byte, error) {
	return json.Marshal(c.value)
}

func (c *CreateImageSyncRepoRequestContentType) UnmarshalJSON(b []byte) error {
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
