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

func Test_setupMasterSlave(t *T.T) {
	t.Skip("skip test based on real redis master-slave")
	ipt := defaultInput()

	ipt.MasterSlave = &redisMasterSlave{
		Hosts: []string{
			"192.168.139.225:6379", // master
			"192.168.139.225:6380", // slave-1
			"192.168.139.225:6381", // slave-2
		},
	}

	ipt.Password = "abc123456"

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	inst, err := ipt.setupMasterSlave(ctx)
	require.NoError(t, err)
	assert.NotNil(t, inst)

	t.Logf("master: %s", inst.addr)
	assert.Equal(t, 2, len(inst.replicas), "should have 2 replicas")
	for i, replica := range inst.replicas {
		t.Logf("replica-%d: %s", i+1, replica.addr)
	}
}
