//go:build linux
// +build linux

package netflow

import (
	"bytes"
	"fmt"
	"math"
	"net"
	"os"
	"sync"
	"sync/atomic"
	"time"

	manager "github.com/DataDog/ebpf-manager"
	"github.com/GuanceCloud/cliutils/logger"
	"github.com/GuanceCloud/cliutils/point"
	"github.com/cilium/ebpf"
	dkebpf "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/externals/ebpf/internal/c"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/externals/ebpf/pkg/cli"
	"golang.org/x/sys/unix"
)

const (
	NoValue           = "N/A"
	DirectionOutgoing = "outgoing"
	DirectionIncoming = "incoming"
	DirectionUnknown  = "unknown"
)

var ephemeralPortMin int32 = 0 // 10_001

func SetEphemeralPortMin(val int32) {
	if val < 0 {
		val = 0
	}
	l.Debugf("ephemeral port start from %d", val)
	atomic.StoreInt32(&ephemeralPortMin, val)
}

var l = logger.DefaultSLogger("ebpf")

type dnsRecorder interface {
	LookupAddr(ip string) string
}

var dnsRecord dnsRecorder

var k8sNetInfo *cli.K8sInfo

func SetDNSRecord(r dnsRecorder) {
	dnsRecord = r
}

func SetLogger(nl *logger.Logger) {
	l = nl
}

func SetK8sNetInfo(n *cli.K8sInfo) {
	k8sNetInfo = n
}

var SrcIPPortRecorder = func() *srcIPPortRecorder {
	ptr := &srcIPPortRecorder{
		Record: map[[4]uint32]IPPortRecord{},
	}
	go ptr.AutoClean()
	return ptr
}()

type IPPortRecord struct {
	IP [4]uint32
	TS time.Time
}

// Assist httpflow to judge server ip.
type srcIPPortRecorder struct {
	sync.RWMutex
	Record map[[4]uint32]IPPortRecord
}

func (record *srcIPPortRecorder) InsertAndUpdate(ip [4]uint32) {
	record.Lock()
	defer record.Unlock()
	record.Record[ip] = IPPortRecord{
		IP: ip,
		TS: time.Now(),
	}
}

func (record *srcIPPortRecorder) Query(ip [4]uint32) (*IPPortRecord, error) {
	record.RLock()
	defer record.RUnlock()
	if v, ok := record.Record[ip]; ok {
		return &v, nil
	} else {
		return nil, fmt.Errorf("not found")
	}
}

const (
	cleanTickerIPPortDur = time.Minute * 3
	cleanIPPortDur       = time.Minute * 5
)

func (record *srcIPPortRecorder) CleanOutdateData() {
	record.Lock()
	defer record.Unlock()
	ts := time.Now()
	needDelete := [][4]uint32{}
	for k, v := range record.Record {
		if ts.Sub(v.TS) > cleanIPPortDur {
			needDelete = append(needDelete, k)
		}
	}
	for _, v := range needDelete {
		delete(record.Record, v)
	}
}

func (record *srcIPPortRecorder) AutoClean() {
	ticker := time.NewTicker(cleanTickerIPPortDur)
	for {
		<-ticker.C
		record.CleanOutdateData()
	}
}

func NewNetFlowManger(constEditor []manager.ConstantEditor, ctMap map[string]*ebpf.Map, closedEventHandler func(cpu int, data []byte,
	perfmap *manager.PerfMap, manager *manager.Manager),
) (*manager.Manager, error) {
	// Some kretprobe type programs need to set maxactive， https://www.kernel.org/doc/Documentation/kprobes.txt.
	m := &manager.Manager{
		Probes: []*manager.Probe{
			{
				ProbeIdentificationPair: manager.ProbeIdentificationPair{
					EBPFFuncName: "kprobe__sockfd_lookup_light",
				},
				KProbeMaxActive: 128,
			}, {
				ProbeIdentificationPair: manager.ProbeIdentificationPair{
					EBPFFuncName: "kretprobe__sockfd_lookup_light",
				},
				KProbeMaxActive: 128,
			}, {
				ProbeIdentificationPair: manager.ProbeIdentificationPair{
					EBPFFuncName: "kprobe__do_sendfile",
				},
			}, {
				ProbeIdentificationPair: manager.ProbeIdentificationPair{
					EBPFFuncName: "kretprobe__do_sendfile",
				},
				KProbeMaxActive: 128,
			}, {
				ProbeIdentificationPair: manager.ProbeIdentificationPair{
					EBPFFuncName: "kprobe__tcp_set_state",
				},
			}, {
				ProbeIdentificationPair: manager.ProbeIdentificationPair{
					EBPFFuncName: "kretprobe__inet_csk_accept",
				},
				KProbeMaxActive: 128,
			}, {
				ProbeIdentificationPair: manager.ProbeIdentificationPair{
					EBPFFuncName: "kprobe__inet_csk_listen_stop",
				},
			}, {
				ProbeIdentificationPair: manager.ProbeIdentificationPair{
					EBPFFuncName: "kprobe__tcp_close",
				},
			}, {
				ProbeIdentificationPair: manager.ProbeIdentificationPair{
					EBPFFuncName: "kprobe__tcp_retransmit_skb",
				},
			}, {
				ProbeIdentificationPair: manager.ProbeIdentificationPair{
					EBPFFuncName: "kprobe__tcp_sendmsg",
				},
			}, {
				ProbeIdentificationPair: manager.ProbeIdentificationPair{
					EBPFFuncName: "kprobe__tcp_cleanup_buf",
				},
			}, {
				ProbeIdentificationPair: manager.ProbeIdentificationPair{
					EBPFFuncName: "kprobe__ip_make_skb",
				},
			}, {
				ProbeIdentificationPair: manager.ProbeIdentificationPair{
					EBPFFuncName: "kprobe__udp_recvmsg",
				},
			}, {
				ProbeIdentificationPair: manager.ProbeIdentificationPair{
					EBPFFuncName: "kretprobe__udp_recvmsg",
				},
				KProbeMaxActive: 128,
			}, {
				ProbeIdentificationPair: manager.ProbeIdentificationPair{
					EBPFFuncName: "kprobe__inet_bind",
				},
			}, {
				ProbeIdentificationPair: manager.ProbeIdentificationPair{
					EBPFFuncName: "kretprobe__inet_bind",
				},
				KProbeMaxActive: 128,
			}, {
				ProbeIdentificationPair: manager.ProbeIdentificationPair{
					EBPFFuncName: "kprobe__inet6_bind",
				},
			}, {
				ProbeIdentificationPair: manager.ProbeIdentificationPair{
					EBPFFuncName: "kretprobe__inet6_bind",
				},
				KProbeMaxActive: 128,
			}, {
				ProbeIdentificationPair: manager.ProbeIdentificationPair{
					EBPFFuncName: "kprobe__udp_destroy_sock",
				},
			},
		},
		PerfMaps: []*manager.PerfMap{
			{
				Map: manager.Map{
					Name: "bpfmap_closed_event",
				},
				PerfMapOptions: manager.PerfMapOptions{
					// sizeof(connection_closed_info) > 112 Byte, pagesize ~= 4k,
					// if cpus = 8, 5 conn/per connection_closed_info
					PerfRingBufferSize: 32 * os.Getpagesize(),
					DataHandler:        closedEventHandler,
				},
			},
		},
	}
	mOpts := manager.Options{
		RLimit: &unix.Rlimit{
			Cur: math.MaxUint64,
			Max: math.MaxUint64,
		},
		ConstantEditors: constEditor,
	}

	if ctMap != nil {
		mOpts.MapEditors = ctMap
	}

	if buf, err := dkebpf.NetFlowBin(); err != nil {
		return nil, fmt.Errorf("netflow.o: %w", err)
	} else if err := m.InitWithOptions((bytes.NewReader(buf)), mOpts); err != nil {
		return nil, err
	}

	return m, nil
}

func ConvConn2M(k ConnectionInfo, v ConnFullStats, name string,
	gTags map[string]string, ptTime time.Time, pidMap map[int][2]string,
) (*point.Point, error) {
	mFields := map[string]interface{}{}
	mTags := map[string]string{}

	for k, v := range gTags {
		mTags[k] = v
	}

	mTags["status"] = "info"
	mTags["pid"] = fmt.Sprint(k.Pid)
	if procName, ok := pidMap[int(k.Pid)]; ok {
		mTags["process_name"] = procName[0]
	} else {
		mTags["process_name"] = NoValue
	}

	isV6 := !ConnAddrIsIPv4(k.Meta)
	if k.Saddr[0] == 0 && k.Saddr[1] == 0 && k.Daddr[0] == 0 && k.Daddr[1] == 0 {
		if k.Saddr[2] == 0xffff0000 && k.Daddr[2] == 0xffff0000 {
			isV6 = false
		} else if k.Saddr[2] == 0 && k.Daddr[2] == 0 && k.Saddr[3] > 1 && k.Daddr[3] > 1 {
			isV6 = false
		}
	}

	if !isV6 {
		mTags["src_ip_type"] = ConnIPv4Type(k.Saddr[3])
		mTags["dst_ip_type"] = ConnIPv4Type(k.Daddr[3])
		mTags["family"] = "IPv4"
	} else {
		mTags["src_ip_type"] = ConnIPv6Type(k.Saddr)

		mTags["dst_ip_type"] = ConnIPv6Type(k.Daddr)
		mTags["family"] = "IPv6"
	}

	srcIP := U32BEToIP(k.Saddr, isV6).String()
	mTags["src_ip"] = srcIP

	dstIP := U32BEToIP(k.Daddr, isV6).String()
	mTags["dst_ip"] = dstIP

	if dnsRecord != nil {
		mTags["dst_domain"] = dnsRecord.LookupAddr(dstIP)
	}

	if k.Sport == math.MaxUint32 {
		mTags["src_port"] = "*"
	} else {
		mTags["src_port"] = fmt.Sprintf("%d", k.Sport)
	}

	if k.Dport == math.MaxUint32 {
		mTags["dst_port"] = "*"
	} else {
		mTags["dst_port"] = fmt.Sprintf("%d", k.Dport)
	}

	mFields["bytes_read"] = int64(v.Stats.RecvBytes)
	mFields["bytes_written"] = int64(v.Stats.SentBytes)

	var l4proto string
	if ConnProtocolIsTCP(k.Meta) {
		l4proto = "tcp"
		mTags["transport"] = l4proto
		mFields["retransmits"] = int64(v.TCPStats.Retransmits)
		mFields["rtt"] = int64(v.TCPStats.Rtt)
		mFields["rtt_var"] = int64(v.TCPStats.RttVar)
		mFields["tcp_closed"] = v.TotalClosed
		mFields["tcp_established"] = v.TotalEstablished
	} else {
		l4proto = "udp"
		mTags["transport"] = l4proto
	}
	mTags["direction"] = ConnDirection2Str(v.Stats.Direction)

	// add K8s tags

	mTags = AddK8sTags2Map(k8sNetInfo, &BaseKey{
		SAddr:     srcIP,
		DAddr:     dstIP,
		SPort:     k.Sport,
		DPort:     k.Dport,
		Transport: l4proto,
	}, mTags)

	kvs := point.NewTags(mTags)
	kvs = append(kvs, point.NewKVs(mFields)...)
	pt := point.NewPointV2(name, kvs, append(
		point.CommonLoggingOptions(), point.WithTime(ptTime))...)
	return pt, nil
}

func IsIncomingFromK8s(k8sNetInfo *cli.K8sInfo, pid int, srcIP string,
	srcPort uint32, transport string,
) bool {
	if k8sNetInfo != nil {
		if t, ok := k8sNetInfo.IsServer(pid,
			transport, srcIP, int(srcPort)); ok {
			return t
		}
	}
	return false
}

func AddK8sTags2Map(k8sNetInfo *cli.K8sInfo,
	basekey *BaseKey, mTags map[string]string,
) map[string]string {
	if mTags == nil {
		mTags = map[string]string{}
	}

	if basekey == nil {
		return mTags
	}

	if k8sNetInfo != nil {
		srcK8sFlag := false
		dstK8sFlag := false
		if t, ok := k8sNetInfo.IsServer(basekey.PID, basekey.Transport,
			basekey.SAddr, int(basekey.SPort)); ok && t {
			mTags["direction"] = DirectionIncoming
		}
		if t, ok := k8sNetInfo.QueryPodInfo(
			basekey.PID, basekey.SAddr, int(basekey.SPort), basekey.Transport); ok && t != nil {
			srcK8sFlag = true
			mTags["src_k8s_namespace"] = t.NS
			mTags["src_k8s_pod_name"] = t.PodName
			mTags["src_k8s_service_name"] = t.SvcName
			mTags["src_k8s_deployment_name"] = t.WorkloadName
			mTags["src_k8s_workload"] = t.WorkloadName
			mTags["src_k8s_workload_type"] = t.Kind.String()

			for k, v := range t.Labels {
				if _, ok := mTags[k]; !ok {
					mTags[k] = v
				}
			}
		} else if t, ok := k8sNetInfo.QuerySvcInfo(
			basekey.Transport, basekey.SAddr, int(basekey.SPort)); ok && t != nil {
			srcK8sFlag = true
			mTags["src_k8s_namespace"] = t.Chain.Tag.NS
			mTags["src_k8s_pod_name"] = NoValue
			mTags["src_k8s_service_name"] = t.Svc.Name
			mTags["src_k8s_deployment_name"] = t.Chain.Tag.WorkloadName
			mTags["src_k8s_workload"] = t.Chain.Tag.WorkloadName
			mTags["src_k8s_workload_type"] = t.Chain.Tag.Kind.String()
		}

		if basekey.DNATAddr != "" && basekey.DNATPort != 0 {
			if t, ok := k8sNetInfo.QueryPodInfo(
				0, basekey.DNATAddr, int(basekey.DNATPort), basekey.Transport); ok && t != nil {
				dstK8sFlag = true
				mTags["dst_k8s_namespace"] = t.NS
				mTags["dst_k8s_pod_name"] = t.PodName
				mTags["dst_k8s_service_name"] = t.SvcName
				mTags["dst_k8s_deployment_name"] = t.WorkloadName
				mTags["dst_k8s_workload"] = t.WorkloadName
				mTags["dst_k8s_workload_type"] = t.Kind.String()
				goto skip_dst
			}
		}

		if t, ok := k8sNetInfo.QueryPodInfo(0,
			basekey.DAddr, int(basekey.DPort), basekey.Transport); ok && t != nil {
			// k.dport
			dstK8sFlag = true
			mTags["dst_k8s_namespace"] = t.NS
			mTags["dst_k8s_pod_name"] = t.PodName
			mTags["dst_k8s_service_name"] = t.SvcName
			mTags["dst_k8s_deployment_name"] = t.WorkloadName
			mTags["dst_k8s_workload"] = t.WorkloadName
			mTags["dst_k8s_workload_type"] = t.Kind.String()
		} else {
			if t, ok := k8sNetInfo.QuerySvcInfo(basekey.Transport,
				basekey.DAddr, int(basekey.DPort)); ok && t != nil {
				dstK8sFlag = true
				mTags["dst_k8s_namespace"] = t.Chain.Tag.NS
				mTags["dst_k8s_pod_name"] = NoValue
				mTags["dst_k8s_service_name"] = t.Svc.Name
				mTags["dst_k8s_deployment_name"] = t.Chain.Tag.WorkloadName
				mTags["dst_k8s_workload"] = t.Chain.Tag.WorkloadName
				mTags["dst_k8s_workload_type"] = t.Chain.Tag.Kind.String()
			}
		}

	skip_dst:
		if srcK8sFlag || dstK8sFlag {
			mTags["sub_source"] = "K8s"
			if !srcK8sFlag {
				mTags["src_k8s_namespace"] = NoValue
				mTags["src_k8s_pod_name"] = NoValue
				mTags["src_k8s_service_name"] = NoValue
				mTags["src_k8s_deployment_name"] = NoValue
				mTags["src_k8s_workload"] = NoValue
				mTags["src_k8s_workload_type"] = NoValue
			}
			if !dstK8sFlag {
				mTags["dst_k8s_namespace"] = NoValue
				mTags["dst_k8s_pod_name"] = NoValue
				mTags["dst_k8s_service_name"] = NoValue
				mTags["dst_k8s_deployment_name"] = NoValue
				mTags["dst_k8s_workload"] = NoValue
				mTags["dst_k8s_workload_type"] = NoValue
			}
		}
	}
	return mTags
}

func U32BEToIPv4Array(addr uint32) [4]uint8 {
	var ip [4]uint8
	for x := 0; x < 4; x++ {
		ip[x] = uint8(addr & 0xff)
		addr >>= 8
	}
	return ip
}

func SwapU16(v uint16) uint16 {
	return ((v & 0x00ff) << 8) | ((v & 0xff00) >> 8)
}

func U32BEToIPv6Array(addr [4]uint32) [8]uint16 {
	var ip [8]uint16
	for x := 0; x < 4; x++ {
		ip[(x * 2)] = SwapU16(uint16(addr[x] & 0xffff))
		ip[(x*2)+1] = SwapU16(uint16((addr[x] >> 16) & 0xffff))
	}
	return ip
}

func U32BEToIP(addr [4]uint32, isIPv6 bool) net.IP {
	ip := net.IP{}
	if !isIPv6 {
		v4 := U32BEToIPv4Array(addr[3])
		for _, v := range v4 {
			ip = append(ip, v)
		}
	} else {
		v6 := U32BEToIPv6Array(addr)
		for _, v := range v6 {
			ip = append(ip, byte((v&0xff00)>>8), byte(v&0x00ff)) // SwapU16(v)
		}
	}
	return ip
}

// ConnNotNeedToFilter rules: 1. Filter connections with the same source IP and destination IP;
// 2. Filter the connection of loopback ip;
// 3. Filter connections without data sending and receiving within a collection cycle;
// 4. Filter connections with port 0 or ip address :: or 0.0.0.0;
// Need to filter, the function returns False.
func ConnNotNeedToFilter(conn *ConnectionInfo, connStats *ConnFullStats) bool {
	if !enableUDP && !ConnProtocolIsTCP(conn.Meta) {
		return false
	}
	if (conn.Saddr[0]|conn.Saddr[1]|conn.Saddr[2]|conn.Saddr[3]) == 0 ||
		(conn.Daddr[0]|conn.Daddr[1]|conn.Daddr[2]|conn.Daddr[3]) == 0 ||
		conn.Sport == 0 || conn.Dport == 0 {
		return false
	}
	if ConnAddrIsIPv4(conn.Meta) { // IPv4
		if (conn.Saddr[3]&0xff) == 127 && (conn.Daddr[3]&0xff) == 127 {
			return false
		}
	} else { // IPv6
		if conn.Saddr[2] == 0xffff0000 && conn.Daddr[2] == 0xffff0000 {
			if (conn.Saddr[3]&0xff) == 127 && (conn.Daddr[3]&0xff) == 127 {
				return false
			}
		} else if (conn.Saddr[0]|conn.Saddr[1]|conn.Saddr[2]) == 0 && conn.Saddr[3] == 1 &&
			(conn.Daddr[0]|conn.Daddr[1]|conn.Daddr[2]) == 0 && conn.Daddr[3] == 1 {
			return false
		}
	}

	// Filter connections that have not changed in the previous cycle
	if connStats.Stats.RecvBytes == 0 && connStats.Stats.SentBytes == 0 &&
		connStats.TotalClosed == 0 && connStats.TotalEstablished == 0 {
		return false
	}

	return true
}

func ConnCmpNoSPort(expected, actual ConnectionInfo) bool {
	expected.Sport = 0
	actual.Sport = 0
	return expected == actual
}

func ConnCmpNoPid(expected, actual ConnectionInfo) bool {
	expected.Pid = 0
	actual.Pid = 0
	return expected == actual
}

const (
	EphemeralPortMin = 32768
	EphemeralPortMax = 60999
)

func IsEphemeralPort(port uint32) bool {
	return port >= uint32(ephemeralPortMin)
}

func IPPortFilterIn(conn *ConnectionInfo) bool {
	if conn.Sport == 0 || conn.Dport == 0 {
		return false
	}

	if ConnAddrIsIPv4(conn.Meta) {
		if (conn.Saddr[3]&0xFF == 0x7F) || (conn.Daddr[3]&0xFF == 0x7F) {
			return false
		}
	} else if (conn.Saddr[0]|conn.Saddr[1]) == 0x00 || (conn.Daddr[0]|conn.Daddr[1]) == 0x00 {
		if (conn.Saddr[2] == 0xffff0000 && conn.Saddr[3]&0xFF == 0x7F) ||
			(conn.Daddr[2] == 0xffff0000 && conn.Daddr[3]&0xFF == 0x7F) {
			return false
		} else if (conn.Saddr[2] == 0x0 && conn.Saddr[3] == 0x01000000) ||
			(conn.Daddr[2] == 0x0 && conn.Daddr[3] == 0x01000000) {
			return false
		}
	}
	return true
}
