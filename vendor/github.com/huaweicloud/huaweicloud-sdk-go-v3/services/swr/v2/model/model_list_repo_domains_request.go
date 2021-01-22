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
type ListRepoDomainsRequest struct {
	ContentType ListRepoDomainsRequestContentType `json:"Content-Type"`
	Namespace   string                            `json:"namespace"`
	Repository  string                            `json:"repository"`
}

func (o ListRepoDomainsRequest) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "ListRepoDomainsRequest struct{}"
	}

	return strings.Join([]string{"ListRepoDomainsRequest", string(data)}, " ")
}

type ListRepoDomainsRequestContentType struct {
	value string
}

type ListRepoDomainsRequestContentTypeEnum struct {
	APPLICATION_JSONCHARSETUTF_8 ListRepoDomainsRequestContentType
	APPLICATION_JSON             ListRepoDomainsRequestContentType
}

func GetListRepoDomainsRequestContentTypeEnum() ListRepoDomainsRequestContentTypeEnum {
	return ListRepoDomainsRequestContentTypeEnum{
		APPLICATION_JSONCHARSETUTF_8: ListRepoDomainsRequestContentType{
			value: "application/json;charset=utf-8",
		},
		APPLICATION_JSON: ListRepoDomainsRequestContentType{
			value: "application/json",
		},
	}
}

func (c ListRepoDomainsRequestContentType) MarshalJSON() ([]byte, error) {
	return json.Marshal(c.value)
}

func (c *ListRepoDomainsRequestContentType) UnmarshalJSON(b []byte) error {
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
