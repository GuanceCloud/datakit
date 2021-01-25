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

type SnapshotData struct {
	// 每小时内截图总数，单位为张。
	Count *int64 `json:"count,omitempty"`
	// 采样时间。日期格式按照ISO8601表示法，并使用UTC时间。 格式为：YYYY-MM-DDThh:mm:ssZ 。
	Time *string `json:"time,omitempty"`
}

func (o SnapshotData) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "SnapshotData struct{}"
	}

	return strings.Join([]string{"SnapshotData", string(data)}, " ")
}
