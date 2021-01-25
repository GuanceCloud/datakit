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
type HibernateClusterRequest struct {
	ClusterId   string `json:"cluster_id"`
	ContentType string `json:"Content-Type"`
}

func (o HibernateClusterRequest) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "HibernateClusterRequest struct{}"
	}

	return strings.Join([]string{"HibernateClusterRequest", string(data)}, " ")
}
