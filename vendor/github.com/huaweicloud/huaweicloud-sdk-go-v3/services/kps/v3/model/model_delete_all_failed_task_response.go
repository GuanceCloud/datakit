/*
 * kps
 *
 * kps v3 版本API
 *
 */

package model

import (
	"encoding/json"

	"strings"
)

// Response Object
type DeleteAllFailedTaskResponse struct {
	HttpStatusCode int `json:"-"`
}

func (o DeleteAllFailedTaskResponse) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "DeleteAllFailedTaskResponse struct{}"
	}

	return strings.Join([]string{"DeleteAllFailedTaskResponse", string(data)}, " ")
}
