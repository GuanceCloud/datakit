//go:build linux
// +build linux

package l4log

import (
	"context"
	"fmt"
	"os"
	"runtime"
	"time"

	"github.com/GuanceCloud/platypus/pkg/ast"
	"github.com/google/gopacket/afpacket"
	"github.com/vishvananda/netns"
	cruntime "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/container/runtime"
)

type cbRawSocket struct {
	// nic_name, nic_mac
	tps map[[2]string]*afpacket.TPacket

	hostNS       bool
	ns           string
	newSocketErr map[[2]string]error
}

func (cb *cbRawSocket) cbNewRawSocket() {
	for nameAndMac, skt := range cb.tps {
		if skt != nil {
			continue
		}

		opt := []any{afpacket.OptInterface(nameAndMac[0])}
		if cb.hostNS {
			// https://www.kernel.org/doc/Documentation/networking/packet_mmap.txt
			// frame nr: 64 * 128, frame size 4096Byte, total ~ 32MiB
			opt = append(opt, afpacket.OptNumBlocks(64))
		}

		if h, err := newRawsocket(newBPFFilter(), opt...); err != nil {
			if cb.newSocketErr == nil {
				cb.newSocketErr = make(map[[2]string]error)
			}
			cb.newSocketErr[nameAndMac] = fmt.Errorf("new raw socket: %w", err)
			continue
		} else {
			log.Infof("new raw socket %s %s %s", nameAndMac[0], nameAndMac[1], cb.ns)
			time.Sleep(time.Millisecond * 100)
			cb.tps[nameAndMac] = h
		}
	}
}

type ifaceInfomation struct {
	// h     *afpacket.TPacket
	conns *TCPConns
	cacel context.CancelFunc

	ifaces [2]string
}

type netnsInformation struct {
	hostNS      bool
	nsUID       string
	contianerID string

	// netns netns.NsHandle
	nns *netnsHandle

	pid map[int]struct{}
	// mac:
	ifaceInf map[[2]string]*ifaceInfomation

	tags map[string]string
}

func (nsInf *netnsInformation) Close() {
	for _, v := range nsInf.ifaceInf {
		if v.cacel != nil {
			v.cacel()
		}
	}
	if nsInf.nns != nil {
		nsInf.nns.close()
	}
}

type netlogMonitor struct {
	netnsInfo map[string]*netnsInformation

	gtags map[string]string

	transportBlacklist ast.Stmts
	filterRuntime      *filterRuntime

	portListen *portListen
}

func newNetlogMonitor(gtags map[string]string, blacklist string, fnG *fnGroup,
) (*netlogMonitor, error) {
	m := &netlogMonitor{
		netnsInfo:     map[string]*netnsInformation{},
		gtags:         gtags,
		portListen:    &portListen{},
		filterRuntime: &filterRuntime{fnG: fnG},
	}

	if blacklist != "" {
		stmts, err := parseFilter(blacklist)
		if err != nil {
			return nil, fmt.Errorf("parse filter: %w", err)
		}
		err = m.filterRuntime.checkStmts(stmts, &netParams{})
		if err != nil {
			return nil, fmt.Errorf("check filter: %w", err)
		}

		m.transportBlacklist = stmts
		log.Infof("transport blacklist: \n\n%s\n", blacklist)
	}

	return m, nil
}

func (m *netlogMonitor) Run(ctx context.Context, containerCtr []cruntime.ContainerRuntime,
) {
	ticker := time.NewTicker(time.Second * 20)
	defer ticker.Stop()
	var allowLo bool

	for {
		netnsInfo := ListContainersAndHostNetNS(containerCtr, allowLo)
		m.CmpAndCleanNetNsNIC(netnsInfo)
		m.CmpAndAddNIC(netnsInfo)
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
		}
	}
}

func (m *netlogMonitor) CmpAndCleanNetNsNIC(netnsInfo map[string]*netnsInformation) {
	// 如果该命名空间在当前不存在：
	// 移除整个网络命名空间下的 raw_socket
	for nsUID, infom := range m.netnsInfo {
		if _, ok := netnsInfo[nsUID]; !ok {
			for _, v := range infom.ifaceInf {
				if v.cacel != nil {
					v.cacel()
				}
			}

			// 删除当前 ns 的信息
			delete(m.netnsInfo, nsUID)
		}
	}
}

func (m *netlogMonitor) CmpAndAddNIC(netnsInfo map[string]*netnsInformation) {
	for nsUID, nsInf := range netnsInfo {
		diffTps := map[[2]string]*afpacket.TPacket{}
		diffTpsNicInfo := map[[2]string]*NICInfo{}

		nicIP := []string{}
		if v, err := nsInf.nns.nicInfo(); err != nil {
			log.Errorf("get network interface info: %w, ns: %s", err, nsUID)
			continue
		} else {
			for _, v := range v {
				for _, v := range v.Addrs {
					nicIP = append(nicIP, v.IP.String())
				}
				diffTps[[2]string{v.Name, v.MAC}] = nil
				diffTpsNicInfo[[2]string{v.Name, v.MAC}] = v
			}
		}

		preNsInf, ok := m.netnsInfo[nsUID]
		if !ok {
			preNsInf = nsInf
			m.netnsInfo[nsUID] = preNsInf
		} else {
			nsInf.Close()
			for k, v := range preNsInf.ifaceInf {
				// 当前该网卡可能被移除或者 down 了
				if _, ok := diffTps[k]; !ok {
					if v.cacel != nil {
						// 取消抓包任务并关闭该 raw_socket
						v.cacel()
					}
					delete(preNsInf.ifaceInf, k)
				}
			}

			for k := range diffTps {
				// 已经开启采集的则不再采集
				if _, ok := preNsInf.ifaceInf[k]; ok {
					delete(diffTps, k)
				}
			}
		}

		// 生成新的回调用于创建 raw_socket
		cbRawSkt := &cbRawSocket{
			hostNS: preNsInf.hostNS,
			ns:     NSInode(preNsInf.nns.netns),
			tps:    diffTps,
		}

		// 创建 raw_socket
		if err := CallWithNetNS(preNsInf.nns.netns, cbRawSkt.cbNewRawSocket); err != nil {
			log.Errorf("call with netns: %w", err)
			preNsInf.Close()
			continue
		}

		for _, v := range cbRawSkt.newSocketErr {
			log.Error(v)
		}

		for idx, h := range diffTps {
			// 未被采集且 raw_socket fd 存在
			if _, ok := preNsInf.ifaceInf[idx]; !ok && h != nil {
				ctx, cacel := context.WithCancel(context.Background())
				tags := map[string]string{}
				for k, v := range m.gtags {
					tags[k] = v
				}
				for k, v := range preNsInf.tags {
					tags[k] = v
				}
				conns := NewTCPConns(tags, preNsInf.contianerID, preNsInf.nsUID,
					idx, m.portListen, m.transportBlacklist, m.filterRuntime)

				if inf, ok := diffTpsNicInfo[idx]; ok && inf != nil {
					conns.virtualNIC = inf.VIface
					conns.hostNetwork = inf.HostIface
				}
				preNsInf.ifaceInf[idx] = &ifaceInfomation{
					cacel:  cacel,
					ifaces: idx,
					conns:  conns,
				}
				time.Sleep(time.Millisecond * 100)
				go conns.CapturePacket(ctx, idx[0], idx[1], nsUID, h)
				go conns.Gather(context.Background(), nicIP)

				pids := map[int]struct{}{}
				for k := range preNsInf.pid {
					pids[k] = struct{}{}
				}
				go preNsInf.nns.tcpPortListenWatcher(ctx, m.portListen, pids)
			} else if h != nil {
				// 不使用则关闭 raw_socket
				h.Close()
			}
		}
	}
}

func CallWithNetNS(newNS netns.NsHandle, fn func()) error {
	if !newNS.IsOpen() {
		return fmt.Errorf("ns fd closed")
	}

	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	prevNS, err := netns.Get()
	if err != nil {
		return err
	}

	defer prevNS.Close() //nolint:errcheck

	if newNS.Equal(prevNS) {
		// call function from param
		fn()
	} else {
		// switch to new network namespace
		if err := netns.Set(newNS); err != nil {
			return fmt.Errorf("switch netns failed: %w", err)
		}
		// call function from param
		fn()

		// revert to previous network namespace
		if err := netns.Set(prevNS); err != nil {
			return err
		}
	}

	return nil
}

func ListContainersAndHostNetNS(ctrLi []cruntime.ContainerRuntime, allowLo bool,
) map[string]*netnsInformation {
	netnsInfo := map[string]*netnsInformation{}
	var curNetnsStr string
	curNetns, err := netns.GetFromPid(os.Getpid())
	if err != nil {
		log.Errorf("get netns from pid: %w", err)
	} else {
		curNetnsStr = NSInode(curNetns)

		// add host network namespace
		if _, ok := netnsInfo[curNetnsStr]; !ok {
			netnsInfo[curNetnsStr] = &netnsInformation{
				hostNS:      true,
				nsUID:       curNetnsStr,
				nns:         newNetNsHandle(true, allowLo, curNetns),
				contianerID: "",
				ifaceInf:    map[[2]string]*ifaceInfomation{},
			}
		} else {
			if err := curNetns.Close(); err != nil {
				log.Error(err)
			}
		}
	}

	for _, containerdCtr := range ctrLi {
		if containerdCtr != nil {
			ctrs, err := containerdCtr.ListContainers()
			if err != nil {
				log.Errorf("get containers: %s", err.Error())
			}
			for _, c := range ctrs {
				nsH, err := netns.GetFromPid(c.Pid)
				if err != nil {
					log.Errorf("get netns from pid: %w", err)
					continue
				}
				nsHStr := NSInode(nsH)
				if nsHStr == curNetnsStr { // skip host network
					if err := nsH.Close(); err != nil {
						log.Error(err)
					}
					continue
				}

				k8sTags := getK8sTags(c.Labels)
				if v, ok := netnsInfo[nsHStr]; !ok {
					netnsInfo[nsHStr] = &netnsInformation{
						nsUID:       nsHStr,
						nns:         newNetNsHandle(false, allowLo, nsH),
						contianerID: c.ID,
						ifaceInf:    map[[2]string]*ifaceInfomation{},
						pid:         map[int]struct{}{c.Pid: {}},
						tags:        k8sTags,
					}
				} else {
					v.pid[c.Pid] = struct{}{}
					if err := nsH.Close(); err != nil {
						log.Error(err)
					}
				}
			}
		}
	}
	return netnsInfo
}

func getK8sTags(labels map[string]string) map[string]string {
	if len(labels) == 0 {
		return nil
	}
	v := map[string]string{}
	v["k8s_namespace"] = labels["io.kubernetes.pod.namespace"]
	v["k8s_pod_name"] = labels["io.kubernetes.pod.name"]
	v["k8s_container_name"] = labels["io.kubernetes.container.name"]
	return v
}
