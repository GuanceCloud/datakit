/*
 * DDS
 *
 * API v3
 *
 */

package model

import (
	"encoding/json"

	"strings"
)

// volume信息。
type Volume struct {
	// 磁盘大小。单位：GB。
	Size string `json:"size"`
	// 磁盘使用量。单位：GB。
	Used string `json:"used"`
}

func (o Volume) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "Volume struct{}"
	}

	return strings.Join([]string{"Volume", string(data)}, " ")
}
