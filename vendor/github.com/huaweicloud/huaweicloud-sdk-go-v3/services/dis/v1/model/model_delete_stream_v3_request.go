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

// Request Object
type DeleteStreamV3Request struct {
	StreamName string `json:"stream_name"`
}

func (o DeleteStreamV3Request) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "DeleteStreamV3Request struct{}"
	}

	return strings.Join([]string{"DeleteStreamV3Request", string(data)}, " ")
}
