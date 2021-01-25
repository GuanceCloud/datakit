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

type TranscodeData struct {
	// 每个采样时间中的转码时长信息。
	SpecList *[]TranscodeSpec `json:"spec_list,omitempty"`
	// 采样时间。日期格式按照ISO8601表示法，并使用UTC时间。 格式为：YYYY-MM-DDThh:mm:ssZ 。
	Time *string `json:"time,omitempty"`
}

func (o TranscodeData) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "TranscodeData struct{}"
	}

	return strings.Join([]string{"TranscodeData", string(data)}, " ")
}
