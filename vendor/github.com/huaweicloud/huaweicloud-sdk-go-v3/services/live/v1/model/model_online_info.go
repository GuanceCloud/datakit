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

type OnlineInfo struct {
	// 域名
	PublishDomain string `json:"publish_domain"`
	// 应用名
	App string `json:"app"`
	// 流名
	Stream string `json:"stream"`
	// 视频编码方式 - H264 - H265
	VideoCodec OnlineInfoVideoCodec `json:"video_codec"`
	// 音频编码方式 - AAC
	AudioCodec OnlineInfoAudioCodec `json:"audio_codec"`
	// 推流设备的ip
	ClientIp string `json:"client_ip"`
	// 开始推流时刻 UTC格式 2006-01-02T15:04:05Z
	StartTime string `json:"start_time"`
}

func (o OnlineInfo) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "OnlineInfo struct{}"
	}

	return strings.Join([]string{"OnlineInfo", string(data)}, " ")
}

type OnlineInfoVideoCodec struct {
	value string
}

type OnlineInfoVideoCodecEnum struct {
	H264 OnlineInfoVideoCodec
	H265 OnlineInfoVideoCodec
}

func GetOnlineInfoVideoCodecEnum() OnlineInfoVideoCodecEnum {
	return OnlineInfoVideoCodecEnum{
		H264: OnlineInfoVideoCodec{
			value: "H264",
		},
		H265: OnlineInfoVideoCodec{
			value: "H265",
		},
	}
}

func (c OnlineInfoVideoCodec) MarshalJSON() ([]byte, error) {
	return json.Marshal(c.value)
}

func (c *OnlineInfoVideoCodec) UnmarshalJSON(b []byte) error {
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

type OnlineInfoAudioCodec struct {
	value string
}

type OnlineInfoAudioCodecEnum struct {
	AAC OnlineInfoAudioCodec
}

func GetOnlineInfoAudioCodecEnum() OnlineInfoAudioCodecEnum {
	return OnlineInfoAudioCodecEnum{
		AAC: OnlineInfoAudioCodec{
			value: "AAC",
		},
	}
}

func (c OnlineInfoAudioCodec) MarshalJSON() ([]byte, error) {
	return json.Marshal(c.value)
}

func (c *OnlineInfoAudioCodec) UnmarshalJSON(b []byte) error {
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
