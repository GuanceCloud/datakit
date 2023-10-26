// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package netflow

import (
	"net"
	"strconv"
	"testing"

	"github.com/BurntSushi/toml"
	"github.com/netsampler/goflow2/utils"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs/netflow/common"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs/netflow/goflowlib"
)

////////////////////////////////////////////////////////////////////////////////

func TestStartServerAndStopServer(t *testing.T) {
	port := GetFreePort()
	ipt := &Input{
		tagger: datakit.DefaultGlobalTagger(),
		Listeners: []common.FlowOpt{
			{
				Type: common.TypeNetFlow5,
				Port: port,
			},
		},
	}
	err := StartServer(ipt)
	require.NoError(t, err)
	require.NotNil(t, serverInstance)

	replaceWithDummyFlowProcessor(serverInstance, 123)

	StopServer()
	require.Nil(t, serverInstance)
}

func TestServer_Stop(t *testing.T) {
	port := GetFreePort()
	ipt := &Input{
		tagger: datakit.DefaultGlobalTagger(),
		Listeners: []common.FlowOpt{
			{
				Type: common.TypeNetFlow5,
				Port: port,
			},
		},
	}
	server, err := NewNetflowServer(ipt)
	require.NoError(t, err, "cannot start Netflow Server")
	assert.NotNil(t, server)

	flowProcessor := replaceWithDummyFlowProcessor(server, port)

	// Stops server
	server.stop()

	// Assert logs present
	assert.Equal(t, flowProcessor.stopped, true)
}

func TestParse(t *testing.T) {
	t.Run("Unmarshal", func(t *testing.T) {
		cfg := `
	namespace = "namespace"

	#[[listeners]]
	#flow_type = "netflow9"
	#port      = 2055

	[[listeners]]
	flow_type = "netflow5"
	port      = 2056

	#[[listeners]]
	#flow_type = "ipfix"
	#port      = 4739

	#[[listeners]]
	#flow_type = "sflow5"
	#port      = 6343

	[inputs.netflow.tags]
	# some_tag = "some_value"
	# more_tag = "some_other_value"
`
		ipt := &Input{}
		err := toml.Unmarshal([]byte(cfg), ipt)
		require.NoError(t, err)
		require.Equal(t, len(ipt.Listeners), 1)
	})
}

////////////////////////////////////////////////////////////////////////////////

type dummyFlowProcessor struct {
	receivedMessages chan interface{}
	stopped          bool
}

func (d *dummyFlowProcessor) FlowRoutine(workers int, addr string, port int, reuseport bool) error {
	return utils.UDPStoppableRoutine(make(chan struct{}), "test_udp", func(msg interface{}) error {
		d.receivedMessages <- msg
		return nil
	}, 3, addr, port, false, logrus.StandardLogger())
}

func (d *dummyFlowProcessor) Shutdown() {
	d.stopped = true
}

func replaceWithDummyFlowProcessor(server *Server, port uint16) *dummyFlowProcessor {
	// Testing using a dummyFlowProcessor since we can't test using real goflow flow processor
	// due to this race condition https://github.com/netsampler/goflow2/issues/83
	flowProcessor := &dummyFlowProcessor{}
	listener := server.listeners[0]
	listener.flowState = &goflowlib.FlowStateWrapper{
		State:    flowProcessor,
		Hostname: "abc",
		Port:     port,
	}
	return flowProcessor
}

////////////////////////////////////////////////////////////////////////////////

func GetFreePort() uint16 {
	var port uint16
	for i := 0; i < 5; i++ {
		conn, err := net.ListenPacket("udp", ":0")
		if err != nil {
			continue
		}
		conn.Close()
		port, err = parsePort(conn.LocalAddr().String())
		if err != nil {
			continue
		}
		return port
	}
	panic("unable to find free port for starting the trap listener")
}

func parsePort(addr string) (uint16, error) {
	_, portString, err := net.SplitHostPort(addr)
	if err != nil {
		return 0, err
	}

	port, err := strconv.ParseUint(portString, 10, 16)
	if err != nil {
		return 0, err
	}
	return uint16(port), nil
}

////////////////////////////////////////////////////////////////////////////////
