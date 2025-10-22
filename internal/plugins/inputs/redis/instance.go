// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package redis

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

type instance struct {
	mode,
	id,
	addr,
	host string

	curRepplica *replica
	replicas    []*replica

	mergedTags,
	infoTags map[string]string // tags during parsing redis info

	latencyLastTime map[string]time.Time

	infoElapsed time.Duration

	version, role string

	slowlogHash [][16]byte

	hbScanners      []*hotbigkeyScanner
	connectedSlaves int64

	ipt *Input

	infoCPULast map[string]*redisCPUUsage

	curCli, cc collectorClient
}

// node represents a concrete Redis node (master or replica) to collect from.
// It contains only the required information for a single collection run.
type node struct {
	host string
	addr string
	cli  collectorClient
	rep  *replica
}

func newInstance() *instance {
	return &instance{
		latencyLastTime: map[string]time.Time{},
		infoTags:        map[string]string{},
		mergedTags:      map[string]string{},
		infoCPULast:     map[string]*redisCPUUsage{},
	}
}

func (i *instance) setup() {
	i.slowlogHash = make([][16]byte, i.ipt.SlowlogMaxLen)

	i.mergedTags["host"] = i.host
	i.mergedTags["server"] = i.addr

	if i.ipt != nil {
		for k, v := range i.ipt.Tags { // add input's tags
			if _, ok := i.mergedTags[k]; !ok {
				i.mergedTags[k] = v
			}
		}
	}
}

func (i *instance) stop() {
	// close master client
	if i.cc != nil {
		if err := i.cc.close(); err != nil {
			l.Warnf("failed to close main client for %s: %s", i.String(), err)
		}
	}

	// close replica clients
	for _, replica := range i.replicas {
		if replica.cc != nil {
			if err := replica.cc.close(); err != nil {
				l.Warnf("failed to close replica client for %s: %s", replica.String(), err)
			}
		}
	}

	l.Infof("redis instance %s closed all connections", i.String())
}

func (i *instance) String() string {
	return fmt.Sprintf("mode: %s | id: %s | addr: %s | replicas: %d",
		i.mode, i.id, i.addr, len(i.replicas))
}

type replica struct {
	id,
	masterID,
	addr,
	host string
	cc collectorClient
}

func (r *replica) String() string {
	return fmt.Sprintf("id: %s | master %s | addr: %s", r.id, r.masterID, r.addr)
}

func (i *instance) resetReplica() {
	i.curRepplica = nil
	i.curCli = nil
}

func (i *instance) setCurrentNode(cli collectorClient, rep *replica, host, addr string) {
	i.curCli = cli
	i.curRepplica = rep
	i.mergedTags["host"] = host
	i.mergedTags["server"] = addr
}

func (i *instance) collect(ctx context.Context) error {
	// Collect metrics on master and all replicas.
	for _, n := range i.nodes() {
		i.setCurrentNode(n.cli, n.rep, n.host, n.addr)

		// metrics
		i.collectInfo(ctx)
		i.collectClientList(ctx)
		i.collectCommandStats(ctx)
		i.collectReplica(ctx)
		i.collectDB(ctx)

		// logging
		// i.collectConfig(ctx)
		i.collectLatency(ctx)
		i.collectSlowLog(ctx)
	}

	// Collect cluster info from master
	i.setCurrentNode(i.cc, nil, i.host, i.addr)
	i.collectCluster(ctx)

	i.resetReplica()

	return nil
}

// nodes enumerates the master and all valid replicas as a flat slice of node.
func (i *instance) nodes() []node {
	out := make([]node, 0, 1+len(i.replicas))
	if i.cc != nil {
		out = append(out, node{host: i.host, addr: i.addr, cli: i.cc, rep: nil})
	}
	for _, r := range i.replicas {
		if r != nil && r.cc != nil {
			out = append(out, node{host: r.host, addr: r.addr, cli: r.cc, rep: r})
		}
	}
	return out
}

// collectorClient defines interfaces that redis client and redis cluster client using.
type collectorClient interface {
	ping(ctx context.Context) (string, error)
	info(ctx context.Context, sections ...string) (string, error)
	clientList(ctx context.Context) (string, error)
	configGet(ctx context.Context, params string) (map[string]string, error)
	scanKeys(ctx context.Context, cursor uint64, match string, batchSize int64) ([]string, uint64, error)
	newPipeline() redis.Pipeliner
	clusterInfo(ctx context.Context) (string, error)
	do(ctx context.Context, args ...any) *redis.Cmd

	close() error
}

type colclient struct{ rdb *redis.Client }

func (c *colclient) ping(ctx context.Context) (string, error)       { return c.rdb.Ping(ctx).Result() }
func (c *colclient) close() error                                   { return c.rdb.Close() }
func (c *colclient) do(ctx context.Context, args ...any) *redis.Cmd { return c.rdb.Do(ctx, args...) }
func (c *colclient) info(ctx context.Context, sections ...string) (string, error) {
	return c.rdb.Info(ctx, sections...).Result()
}

func (c *colclient) clientList(ctx context.Context) (string, error) {
	return c.rdb.ClientList(ctx).Result()
}

func (c *colclient) configGet(ctx context.Context, params string) (map[string]string, error) {
	return c.rdb.ConfigGet(ctx, params).Result()
}

func (c *colclient) scanKeys(ctx context.Context, cursor uint64, match string, batchSize int64) ([]string, uint64, error) {
	return c.rdb.Scan(ctx, cursor, match, batchSize).Result()
}
func (c *colclient) newPipeline() redis.Pipeliner { return c.rdb.Pipeline() }
func (c *colclient) clusterInfo(ctx context.Context) (string, error) {
	return c.rdb.ClusterInfo(ctx).Result()
}
