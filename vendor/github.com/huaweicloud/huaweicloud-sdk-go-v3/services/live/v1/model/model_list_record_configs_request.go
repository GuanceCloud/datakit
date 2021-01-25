/*
 * Live
 *
 * 直播服务源站所有接口
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
type ListRecordConfigsRequest struct {
	Domain     string                              `json:"domain"`
	AppName    *string                             `json:"app_name,omitempty"`
	StreamName *string                             `json:"stream_name,omitempty"`
	Page       *int32                              `json:"page,omitempty"`
	Size       *int32                              `json:"size,omitempty"`
	RecordType *ListRecordConfigsRequestRecordType `json:"record_type,omitempty"`
}

func (o ListRecordConfigsRequest) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "ListRecordConfigsRequest struct{}"
	}

	return strings.Join([]string{"ListRecordConfigsRequest", string(data)}, " ")
}

type ListRecordConfigsRequestRecordType struct {
	value string
}

type ListRecordConfigsRequestRecordTypeEnum struct {
	CONFIGER_RECORD ListRecordConfigsRequestRecordType
}

func GetListRecordConfigsRequestRecordTypeEnum() ListRecordConfigsRequestRecordTypeEnum {
	return ListRecordConfigsRequestRecordTypeEnum{
		CONFIGER_RECORD: ListRecordConfigsRequestRecordType{
			value: "configer_record",
		},
	}
}

func (c ListRecordConfigsRequestRecordType) MarshalJSON() ([]byte, error) {
	return json.Marshal(c.value)
}

func (c *ListRecordConfigsRequestRecordType) UnmarshalJSON(b []byte) error {
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
