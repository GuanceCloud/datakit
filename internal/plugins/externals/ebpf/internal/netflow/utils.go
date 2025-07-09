//go:build linux
// +build linux

package netflow

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"math"
	"net"
	"os"
	"sync"
	"sync/atomic"
	"time"

	manager "github.com/DataDog/ebpf-manager"
	"github.com/GuanceCloud/cliutils/logger"
	"github.com/cilium/ebpf"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/ntp"
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
		TS: ntp.Now(),
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
	ts := ntp.Now()
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

func AddClientServerInf(mtags map[string]string, mfields map[string]any) (map[string]string, map[string]any) {
	direction, ok := mtags["direction"]
	if !ok {
		return mtags, mfields
	}

	if v, ok := mtags["dst_domain"]; ok {
		mtags["server_domain"] = v
	}

	switch direction {
	case DirectionIncoming:
		mtags["conn_side"] = "server"

		mtags["server_ip"] = mtags["src_ip"]
		mtags["server_ip_type"] = mtags["dst_ip_type"]
		mtags["server_port"] = mtags["src_port"]

		mtags["client_ip"] = mtags["dst_ip"]
		mtags["client_ip_type"] = mtags["dst_ip_type"]
		mtags["client_port"] = mtags["dst_port"]

		if v, ok := mfields["bytes_read"]; ok {
			mfields["client_sent"] = v
		}
		if v, ok := mfields["bytes_written"]; ok {
			mfields["server_sent"] = v
		}
	default:
		mtags["conn_side"] = "client"

		mtags["server_ip"] = mtags["dst_ip"]
		mtags["server_ip_type"] = mtags["dst_ip_type"]
		mtags["server_port"] = mtags["dst_port"]

		mtags["client_ip"] = mtags["src_ip"]
		mtags["client_ip_type"] = mtags["src_ip_type"]
		mtags["client_port"] = mtags["src_port"]

		if v, ok := mfields["bytes_written"]; ok {
			mfields["client_sent"] = v
		}
		if v, ok := mfields["bytes_read"]; ok {
			mfields["server_sent"] = v
		}
	}

	return mtags, mfields
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

func addNoValueK8s(src bool, mTags map[string]string) map[string]string {
	if mTags == nil {
		mTags = map[string]string{}
	}

	client := true
	if v, ok := mTags["direction"]; ok {
		if v == DirectionIncoming {
			client = false
		}
	}

	if src {
		mTags["src_k8s_namespace"] = NoValue
		mTags["src_k8s_pod_name"] = NoValue
		mTags["src_k8s_service_name"] = NoValue
		mTags["src_k8s_deployment_name"] = NoValue
		mTags["src_k8s_workload_name"] = NoValue
		mTags["src_k8s_workload_type"] = NoValue
	} else {
		mTags["dst_k8s_namespace"] = NoValue
		mTags["dst_k8s_pod_name"] = NoValue
		mTags["dst_k8s_service_name"] = NoValue
		mTags["dst_k8s_deployment_name"] = NoValue
		mTags["dst_k8s_workload_name"] = NoValue
		mTags["dst_k8s_workload_type"] = NoValue
	}

	if (client && src) || (!client && !src) {
		mTags["client_k8s_namespace"] = NoValue
		mTags["client_k8s_pod_name"] = NoValue
		mTags["client_k8s_service_name"] = NoValue
		mTags["client_k8s_deployment_name"] = NoValue
		mTags["client_k8s_workload_name"] = NoValue
		mTags["client_k8s_workload_type"] = NoValue
	} else {
		mTags["server_k8s_namespace"] = NoValue
		mTags["server_k8s_pod_name"] = NoValue
		mTags["server_k8s_service_name"] = NoValue
		mTags["server_k8s_deployment_name"] = NoValue
		mTags["server_k8s_workload_name"] = NoValue
		mTags["server_k8s_workload_type"] = NoValue
	}

	return mTags
}

func addPodTag(src bool, k8stag *cli.K8sTag, mTags map[string]string) map[string]string {
	if mTags == nil {
		mTags = map[string]string{}
	}
	client := true
	if v, ok := mTags["direction"]; ok {
		if v == DirectionIncoming {
			client = false
		}
	}

	if src {
		mTags["src_k8s_namespace"] = k8stag.NS
		mTags["src_k8s_pod_name"] = k8stag.PodName
		mTags["src_k8s_service_name"] = k8stag.SvcName
		mTags["src_k8s_deployment_name"] = k8stag.WorkloadName
		mTags["src_k8s_workload_name"] = k8stag.WorkloadName
		mTags["src_k8s_workload_type"] = k8stag.Kind.String()
		for k, v := range k8stag.Labels {
			mTags[k] = v
		}
	} else {
		mTags["dst_k8s_namespace"] = k8stag.NS
		mTags["dst_k8s_pod_name"] = k8stag.PodName
		mTags["dst_k8s_service_name"] = k8stag.SvcName
		mTags["dst_k8s_deployment_name"] = k8stag.WorkloadName
		mTags["dst_k8s_workload_name"] = k8stag.WorkloadName
		mTags["dst_k8s_workload_type"] = k8stag.Kind.String()
	}

	if (client && src) || (!client && !src) {
		mTags["client_k8s_namespace"] = k8stag.NS
		mTags["client_k8s_pod_name"] = k8stag.PodName
		mTags["client_k8s_service_name"] = k8stag.SvcName
		mTags["client_k8s_deployment_name"] = k8stag.WorkloadName
		mTags["client_k8s_workload_name"] = k8stag.WorkloadName
		mTags["client_k8s_workload_type"] = k8stag.Kind.String()
	} else {
		mTags["server_k8s_namespace"] = k8stag.NS
		mTags["server_k8s_pod_name"] = k8stag.PodName
		mTags["server_k8s_service_name"] = k8stag.SvcName
		mTags["server_k8s_deployment_name"] = k8stag.WorkloadName
		mTags["server_k8s_workload_name"] = k8stag.WorkloadName
		mTags["server_k8s_workload_type"] = k8stag.Kind.String()
	}
	return mTags
}

func addSvcTag(src bool, t *cli.PodChainSvc, mTags map[string]string) map[string]string {
	if mTags == nil {
		mTags = map[string]string{}
	}

	client := true
	if v, ok := mTags["direction"]; ok {
		if v == DirectionIncoming {
			client = false
		}
	}

	if src {
		mTags["src_k8s_namespace"] = t.Chain.Tag.NS
		mTags["src_k8s_pod_name"] = NoValue
		mTags["src_k8s_service_name"] = t.Svc.Name
		mTags["src_k8s_deployment_name"] = t.Chain.Tag.WorkloadName
		mTags["src_k8s_workload"] = t.Chain.Tag.WorkloadName
		mTags["src_k8s_workload_type"] = t.Chain.Tag.Kind.String()
	} else {
		mTags["dst_k8s_namespace"] = t.Chain.Tag.NS
		mTags["dst_k8s_pod_name"] = NoValue
		mTags["dst_k8s_service_name"] = t.Svc.Name
		mTags["dst_k8s_deployment_name"] = t.Chain.Tag.WorkloadName
		mTags["dst_k8s_workload"] = t.Chain.Tag.WorkloadName
		mTags["dst_k8s_workload_type"] = t.Chain.Tag.Kind.String()
	}

	if (client && src) || (!client && !src) {
		mTags["client_k8s_namespace"] = t.Chain.Tag.NS
		mTags["client_k8s_pod_name"] = NoValue
		mTags["client_k8s_service_name"] = t.Svc.Name
		mTags["client_k8s_deployment_name"] = t.Chain.Tag.WorkloadName
		mTags["client_k8s_workload"] = t.Chain.Tag.WorkloadName
		mTags["client_k8s_workload_type"] = t.Chain.Tag.Kind.String()
	} else {
		mTags["server_k8s_namespace"] = t.Chain.Tag.NS
		mTags["server_k8s_pod_name"] = NoValue
		mTags["server_k8s_service_name"] = t.Svc.Name
		mTags["server_k8s_deployment_name"] = t.Chain.Tag.WorkloadName
		mTags["server_k8s_workload"] = t.Chain.Tag.WorkloadName
		mTags["server_k8s_workload_type"] = t.Chain.Tag.Kind.String()
	}

	return mTags
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
			mTags = addPodTag(true, t, mTags)
		} else if t, ok := k8sNetInfo.QuerySvcInfo(
			basekey.Transport, basekey.SAddr, int(basekey.SPort)); ok && t != nil {
			srcK8sFlag = true
			mTags = addSvcTag(true, t, mTags)
		}

		if basekey.DNATAddr != "" && basekey.DNATPort != 0 {
			if t, ok := k8sNetInfo.QueryPodInfo(
				0, basekey.DNATAddr, int(basekey.DNATPort), basekey.Transport); ok && t != nil {
				dstK8sFlag = true
				mTags = addPodTag(false, t, mTags)
				goto skip_dst
			}
		}

		if t, ok := k8sNetInfo.QueryPodInfo(0,
			basekey.DAddr, int(basekey.DPort), basekey.Transport); ok && t != nil {
			// k.dport
			dstK8sFlag = true
			addPodTag(false, t, mTags)
		} else {
			if t, ok := k8sNetInfo.QuerySvcInfo(basekey.Transport,
				basekey.DAddr, int(basekey.DPort)); ok && t != nil {
				dstK8sFlag = true
				mTags = addSvcTag(false, t, mTags)
			}
		}

	skip_dst:
		if srcK8sFlag || dstK8sFlag {
			mTags["sub_source"] = "K8s"
			if !srcK8sFlag {
				addNoValueK8s(true, mTags)
			}
			if !dstK8sFlag {
				addNoValueK8s(false, mTags)
			}
		}
	}
	return mTags
}

func U32BEToIPv4Array(addr uint32) [4]byte {
	ip := [4]byte{}
	binary.LittleEndian.PutUint32(ip[:], addr)
	return ip
}

func U32IP4(addr uint32) net.IP {
	ip4 := U32BEToIPv4Array(addr)
	return net.IP(ip4[:])
}

func U32BEToIPv6Array(addr [4]uint32) [16]byte {
	var ip [16]byte
	for x := 0; x < 4; x++ {
		binary.LittleEndian.PutUint32(ip[x*4:], addr[x])
	}
	return ip
}

func U32IP6(addr [4]uint32) net.IP {
	ip6 := U32BEToIPv6Array(addr)
	return net.IP(ip6[:])
}

func U32BEToIP(addr [4]uint32, isIPv6 bool) net.IP {
	if isIPv6 {
		return U32IP6(addr)
	} else {
		return U32IP4(addr[3])
	}
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
