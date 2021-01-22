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
type UpdateStreamV3Response struct {
	HttpStatusCode int `json:"-"`
}

func (o UpdateStreamV3Response) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "UpdateStreamV3Response struct{}"
	}

	return strings.Join([]string{"UpdateStreamV3Response", string(data)}, " ")
}
