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

type StreamCountData struct {
	// 采样时间点的推流路数。
	Value *int32 `json:"value,omitempty"`
	// 采样时间。日期格式按照ISO8601表示法，并使用UTC时间。 格式为：YYYY-MM-DDThh:mm:ssZ。
	Time *string `json:"time,omitempty"`
}

func (o StreamCountData) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "StreamCountData struct{}"
	}

	return strings.Join([]string{"StreamCountData", string(data)}, " ")
}
