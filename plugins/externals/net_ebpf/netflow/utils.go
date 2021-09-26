// +build linux

package netflow

import (
	"bytes"
	"context"
	"fmt"
	"io/ioutil"
	"math"
	"net/http"
	"sort"
	"strings"
	"time"

	"github.com/DataDog/ebpf/manager"
	"gitlab.jiagouyun.com/cloudcare-tools/cliutils/logger"
	dkebpf "gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/externals/net_ebpf/c"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"
	"golang.org/x/net/context/ctxhttp"
	"golang.org/x/sys/unix"
)

var l = logger.DefaultSLogger("net_ebpf")

func SetLogger(nl *logger.Logger) {
	l = nl
}

func NewNetFlowManger(constEditor []manager.ConstantEditor) (*manager.Manager, error) {
	m := &manager.Manager{
		Probes: []*manager.Probe{
			{
				Section: "kprobe/sockfd_lookup_light",
			}, {
				Section: "kretprobe/sockfd_lookup_light",
			}, {
				Section: "kprobe/tcp_set_state",
			}, {
				Section: "kretprobe/inet_csk_accept",
			}, {
				Section: "kprobe/inet_csk_listen_stop",
			}, {
				Section: "kprobe/tcp_close",
			}, {
				Section: "kprobe/tcp_retransmit_skb",
			}, {
				Section: "kprobe/tcp_sendmsg",
			}, {
				Section: "kprobe/tcp_cleanup_rbuf",
			}, {
				Section: "kprobe/ip_make_skb",
			}, {
				Section: "kprobe/udp_recvmsg",
			}, {
				Section: "kretprobe/udp_recvmsg",
			}, {
				Section: "kprobe/inet_bind",
			}, {
				Section: "kretprobe/inet_bind",
			}, {
				Section: "kprobe/inet6_bind",
			}, {
				Section: "kretprobe/inet6_bind",
			}, {
				Section: "kprobe/udp_destroy_sock",
			},
		},
		PerfMaps: []*manager.PerfMap{
			{
				Map: manager.Map{
					Name: "bpfmap_closed_event",
				},
				PerfMapOptions: manager.PerfMapOptions{
					DataHandler: closedEventHandler,
				},
			},
		},
	}
	m_opts := manager.Options{
		RLimit: &unix.Rlimit{
			Cur: math.MaxUint64,
			Max: math.MaxUint64,
		},
		ConstantEditors: constEditor,
	}
	if buf, err := dkebpf.Asset("netflow.o"); err != nil {
		return nil, err
	} else {
		if err := m.InitWithOptions((bytes.NewReader(buf)), m_opts); err != nil {
			return nil, err
		}
	}
	return m, nil
}

func WriteData(data []byte, urlPath string) error {
	// dataway path
	ctx, ctxCancel := context.WithCancel(context.Background())
	defer ctxCancel()
	httpReq, err := http.NewRequest("POST", urlPath, bytes.NewBuffer(data))
	if err != nil {
		l.Errorf("[error] %s", err.Error())
		return err
	}

	httpReq = httpReq.WithContext(ctx)
	tmctx, timeoutCancel := context.WithTimeout(context.Background(), time.Second*10)
	defer timeoutCancel()

	resp, err := ctxhttp.Do(tmctx, http.DefaultClient, httpReq)
	if err != nil {
		l.Errorf("[error] %s", err.Error())
		return err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		l.Error(err)
		return err
	}

	l.Debug(urlPath, resp.StatusCode)
	switch resp.StatusCode / 100 {
	case 2:
		return nil
	default:
		l.Debugf("post to %s HTTP: %d: %s", urlPath, resp.StatusCode, string(body))
		return fmt.Errorf("post to %s failed(HTTP: %d): %s", urlPath, resp.StatusCode, string(body))
	}
}

func FeedMeasurement(measurements *[]inputs.Measurement, path string) error {
	lines := [][]byte{}
	for _, m := range *measurements {
		if pt, err := m.LineProto(); err != nil {
			l.Warn(err)
		} else {
			ptstr := pt.String()
			lines = append(lines, []byte(ptstr))
		}
	}

	if err := WriteData(bytes.Join(lines, []byte("\n")), path); err != nil {
		return err
	}
	return nil
}

func convertConn2Measurement(connR *ConnResult, name string) *[]inputs.Measurement {
	collectCache := []inputs.Measurement{}

	for k, v := range connR.result {
		m := convConn2M(k, v, name, connR.tags, connR.ts)
		collectCache = append(collectCache, m)
	}
	return &collectCache
}

func convConn2M(k ConnectionInfo, v ConnFullStats, name string, tags map[string]string, ts time.Time) inputs.Measurement {
	m := measurement{
		name:   name,
		tags:   map[string]string{},
		fields: map[string]interface{}{},
		ts:     ts,
	}
	for k, v := range tags {
		m.tags[k] = v
	}

	m.tags["source"] = "netflow"
	m.tags["status"] = "info"
	m.tags["pid"] = fmt.Sprint(k.Pid)
	if connAddrIsIPv4(k.Meta) {
		m.tags["src_ip"] = U32BEToIPv4(k.Saddr[3])
		m.tags["dst_ip"] = U32BEToIPv4(k.Daddr[3])

		m.tags["src_ip_type"] = connIPv4Type(k.Saddr[3])
		m.tags["dst_ip_type"] = connIPv4Type(k.Daddr[3])

		m.tags["family"] = "IPv4"
	} else {
		m.tags["src_ip"] = U32BEToIPv6(k.Saddr)
		m.tags["dst_ip"] = U32BEToIPv6(k.Daddr)

		m.tags["src_ip_type"] = "other"
		m.tags["dst_ip_type"] = "other"

		m.tags["family"] = "IPv6"
	}
	if k.Sport == math.MaxUint32 {
		m.tags["src_port"] = "*"
	} else {
		m.tags["src_port"] = fmt.Sprintf("%d", k.Sport)
	}
	m.tags["dst_port"] = fmt.Sprintf("%d", k.Dport)

	m.fields["bytes_read"] = int64(v.Stats.Recv_bytes)
	m.fields["bytes_written"] = int64(v.Stats.Sent_bytes)

	if connProtocolIsTCP(k.Meta) {
		m.tags["transport"] = "tcp"
		m.fields["retransmits"] = int64(v.TcpStats.Retransmits)
		m.fields["rtt"] = int64(v.TcpStats.Rtt)
		m.fields["rtt_var"] = int64(v.TcpStats.Rtt_var)
		m.fields["tcp_closed"] = v.TotalClosed
		m.fields["tcp_established"] = v.TotalEstablished
	} else {
		m.tags["transport"] = "udp"
	}
	m.tags["direction"] = connDirection2Str(v.Stats.Direction)

	if connProtocolIsTCP(k.Meta) {
		l.Debug(fmt.Sprintf("pid %s: %s:%s->%s:%s r/w: %d/%d e/c: %d/%d re: %d rtt/rttvar: %.2fms/%.2fms (%s, %s)",
			m.tags["pid"], m.tags["src_ip"], m.tags["src_port"], m.tags["dst_ip"], m.tags["dst_port"], m.fields["bytes_read"], m.fields["bytes_written"],
			m.fields["tcp_established"], m.fields["tcp_closed"], m.fields["retransmits"], float64(v.TcpStats.Rtt)/1000., float64(v.TcpStats.Rtt_var)/1000, m.tags["transport"], m.tags["direction"]))
	} else {
		l.Debug(fmt.Sprintf("pid %s: %s:%s->%s:%s r/w: %d/%d (%s, %s)",
			m.tags["pid"], m.tags["src_ip"], m.tags["src_port"], m.tags["dst_ip"], m.tags["dst_port"], m.fields["bytes_read"], m.fields["bytes_written"],
			m.tags["transport"], m.tags["direction"]))
	}
	return &m
}

func U32BEToIPv4Array(addr uint32) [4]int {
	var ip [4]int
	for x := 0; x < 4; x++ {
		ip[x] = int(addr & 0xff)
		addr = addr >> 8
	}
	return ip
}

func U32BEToIPv4(addr uint32) string {
	ip := U32BEToIPv4Array(addr)
	return fmt.Sprintf("%d.%d.%d.%d", ip[0], ip[1], ip[2], ip[3])
}

func U32BEToIPv6(addr [4]uint32) string {
	// addr byte order: big endian
	var ip [8]int
	for x := 1; x < 4; x++ {
		ip[(x * 2)] = int(addr[x] & 0xffff)         // uint32 低16位
		ip[(x*2)+1] = int((addr[x] >> 16) & 0xffff) //	高16位
	}
	ipStr := fmt.Sprintf("%x:%x:%x:%x:%x:%x",
		ip[0], ip[1], ip[2], ip[3], ip[4], ip[5])
	// IPv4-mapped IPv6 address
	if ipStr == "0:0:0:0:0:ffff" {
		ipStr += ":" + U32BEToIPv4(addr[3])
	} else {
		ipStr += fmt.Sprintf(":%x:%x", ip[6], ip[7])
	}
	return ipStr
}

// 规则: 1. 过滤源 IP 和目标 IP 相同的连接;
// 3. 过滤 loopback
// 2. 过滤一个采集周期内的无数据收发的连接；
// 需过滤，函数返回 False.
func ConnFilter(conn ConnectionInfo, connStats ConnFullStats) bool {
	// 过滤同 IP 地址的连接，适用于 UDP 和 TCP
	if connAddrIsIPv4(conn.Meta) {
		saddr := U32BEToIPv4(conn.Saddr[3])
		daddr := U32BEToIPv4(conn.Daddr[3])
		if strings.EqualFold(saddr, daddr) {
			return false
		} else if strings.Split(saddr, ".")[0] == "127" && strings.Split(daddr, ".")[0] == "127" {
			return false
		}
	} else {
		for x := 0; x < 4; x++ {
			if conn.Daddr[x] != conn.Saddr[x] {
				break
			}
			if x == 3 {
				return false
			}
		}
	}

	// 过滤无数据收发的连接
	if connStats.Stats.Recv_bytes == 0 && connStats.Stats.Sent_bytes == 0 {
		return false
	}

	return true
}

// 聚合 src port 为临时端口(32768 ~ 60999)的连接,
// 被聚合的端口号被设置为
// cat /proc/sys/net/ipv4/ip_local_port_range.
func connMerge(preResult *ConnResult) {
	resultTmpConn := map[ConnectionInfo]ConnFullStats{}
	if len(preResult.result) < 1 {
		return
	}

	connInfoList := ConnInfoList{}

	for k := range preResult.result {
		connInfoList = append(connInfoList, k)
	}
	sort.Sort(connInfoList)
	connRecord := map[ConnectionInfo]bool{}
	lastIndex := -1
	for k := 0; k < len(connInfoList); k++ {
		if !isEphemeralPort(connInfoList[k].Sport) {
			continue
		}
		if lastIndex < 0 {
			lastIndex = k
			resultTmpConn[connInfoList[k]] = preResult.result[connInfoList[k]]
			delete(preResult.result, connInfoList[k])
		} else if connCmpNoSPort(connInfoList[lastIndex], connInfoList[k]) {
			connRecord[connInfoList[lastIndex]] = true
			resultTmpConn[connInfoList[lastIndex]] = statsTCPOp("+", resultTmpConn[connInfoList[lastIndex]],
				preResult.result[connInfoList[k]].Stats, preResult.result[connInfoList[k]].TcpStats)

			connfull := resultTmpConn[connInfoList[lastIndex]]
			connfull.TotalEstablished += preResult.result[connInfoList[k]].TotalEstablished
			connfull.TotalClosed += preResult.result[connInfoList[k]].TotalClosed
			resultTmpConn[connInfoList[lastIndex]] = connfull

			delete(preResult.result, connInfoList[k])
		} else {
			k--
			lastIndex = -1
		}
	}
	for k, v := range resultTmpConn {
		if _, ok := connRecord[k]; ok {
			k.Sport = math.MaxUint32
		}
		preResult.result[k] = v
	}
}

func connCmpNoSPort(expected, actual ConnectionInfo) bool {
	expected.Sport = 0
	actual.Sport = 0
	return expected == actual
}

type ConnInfoList []ConnectionInfo

func (l ConnInfoList) Len() int {
	return len(l)
}

func (l ConnInfoList) Less(i, j int) bool {
	metaI := l[i].Meta
	metaJ := l[j].Meta

	// family (ipv4)
	if metaI&CONN_L3_MASK != metaJ&CONN_L3_MASK {
		return metaI&CONN_L3_MASK == CONN_L3_IPv4
	}

	// transport (tcp)
	if metaI&CONN_L4_MASK != metaJ&CONN_L4_MASK {
		return metaI&CONN_L4_MASK == CONN_L4_TCP
	}

	// pid
	if l[i].Pid != l[j].Pid {
		return l[i].Pid < l[j].Pid
	}

	// src ip, dst ip
	if metaI&CONN_L3_MASK == CONN_L3_IPv4 { // ipv4
		if l[i].Saddr[3] != l[j].Saddr[3] {
			return l[i].Saddr[3] < l[j].Saddr[3]
		}
		if l[i].Daddr[3] != l[j].Daddr[3] { // dst ip
			return l[i].Daddr[3] < l[j].Daddr[3]
		}
	} else { // ipv6
		if l[i].Saddr != l[j].Saddr {
			for x := 0; x < 4; x++ {
				if l[i].Saddr[x] > l[j].Saddr[x] {
					return false
				}
			}
			return true
		}
		if l[i].Daddr != l[j].Daddr {
			for x := 0; x < 4; x++ {
				if l[i].Daddr[x] > l[j].Daddr[x] {
					return false
				}
			}
			return true
		}
	}

	// dst port
	if l[i].Dport != l[j].Dport {
		return l[i].Dport < l[j].Dport
	}

	// src port
	if l[i].Sport != l[j].Sport {
		return l[i].Sport < l[j].Sport
	}

	// all equal
	return false
}

func (l ConnInfoList) Swap(i, j int) {
	l[i], l[j] = l[j], l[i]
}

const (
	EphemeralPortMin = 32768
	EphemeralPortMax = 60999
)

func isEphemeralPort(port uint32) bool {
	return port >= EphemeralPortMin && port <= EphemeralPortMax
}
