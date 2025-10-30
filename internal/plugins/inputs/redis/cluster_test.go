// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package redis

import (
	"context"
	T "testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_clusterSetup(t *T.T) {
	t.Skip("skip-on-real-redis")
	ipt := defaultInput()
	ipt.Password = "abc123456"
	ipt.Cluster = &redisCluster{
		Hosts: []string{
			"centos.orb.local:7003",
		},
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	require.NoError(t, ipt.setupCluster(ctx))
	arr, err := ipt.scanClusterMasters(ctx)
	assert.NoError(t, err)

	t.Logf("scanned %d tartget", len(arr))
	for _, x := range arr {
		t.Logf("target: %s", x)
	}
}
