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
type CreateIterationV4Response struct {
	// 迭代id
	Id             *int32 `json:"id,omitempty"`
	HttpStatusCode int    `json:"-"`
}

func (o CreateIterationV4Response) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "CreateIterationV4Response struct{}"
	}

	return strings.Join([]string{"CreateIterationV4Response", string(data)}, " ")
}
