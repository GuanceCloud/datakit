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
type ListSingleStreamFramerateResponse struct {
	// 用量详情。
	FramerateInfoList *[]V2FramerateInfo `json:"framerate_info_list,omitempty"`
	XRequestId        *string            `json:"X-request-id,omitempty"`
	HttpStatusCode    int                `json:"-"`
}

func (o ListSingleStreamFramerateResponse) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "ListSingleStreamFramerateResponse struct{}"
	}

	return strings.Join([]string{"ListSingleStreamFramerateResponse", string(data)}, " ")
}
