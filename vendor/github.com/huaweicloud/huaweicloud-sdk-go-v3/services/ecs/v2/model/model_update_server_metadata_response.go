/*
 * ECS
 *
 * ECS Open API
 *
 */

package model

import (
	"encoding/json"

	"strings"
)

// Response Object
type UpdateServerMetadataResponse struct {
	// 用户自定义metadata键值对。
	Metadata       map[string]string `json:"metadata,omitempty"`
	HttpStatusCode int               `json:"-"`
}

func (o UpdateServerMetadataResponse) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "UpdateServerMetadataResponse struct{}"
	}

	return strings.Join([]string{"UpdateServerMetadataResponse", string(data)}, " ")
}
