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
type CreateAppV3Response struct {
	HttpStatusCode int `json:"-"`
}

func (o CreateAppV3Response) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "CreateAppV3Response struct{}"
	}

	return strings.Join([]string{"CreateAppV3Response", string(data)}, " ")
}
