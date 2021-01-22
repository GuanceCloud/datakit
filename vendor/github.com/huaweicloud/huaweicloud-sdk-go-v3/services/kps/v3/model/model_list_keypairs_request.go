/*
 * kps
 *
 * kps v3 版本API
 *
 */

package model

import (
	"encoding/json"

	"strings"
)

// Request Object
type ListKeypairsRequest struct {
}

func (o ListKeypairsRequest) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "ListKeypairsRequest struct{}"
	}

	return strings.Join([]string{"ListKeypairsRequest", string(data)}, " ")
}
