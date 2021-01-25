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

type QualityInfo struct {
	// 模板名称。
	TemplateName *string `json:"templateName,omitempty"`
	// 包含如下取值： - FHD： 超高清，系统缺省名称 - HD： 高清，系统缺省名称 - SD： 标清，系统缺省名称 - LD： 流畅，系统缺省名称 - XXX： 租户自定义名称。用户自定义名称不能与系统缺省名称冲突；多个自定义名称不能重复
	Quality string `json:"quality"`
	// 是否使用窄带高清转码，模板组里不同模板的PVC选项必须相同。 - on：启用。 - off：不启用。 默认为off
	Pvc *QualityInfoPvc `json:"PVC,omitempty"`
	// 是否启用高清低码，较PVC相比画质增强。 - on：启用。 - off：不启用。 默认为off。
	Hdlb *QualityInfoHdlb `json:"hdlb,omitempty"`
	// 视频编码格式，模板组里不同模板的编码格式必须相同。 - H264：使用H.264。 - H265：使用H.265。 默认为H264。
	Codec *QualityInfoCodec `json:"codec,omitempty"`
	// 视频宽度（单位：像素） - H264   取值范围：32-3840，必须为2的倍数 。 - H265   取值范围：320-3840 ，必须为4的倍数。
	Width int32 `json:"width"`
	// 视频高度（单位：像素） - H264   取值范围：32-2160，必须为2的倍数。 - H265   取值范围：240-2160，必须为4的倍数。
	Height int32 `json:"height"`
	// 转码视频的码率（单位：Kbps）。 取值范围：40-30000。
	Bitrate int32 `json:"bitrate"`
	// 转码视频帧率（单位：fps）。 取值范围：0-30，0表示保持帧率不变。
	VideoFrameRate *int32 `json:"video_frame_rate,omitempty"`
	// 转码输出支持的协议类型。当前只支持RTMP和HLS，且模板组里不同模板的输出协议类型必须相同。 - RTMP - HLS - DASH  默认为RTMP。
	Protocol *QualityInfoProtocol `json:"protocol,omitempty"`
	// I帧间隔（单位：帧）。  取值范围：0-500。  默认为25。
	IFrameInterval *int32 `json:"iFrameInterval,omitempty"`
	// 按时间设置I帧间隔，与“iFrameInterval”选择一个设置即可。  取值范围：[0,10]  默认值：4
	Gop *int32 `json:"gop,omitempty"`
}

func (o QualityInfo) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "QualityInfo struct{}"
	}

	return strings.Join([]string{"QualityInfo", string(data)}, " ")
}

type QualityInfoPvc struct {
	value string
}

type QualityInfoPvcEnum struct {
	ON  QualityInfoPvc
	OFF QualityInfoPvc
}

func GetQualityInfoPvcEnum() QualityInfoPvcEnum {
	return QualityInfoPvcEnum{
		ON: QualityInfoPvc{
			value: "on",
		},
		OFF: QualityInfoPvc{
			value: "off",
		},
	}
}

func (c QualityInfoPvc) MarshalJSON() ([]byte, error) {
	return json.Marshal(c.value)
}

func (c *QualityInfoPvc) UnmarshalJSON(b []byte) error {
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

type QualityInfoHdlb struct {
	value string
}

type QualityInfoHdlbEnum struct {
	ON  QualityInfoHdlb
	OFF QualityInfoHdlb
}

func GetQualityInfoHdlbEnum() QualityInfoHdlbEnum {
	return QualityInfoHdlbEnum{
		ON: QualityInfoHdlb{
			value: "on",
		},
		OFF: QualityInfoHdlb{
			value: "off",
		},
	}
}

func (c QualityInfoHdlb) MarshalJSON() ([]byte, error) {
	return json.Marshal(c.value)
}

func (c *QualityInfoHdlb) UnmarshalJSON(b []byte) error {
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

type QualityInfoCodec struct {
	value string
}

type QualityInfoCodecEnum struct {
	H264 QualityInfoCodec
	H265 QualityInfoCodec
}

func GetQualityInfoCodecEnum() QualityInfoCodecEnum {
	return QualityInfoCodecEnum{
		H264: QualityInfoCodec{
			value: "H264",
		},
		H265: QualityInfoCodec{
			value: "H265",
		},
	}
}

func (c QualityInfoCodec) MarshalJSON() ([]byte, error) {
	return json.Marshal(c.value)
}

func (c *QualityInfoCodec) UnmarshalJSON(b []byte) error {
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

type QualityInfoProtocol struct {
	value string
}

type QualityInfoProtocolEnum struct {
	RTMP QualityInfoProtocol
	HLS  QualityInfoProtocol
	DASH QualityInfoProtocol
}

func GetQualityInfoProtocolEnum() QualityInfoProtocolEnum {
	return QualityInfoProtocolEnum{
		RTMP: QualityInfoProtocol{
			value: "RTMP",
		},
		HLS: QualityInfoProtocol{
			value: "HLS",
		},
		DASH: QualityInfoProtocol{
			value: "DASH",
		},
	}
}

func (c QualityInfoProtocol) MarshalJSON() ([]byte, error) {
	return json.Marshal(c.value)
}

func (c *QualityInfoProtocol) UnmarshalJSON(b []byte) error {
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
