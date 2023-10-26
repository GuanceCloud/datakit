// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package testutils

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestExternalIP(t *testing.T) {
	t.Run("external-ip", func(t *testing.T) {
		ip, err := ExternalIP()

		assert.NoError(t, err)

		t.Logf("external IP: %s", ip)
	})
}

func TestGetPort(t *testing.T) {
	t.Run("base-100", func(t *testing.T) {
		for i := 0; i < 10; i++ {
			get := RandPort("tcp")
			assert.True(t, get > baseOffset)
			t.Logf("get: %d -> %d", i, get)
		}

		for _, i := range []int{1234, 6379, 3306, 2375} {
			get := RandPort("tcp")
			assert.True(t, get > baseOffset)
			t.Logf("get: %d -> %d", i, get)
		}
	})

	t.Run("larger-than-base", func(t *testing.T) {
		for i := baseOffset; i < baseOffset+10; i++ {
			get := RandPort("tcp")
			assert.True(t, get > baseOffset)
			t.Logf("get: %d -> %d", i, get)
		}
	})

	t.Run("larger-than-max", func(t *testing.T) {
		for i := maxPort; i < maxPort+10; i++ {
			get := RandPort("tcp")
			assert.True(t, get > baseOffset)
			t.Logf("get: %d -> %d", i, get)
		}
	})

	t.Run("get-udp-port", func(t *testing.T) {
		conn, get, err := RandPortUDP()
		assert.NoError(t, err)
		defer conn.Close()
		assert.True(t, get > 0)
		t.Logf("get: %d", get)
	})
}

func TestGetContainerName(t *testing.T) {
	cases := []struct {
		name string
	}{
		{
			name: "myrepo/nginx:1.8.0",
		},
		{
			name: "nginx",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			out := GetUniqueContainerName(tc.name)

			t.Logf("name: %s", out)
		})
	}
}

func TestPurgeRemoteByName(t *testing.T) {
	if !CheckIntegrationTestingRunning() {
		t.Skip()
	}

	cases := []struct {
		name string
	}{
		{
			name: "nginx",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := PurgeRemoteByName(tc.name)
			require.NoError(t, err)
		})
	}
}
