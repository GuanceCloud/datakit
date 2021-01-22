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

// Response Object
type ListSingleStreamBitrateResponse struct {
	// 用量详情。
	BitrateInfoList *[]V2BitrateInfo `json:"bitrate_info_list,omitempty"`
	XRequestId      *string          `json:"X-request-id,omitempty"`
	HttpStatusCode  int              `json:"-"`
}

func (o ListSingleStreamBitrateResponse) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "ListSingleStreamBitrateResponse struct{}"
	}

	return strings.Join([]string{"ListSingleStreamBitrateResponse", string(data)}, " ")
}
