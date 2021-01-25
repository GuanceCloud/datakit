/*
 * VPC
 *
 * VPC Open API
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
type ListVpcRoutesRequest struct {
	Limit       *int32                    `json:"limit,omitempty"`
	Marker      *string                   `json:"marker,omitempty"`
	Id          *string                   `json:"id,omitempty"`
	Type        *ListVpcRoutesRequestType `json:"type,omitempty"`
	VpcId       *string                   `json:"vpc_id,omitempty"`
	Destination *string                   `json:"destination,omitempty"`
	TenantId    *string                   `json:"tenant_id,omitempty"`
}

func (o ListVpcRoutesRequest) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "ListVpcRoutesRequest struct{}"
	}

	return strings.Join([]string{"ListVpcRoutesRequest", string(data)}, " ")
}

type ListVpcRoutesRequestType struct {
	value string
}

type ListVpcRoutesRequestTypeEnum struct {
	PEERING ListVpcRoutesRequestType
}

func GetListVpcRoutesRequestTypeEnum() ListVpcRoutesRequestTypeEnum {
	return ListVpcRoutesRequestTypeEnum{
		PEERING: ListVpcRoutesRequestType{
			value: "peering",
		},
	}
}

func (c ListVpcRoutesRequestType) MarshalJSON() ([]byte, error) {
	return json.Marshal(c.value)
}

func (c *ListVpcRoutesRequestType) UnmarshalJSON(b []byte) error {
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
