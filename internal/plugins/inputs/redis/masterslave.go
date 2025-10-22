// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package redis

import (
	"context"
	"fmt"
	"net"
)

func (ipt *Input) setupMasterSlave(ctx context.Context) (*instance, error) {
	if len(ipt.MasterSlave.Hosts) == 0 {
		return nil, fmt.Errorf("master_slave.hosts is empty")
	}

	// First host is master
	masterAddr := ipt.MasterSlave.Hosts[0]
	host, _, err := net.SplitHostPort(masterAddr)
	if err != nil {
		return nil, fmt.Errorf("invalid master address %s: %w", masterAddr, err)
	}

	inst := newInstance()
	inst.ipt = ipt
	inst.mode = modeMasterSlave
	inst.addr = masterAddr
	inst.host = host
	inst.setup()

	// Add replicas (clients will be initialized by initializeClients)
	for i := 1; i < len(ipt.MasterSlave.Hosts); i++ {
		replicaAddr := ipt.MasterSlave.Hosts[i]
		repHost, _, err := net.SplitHostPort(replicaAddr)
		if err != nil {
			l.Warnf("invalid replica address %s: %s, skipped", replicaAddr, err)
			continue
		}

		rep := &replica{
			host: repHost,
			addr: replicaAddr,
		}

		inst.replicas = append(inst.replicas, rep)
	}

	return inst, nil
}
