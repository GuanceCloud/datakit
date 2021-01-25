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
type CreateTransferTaskV3Response struct {
	HttpStatusCode int `json:"-"`
}

func (o CreateTransferTaskV3Response) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "CreateTransferTaskV3Response struct{}"
	}

	return strings.Join([]string{"CreateTransferTaskV3Response", string(data)}, " ")
}
