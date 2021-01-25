/*
 * SMN
 *
 * SMN Open API
 *
 */

package model

import (
	"encoding/json"

	"strings"
)

type CreateResourceTagRequestBody struct {
	Tag *CreateResourceTagRequestBodyTag `json:"tag"`
}

func (o CreateResourceTagRequestBody) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "CreateResourceTagRequestBody struct{}"
	}

	return strings.Join([]string{"CreateResourceTagRequestBody", string(data)}, " ")
}
