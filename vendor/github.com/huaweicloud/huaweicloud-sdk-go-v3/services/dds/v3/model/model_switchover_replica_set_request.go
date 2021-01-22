/*
 * DDS
 *
 * API v3
 *
 */

package model

import (
	"encoding/json"

	"strings"
)

// Request Object
type SwitchoverReplicaSetRequest struct {
	InstanceId string `json:"instance_id"`
}

func (o SwitchoverReplicaSetRequest) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "SwitchoverReplicaSetRequest struct{}"
	}

	return strings.Join([]string{"SwitchoverReplicaSetRequest", string(data)}, " ")
}
