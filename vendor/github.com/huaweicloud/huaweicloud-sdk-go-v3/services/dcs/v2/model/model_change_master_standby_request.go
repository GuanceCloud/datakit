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
type ChangeMasterStandbyRequest struct {
	InstanceId string `json:"instance_id"`
}

func (o ChangeMasterStandbyRequest) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "ChangeMasterStandbyRequest struct{}"
	}

	return strings.Join([]string{"ChangeMasterStandbyRequest", string(data)}, " ")
}
