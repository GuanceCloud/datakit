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

type V3Cluster struct {
	// API版本，固定值“v3”，该值不可修改。
	ApiVersion string `json:"apiVersion"`
	// API类型，固定值“Cluster”或“cluster”，该值不可修改。
	Kind     string           `json:"kind"`
	Metadata *ClusterMetadata `json:"metadata"`
	Spec     *V3ClusterSpec   `json:"spec"`
	Status   *ClusterStatus   `json:"status,omitempty"`
}

func (o V3Cluster) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "V3Cluster struct{}"
	}

	return strings.Join([]string{"V3Cluster", string(data)}, " ")
}
