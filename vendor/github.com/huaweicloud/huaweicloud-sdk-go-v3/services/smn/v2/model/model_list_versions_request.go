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

// Request Object
type ListVersionsRequest struct {
}

func (o ListVersionsRequest) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "ListVersionsRequest struct{}"
	}

	return strings.Join([]string{"ListVersionsRequest", string(data)}, " ")
}
