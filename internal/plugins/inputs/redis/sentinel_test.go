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

func Test_setupSentinel(t *T.T) {
	t.Skip("skip test based on real redis sentinel")
	ipt := defaultInput()

	ipt.MasterSlave = &redisMasterSlave{
		Hosts: []string{},
		Sentinel: &redisSentinel{
			Hosts: []string{
				"centos.orb.local:26380",
				"centos.orb.local:26381",
				"centos.orb.local:26382",
			},
			Password:   "123456abc",
			MasterName: "mymaster",
		},
	}

	ipt.Password = "abc123456"

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	require.NoError(t, ipt.setupSentinel(ctx))
	master, err := ipt.sentinelDiscoverMaster(ctx)
	assert.NoError(t, err)

	t.Logf("master: %s", master.addr)
	for _, x := range master.replicas {
		t.Logf("replica: %+#v", x)
	}
}
