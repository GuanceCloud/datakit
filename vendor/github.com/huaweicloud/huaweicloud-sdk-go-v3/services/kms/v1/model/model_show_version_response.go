/*
 * kms
 *
 * KMS v1.0 API, open API
 *
 */

package model

import (
	"encoding/json"

	"strings"
)

// Response Object
type ShowVersionResponse struct {
	// 描述version 对象的列表，详情请参见 ApiVersionDetail字段数据结构说明。
	Version        *interface{} `json:"version,omitempty"`
	HttpStatusCode int          `json:"-"`
}

func (o ShowVersionResponse) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "ShowVersionResponse struct{}"
	}

	return strings.Join([]string{"ShowVersionResponse", string(data)}, " ")
}
