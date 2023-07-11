// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package container

import (
	"encoding/json"
	"fmt"
)

type criInfo struct {
	SandboxID   string        `json:"sandboxID"`
	Pid         int64         `json:"pid"`
	RuntimeType string        `json:"runtimeType"`
	Config      criInfoConfig `json:"config"`
}

type criInfoConfig struct {
	Envs envVars `json:"envs"`
}

type envVars []envVar

func (ev envVars) Find(key string) string {
	for _, e := range ev {
		if e.Key == key {
			return e.Value
		}
	}
	return ""
}

type envVar struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

func parseCriInfo(in string) (*criInfo, error) {
	if in == "" {
		return nil, fmt.Errorf("info cannot be empty")
	}
	var info criInfo
	if err := json.Unmarshal([]byte(in), &info); err != nil {
		return nil, err
	}
	return &info, nil
}
