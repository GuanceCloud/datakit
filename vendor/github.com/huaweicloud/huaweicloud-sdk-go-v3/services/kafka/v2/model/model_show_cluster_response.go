/*
 * Kafka
 *
 * Kafka Document API
 *
 */

package model

import (
	"encoding/json"

	"strings"
)

// Response Object
type ShowClusterResponse struct {
	Cluster        *ShowClusterRespCluster `json:"cluster,omitempty"`
	HttpStatusCode int                     `json:"-"`
}

func (o ShowClusterResponse) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "ShowClusterResponse struct{}"
	}

	return strings.Join([]string{"ShowClusterResponse", string(data)}, " ")
}
