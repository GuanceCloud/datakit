// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package redis

import (
	T "testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_flagLog(t *T.T) {
	t.Run(`basic`, func(t *T.T) {
		flags := `bxOPS`
		assert.Equal(t, lCritical, flagLogLevel(flags))

		flags = `N`
		ll := flagLogLevel(flags)
		assert.Equal(t, lOk, ll)

		flags = `B`
		ll = flagLogLevel(flags)
		assert.Equal(t, lInfo, ll)

		flags = `-` // unknown
		ll = flagLogLevel(flags)
		assert.Equal(t, lDebug, ll)
	})
}

func Test_doParseClientList(t *T.T) {
	t.Run(`basic`, func(t *T.T) {
		list := `id=1320 addr=172.20.0.1:36302 laddr=172.20.0.5:6379 fd=14 name= age=41 idle=11 flags=N db=0 sub=1 psub=1 ssub=1 multi=-1 watch=1 qbuf=1 qbuf-free=1 argv-mem=1 multi-mem=1 rbs=1024 rbp=1 obl=1 oll=1 omem=1 tot-mem=1928 events=r cmd=sismember user=default redir=-1 resp=2 lib-name= lib-ver=`
		clist, err := doParseClientList(list)
		assert.NoError(t, err)

		assert.Len(t, clist, 1)
		x := clist[0]

		assert.Equal(t, int64(1320), x.id)
		assert.Equal(t, "172.20.0.1:36302", x.addr)
		assert.Equal(t, "172.20.0.5:6379", x.laddr)
		assert.Equal(t, int64(14), x.fd)
		assert.Empty(t, x.name)
		assert.Equal(t, int64(11), x.idle)
		assert.Equal(t, int64(41), x.age)
		assert.Equal(t, "N", x.flags)
		assert.Equal(t, "db0", x.db)
		assert.Equal(t, int64(1), x.sub)
		assert.Equal(t, int64(1), x.psub)
		assert.Equal(t, int64(1), x.ssub)
		assert.Equal(t, int64(-1), x.multi)
		assert.Equal(t, int64(1), x.watch)
		assert.Equal(t, int64(1), x.qbuf)
		assert.Equal(t, int64(1), x.qbuffree)
		assert.Equal(t, int64(1), x.argvmem)
		assert.Equal(t, int64(1), x.multimem)
		assert.Equal(t, int64(1024), x.rbs)
		assert.Equal(t, int64(1), x.rbp)
		assert.Equal(t, int64(1), x.obl)
		assert.Equal(t, int64(1), x.oll)
		assert.Equal(t, int64(1), x.omem)
		assert.Equal(t, int64(1928), x.totmem)
		assert.Equal(t, "r", x.events)
		assert.Equal(t, "sismember", x.cmd)
		assert.Equal(t, "default", x.user)
		assert.Equal(t, int64(-1), x.redir)
		assert.Equal(t, int64(2), x.resp)
	})

	t.Run(`buildcli-logging-point`, func(t *T.T) {
		ipt := defaultInput()

		list := `
id=1320 addr=172.20.0.1:36302 laddr=172.20.0.5:6379 fd=14 name= age=41 idle=30 flags=N db=0 sub=3 psub=2 ssub=1 multi=-1 watch=1 qbuf=1 qbuf-free=1 argv-mem=1 multi-mem=1 rbs=1024 rbp=1 obl=1 oll=1 omem=1 tot-mem=1920 events=r cmd=sismember user=default redir=-1 resp=2 lib-name= lib-ver= tot-cmds=2 tot-net-in=1023 tot-net-out=1023
id=1320 addr=172.20.0.1:36302 laddr=172.20.0.5:6379 fd=14 name= age=41 idle=30 flags=N db=0 sub=3 psub=2 ssub=1 multi=3 watch=1 qbuf=1 qbuf-free=1 argv-mem=1 multi-mem=1 rbs=1024 rbp=1 obl=1 oll=1 omem=1 tot-mem=1921 events=r cmd=sismember user=default redir=-1 resp=2 lib-name= lib-ver= tot-cmds=2 tot-net-in=1021 tot-net-out=1021
id=1320 addr=172.20.0.1:36302 laddr=172.20.0.5:6379 fd=14 name= age=41 idle=41 flags=bxOPS db=0 sub=3 psub=2 ssub=1 multi=1 watch=1 qbuf=2 qbuf-free=2 argv-mem=1 multi-mem=1 rbs=1024 rbp=1 obl=2 oll=2 omem=2 tot-mem=1922 events=r cmd=sismember user=default redir=-1 resp=2 lib-name= lib-ver= tot-cmds=1 tot-net-in=1020 tot-net-out=1020
`
		clist, err := doParseClientList(list)
		assert.NoError(t, err)

		assert.Len(t, clist, 3)

		inst := newInstance()
		inst.ipt = ipt
		mpts, lpts := inst.buildcliPoints(clist)

		require.Len(t, lpts, 1) // 1 block client
		require.Len(t, mpts, 1) // always 1 metric point on multi-line client list

		assert.Equal(t, int64(41), mpts[0].Get("max_idle"))
		assert.Equal(t, int64(2), mpts[0].Get("max_qbuf"))
		assert.Equal(t, int64(2), mpts[0].Get("max_obl"))
		assert.Equal(t, float64(5), mpts[0].Get("total_cmds"))
		assert.Equal(t, int64(3), mpts[0].Get("max_multi"))
		assert.Equal(t, float64(4), mpts[0].Get("multi_total"))
		assert.Equal(t, 4.0/3.0, mpts[0].Get("multi_avg"))
		assert.Equal(t, float64(1023+1021+1020), mpts[0].Get("total_netin"))
		assert.Equal(t, float64(1023+1021+1020), mpts[0].Get("total_netout"))
		assert.Equal(t, float64(1+1+2), mpts[0].Get("total_obl"))
		assert.Equal(t, int64(1922), mpts[0].Get("max_totmem"))
		assert.Equal(t, float64(1922+1920+1921), mpts[0].Get("total_totmem"))
		assert.Equal(t, float64(1+1+1), mpts[0].Get("total_ssub"))
		assert.Equal(t, float64(2+2+2), mpts[0].Get("total_psub"))
		assert.Equal(t, float64(3+3+3), mpts[0].Get("total_sub"))
		assert.Equal(t, int64(1), mpts[0].Get(flagKeys['x']))
		assert.Equal(t, int64(1), mpts[0].Get(flagKeys['b']))
		assert.Equal(t, int64(1), mpts[0].Get(flagKeys['O']))
		assert.Equal(t, int64(1), mpts[0].Get(flagKeys['P']))
		assert.Equal(t, int64(1), mpts[0].Get(flagKeys['S']))
		assert.Equal(t, int64(2), mpts[0].Get(flagKeys['N']))

		assert.Equal(t, lCritical.String(), lpts[0].Get("status"))

		t.Logf("pt: %s", mpts[0].Pretty())
		t.Logf("pt: %s", lpts[0].Pretty())
	})

	t.Run(`multiple-client-log`, func(t *T.T) {
		ipt := defaultInput()

		ipt.ClientListCollector.CollectLogOnFlags = "bN" // enable N flag as logging point

		list := `
id=1320 addr=172.20.0.1:36302 laddr=172.20.0.5:6379 fd=14 name= age=41 idle=41 flags=N db=0 sub=3 psub=2 ssub=1 multi=1 watch=1 qbuf=2 qbuf-free=2 argv-mem=1 multi-mem=1 rbs=1024 rbp=1 obl=2 oll=2 omem=2 tot-mem=1922 events=r cmd=sismember user=default redir=-1 resp=2 lib-name= lib-ver= tot-cmds=1 tot-net-in=1020 tot-net-out=1020
id=1321 addr=172.20.0.1:36302 laddr=172.20.0.5:6379 fd=14 name= age=41 idle=41 flags=bxOPS db=0 sub=3 psub=2 ssub=1 multi=1 watch=1 qbuf=2 qbuf-free=2 argv-mem=1 multi-mem=1 rbs=1024 rbp=1 obl=2 oll=2 omem=2 tot-mem=1922 events=r cmd=sismember user=default redir=-1 resp=2 lib-name= lib-ver= tot-cmds=1 tot-net-in=1020 tot-net-out=1020
# client with unknown-flag: q
id=1321 addr=172.20.0.1:36302 laddr=172.20.0.5:6379 fd=14 name= age=41 idle=41 flags=q db=0 sub=3 psub=2 ssub=1 multi=1 watch=1 qbuf=2 qbuf-free=2 argv-mem=1 multi-mem=1 rbs=1024 rbp=1 obl=2 oll=2 omem=2 tot-mem=1922 events=r cmd=sismember user=default redir=-1 resp=2 lib-name= lib-ver= tot-cmds=1 tot-net-in=1020 tot-net-out=1020
`

		inst := newInstance()
		inst.ipt = ipt

		clist, err := doParseClientList(list)
		assert.NoError(t, err)

		assert.Len(t, clist, 3)
		mpts, lpts := inst.buildcliPoints(clist)

		require.Len(t, lpts, 2)
		require.Len(t, mpts, 1) // always 1 metric point on multi-line client list

		require.Equal(t, int64(1320), lpts[0].Get("id"))
		require.Equal(t, int64(1321), lpts[1].Get("id"))
		require.Equal(t, int64(1), mpts[0].Get("flag_unknown_q"))

		t.Logf("pt: %s", mpts[0].Pretty())
		t.Logf("pt: %s", lpts[0].Pretty())
	})
}
