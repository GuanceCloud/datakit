// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.
// Some code modified from project Datadog (https://www.datadoghq.com/).

// Package traps contains snmp traps server logical.
package traps

import (
	"errors"
	"fmt"
	"hash/fnv"
	"net"
	"time"

	"github.com/gosnmp/gosnmp"
	"gitlab.jiagouyun.com/cloudcare-tools/cliutils/logger"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs/snmp/snmputil"
)

const (
	// Make sure to have at least some unique bytes for the authoritative engineID.
	// Unlikely to happen since the agent cannot start without a hostname.
	defaultAgentHostname = "datakit-agent"
	defaultNamespace     = "default"
)

// SnmpPacket is the type of packets yielded by server listeners.
type SnmpPacket struct {
	Content   *gosnmp.SnmpPacket
	Addr      *net.UDPAddr
	Timestamp int64
}

// PacketsChannel is the type of channels of trap packets.
type PacketsChannel = chan *SnmpPacket

// TrapServer manages an SNMP trap listener.
type TrapServer struct {
	Addr     string
	config   TrapsServerOpt
	listener *TrapListener
	sender   *TrapForwarder
}

var (
	serverInstance *TrapServer
	errStart       error
)

// StartServer starts the global trap server.
func StartServer(c *TrapsServerOpt) error {
	l = logger.SLogger(packageName)

	// internal initialize
	if err := checkDefaultConfig(c, defaultAgentHostname); err != nil {
		return err
	}

	oidResolver, err := NewMultiFilesOIDResolver()
	if err != nil {
		return err
	}
	formatter, err := NewJSONFormatter(oidResolver, c.Namespace)
	if err != nil {
		return err
	}
	server, err := NewTrapServer(*c, formatter)
	serverInstance = server
	errStart = err
	return err
}

func checkDefaultConfig(c *TrapsServerOpt, agentHostName string) error {
	if !c.Enabled {
		return errors.New("traps listener is disabled")
	}

	// gosnmp only supports one v3 user at the moment.
	if len(c.Users) > 1 {
		return errors.New("only one user is currently supported in SNMP Traps Listener configuration")
	}

	// Set defaults.
	if c.Port == 0 {
		c.Port = defaultPort
	}
	if c.BindHost == "" {
		// Default to global bind_host option.
		c.BindHost = "0.0.0.0"
	}
	if c.StopTimeout == 0 {
		c.StopTimeout = defaultStopTimeout
	}

	if agentHostName == "" {
		// Make sure to have at least some unique bytes for the authoritative engineID.
		// Unlikely to happen since the agent cannot start without a hostname
		agentHostName = "unknown-datakit-agent"
	}
	h := fnv.New128()
	h.Write([]byte(agentHostName))
	// First byte is always 0x80
	// Next four bytes are the Private Enterprise Number (set to an invalid value here)
	// The next 16 bytes are the hash of the agent hostname
	engineID := h.Sum([]byte{0x80, 0xff, 0xff, 0xff, 0xff})
	c.authoritativeEngineID = string(engineID)

	if c.Namespace == "" {
		c.Namespace = defaultNamespace
	}
	var err error
	c.Namespace, err = snmputil.NormalizeNamespace(c.Namespace)
	if err != nil {
		return fmt.Errorf("unable to load config: %w", err)
	}

	if c.CommunityStrings == nil {
		c.CommunityStrings = []string{}
	}

	if c.Users == nil {
		c.Users = []UserV3{}
	}

	return nil
}

// StopServer stops the global trap server, if it is running.
func StopServer() {
	if serverInstance != nil {
		serverInstance.Stop()
		serverInstance = nil
		errStart = nil
	}
}

// IsRunning returns whether the trap server is currently running.
func IsRunning() bool {
	return serverInstance != nil
}

// NewTrapServer configures and returns a running SNMP traps server.
func NewTrapServer(opt TrapsServerOpt, formatter Formatter) (*TrapServer, error) {
	packets := make(PacketsChannel, packetsChanSize)

	listener, err := startSNMPTrapListener(opt, packets)
	if err != nil {
		return nil, err
	}

	trapForwarder, err := startSNMPTrapForwarder(formatter, packets, opt.Election)
	if err != nil {
		return nil, fmt.Errorf("unable to start trapForwarder: %w. Will not listen for SNMP traps", err)
	}
	server := &TrapServer{
		listener: listener,
		config:   opt,
		sender:   trapForwarder,
	}

	return server, nil
}

func startSNMPTrapForwarder(formatter Formatter, packets PacketsChannel, election bool) (*TrapForwarder, error) {
	trapForwarder, err := NewTrapForwarder(formatter, packets, election)
	if err != nil {
		return nil, err
	}
	trapForwarder.Start()
	return trapForwarder, nil
}

func startSNMPTrapListener(opt TrapsServerOpt, packets PacketsChannel) (*TrapListener, error) {
	trapListener, err := NewTrapListener(opt, packets)
	if err != nil {
		return nil, err
	}
	err = trapListener.Start()
	if err != nil {
		return nil, err
	}
	return trapListener, nil
}

// Stop stops the TrapServer.
func (s *TrapServer) Stop() {
	stopped := make(chan interface{})

	go func() {
		l.Infof("Stop listening on %s", s.config.Addr())
		s.listener.Stop()
		s.sender.Stop()
		close(stopped)
	}()

	select {
	case <-stopped:
	case <-time.After(time.Duration(s.config.StopTimeout) * time.Second):
		l.Errorf("Stopping server. Timeout after %d seconds", s.config.StopTimeout)
	}
}
