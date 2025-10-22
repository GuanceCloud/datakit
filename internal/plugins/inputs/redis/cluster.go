// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package redis

import (
	"context"
	"fmt"
	"net/netip"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"
)

func (ipt *Input) setupCluster(ctx context.Context) error {
	crdb := redis.NewClusterClient(&redis.ClusterOptions{
		Addrs:        ipt.Cluster.Hosts,
		MaxRedirects: 3,
		ClientName:   "datakit",
		TLSConfig:    ipt.tlsConf,
		Username:     ipt.Username,
		Password:     ipt.Password,
		PoolSize:     3,
		MinIdleConns: 1,
		ReadTimeout:  3 * time.Second,
		WriteTimeout: 3 * time.Second,
		DialTimeout:  5 * time.Second,
	})

	if _, err := crdb.Ping(ctx).Result(); err != nil {
		l.Warnf("cluster rdb.Ping: %s", err.Error())
		_ = crdb.Close()
		return err
	}

	l.Infof("setup cluster on %+#v ok", ipt.Cluster.Hosts)

	ipt.crdb = crdb
	return nil
}

func (ipt *Input) scanClusterMasters(ctx context.Context) ([]*instance, error) {
	var instances []*instance

	err := ipt.crdb.ForEachMaster(ctx, func(ctx context.Context, master *redis.Client) error {
		mi, err := master.ClusterNodes(ctx).Result()
		if err != nil {
			return fmt.Errorf("failed to get master node %s(CLUSTER NDOES): %w", master.Options().Addr, err)
		}

		masterID, err := parseNodeID(mi, master.Options().Addr)
		if err != nil {
			return fmt.Errorf("failed to parse master node %s ID: %w", master.Options().Addr, err)
		}

		host, err := parseIPPort(master.Options().Addr)
		if err != nil {
			l.Errorf("parseIPPort: %s", err.Error())
			return err
		}

		inst := newInstance()

		inst.ipt = ipt
		inst.mode = modeCluster
		inst.id = masterID
		inst.addr = master.Options().Addr
		inst.host = host
		inst.setup()

		instances = append(instances, inst)

		l.Infof("get master node %q/%q", inst.addr, inst.id)

		replicas, err := ipt.crdb.Do(ctx, "CLUSTER", "REPLICAS", inst.id).StringSlice()
		if err != nil {
			// Fallback to CLUSTER SLAVES for Redis 4.0 compatibility
			replicas, err = ipt.crdb.Do(ctx, "CLUSTER", "SLAVES", inst.id).StringSlice()
			if err != nil {
				return fmt.Errorf("failed to get cluster replicas of %q: %w", inst.id, err)
			}
			l.Debugf("CLUSTER SLAVES %q: %q", inst.id, replicas)
		} else {
			l.Debugf("CLUSTER REPLICAS %q: %q", inst.id, replicas)
		}

		// Parse replicas info
		for _, r := range replicas {
			arr := strings.Fields(r)
			// replica example:
			//  "<replica-id> 172.19.0.5:6379@16379 slave,fail? <master-id> 1754373467516 1754373464341 2 connected"
			if len(arr) != 8 {
				l.Warnf("invalid replica: %q, skipped", r)
				continue
			}

			replicaID, replicaAddr, replicaFlags, linkState := arr[0], arr[1], arr[2], arr[7]
			if atIndex := strings.LastIndex(replicaAddr, "@"); atIndex != -1 {
				replicaAddr = replicaAddr[:atIndex]
			}

			// TODO: check flags and state
			_ = replicaFlags
			_ = linkState
			host, err := parseIPPort(replicaAddr)
			if err != nil {
				l.Errorf("parseIPPort: %s", err.Error())
				return err
			}

			l.Infof("found %q replica %s", inst.addr, replicaAddr)
			inst.replicas = append(inst.replicas, &replica{
				id:       replicaID,
				masterID: inst.id,
				addr:     replicaAddr,
				host:     host,
			})
		}
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("failed to loop master nodes: %w", err)
	}

	return instances, nil
}

func parseIPPort(s string) (string, error) {
	addr, err := netip.ParseAddrPort(s)
	if err != nil {
		return "", err
	}

	return addr.Addr().String(), nil
}

func parseNodeID(s, addr string) (string, error) {
	lines := strings.Split(s, "\n")
	for _, ln := range lines {
		if strings.Contains(ln, "myself") && strings.Contains(ln, addr) {
			arr := strings.Fields(ln)
			if len(arr) > 0 {
				return arr[0], nil
			}
		}
	}

	return "", fmt.Errorf("node ID of %q not founds", addr)
}
