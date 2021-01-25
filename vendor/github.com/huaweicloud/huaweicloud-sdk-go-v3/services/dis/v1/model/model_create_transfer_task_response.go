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
type CreateTransferTaskResponse struct {
	HttpStatusCode int `json:"-"`
}

func (o CreateTransferTaskResponse) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "CreateTransferTaskResponse struct{}"
	}

	return strings.Join([]string{"CreateTransferTaskResponse", string(data)}, " ")
}
