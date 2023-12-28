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
