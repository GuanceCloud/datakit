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
type UpdateClusterRequest struct {
	ClusterId   string              `json:"cluster_id"`
	ContentType string              `json:"Content-Type"`
	ErrorStatus *string             `json:"errorStatus,omitempty"`
	Body        *ClusterInformation `json:"body,omitempty"`
}

func (o UpdateClusterRequest) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "UpdateClusterRequest struct{}"
	}

	return strings.Join([]string{"UpdateClusterRequest", string(data)}, " ")
}
