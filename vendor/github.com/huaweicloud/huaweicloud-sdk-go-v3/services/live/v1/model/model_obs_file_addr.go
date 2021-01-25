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

type ObsFileAddr struct {
	// OBS的bucket名称
	Bucket string `json:"bucket"`
	// OBS Bucket所在数据中心（OBS Location）
	Location string `json:"location"`
	// OBS对象路径，遵守OSS Object定义，当用于指示input时,需要指定到具体对象；当用于指示output时, 只需指定到转码结果期望存放的路径
	Object string `json:"object"`
}

func (o ObsFileAddr) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "ObsFileAddr struct{}"
	}

	return strings.Join([]string{"ObsFileAddr", string(data)}, " ")
}
