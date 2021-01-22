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
type DeleteQueueResponse struct {
	HttpStatusCode int `json:"-"`
}

func (o DeleteQueueResponse) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "DeleteQueueResponse struct{}"
	}

	return strings.Join([]string{"DeleteQueueResponse", string(data)}, " ")
}
