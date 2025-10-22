// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package redis

import (
	"bufio"
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/GuanceCloud/cliutils/point"
	dkio "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/io"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs"
)

type clientListCollector struct {
	CollectLogOnFlags string `toml:"log_on_flags"`
}

type logLevel int

const (
	lDebug logLevel = iota
	lOk
	lInfo
	lNotice
	lWarn
	lError
	lCritical
)

var (
	// [flag:point-key] name mapping.
	flagKeys = map[uint8]string{
		'A': `flag_closing_asap`,
		'B': `flag_tracking_broadcast_mode`,
		'M': `flag_master_connection`,
		'N': `flag_normal`,
		'O': `flag_monitor_mode`,
		'P': `flag_pubsub_subscriber`,
		'R': `flag_tracking_redir_invalid`,
		'S': `flag_replica_connection`,
		'T': `flag_tracking_no_touch`,
		'U': `flag_unix_socket`,
		'b': `flag_blocked_on_command`,
		'c': `flag_closing_after_reply`,
		'd': `flag_transaction_dirty`,
		'e': `flag_eviction_exempt`,
		'i': `flag_vm_io_wait`,
		'r': `flag_cluster_readonly`,
		't': `flag_client_side_caching`,
		'u': `flag_unblocked`,
		'x': `flag_in_transaction`,
	}

	flagLogLevels = map[uint8]logLevel{
		'A': lOk,
		'B': lInfo,
		'M': lOk,
		'N': lOk,
		'O': lWarn,
		'P': lOk,
		'R': lWarn,
		'S': lOk,
		'T': lInfo,
		'U': lInfo,
		'b': lCritical,
		'c': lDebug,
		'd': lDebug,
		'e': lInfo,
		'i': lDebug,
		'r': lInfo,
		't': lInfo,
		'u': lDebug,
		'x': lWarn,
	}
)

func (l logLevel) String() string {
	switch l {
	case lDebug:
		return "debug"
	case lOk:
		return "ok"
	case lInfo:
		return "info"
	case lNotice:
		return "notice"
	case lWarn:
		return "warn"
	case lError:
		return "error"
	case lCritical:
		return "critical"
	default:
		return "info"
	}
}

// logLevel get highest log level for client list flags.
func flagLogLevel(flags string) logLevel {
	ll := lDebug
	for _, f := range flags {
		if x, ok := flagLogLevels[uint8(f)]; ok && x > ll {
			ll = x // use higher log level
		}
	}

	return ll
}

func (clc *clientListCollector) asLogging(flags string) bool {
	for _, f := range flags { // NOTE: same client may have multiple flags here.
		for _, x := range clc.CollectLogOnFlags {
			if f == x {
				return true
			}
		}
	}

	l.Debugf("flag %q not as logging(expect %q)", flags, clc.CollectLogOnFlags)
	return false
}

type redisClientList struct {
	raw,
	addr, // address/port of the client
	laddr, // address/port of local address client connected to (bind address)
	flags, // client flags (see below)
	db, // current database ID
	events, // file descriptor events (see below)
	cmd, // last command played
	user, // the authenticated username of the client
	name string // the name set by the client with CLIENT SETNAME

	fd, // file descriptor corresponding to the socket
	id, // a unique 64-bit client ID
	age, // total duration of the connection in seconds
	idle, // idle time of the connection in seconds
	sub, // number of channel subscriptions
	psub, // number of pattern matching subscriptions
	ssub, // number of shard channel subscriptions. Added in Redis 7.0.3
	multi, // number of commands in a MULTI/EXEC context
	watch, // number of keys this client is currently watching. Added in Redis 7.4
	qbuf, // query buffer length (0 means no query pending)
	qbuffree, // free space of the query buffer (0 means the buffer is full)
	argvmem, // incomplete arguments for the next command (already extracted from query buffer)
	multimem, // memory is used up by buffered multi commands. Added in Redis 7.0
	obl, // output buffer length
	oll, // output list length (replies are queued in this list when the buffer is full)
	omem, // output buffer memory usage
	totmem, // total memory consumed by this client in its various buffers
	redir, // client id of current client tracking redirection
	resp, // client RESP protocol version. Added in Redis 7.0
	rbp, // peak size of the client's read buffer since the client connected. Added in Redis 7.0
	rbs, // current size of the client's read buffer in bytes. Added in Redis 7.0
	iothread, // id of I/O thread assigned to the client. Added in Redis 8.0
	totnetin, // total network input bytes read from this client.
	totnetout, // total network output bytes sent to this client.
	totcmds int64 // total count of commands this client executed.
}

func (i *instance) cliLoggingPoint(rc *redisClientList, opts ...point.Option) *point.Point {
	var kvs point.KVs

	kvs = kvs.Set("addr", rc.addr).
		Set("laddr", rc.laddr).
		Set("flags", rc.flags).
		Set("db", rc.db).
		Set("events", rc.events).
		Set("cmd", rc.cmd).
		Set("user", rc.user).
		Set("name", rc.name).
		Set("fd", rc.fd).
		Set("id", rc.id).
		Set("age", rc.age).
		Set("idle", rc.idle).
		Set("sub", rc.sub).
		Set("psub", rc.psub).
		Set("ssub", rc.ssub).
		Set("multi", rc.multi).
		Set("watch", rc.watch).
		Set("qbuf", rc.qbuf).
		Set("qbuf-free", rc.qbuffree).
		Set("argv-mem", rc.argvmem).
		Set("multi-mem", rc.multimem).
		Set("obl", rc.obl).
		Set("oll", rc.oll).
		Set("omem", rc.omem).
		Set("tot-mem", rc.totmem).
		Set("redir", rc.redir).
		Set("resp", rc.resp).
		Set("rbp", rc.rbp).
		Set("rbs", rc.rbs).
		Set("io-thread", rc.iothread).
		Set("tot-netin", rc.totnetin).
		Set("tot-netout", rc.totnetout).
		Set("tot-cmds", rc.totcmds).
		Set("status", flagLogLevel(rc.flags).String()).
		Set("message", rc.raw)

	for k, v := range i.mergedTags {
		kvs = kvs.AddTag(k, v)
	}

	return point.NewPoint(measureuemtRedisClientLogging, kvs, opts...)
}

func doParseClientList(list string) (clilist []*redisClientList, err error) {
	scanner := bufio.NewScanner(strings.NewReader(list))

	for scanner.Scan() {
		line := scanner.Text()
		if len(line) == 0 || line[0] == '#' {
			continue
		}

		parts := strings.Split(line, " ")
		rc := redisClientList{
			raw: line,
		}

		for _, p := range parts {
			kv := strings.Split(p, "=")
			if len(kv) != 2 {
				continue
			}

			if kv[1] == "" {
				continue // ignored
			}

			v, _ := strconv.ParseInt(kv[1], 10, 64)

			switch kv[0] {
			case "addr":
				rc.addr = kv[1]
			case "laddr":
				rc.laddr = kv[1]
			case "name":
				rc.name = kv[1]
			case "events":
				rc.events = kv[1]
			case "cmd":
				rc.cmd = kv[1]
			case "user":
				rc.user = kv[1]
			case "flags":
				rc.flags = kv[1]
			case "db":
				rc.db = fmt.Sprintf("db%s", kv[1])

			case "io-thread":
				rc.iothread = v
			case "fd":
				rc.fd = v
			case "id":
				rc.id = v
			case "tot-cmds":
				rc.totcmds = v
			case "tot-net-out":
				rc.totnetout = v
			case "tot-net-in":
				rc.totnetin = v
			case "rbs":
				rc.rbs = v
			case "rbp":
				rc.rbp = v
			case "resp":
				rc.resp = v
			case "redir":
				rc.redir = v
			case "age":
				rc.age = v
			case "idle":
				rc.idle = v
			case "sub":
				rc.sub = v
			case "psub":
				rc.psub = v
			case "ssub":
				rc.ssub = v
			case "multi":
				rc.multi = v
			case "watch":
				rc.watch = v
			case "qbuf":
				rc.qbuf = v
			case "qbuf-free":
				rc.qbuffree = v
			case "argv-mem":
				rc.argvmem = v
			case "multi-mem":
				rc.multimem = v
			case "obl":
				rc.obl = v
			case "oll":
				rc.oll = v
			case "omem":
				rc.omem = v
			case "tot-mem":
				rc.totmem = v
			default: // ignored
				l.Warnf("ignored key %q=%q", kv[0], v)
			}
		}

		clilist = append(clilist, &rc)
	}

	return clilist, nil
}

func (i *instance) buildcliPoints(arr []*redisClientList) (mpts, lpts []*point.Point) {
	var (
		totmemsum,
		totalnetout,
		totalsub,
		totalpsub,
		totalssub,
		totalnetin,
		totalobl,
		totalcmds,
		totalqbuf,
		totalmulti float64

		maxmulti,
		totmemmax,
		maxqbuf,
		maxobl,
		maxidle int64

		avgmulti float64

		lopts = append(point.DefaultLoggingOptions(), point.WithTime(i.ipt.ptsTime))
		mopts = append(point.DefaultMetricOptions(), point.WithTime(i.ipt.ptsTime))
	)

	flags := map[uint8]int64{}

	// See:
	//   https://redis.io/docs/latest/commands/client-list/
	// for more field info.

	for _, x := range arr {
		if i.ipt.ClientListCollector.asLogging(x.flags) {
			lpts = append(lpts, i.cliLoggingPoint(x, lopts...))
		}

		for _, f := range x.flags {
			flags[uint8(f)]++
		}

		if x.sub > 0 {
			totalsub += float64(x.sub)
		}

		if x.psub > 0 {
			totalpsub += float64(x.psub)
		}

		if x.ssub > 0 {
			totalssub += float64(x.ssub)
		}

		if x.multi > 0 {
			totalmulti += float64(x.multi)
			if x.multi > maxmulti {
				maxmulti = x.multi // find max-multi across current list
			}
		}

		if x.totmem > 0 {
			totmemsum += float64(x.totmem)
			if x.totmem > totmemmax {
				totmemmax = x.totmem
			}
		}

		if x.totnetin > 0 {
			totalnetin += float64(x.totnetin)
		}

		if x.totnetout > 0 {
			totalnetout += float64(x.totnetout)
		}

		if x.totcmds > 0 {
			totalcmds += float64(x.totcmds)
		}

		totalobl += float64(x.obl)
		if x.obl > maxobl {
			maxobl = x.obl
		}

		totalqbuf += float64(x.qbuf)
		if x.qbuf > maxqbuf {
			maxqbuf = x.qbuf
		}

		if x.idle > maxidle {
			maxidle = x.idle
		}
	}

	avgmulti = (totalmulti) / float64(len(arr))

	var kvs point.KVs

	kvs = kvs.
		Set("max_idle", maxidle).
		Set("max_multi", maxmulti).
		Set("max_obl", maxobl).
		Set("max_qbuf", maxqbuf).
		Set("max_totmem", totmemmax).
		Set("multi_avg", avgmulti).
		Set("multi_total", totalmulti).
		Set("total_cmds", totalcmds).
		Set("total_netin", totalnetin).
		Set("total_netout", totalnetout).
		Set("total_obl", totalobl).
		Set("total_psub", totalpsub).
		Set("total_qbuf", totalqbuf).
		Set("total_ssub", totalssub).
		Set("total_sub", totalsub).
		Set("total_totmem", totmemsum)

	for k, v := range flags {
		if pkey, ok := flagKeys[k]; ok {
			kvs = kvs.Set(pkey, v) // add human-readable flags
		} else {
			kvs = kvs.Add("flag_unknown_"+string(k), v)
		}
	}

	for k, v := range i.mergedTags {
		kvs = kvs.AddTag(k, v)
	}

	mpts = append(mpts, point.NewPoint(measureuemtRedisClientsStat, kvs, mopts...))

	return mpts, lpts
}

func (i *instance) collectClientList(ctx context.Context) {
	collectStart := time.Now()

	list, err := i.curCli.clientList(ctx)
	if err != nil {
		l.Error("client list get error: %s, ignored", err)
		return
	}

	clist, err := doParseClientList(list)
	if err != nil {
		l.Warnf("doParseClientList: %s, ignored", err.Error())
		return
	}

	mpts, lpts := i.buildcliPoints(clist)
	ccost := time.Since(collectStart)

	if len(mpts) > 0 {
		if err := i.ipt.feeder.Feed(point.Metric, mpts,
			dkio.WithCollectCost(ccost),
			dkio.WithElection(i.ipt.Election),
			dkio.WithSource(dkio.FeedSource(inputName, "client-stat")),
			dkio.WithMeasurement(inputs.GetOverrideMeasurement(i.ipt.MeasurementVersion, measureuemtRedis))); err != nil {
			l.Warnf("feed: %s, ignored", err)
		}
	}

	if len(lpts) > 0 {
		if err := i.ipt.feeder.Feed(point.Logging, lpts,
			dkio.WithCollectCost(ccost),
			dkio.WithElection(i.ipt.Election),
			dkio.WithSource(dkio.FeedSource(inputName, "client"))); err != nil {
			l.Warnf("feed: %s, ignored", err)
		}
	}
}

type clientLoggingMeasurement struct{}

func (clientLoggingMeasurement) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name: measureuemtRedisClient,
		Desc: "Client details comes from `CLIENT LIST` that on specific flags. " +
			"See [here](https://redis.io/docs/latest/commands/client-list){:target=\"_blank\"}.",
		DescZh: "`CLIENT LIST` 返回中带特定标志位（`flags`）的客户端详情，详细字段参见[这里](https://redis.io/docs/latest/commands/client-list){:target=\"_blank\"}。",
		Cat:    point.Logging,
		Tags: map[string]interface{}{
			"host":   &inputs.TagInfo{Desc: "Hostname."},
			"server": &inputs.TagInfo{Desc: "Server addr."},
		},
		Fields: map[string]interface{}{
			"status": &inputs.FieldInfo{
				DataType: inputs.String,
				Type:     inputs.NoType,
				Unit:     inputs.NoUnit,
				Desc:     "current database ID",
			},
			"message": &inputs.FieldInfo{
				DataType: inputs.String,
				Type:     inputs.NoType,
				Unit:     inputs.NoUnit,
				Desc:     "Raw text of origin client list",
			},
			"more-fields...": &inputs.FieldInfo{
				DataType: inputs.NoDataType,
				Type:     inputs.NoType,
				Unit:     inputs.NoUnit,
				Desc:     ``,
			},
		},
	}
}
