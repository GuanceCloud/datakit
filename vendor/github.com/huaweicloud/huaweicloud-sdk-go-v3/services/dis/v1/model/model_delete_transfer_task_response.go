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
type DeleteTransferTaskResponse struct {
	HttpStatusCode int `json:"-"`
}

func (o DeleteTransferTaskResponse) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "DeleteTransferTaskResponse struct{}"
	}

	return strings.Join([]string{"DeleteTransferTaskResponse", string(data)}, " ")
}
