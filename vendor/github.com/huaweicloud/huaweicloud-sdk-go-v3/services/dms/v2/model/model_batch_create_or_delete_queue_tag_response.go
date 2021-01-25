/*
 * DMS
 *
 * DMS Document API
 *
 */

package model

import (
	"encoding/json"

	"strings"
)

// Response Object
type BatchCreateOrDeleteQueueTagResponse struct {
	HttpStatusCode int `json:"-"`
}

func (o BatchCreateOrDeleteQueueTagResponse) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "BatchCreateOrDeleteQueueTagResponse struct{}"
	}

	return strings.Join([]string{"BatchCreateOrDeleteQueueTagResponse", string(data)}, " ")
}
