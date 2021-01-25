/*
 * DCS
 *
 * DCS V2版本API
 *
 */

package model

import (
	"encoding/json"

	"strings"
)

// Request Object
type ListMonitoredObjectsRequest struct {
	DimName string `json:"dim_name"`
	Offset  *int32 `json:"offset,omitempty"`
	Limit   *int32 `json:"limit,omitempty"`
}

func (o ListMonitoredObjectsRequest) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "ListMonitoredObjectsRequest struct{}"
	}

	return strings.Join([]string{"ListMonitoredObjectsRequest", string(data)}, " ")
}
