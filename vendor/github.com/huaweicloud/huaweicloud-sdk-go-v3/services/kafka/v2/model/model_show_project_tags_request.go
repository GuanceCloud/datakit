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

// Request Object
type ShowProjectTagsRequest struct {
	ProjectId string `json:"project_id"`
}

func (o ShowProjectTagsRequest) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "ShowProjectTagsRequest struct{}"
	}

	return strings.Join([]string{"ShowProjectTagsRequest", string(data)}, " ")
}
