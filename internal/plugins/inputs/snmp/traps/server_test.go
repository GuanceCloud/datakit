// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.
// Some code modified from project Datadog (https://www.datadoghq.com/).

package traps

import (
	"net"
	"testing"

	"github.com/stretchr/testify/require"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs/snmp/snmprefiles"
)

func TestStopRemovesRegisteredServer(t *testing.T) {
	oldConfdDir := datakit.ConfdDir
	datakit.ConfdDir = t.TempDir()
	StopServer()
	t.Cleanup(func() {
		StopServer()
		datakit.ConfdDir = oldConfdDir
	})
	require.NoError(t, snmprefiles.ReleaseFiles())

	conn, err := net.ListenUDP("udp4", &net.UDPAddr{IP: net.ParseIP("127.0.0.1")})
	require.NoError(t, err)
	port := uint16(conn.LocalAddr().(*net.UDPAddr).Port)
	require.NoError(t, conn.Close())

	options := &TrapsServerOpt{
		Enabled:          true,
		BindHost:         "127.0.0.1",
		Port:             port,
		CommunityStrings: []string{"public"},
		StopTimeout:      1,
	}

	server, err := StartServer(options)
	require.NoError(t, err)
	require.NotNil(t, server)

	duplicate, err := StartServer(options)
	require.NoError(t, err)
	require.Nil(t, duplicate)

	server.Stop()

	restarted, err := StartServer(options)
	require.NoError(t, err)
	require.NotNil(t, restarted)
	restarted.Stop()
}
