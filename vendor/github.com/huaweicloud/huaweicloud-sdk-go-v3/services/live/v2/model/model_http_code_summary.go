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

type HttpCodeSummary struct {
	// 状态码信息
	HttpCodes *[]HttpCode `json:"http_codes,omitempty"`
	// 采样时间。日期格式按照ISO8601表示法，并使用UTC时间。 格式为：YYYY-MM-DDThh:mm:ssZ 。
	Time *string `json:"time,omitempty"`
}

func (o HttpCodeSummary) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "HttpCodeSummary struct{}"
	}

	return strings.Join([]string{"HttpCodeSummary", string(data)}, " ")
}
