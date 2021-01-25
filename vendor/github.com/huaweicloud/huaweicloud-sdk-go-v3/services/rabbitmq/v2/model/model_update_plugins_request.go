/*
 * RabbitMQ
 *
 * RabbitMQ Document API
 *
 */

package model

import (
	"encoding/json"

	"strings"
)

// Request Object
type UpdatePluginsRequest struct {
	ProjectId  string            `json:"project_id"`
	InstanceId string            `json:"instance_id"`
	Body       *UpdatePluginsReq `json:"body,omitempty"`
}

func (o UpdatePluginsRequest) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "UpdatePluginsRequest struct{}"
	}

	return strings.Join([]string{"UpdatePluginsRequest", string(data)}, " ")
}
