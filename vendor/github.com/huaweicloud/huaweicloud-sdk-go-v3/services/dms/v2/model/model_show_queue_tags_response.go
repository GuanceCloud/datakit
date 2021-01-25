/*
 * DMS
 *
 * DMS Document API
 *
 */

package model

import (
	"encoding/json"

	"strings"
)

// Response Object
type ShowQueueTagsResponse struct {
	// 标签列表
	Tags           *[]BatchCreateOrDeleteTagReqTags `json:"tags,omitempty"`
	HttpStatusCode int                              `json:"-"`
}

func (o ShowQueueTagsResponse) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "ShowQueueTagsResponse struct{}"
	}

	return strings.Join([]string{"ShowQueueTagsResponse", string(data)}, " ")
}
