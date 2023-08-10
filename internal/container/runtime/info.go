// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package runtime

import (
	"encoding/json"
	"fmt"
	"strings"
)

type criInfo struct {
	Pid    int `json:"pid"`
	Config struct {
		Envs []env `json:"envs"`
	} `json:"config"`
}

type env struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

func (c *criInfo) getPid() int {
	return c.Pid
}

func (c *criInfo) getConfigEnvs() map[string]string {
	m := make(map[string]string, len(c.Config.Envs))
	for _, env := range c.Config.Envs {
		m[env.Key] = m[env.Value]
	}
	return m
}

// parseCriInfo return CRI info struct,
// https://github.com/cri-o/cri-o/issues/3604
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

// parseDockerEnv, return the value corresponding to this key in 'envs'
//    example: "PATH=/usr/local/sbin:/usr/local/bin", return "PATH : /usr/local/sbin:/usr/local/bin"
func parseDockerEnv(envs []string) map[string]string {
	m := make(map[string]string, len(envs))
	for _, envStr := range envs {
		array := strings.Split(envStr, "=")
		if len(array) < 1 {
			continue
		}
		key := array[0]
		value := strings.Join(array[1:], "")
		m[key] = value
	}
	return m
}
