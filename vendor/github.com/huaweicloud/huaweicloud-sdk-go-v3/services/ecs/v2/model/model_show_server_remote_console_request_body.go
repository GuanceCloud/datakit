/*
 * ECS
 *
 * ECS Open API
 *
 */

package model

import (
	"encoding/json"

	"strings"
)

// This is a auto create Body Object
type ShowServerRemoteConsoleRequestBody struct {
	RemoteConsole *GetServerRemoteConsoleOption `json:"remote_console"`
}

func (o ShowServerRemoteConsoleRequestBody) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "ShowServerRemoteConsoleRequestBody struct{}"
	}

	return strings.Join([]string{"ShowServerRemoteConsoleRequestBody", string(data)}, " ")
}
