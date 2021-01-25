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

type RecordConfigInfo struct {
	// 直播播放域名
	Domain string `json:"domain"`
	// 应用名称。 默认为“live”，若您需要自定义应用名称，请先提交工单申请。
	AppName string `json:"app_name"`
	// 秒，周期录制时常，最小15分钟，最大6小时，默认1小时
	RecordDuration *int32 `json:"record_duration,omitempty"`
	// 录制格式flv，hls，mp4，默认flv（目前仅支持flv）
	RecordFormat *RecordConfigInfoRecordFormat `json:"record_format,omitempty"`
	// 录制类型，configer_record，默认configer_record。表示录制配置好后，只要有流就录制
	RecordType *RecordConfigInfoRecordType `json:"record_type,omitempty"`
	// 录制位置vod， 默认vod（目前暂只支持vod）
	RecordLocation *RecordConfigInfoRecordLocation `json:"record_location,omitempty"`
	// 录制文件前缀， DomainName，AppName，StreamName必须，默认{DomainName}/{AppName}/{StreamName}/{StartTime}-{EndTime}
	RecordPrefix *string      `json:"record_prefix,omitempty"`
	ObsAddr      *ObsFileAddr `json:"obs_addr,omitempty"`
	// 录制配置创建时间，样例2020-02-14T10:45:16.947+08:00
	CreateTime *string `json:"create_time,omitempty"`
}

func (o RecordConfigInfo) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "RecordConfigInfo struct{}"
	}

	return strings.Join([]string{"RecordConfigInfo", string(data)}, " ")
}

type RecordConfigInfoRecordFormat struct {
	value string
}

type RecordConfigInfoRecordFormatEnum struct {
	FLV RecordConfigInfoRecordFormat
	HLS RecordConfigInfoRecordFormat
	MP4 RecordConfigInfoRecordFormat
}

func GetRecordConfigInfoRecordFormatEnum() RecordConfigInfoRecordFormatEnum {
	return RecordConfigInfoRecordFormatEnum{
		FLV: RecordConfigInfoRecordFormat{
			value: "flv",
		},
		HLS: RecordConfigInfoRecordFormat{
			value: "hls",
		},
		MP4: RecordConfigInfoRecordFormat{
			value: "mp4",
		},
	}
}

func (c RecordConfigInfoRecordFormat) MarshalJSON() ([]byte, error) {
	return json.Marshal(c.value)
}

func (c *RecordConfigInfoRecordFormat) UnmarshalJSON(b []byte) error {
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

type RecordConfigInfoRecordType struct {
	value string
}

type RecordConfigInfoRecordTypeEnum struct {
	CONFIGER_RECORD RecordConfigInfoRecordType
}

func GetRecordConfigInfoRecordTypeEnum() RecordConfigInfoRecordTypeEnum {
	return RecordConfigInfoRecordTypeEnum{
		CONFIGER_RECORD: RecordConfigInfoRecordType{
			value: "configer_record",
		},
	}
}

func (c RecordConfigInfoRecordType) MarshalJSON() ([]byte, error) {
	return json.Marshal(c.value)
}

func (c *RecordConfigInfoRecordType) UnmarshalJSON(b []byte) error {
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

type RecordConfigInfoRecordLocation struct {
	value string
}

type RecordConfigInfoRecordLocationEnum struct {
	VOD RecordConfigInfoRecordLocation
	OBS RecordConfigInfoRecordLocation
}

func GetRecordConfigInfoRecordLocationEnum() RecordConfigInfoRecordLocationEnum {
	return RecordConfigInfoRecordLocationEnum{
		VOD: RecordConfigInfoRecordLocation{
			value: "vod",
		},
		OBS: RecordConfigInfoRecordLocation{
			value: "obs",
		},
	}
}

func (c RecordConfigInfoRecordLocation) MarshalJSON() ([]byte, error) {
	return json.Marshal(c.value)
}

func (c *RecordConfigInfoRecordLocation) UnmarshalJSON(b []byte) error {
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
