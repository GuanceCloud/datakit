//go:build linux
// +build linux

package l4log

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParser(t *testing.T) {
	a, err := parseFilter(`

	ipnet_contains("127.0.0.1", "127.0.0.1")
 (src_port==223)
ip_saddr == "1" ; tcp || udp; !ipv6

`)
	if err != nil {
		t.Fatal(err)
	}

	runt := filterRuntime{
		fnG: _fnList,
	}
	err = runt.checkStmts(a, &netParams{})
	if err != nil {
		t.Error(err)
	}

	drop := runt.runNetFilterDrop(a, &netParams{})
	if drop {
		t.Log("drop")
	}
	t.Log()
}

type caseElem struct {
	name   string
	rule   string
	data   *netParams
	result bool
}

func BenchmarkBlacklistFilter(b *testing.B) {
	c := []caseElem{
		{
			name: "r1",
			rule: `(ipnet_contains('10.224.10.0/16', ip_saddr) || 
						ipnet_contains('10.224.10.0/16', ip_daddr))`,
			data: &netParams{
				ipDAddr: "244.178.44.111",
				ipSAddr: "244.178.44.111",
			},
		},
		{
			name: "r2",
			rule: `ip_saddr == "1" && udp && ipv6 || (src_port==223) || dst_port == 123 || ip6_daddr == "2"`,
			data: &netParams{},
		},

		{
			name: "r3",
			rule: `(has_prefix(k8s_src_pod, 'datakit-data') || has_prefix(k8s_src_pod, 'datakit-data'))`,
			data: &netParams{},
		},
		{
			name: "r4",
			rule: `ip_saddr == "1" && udp && ipv6 || (src_port==223) || dst_port == 123 || ip6_daddr == "2"`,
			data: &netParams{},
		},
	}

	for _, c := range c {
		b.Run(c.name, func(b *testing.B) {
			a, err := parseFilter(c.rule)
			if err != nil {
				b.Fatal(err)
			}
			runt := filterRuntime{
				fnG: _fnList,
			}
			err = runt.checkStmts(a, c.data)
			if err != nil {
				b.Fatal(err)
			}
			b.ResetTimer()

			for i := 0; i < b.N; i++ {
				runt.runNetFilterDrop(a, c.data)
			}
			b.StopTimer()
		})
	}
}

func TestBlacklist(t *testing.T) {
	cases := []caseElem{
		{
			name: "ipnet",
			rule: `(ipnet_contains('10.224.10.0/16', ip_saddr) || ipnet_contains('10.224.10.0/16', ip_daddr))`,
			data: &netParams{
				ipSAddr: "10.223.10.1",
				ipDAddr: "10.223.10.1",
			},
			result: false,
		},
		{
			name: "ipnet2",
			rule: `(ipnet_contains('10.224.10.0/16', ip_saddr) || ipnet_contains('10.224.10.0/16', ip_daddr))`,
			data: &netParams{
				ipSAddr: "10.224.10.1",
				ipDAddr: "10.223.10.1",
			},
			result: true,
		},
		{
			name: "ipnet3",
			rule: `(ipnet_contains('10.224.10.0/16', ip_saddr) || ipnet_contains('10.224.10.0/16', ip_daddr))`,
			data: &netParams{
				ipSAddr: "10.223.10.1",
				ipDAddr: "10.224.10.1",
			},
			result: true,
		},
		{
			name: "prefix",
			rule: `(has_prefix(k8s_src_pod, 'datakit-data') || has_prefix(k8s_src_pod, 'datakit-data'))`,
			data: &netParams{
				k8sSrcPod: "1datakit-data",
				k8sDstPod: "1datakit-data",
			},
			result: false,
		},
		{
			name: "prefix1",
			rule: `(has_prefix(k8s_src_pod, 'datakit-data') || has_prefix(k8s_dst_pod, 'datakit-data'))`,
			data: &netParams{
				k8sSrcPod: "1datakit-data",
				k8sDstPod: "datakit-data",
			},
			result: true,
		},
		{
			name: "prefix2",
			rule: `(has_prefix(k8s_src_pod, 'datakit-data') || has_prefix(k8s_dst_pod, 'datakit-data'))`,
			data: &netParams{
				k8sSrcPod: "datakit-data",
				k8sDstPod: "1datakit-data",
			},
			result: true,
		},
		{
			name: "others",
			rule: `
			ip_saddr == "1"
			ip_daddr == "2"
			ip6_saddr == "3"
			ip6_daddr == "4"
			src_port >= 10;dst_port < 20 && dst_port > 0
			tcp
			!udp
			ipv4
			!ipv6
			`,
			data:   &netParams{},
			result: false,
		},
		{
			name: "others1",
			rule: `
			ip_saddr == "1"
			ip_daddr == "2"
			ip6_saddr == "3"
			ip6_daddr == "4"
			src_port >= 10
			dst_port < 20 && dst_port > 0
			tcp
			!udp
			ipv4
			!ipv6
			`,
			data: &netParams{
				ip6SAddr: "3",
			},
			result: true,
		},
	}

	for _, c := range cases {
		ast, err := parseFilter(c.rule)
		if err != nil {
			t.Fatal(err)
		}
		runt := filterRuntime{
			fnG: _fnList,
		}
		err = runt.checkStmts(ast, c.data)
		if err != nil {
			t.Error(err)
		}
		v := runt.runNetFilterDrop(ast, c.data)
		assert.Equal(t, c.result, v)
	}
}

func TestBlacklistDecisionCache(t *testing.T) {
	rt := &filterRuntime{fnG: _fnList}

	rule80, err := parseFilter(`src_port == 80`)
	if err != nil {
		t.Fatal(err)
	}

	conns := &TCPConns{
		runtime:   rt,
		blacklist: rule80,
	}

	key80 := &PMeta{
		SrcIP:   "10.0.0.1",
		DstIP:   "10.0.0.2",
		SrcPort: 80,
		DstPort: 8080,
	}

	if !conns.shouldDropByBlacklist(key80, false, 1) {
		t.Fatalf("expected first rule to drop src_port 80")
	}

	rule81, err := parseFilter(`src_port == 81`)
	if err != nil {
		t.Fatal(err)
	}
	conns.blacklist = rule81

	if conns.shouldDropByBlacklist(key80, false, 2) {
		t.Fatalf("expected blacklist cache to be invalidated after rule change")
	}

	key81 := &PMeta{
		SrcIP:   "10.0.0.1",
		DstIP:   "10.0.0.2",
		SrcPort: 81,
		DstPort: 8080,
	}

	if !conns.shouldDropByBlacklist(key81, false, 3) {
		t.Fatalf("expected new flow key to use updated rule")
	}

	assert.Len(t, conns.blacklistCache, 2)
	entry80, ok80 := conns.blacklistCache[blacklistCacheKey{meta: *key80}]
	entry81, ok81 := conns.blacklistCache[blacklistCacheKey{meta: *key81}]
	assert.True(t, ok80)
	assert.True(t, ok81)
	assert.False(t, entry80.drop)
	assert.True(t, entry81.drop)
	assert.Equal(t, int64(2), entry80.lastTS)
}
