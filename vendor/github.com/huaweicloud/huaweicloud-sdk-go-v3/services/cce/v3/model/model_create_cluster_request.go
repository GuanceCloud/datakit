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
type CreateClusterRequest struct {
	ContentType string     `json:"Content-Type"`
	Body        *V3Cluster `json:"body,omitempty"`
}

func (o CreateClusterRequest) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "CreateClusterRequest struct{}"
	}

	return strings.Join([]string{"CreateClusterRequest", string(data)}, " ")
}
