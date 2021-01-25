/*
 * DIS
 *
 * DIS v1 API
 *
 */

package model

import (
	"encoding/json"

	"strings"
)

// Response Object
type DeleteStreamResponse struct {
	HttpStatusCode int `json:"-"`
}

func (o DeleteStreamResponse) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "DeleteStreamResponse struct{}"
	}

	return strings.Join([]string{"DeleteStreamResponse", string(data)}, " ")
}
