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
type CreateRepoDomainsRequest struct {
	ContentType CreateRepoDomainsRequestContentType `json:"Content-Type"`
	Namespace   string                              `json:"namespace"`
	Repository  string                              `json:"repository"`
	Body        *CreateRepoDomainsRequestBody       `json:"body,omitempty"`
}

func (o CreateRepoDomainsRequest) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "CreateRepoDomainsRequest struct{}"
	}

	return strings.Join([]string{"CreateRepoDomainsRequest", string(data)}, " ")
}

type CreateRepoDomainsRequestContentType struct {
	value string
}

type CreateRepoDomainsRequestContentTypeEnum struct {
	APPLICATION_JSONCHARSETUTF_8 CreateRepoDomainsRequestContentType
	APPLICATION_JSON             CreateRepoDomainsRequestContentType
}

func GetCreateRepoDomainsRequestContentTypeEnum() CreateRepoDomainsRequestContentTypeEnum {
	return CreateRepoDomainsRequestContentTypeEnum{
		APPLICATION_JSONCHARSETUTF_8: CreateRepoDomainsRequestContentType{
			value: "application/json;charset=utf-8",
		},
		APPLICATION_JSON: CreateRepoDomainsRequestContentType{
			value: "application/json",
		},
	}
}

func (c CreateRepoDomainsRequestContentType) MarshalJSON() ([]byte, error) {
	return json.Marshal(c.value)
}

func (c *CreateRepoDomainsRequestContentType) UnmarshalJSON(b []byte) error {
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
