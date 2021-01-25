/*
 * DevStar
 *
 * DevStar API
 *
 */

package model

import (
	"encoding/json"

	"strings"
)

// Response Object
type DownloadApplicationCodeResponse struct {
	HttpStatusCode int `json:"-"`
}

func (o DownloadApplicationCodeResponse) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "DownloadApplicationCodeResponse struct{}"
	}

	return strings.Join([]string{"DownloadApplicationCodeResponse", string(data)}, " ")
}
