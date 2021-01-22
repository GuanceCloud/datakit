/*
 * DMS
 *
 * DMS Document API
 *
 */

package model

import (
	"encoding/json"

	"strings"
)

// Response Object
type ListQueuesResponse struct {
	// 该租户的所有队列总数。
	Total *int32 `json:"total,omitempty"`
	// 该租户的所有队列数组。
	Queues         *[]ListQueuesRespQueues `json:"queues,omitempty"`
	HttpStatusCode int                     `json:"-"`
}

func (o ListQueuesResponse) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "ListQueuesResponse struct{}"
	}

	return strings.Join([]string{"ListQueuesResponse", string(data)}, " ")
}
