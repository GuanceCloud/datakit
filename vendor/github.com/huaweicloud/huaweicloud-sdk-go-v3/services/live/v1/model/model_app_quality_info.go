/*
 * Live
 *
 * 直播服务源站所有接口
 *
 */

package model

import (
	"encoding/json"

	"strings"
)

type AppQualityInfo struct {
	// 应用名称
	AppName *string `json:"app_name,omitempty"`
	// 视频质量信息
	QualityInfo *[]QualityInfo `json:"quality_info,omitempty"`
}

func (o AppQualityInfo) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "AppQualityInfo struct{}"
	}

	return strings.Join([]string{"AppQualityInfo", string(data)}, " ")
}
