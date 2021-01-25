/*
 * ProjectMan
 *
 * devcloud projectman api
 *
 */

package model

import (
	"encoding/json"

	"strings"
)

// Request Object
type BatchDeleteIterationsV4Request struct {
	ProjectId string                              `json:"project_id"`
	Body      *BatchDeleteIterationsV4RequestBody `json:"body,omitempty"`
}

func (o BatchDeleteIterationsV4Request) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "BatchDeleteIterationsV4Request struct{}"
	}

	return strings.Join([]string{"BatchDeleteIterationsV4Request", string(data)}, " ")
}
