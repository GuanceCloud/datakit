/*
 * BSS
 *
 * Business Support System API
 *
 */

package model

import (
	"encoding/json"

	"strings"
)

// Request Object
type ListCountiesRequest struct {
	XLanguage *string `json:"X-Language,omitempty"`
	CityCode  string  `json:"city_code"`
	Offset    *int32  `json:"offset,omitempty"`
	Limit     *int32  `json:"limit,omitempty"`
}

func (o ListCountiesRequest) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "ListCountiesRequest struct{}"
	}

	return strings.Join([]string{"ListCountiesRequest", string(data)}, " ")
}
