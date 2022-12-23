// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.
// Some code modified from project Datadog (https://www.datadoghq.com/).

package traps

import (
	"context"
	"fmt"
	"net"
	"time"

	"github.com/gosnmp/gosnmp"
)

//------------------------------------------------------------------------------

// TrapListener opens an UDP socket and put all received traps in a channel.
type TrapListener struct {
	config        TrapsServerOpt
	packets       PacketsChannel
	listener      *gosnmp.TrapListener
	errorsChannel chan error
}

// NewTrapListener creates a simple TrapListener instance but does not start it.
func NewTrapListener(opt TrapsServerOpt, packets PacketsChannel) (*TrapListener, error) {
	var err error
	gosnmpListener := gosnmp.NewTrapListener()
	gosnmpListener.Params, err = BuildSNMPParams(&opt)
	if err != nil {
		return nil, err
	}
	errorsChan := make(chan error, 1)
	trapListener := &TrapListener{
		config:        opt,
		packets:       packets,
		listener:      gosnmpListener,
		errorsChannel: errorsChan,
	}

	gosnmpListener.OnNewTrap = trapListener.receiveTrap
	return trapListener, nil
}

// Start the TrapListener instance. Need to be manually Stopped.
func (t *TrapListener) Start() error {
	l.Infof("Start listening for traps on %s", t.config.Addr())
	g.Go(func(ctx context.Context) error {
		t.run()
		return nil
	}) // Go
	return t.blockUntilReady()
}

func (t *TrapListener) run() {
	err := t.listener.Listen(t.config.Addr()) // blocking call
	if err != nil {
		t.errorsChannel <- err
	}
}

func (t *TrapListener) blockUntilReady() error {
	select {
	// Wait for listener to be started and listening to traps.
	// See: https://godoc.org/github.com/gosnmp/gosnmp#TrapListener.Listening
	case <-t.listener.Listening():
		return nil
	// If the listener failed to start (eg because it couldn't bind to a socket),
	// we'll get an error here.
	case err := <-t.errorsChannel:
		return fmt.Errorf("error happened when listening for SNMP Traps: %w", err)
	}
}

// Stop the current TrapListener instance.
func (t *TrapListener) Stop() {
	t.listener.Close()
}

func (t *TrapListener) receiveTrap(p *gosnmp.SnmpPacket, u *net.UDPAddr) {
	if err := validatePacket(p, &(t.config)); err != nil {
		l.Warnf("Invalid credentials from %s on listener %s, dropping traps", u.String(), t.config.Addr())
		trapsPacketsAuthErrors.Add(1)
		return
	}
	l.Debugf("Packet received from %s on listener %s", u.String(), t.config.Addr())
	trapsPackets.Add(1)
	t.packets <- &SnmpPacket{Content: p, Addr: u, Timestamp: time.Now().UnixMilli()}
}
