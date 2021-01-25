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
type ListMonitoredObjectsOfInstanceRequest struct {
	InstanceId string `json:"instance_id"`
	DimName    string `json:"dim_name"`
}

func (o ListMonitoredObjectsOfInstanceRequest) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "ListMonitoredObjectsOfInstanceRequest struct{}"
	}

	return strings.Join([]string{"ListMonitoredObjectsOfInstanceRequest", string(data)}, " ")
}
