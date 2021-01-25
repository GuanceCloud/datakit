/*
 * VPC
 *
 * VPC Open API
 *
 */

package model

import (
	"encoding/json"

	"strings"
)

//
type CreatePortRequestBody struct {
	Port *CreatePortOption `json:"port"`
}

func (o CreatePortRequestBody) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "CreatePortRequestBody struct{}"
	}

	return strings.Join([]string{"CreatePortRequestBody", string(data)}, " ")
}
