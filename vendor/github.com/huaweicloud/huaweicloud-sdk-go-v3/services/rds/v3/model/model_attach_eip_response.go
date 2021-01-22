/*
 * RDS
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
type AttachEipResponse struct {
	HttpStatusCode int `json:"-"`
}

func (o AttachEipResponse) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "AttachEipResponse struct{}"
	}

	return strings.Join([]string{"AttachEipResponse", string(data)}, " ")
}
