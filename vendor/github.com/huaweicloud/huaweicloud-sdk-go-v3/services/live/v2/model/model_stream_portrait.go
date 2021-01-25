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

type StreamPortrait struct {
	// 统计日期，日期格式按照ISO8601表示法，格式：YYYYMMDD，如20200904。统计该统计日期00:00-23:59时段的播放画像信息。
	Time *string `json:"time,omitempty"`
	// 累计流量，单位为byte。
	Flow *int64 `json:"flow,omitempty"`
	// 累计播放时长,单位为秒。
	PlayDuration *int64 `json:"play_duration,omitempty"`
	// 累计请求次数。
	RequestCount *int64 `json:"request_count,omitempty"`
	// 累计观看人数,根据IP去重。
	UserCount *int64 `json:"user_count,omitempty"`
	// 峰值观看人数,flv/rtmp协议是统计Session会话ID，其它协议统计IP,1分钟的采样数据。
	PeakUserCount *int64 `json:"peak_user_count,omitempty"`
	// 峰值带宽，单位bps,5分钟的采样数据。
	PeakBandwidth *int64 `json:"peak_bandwidth,omitempty"`
	// 累计直播(推流)时长,单位为秒。
	PushDuration *int64 `json:"push_duration,omitempty"`
}

func (o StreamPortrait) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "StreamPortrait struct{}"
	}

	return strings.Join([]string{"StreamPortrait", string(data)}, " ")
}
