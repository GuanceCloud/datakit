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

// Response Object
type ListProjectIterationsV4Response struct {
	// 迭代总数
	Total *int32 `json:"total,omitempty"`
	// 迭代信息
	Iterations     *[]ListProjectVersionsV4ResponseBodyIterations `json:"iterations,omitempty"`
	HttpStatusCode int                                            `json:"-"`
}

func (o ListProjectIterationsV4Response) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "ListProjectIterationsV4Response struct{}"
	}

	return strings.Join([]string{"ListProjectIterationsV4Response", string(data)}, " ")
}
