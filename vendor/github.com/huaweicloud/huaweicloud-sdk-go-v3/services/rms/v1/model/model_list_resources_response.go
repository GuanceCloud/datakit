/*
 * RMS
 *
 * Resource Manager Api
 *
 */

package model

import (
	"encoding/json"

	"strings"
)

// Response Object
type ListResourcesResponse struct {
	// 资源列表
	Resources      *[]ResourceEntity `json:"resources,omitempty"`
	PageInfo       *PageInfo         `json:"page_info,omitempty"`
	HttpStatusCode int               `json:"-"`
}

func (o ListResourcesResponse) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "ListResourcesResponse struct{}"
	}

	return strings.Join([]string{"ListResourcesResponse", string(data)}, " ")
}
