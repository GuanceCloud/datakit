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
type ListRepositoryTagsRequest struct {
	ContentType ListRepositoryTagsRequestContentType `json:"Content-Type"`
	Namespace   string                               `json:"namespace"`
	Repository  string                               `json:"repository"`
	Offset      *string                              `json:"offset,omitempty"`
	Limit       *string                              `json:"limit,omitempty"`
	OrderColumn *string                              `json:"order_column,omitempty"`
	OrderType   *ListRepositoryTagsRequestOrderType  `json:"order_type,omitempty"`
	Tag         *string                              `json:"tag,omitempty"`
}

func (o ListRepositoryTagsRequest) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "ListRepositoryTagsRequest struct{}"
	}

	return strings.Join([]string{"ListRepositoryTagsRequest", string(data)}, " ")
}

type ListRepositoryTagsRequestContentType struct {
	value string
}

type ListRepositoryTagsRequestContentTypeEnum struct {
	APPLICATION_JSONCHARSETUTF_8 ListRepositoryTagsRequestContentType
	APPLICATION_JSON             ListRepositoryTagsRequestContentType
}

func GetListRepositoryTagsRequestContentTypeEnum() ListRepositoryTagsRequestContentTypeEnum {
	return ListRepositoryTagsRequestContentTypeEnum{
		APPLICATION_JSONCHARSETUTF_8: ListRepositoryTagsRequestContentType{
			value: "application/json;charset=utf-8",
		},
		APPLICATION_JSON: ListRepositoryTagsRequestContentType{
			value: "application/json",
		},
	}
}

func (c ListRepositoryTagsRequestContentType) MarshalJSON() ([]byte, error) {
	return json.Marshal(c.value)
}

func (c *ListRepositoryTagsRequestContentType) UnmarshalJSON(b []byte) error {
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

type ListRepositoryTagsRequestOrderType struct {
	value string
}

type ListRepositoryTagsRequestOrderTypeEnum struct {
	DESC ListRepositoryTagsRequestOrderType
	ASC  ListRepositoryTagsRequestOrderType
}

func GetListRepositoryTagsRequestOrderTypeEnum() ListRepositoryTagsRequestOrderTypeEnum {
	return ListRepositoryTagsRequestOrderTypeEnum{
		DESC: ListRepositoryTagsRequestOrderType{
			value: "desc",
		},
		ASC: ListRepositoryTagsRequestOrderType{
			value: "asc",
		},
	}
}

func (c ListRepositoryTagsRequestOrderType) MarshalJSON() ([]byte, error) {
	return json.Marshal(c.value)
}

func (c *ListRepositoryTagsRequestOrderType) UnmarshalJSON(b []byte) error {
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
