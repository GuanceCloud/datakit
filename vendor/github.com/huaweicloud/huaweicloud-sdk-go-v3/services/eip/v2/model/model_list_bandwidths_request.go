/*
 * EIP
 *
 * 云服务接口
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
type ListBandwidthsRequest struct {
	Marker              *string                         `json:"marker,omitempty"`
	Limit               *int32                          `json:"limit,omitempty"`
	EnterpriseProjectId *string                         `json:"enterprise_project_id,omitempty"`
	ShareType           *ListBandwidthsRequestShareType `json:"share_type,omitempty"`
}

func (o ListBandwidthsRequest) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "ListBandwidthsRequest struct{}"
	}

	return strings.Join([]string{"ListBandwidthsRequest", string(data)}, " ")
}

type ListBandwidthsRequestShareType struct {
	value string
}

type ListBandwidthsRequestShareTypeEnum struct {
	WHOLE ListBandwidthsRequestShareType
	PER   ListBandwidthsRequestShareType
}

func GetListBandwidthsRequestShareTypeEnum() ListBandwidthsRequestShareTypeEnum {
	return ListBandwidthsRequestShareTypeEnum{
		WHOLE: ListBandwidthsRequestShareType{
			value: "WHOLE",
		},
		PER: ListBandwidthsRequestShareType{
			value: "PER",
		},
	}
}

func (c ListBandwidthsRequestShareType) MarshalJSON() ([]byte, error) {
	return json.Marshal(c.value)
}

func (c *ListBandwidthsRequestShareType) UnmarshalJSON(b []byte) error {
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
