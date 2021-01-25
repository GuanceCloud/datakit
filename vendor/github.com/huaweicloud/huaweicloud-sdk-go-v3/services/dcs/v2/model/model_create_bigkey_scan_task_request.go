/*
 * DCS
 *
 * DCS V2版本API
 *
 */

package model

import (
	"encoding/json"

	"strings"
)

// Request Object
type CreateBigkeyScanTaskRequest struct {
	InstanceId string `json:"instance_id"`
}

func (o CreateBigkeyScanTaskRequest) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "CreateBigkeyScanTaskRequest struct{}"
	}

	return strings.Join([]string{"CreateBigkeyScanTaskRequest", string(data)}, " ")
}
