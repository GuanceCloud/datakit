/*
 * Live
 *
 * 数据分析服务接口
 *
 */

package model

import (
	"encoding/json"

	"strings"
)

type TranscodeSpec struct {
	// 转码规格，格式是“编码格式_分辨率档位”（未开启高清低码）和“编码格式_PVC_分辨率档位”（开启高清低码）。  其中编码格式包括H264、H265，分辨率档位包括：  4K（3840 x 2160）及以下，2K（2560 x 1440）及以下，FHD（1920 x 1080）及以下，HD（1280 x 720）及以下，SD（640 x 480）及以下。
	Type *string `json:"type,omitempty"`
	// 采样时间点转码时长，单位为分钟，保留两位小数。
	Value *float64 `json:"value,omitempty"`
}

func (o TranscodeSpec) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "TranscodeSpec struct{}"
	}

	return strings.Join([]string{"TranscodeSpec", string(data)}, " ")
}
