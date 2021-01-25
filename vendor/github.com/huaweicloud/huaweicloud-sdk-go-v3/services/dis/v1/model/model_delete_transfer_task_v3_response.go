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
type DeleteTransferTaskV3Response struct {
	HttpStatusCode int `json:"-"`
}

func (o DeleteTransferTaskV3Response) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "DeleteTransferTaskV3Response struct{}"
	}

	return strings.Join([]string{"DeleteTransferTaskV3Response", string(data)}, " ")
}
