/*
 * SMN
 *
 * SMN Open API
 *
 */

package model

import (
	"encoding/json"

	"strings"
)

type ApplicationEndpoint struct {
	// 创建application的时间 时间格式为UTC时间，YYYY-MM-DDTHH:MM:SSZ。
	CreateTime string `json:"create_time"`
	// Application endpoint的唯一资源标识。
	EndpointUrn string `json:"endpoint_urn"`
	// 用户自定义数据 最大长度支持UTF-8编码后2048字节。
	UserData string `json:"user_data"`
	// endpoint启用开关 true或false字符串。
	Enabled string `json:"enabled"`
	// 设备token 最大长度512个字节。
	Token string `json:"token"`
}

func (o ApplicationEndpoint) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "ApplicationEndpoint struct{}"
	}

	return strings.Join([]string{"ApplicationEndpoint", string(data)}, " ")
}
