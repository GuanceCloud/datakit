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

// Response Object
type ListAppV3Response struct {
	// 是否还有更多满足条件的App。  - true：是。 - false：否。
	HasMoreApp *bool `json:"has_more_app,omitempty"`
	// AppEntry list that meets the current request.
	Apps           *[]DescribeAppResult `json:"apps,omitempty"`
	HttpStatusCode int                  `json:"-"`
}

func (o ListAppV3Response) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "ListAppV3Response struct{}"
	}

	return strings.Join([]string{"ListAppV3Response", string(data)}, " ")
}
