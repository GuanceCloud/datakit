/*
 * Live
 *
 * 数据分析服务接口
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
type ListBandwidthDetailRequest struct {
	PlayDomains []string                            `json:"play_domains"`
	App         *string                             `json:"app,omitempty"`
	Stream      *string                             `json:"stream,omitempty"`
	Region      *[]string                           `json:"region,omitempty"`
	Isp         *[]string                           `json:"isp,omitempty"`
	Interval    *ListBandwidthDetailRequestInterval `json:"interval,omitempty"`
	StartTime   *string                             `json:"start_time,omitempty"`
	EndTime     *string                             `json:"end_time,omitempty"`
}

func (o ListBandwidthDetailRequest) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "ListBandwidthDetailRequest struct{}"
	}

	return strings.Join([]string{"ListBandwidthDetailRequest", string(data)}, " ")
}

type ListBandwidthDetailRequestInterval struct {
	value int32
}

type ListBandwidthDetailRequestIntervalEnum struct {
	E_300   ListBandwidthDetailRequestInterval
	E_3600  ListBandwidthDetailRequestInterval
	E_86400 ListBandwidthDetailRequestInterval
}

func GetListBandwidthDetailRequestIntervalEnum() ListBandwidthDetailRequestIntervalEnum {
	return ListBandwidthDetailRequestIntervalEnum{
		E_300: ListBandwidthDetailRequestInterval{
			value: 300,
		}, E_3600: ListBandwidthDetailRequestInterval{
			value: 3600,
		}, E_86400: ListBandwidthDetailRequestInterval{
			value: 86400,
		},
	}
}

func (c ListBandwidthDetailRequestInterval) MarshalJSON() ([]byte, error) {
	return json.Marshal(c.value)
}

func (c *ListBandwidthDetailRequestInterval) UnmarshalJSON(b []byte) error {
	myConverter := converter.StringConverterFactory("int32")
	if myConverter != nil {
		val, err := myConverter.CovertStringToInterface(strings.Trim(string(b[:]), "\""))
		if err == nil {
			c.value = val.(int32)
			return nil
		}
		return err
	} else {
		return errors.New("convert enum data to int32 error")
	}
}
