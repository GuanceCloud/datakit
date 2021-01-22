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

// Request Object
type GetRecordsRequest struct {
	PartitionCursor string `json:"partition-cursor"`
	MaxFetchBytes   *int32 `json:"max_fetch_bytes,omitempty"`
}

func (o GetRecordsRequest) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "GetRecordsRequest struct{}"
	}

	return strings.Join([]string{"GetRecordsRequest", string(data)}, " ")
}
