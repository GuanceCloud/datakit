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

// Response Object
type CreateClusterResponse struct {
	// API版本，固定值“v3”，该值不可修改。
	ApiVersion *string `json:"apiVersion,omitempty"`
	// API类型，固定值“Cluster”或“cluster”，该值不可修改。
	Kind           *string          `json:"kind,omitempty"`
	Metadata       *ClusterMetadata `json:"metadata,omitempty"`
	Spec           *V3ClusterSpec   `json:"spec,omitempty"`
	Status         *ClusterStatus   `json:"status,omitempty"`
	HttpStatusCode int              `json:"-"`
}

func (o CreateClusterResponse) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "CreateClusterResponse struct{}"
	}

	return strings.Join([]string{"CreateClusterResponse", string(data)}, " ")
}
