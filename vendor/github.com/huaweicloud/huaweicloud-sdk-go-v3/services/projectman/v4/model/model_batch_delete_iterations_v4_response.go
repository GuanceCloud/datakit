/*
 * ProjectMan
 *
 * devcloud projectman api
 *
 */

package model

import (
	"encoding/json"

	"strings"
)

// Response Object
type BatchDeleteIterationsV4Response struct {
	HttpStatusCode int `json:"-"`
}

func (o BatchDeleteIterationsV4Response) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "BatchDeleteIterationsV4Response struct{}"
	}

	return strings.Join([]string{"BatchDeleteIterationsV4Response", string(data)}, " ")
}
