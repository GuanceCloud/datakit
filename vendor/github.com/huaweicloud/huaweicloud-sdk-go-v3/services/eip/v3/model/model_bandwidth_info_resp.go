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

// 公网IP绑定的带宽信息
type BandwidthInfoResp struct {
	// 带宽名称
	BandwidthName *string `json:"bandwidth_name,omitempty"`
	// 带宽大小
	BandwidthNumber *int32 `json:"bandwidth_number,omitempty"`
	// 带宽类型
	BandwidthType *BandwidthInfoRespBandwidthType `json:"bandwidth_type,omitempty"`
	// 带宽id
	BandwidthId *string `json:"bandwidth_id,omitempty"`
}

func (o BandwidthInfoResp) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "BandwidthInfoResp struct{}"
	}

	return strings.Join([]string{"BandwidthInfoResp", string(data)}, " ")
}

type BandwidthInfoRespBandwidthType struct {
	value string
}

type BandwidthInfoRespBandwidthTypeEnum struct {
	PER   BandwidthInfoRespBandwidthType
	WHOLE BandwidthInfoRespBandwidthType
}

func GetBandwidthInfoRespBandwidthTypeEnum() BandwidthInfoRespBandwidthTypeEnum {
	return BandwidthInfoRespBandwidthTypeEnum{
		PER: BandwidthInfoRespBandwidthType{
			value: "PER",
		},
		WHOLE: BandwidthInfoRespBandwidthType{
			value: "WHOLE",
		},
	}
}

func (c BandwidthInfoRespBandwidthType) MarshalJSON() ([]byte, error) {
	return json.Marshal(c.value)
}

func (c *BandwidthInfoRespBandwidthType) UnmarshalJSON(b []byte) error {
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
