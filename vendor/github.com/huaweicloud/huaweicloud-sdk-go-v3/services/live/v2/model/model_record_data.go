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

type RecordData struct {
	// 最大并发路数。
	ConcurrentCount *int32 `json:"concurrent_count,omitempty"`
	// 采样时间，每小时内最大并发路数时间点。日期格式按照ISO8601表示法，并使用UTC时间。 格式为：YYYY-MM-DDThh:mm:ssZ 。
	Time *string `json:"time,omitempty"`
}

func (o RecordData) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "RecordData struct{}"
	}

	return strings.Join([]string{"RecordData", string(data)}, " ")
}
