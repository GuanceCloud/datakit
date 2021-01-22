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
type DescribeAppRequest struct {
	AppName string `json:"app_name"`
}

func (o DescribeAppRequest) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "DescribeAppRequest struct{}"
	}

	return strings.Join([]string{"DescribeAppRequest", string(data)}, " ")
}
