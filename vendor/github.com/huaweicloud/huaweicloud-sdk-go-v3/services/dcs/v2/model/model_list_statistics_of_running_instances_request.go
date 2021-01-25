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
type ListStatisticsOfRunningInstancesRequest struct {
}

func (o ListStatisticsOfRunningInstancesRequest) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "ListStatisticsOfRunningInstancesRequest struct{}"
	}

	return strings.Join([]string{"ListStatisticsOfRunningInstancesRequest", string(data)}, " ")
}
