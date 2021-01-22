/*
 * CCE
 *
 * CCE开放API
 *
 */

package model

import (
	"encoding/json"

	"strings"
)

// Request Object
type ListAddonInstancesRequest struct {
	ContentType string `json:"Content-Type"`
	ClusterId   string `json:"cluster_id"`
}

func (o ListAddonInstancesRequest) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "ListAddonInstancesRequest struct{}"
	}

	return strings.Join([]string{"ListAddonInstancesRequest", string(data)}, " ")
}
