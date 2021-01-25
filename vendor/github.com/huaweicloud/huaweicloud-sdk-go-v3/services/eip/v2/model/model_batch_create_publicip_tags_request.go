/*
 * EIP
 *
 * 云服务接口
 *
 */

package model

import (
	"encoding/json"

	"strings"
)

// Request Object
type BatchCreatePublicipTagsRequest struct {
	PublicipId string                              `json:"publicip_id"`
	Body       *BatchCreatePublicipTagsRequestBody `json:"body,omitempty"`
}

func (o BatchCreatePublicipTagsRequest) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "BatchCreatePublicipTagsRequest struct{}"
	}

	return strings.Join([]string{"BatchCreatePublicipTagsRequest", string(data)}, " ")
}
