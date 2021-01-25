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
type ShowUpBandwidthRequest struct {
	PublishDomains []string                        `json:"publish_domains"`
	App            *string                         `json:"app,omitempty"`
	Stream         *string                         `json:"stream,omitempty"`
	Region         *[]string                       `json:"region,omitempty"`
	Isp            *[]string                       `json:"isp,omitempty"`
	Interval       *ShowUpBandwidthRequestInterval `json:"interval,omitempty"`
	StartTime      *string                         `json:"start_time,omitempty"`
	EndTime        *string                         `json:"end_time,omitempty"`
}

func (o ShowUpBandwidthRequest) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "ShowUpBandwidthRequest struct{}"
	}

	return strings.Join([]string{"ShowUpBandwidthRequest", string(data)}, " ")
}

type ShowUpBandwidthRequestInterval struct {
	value int32
}

type ShowUpBandwidthRequestIntervalEnum struct {
	E_300   ShowUpBandwidthRequestInterval
	E_3600  ShowUpBandwidthRequestInterval
	E_86400 ShowUpBandwidthRequestInterval
}

func GetShowUpBandwidthRequestIntervalEnum() ShowUpBandwidthRequestIntervalEnum {
	return ShowUpBandwidthRequestIntervalEnum{
		E_300: ShowUpBandwidthRequestInterval{
			value: 300,
		}, E_3600: ShowUpBandwidthRequestInterval{
			value: 3600,
		}, E_86400: ShowUpBandwidthRequestInterval{
			value: 86400,
		},
	}
}

func (c ShowUpBandwidthRequestInterval) MarshalJSON() ([]byte, error) {
	return json.Marshal(c.value)
}

func (c *ShowUpBandwidthRequestInterval) UnmarshalJSON(b []byte) error {
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
