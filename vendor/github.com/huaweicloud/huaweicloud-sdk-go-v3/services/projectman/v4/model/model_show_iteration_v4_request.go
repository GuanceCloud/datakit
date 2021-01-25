/*
 * ProjectMan
 *
 * devcloud projectman api
 *
 */

package model

import (
	"encoding/json"

	"strings"
)

// Request Object
type ShowIterationV4Request struct {
	IterationId int32 `json:"iteration_id"`
}

func (o ShowIterationV4Request) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "ShowIterationV4Request struct{}"
	}

	return strings.Join([]string{"ShowIterationV4Request", string(data)}, " ")
}
