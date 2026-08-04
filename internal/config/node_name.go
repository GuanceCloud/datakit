// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package config

import (
	"fmt"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/datakit"
)

// GetLocalNodeName returns the Kubernetes node name injected into the DataKit
// Pod. ENV_K8S_NODE_NAME takes precedence over the deprecated NODE_NAME.
func GetLocalNodeName() (string, error) {
	if nodeName := localNodeNameFromEnv(); nodeName != "" {
		return nodeName, nil
	}

	return "", fmt.Errorf("invalid ENV_K8S_NODE_NAME environment, cannot be empty")
}

func localNodeNameFromEnv() string {
	for _, env := range []string{
		"ENV_K8S_NODE_NAME",
		"NODE_NAME", // Deprecated
	} {
		if value := datakit.GetEnv(env); value != "" {
			return value
		}
	}

	return ""
}
