// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package redis

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"
)

func (ipt *Input) setupSentinel(ctx context.Context) error {
	for _, addr := range ipt.MasterSlave.Sentinel.Hosts {
		rdb := redis.NewSentinelClient(&redis.Options{
			Addr:         addr,
			ClientName:   "datakit",
			TLSConfig:    ipt.tlsConf,
			Password:     ipt.MasterSlave.Sentinel.Password,
			PoolSize:     3,
			MinIdleConns: 1,
			ReadTimeout:  3 * time.Second,
			WriteTimeout: 3 * time.Second,
			DialTimeout:  5 * time.Second,
		})

		if _, err := rdb.Ping(ctx).Result(); err != nil {
			l.Warnf("rdb.Ping: %s", err.Error())
			_ = rdb.Close()
			return err
		} else {
			l.Infof("connect to sentinel %q ok", addr)
			ipt.srdb = rdb
			return nil
		}
	}

	return fmt.Errorf("sentinel setup failed")
}

func (ipt *Input) sentinelDiscoverMaster(ctx context.Context) (*instance, error) {
	addr, err := ipt.srdb.GetMasterAddrByName(ctx, ipt.MasterSlave.Sentinel.MasterName).Result()
	if err != nil {
		return nil, fmt.Errorf("sentinel failed to get addr of master %q", ipt.MasterSlave.Sentinel.MasterName)
	}

	if len(addr) != 2 {
		return nil, fmt.Errorf("invalid master addr: %v", addr)
	}

	inst := newInstance()

	inst.ipt = ipt
	inst.mode = modeSentinel
	inst.addr = fmt.Sprintf("%s:%s", addr[0], addr[1])
	inst.host = addr[0]
	inst.setup()

	repInfo, err := ipt.srdb.Replicas(ctx, ipt.MasterSlave.Sentinel.MasterName).Result()
	if err != nil {
		// Fallback to SENTINEL SLAVES for Redis 4.0 compatibility
		cmd := redis.NewMapStringStringSliceCmd(ctx, "sentinel", "slaves", ipt.MasterSlave.Sentinel.MasterName)
		err = ipt.srdb.Process(ctx, cmd)
		if err != nil {
			// NOTE: no replicas is ok (single master mode)
			l.Warnf("failed to get replicas of master %s: %s, ignored", inst.addr, err)
			return inst, nil
		}
		repInfo, err = cmd.Result()
		if err != nil {
			l.Warnf("failed to get replicas result of master %s: %s, ignored", inst.addr, err)
			return inst, nil
		}
		l.Debugf("SENTINEL SLAVES for master %s", ipt.MasterSlave.Sentinel.MasterName)
	} else {
		l.Debugf("SENTINEL REPLICAS for master %s", ipt.MasterSlave.Sentinel.MasterName)
	}

	if len(repInfo) == 0 {
		l.Infof("no replicas found for master %s (single master mode)", inst.addr)
		return inst, nil
	}

	for _, kv := range repInfo {
		ip, ipok := kv["ip"]
		port, portok := kv["port"]
		flags, flagok := kv["flags"]

		// flags that not contains `s_down' or `o_down'
		if ipok && portok && flagok && !strings.Contains(flags, "down") {
			repaddr := fmt.Sprintf("%s:%s", ip, port)
			inst.replicas = append(inst.replicas, &replica{
				addr: repaddr,
				host: ip,
				// NOTE: other fields not used.
			})

			l.Infof("add replica %q", repaddr)
		} else {
			l.Warnf("skip invalid replica %+#v", kv)
		}
	}

	return inst, nil
}
