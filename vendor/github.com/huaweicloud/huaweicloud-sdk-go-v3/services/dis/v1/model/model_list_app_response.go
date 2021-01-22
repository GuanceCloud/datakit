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
type ListAppResponse struct {
	// 是否还有更多满足条件的App。  - true：是。 - false：否。
	HasMoreApp *bool `json:"has_more_app,omitempty"`
	// AppEntry list that meets the current request.
	Apps           *[]DescribeAppResult `json:"apps,omitempty"`
	HttpStatusCode int                  `json:"-"`
}

func (o ListAppResponse) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "ListAppResponse struct{}"
	}

	return strings.Join([]string{"ListAppResponse", string(data)}, " ")
}
