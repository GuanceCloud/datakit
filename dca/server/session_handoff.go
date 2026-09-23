// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package server

import (
	"strconv"
	"time"

	"github.com/gorilla/websocket"
)

const sessionProbeTimeout = 3 * time.Second

// prepareRegistration checks an existing socket before rejecting a reconnect.
// A proxy may keep the old server-side socket open after the client reconnects.
// Only an answer to this fresh ping proves that the old socket can still reply.
func (manager *ClientManager) prepareRegistration(incoming *Client) bool {
	owner, ok := manager.getClient(incoming.ID)
	if !ok {
		return true
	}
	if owner.exited.Load() {
		incoming.replacementFor = owner
		return true
	}
	// Bound outstanding probes to one per owner, even during a reconnect storm.
	if !owner.probing.CompareAndSwap(false, true) {
		l.Infof("session probe already pending: incoming={%s} owner={%s}", incoming.logContext, owner.logContext)
		return false
	}
	defer owner.probing.Store(false)

	owner.Lock()
	owner.probeNumber++
	token := "dca-session-probe-" + strconv.FormatUint(owner.probeNumber, 10)
	pong := make(chan struct{}, 1)
	owner.probeToken, owner.probePong = token, pong
	owner.Unlock()
	defer func() {
		owner.Lock()
		owner.probeToken, owner.probePong = "", nil
		owner.Unlock()
	}()

	started := time.Now()
	deadline := started.Add(sessionProbeTimeout)
	l.Infof("probing previous session: incoming={%s} owner={%s}", incoming.logContext, owner.logContext)
	if err := owner.Socket.WriteControl(websocket.PingMessage, []byte(token), deadline); err != nil {
		l.Infof("session probe write failed: %s, owner={%s}", err, owner.logContext)
		incoming.replacementFor = owner
		return true
	}

	timer := time.NewTimer(time.Until(deadline))
	defer timer.Stop()
	select {
	case <-pong:
		l.Infof("duplicate rejected: previous session answered probe in %s, incoming={%s} owner={%s}",
			time.Since(started), incoming.logContext, owner.logContext)
		return false
	case <-incoming.Close:
		return false
	case <-owner.Close:
		l.Infof("previous session exited during probe: %s", owner.logContext)
	case <-timer.C:
		l.Infof("previous session did not answer probe within %s: %s", sessionProbeTimeout, owner.logContext)
	}
	incoming.replacementFor = owner
	return true
}

func (c *Client) receiveProbePong(data string) {
	c.Lock()
	defer c.Unlock()
	if c.probePong != nil && data == c.probeToken {
		select {
		case c.probePong <- struct{}{}:
		default:
		}
	}
}

func (manager *ClientManager) heartbeatClient(client *Client) {
	manager.RLock()
	defer manager.RUnlock()
	if manager.Clients[client.ID] != client || client.exited.Load() {
		return
	}
	if err := datakitDB.Heartbeat(client.ID); err != nil {
		l.Warnf("failed to heartbeat: %s, %s", err.Error(), client.logContext)
	}
}
