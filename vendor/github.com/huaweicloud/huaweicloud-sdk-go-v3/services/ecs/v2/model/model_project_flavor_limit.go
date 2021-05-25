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

//
type ProjectFlavorLimit struct {
}

func (o ProjectFlavorLimit) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "ProjectFlavorLimit struct{}"
	}

	return strings.Join([]string{"ProjectFlavorLimit", string(data)}, " ")
}
