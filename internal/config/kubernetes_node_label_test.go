// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package config

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/datakit"
)

func TestResolveKubernetesNodeLabelHostTags(t *testing.T) {
	originalDocker := datakit.Docker
	t.Cleanup(func() { datakit.Docker = originalDocker })

	datakit.Docker = true
	t.Setenv("KUBERNETES_SERVICE_PORT", "443")
	t.Setenv("ENV_K8S_NODE_NAME", "node-a")

	c := DefaultConfig()
	c.GlobalHostTags = map[string]string{
		"region":    "cn-east-1",
		"node_pool": "__k8s_node_label:example.com/node-pool",
		"missing":   "__k8s_node_label:example.com/missing",
	}

	getCalls := 0
	c.resolveKubernetesNodeLabelHostTags(context.Background(), func(_ context.Context, nodeName string) (map[string]string, error) {
		getCalls++
		require.Equal(t, "node-a", nodeName)
		return map[string]string{"example.com/node-pool": "pool-a"}, nil
	})

	require.Equal(t, 1, getCalls)
	require.Equal(t, map[string]string{
		"region":    "cn-east-1",
		"node_pool": "pool-a",
	}, c.GlobalHostTags)
}
