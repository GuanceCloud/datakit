/*
 * SWR
 *
 * SWR API
 *
 */

package model

import (
	"encoding/json"

	"strings"
)

// Response Object
type CreateRetentionResponse struct {
	// 镜像老化规则id
	Id             *int32 `json:"id,omitempty"`
	HttpStatusCode int    `json:"-"`
}

func (o CreateRetentionResponse) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "CreateRetentionResponse struct{}"
	}

	return strings.Join([]string{"CreateRetentionResponse", string(data)}, " ")
}
