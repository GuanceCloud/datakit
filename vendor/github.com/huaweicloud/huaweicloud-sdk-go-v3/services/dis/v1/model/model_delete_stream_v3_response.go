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
type DeleteStreamV3Response struct {
	HttpStatusCode int `json:"-"`
}

func (o DeleteStreamV3Response) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "DeleteStreamV3Response struct{}"
	}

	return strings.Join([]string{"DeleteStreamV3Response", string(data)}, " ")
}
