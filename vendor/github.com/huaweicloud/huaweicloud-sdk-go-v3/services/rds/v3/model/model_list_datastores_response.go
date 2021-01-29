/*
 * RDS
 *
 * API v3
 *
 */

package model

import (
	"encoding/json"

	"strings"
)

// Response Object
type ListDatastoresResponse struct {
	// 数据库引擎信息。
	DataStores     *[]LDatastore `json:"dataStores,omitempty"`
	HttpStatusCode int           `json:"-"`
}

func (o ListDatastoresResponse) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "ListDatastoresResponse struct{}"
	}

	return strings.Join([]string{"ListDatastoresResponse", string(data)}, " ")
}
