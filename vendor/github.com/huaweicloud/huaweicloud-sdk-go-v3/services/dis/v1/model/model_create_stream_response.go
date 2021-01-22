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
type CreateStreamResponse struct {
	HttpStatusCode int `json:"-"`
}

func (o CreateStreamResponse) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "CreateStreamResponse struct{}"
	}

	return strings.Join([]string{"CreateStreamResponse", string(data)}, " ")
}
