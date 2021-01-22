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

// Response Object
type BatchTagActionResponse struct {
	HttpStatusCode int `json:"-"`
}

func (o BatchTagActionResponse) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "BatchTagActionResponse struct{}"
	}

	return strings.Join([]string{"BatchTagActionResponse", string(data)}, " ")
}
