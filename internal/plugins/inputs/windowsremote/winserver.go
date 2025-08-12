// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

//go:build windows
// +build windows

package windowsremote

import (
	"encoding/json"
	"fmt"
	"net"
	"strings"
	"sync"
	"time"

	"github.com/GuanceCloud/cliutils/point"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/ntp"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs/windowsremote/wmi"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs/windowsremote/wmi/class"
)

type winServer struct {
	ip                     string
	user, pw               string
	host                   string
	conn                   *wmi.SWbemServices
	CPU                    class.Win32_Processor
	System                 class.Win32_OperatingSystem
	Disk                   []class.Win32_LogicalDisk
	Net                    []class.Win32_NetworkAdapterConfiguration
	Process                []class.Process
	EnableNetwork          []string
	cpuPercent             float64
	diskIOReadBytesPerSec  int64
	diskIOWriteBytesPerSec int64
	netRecvBytesPerSec     int64
	netSendBytesPerSec     int64
	tags                   map[string]string
	lock                   sync.RWMutex
	lastLogTime            time.Time
}

func newServer(ip string, user, pw string) *winServer {
	s := &winServer{ip: ip, user: user, pw: pw}
	conn, err := initWmiClient(ip, user, pw)
	if err != nil {
		l.Errorf("initWmiClient ip:%s user=%s ,pw=%s error: %v", ip, user, pw, err)
		return nil
	}
	s.conn = conn
	return s
}

func (s *winServer) toObjectPoints() []*point.Point {
	s.lock.Lock()
	defer s.lock.Unlock()
	var pts []*point.Point
	//组建hostinfo.
	hostInfo := &HostInfo{
		HostMeta:               &HostMetaInfo{},
		CPU:                    make([]*CPUInfo, 0),
		Mem:                    &MemInfo{},
		Net:                    make([]*NetInfo, 0),
		Disk:                   make([]*DiskInfo, 0),
		cpuPercent:             s.cpuPercent,
		diskIOReadBytesPerSec:  s.diskIOReadBytesPerSec,
		diskIOWriteBytesPerSec: s.diskIOWriteBytesPerSec,
		netRecvBytesPerSec:     s.netRecvBytesPerSec,
		netSendBytesPerSec:     s.netSendBytesPerSec,
	}
	hostMate := HostMetaInfo{
		HostName:        strings.ToLower(s.System.CSName),
		BootTime:        uint64(s.System.LastBootUpTime.UnixMilli() / 1000),
		OS:              "windows",
		Platform:        s.System.Caption,
		PlatformFamily:  "",
		PlatformVersion: fmt.Sprintf("%s (Build %s)", s.System.Version, s.System.BuildNumber),
		Kernel:          fmt.Sprintf("%s (Build %s)", s.System.Version, s.System.BuildNumber),
		Arch:            s.System.OSArchitecture,
	}

	hostInfo.HostMeta = &hostMate

	used := s.System.TotalVisibleMemorySize - s.System.FreePhysicalMemory
	perent := float64(used) / float64(s.System.TotalVisibleMemorySize)
	mem := &MemInfo{
		MemoryTotal: s.System.TotalVisibleMemorySize * 1024,
		SwapTotal:   s.System.TotalVirtualMemorySize * 1024,
		usedPercent: perent,
	}
	hostInfo.Mem = mem
	cpu := &CPUInfo{
		Vendor:    s.CPU.Manufacturer,
		Module:    s.CPU.Name,
		Cores:     int32(s.CPU.NumberOfLogicalProcessors),
		Mhz:       float64(s.CPU.MaxClockSpeed),
		CacheSize: int32(s.CPU.L2CacheSize + s.CPU.L3CacheSize),
		Flags:     "",
	}
	hostInfo.CPU = append(hostInfo.CPU, cpu)
	for i, adapter := range s.Net {
		v4, v6, _ := classifyIPs(adapter.IPAddress)
		ip4, ip6 := "", ""
		if len(v4) > 0 {
			ip4 = v4[0]
		}
		if len(v6) > 0 {
			ip6 = v6[0]
		}
		net := &NetInfo{
			Index:        i,
			MTU:          0, //待定
			Name:         adapter.Description,
			HardwareAddr: "",
			Flags:        nil,
			IP4:          ip4,
			IP6:          ip6,
			IP4All:       v4,
			IP6All:       v6,
			Addrs:        adapter.IPAddress,
		}
		hostInfo.Net = append(hostInfo.Net, net)
	}
	total, userd := uint64(0), uint64(0)
	for _, logicalDisk := range s.Disk {
		disk := &DiskInfo{
			Device:     logicalDisk.DeviceID,
			Total:      logicalDisk.Size,
			Fstype:     logicalDisk.FileSystem,
			MountPoint: logicalDisk.Name,
			Opts:       "",
		}
		total += logicalDisk.Size
		userd += logicalDisk.Size - logicalDisk.FreeSpace
		hostInfo.Disk = append(hostInfo.Disk, disk)
	}

	diskUsedPercent := float64(userd) / float64(total)
	message := &HostObjectMessage{
		Host:       hostInfo,
		Collectors: nil,
		Config:     nil,
	}
	messageData, err := json.Marshal(message)
	if err != nil {
		l.Errorf("json marshal err:%s", err.Error())
		return nil
	}

	l.Debugf("messageData len: %d", len(messageData))
	var kvs point.KVs
	for k, v := range s.tags {
		kvs = kvs.AddTag(k, v)
	}
	kvs = kvs.Set("message", string(messageData)).
		Set("start_time", message.Host.HostMeta.BootTime*1000).
		Set("datakit_ver", datakit.Version).
		Set("cpu_usage", message.Host.cpuPercent).
		Set("mem_used_percent", message.Host.Mem.usedPercent).
		// Add("load", message.Host.load5, false, true).
		Set("disk_used_percent", diskUsedPercent).
		Set("diskio_read_bytes_per_sec", s.diskIOReadBytesPerSec).
		Set("diskio_write_bytes_per_sec", s.diskIOWriteBytesPerSec).
		Set("net_recv_bytes_per_sec", s.netRecvBytesPerSec).
		Set("net_send_bytes_per_sec", s.netSendBytesPerSec).
		// Add("logging_level", message.Host.loggingLevel, false, true).
		AddTag("name", message.Host.HostMeta.HostName).
		AddTag("os", message.Host.HostMeta.OS).
		Add("num_cpu", s.CPU.NumberOfLogicalProcessors).
		// AddTag("unicast_ip", message.Config.IP).
		Set("disk_total", total).
		AddTag("arch", message.Host.HostMeta.Arch).
		AddTag("host", s.host).
		AddTag("ip", s.ip)
	opts := point.DefaultObjectOptions()
	pt := point.NewPoint(hostobjectMeasurement, kvs, opts...)
	pt.SetTime(ntp.Now())
	pts = append(pts, pt)

	for _, p := range s.Process {
		pts = append(pts, p.ToPoint(s.tags))
	}
	l.Debugf("toObjectPoints pts %d", len(pts))
	return pts
}

func (s *winServer) beginCollectObject() {
	s.lock.Lock()
	defer s.lock.Unlock()
	defer func() {
		if err := recover(); err != nil {
			l.Errorf("err:%v", err)
			// 捕捉到异常 尝试重连
			s.conn, err = initWmiClient(s.ip, s.user, s.pw)
			if err != nil {
				l.Errorf("initWmiClient ip:%s error: %v", s.ip, err)
			}
		}
	}()
	// 同步进行 因为单个wmi连接不允许并行查询多个class。
	cpus, err := wmi.QueryProcessors(s.conn)
	if err != nil {
		l.Errorf("wmi query cpu error: %s", err)
	} else {
		s.CPU = cpus[0]
	}

	systems, err := wmi.QueryOperatingSystem(s.conn)
	if err != nil {
		l.Errorf("wmi query system error: %s", err)
	} else {
		s.System = systems[0]
		s.host = strings.ToLower(s.System.CSName)
	}

	network, err := wmi.QueryNetworkAdapterConfiguration(s.conn)
	if err != nil {
		l.Errorf("wmi query network adapter configuration error: %s", err)
	} else {
		s.Net = network
		// network 需要收集已经打开的网卡的Name，没有打开的 不采集其指标信息。
		s.EnableNetwork = []string{}
		for _, net := range network {
			l.Debugf("add network Description: %s", net.Description)
			s.EnableNetwork = append(s.EnableNetwork, net.Description)
		}
	}

	disk, err := wmi.QueryDisk(s.conn)
	if err != nil {
		l.Errorf("wmi query disk error: %s", err)
	} else {
		s.Disk = disk
	}

	process, err := wmi.QueryProcess(s.conn)
	if err != nil {
		l.Errorf("wmi query process error: %s", err)
	} else {
		l.Debugf("get process len=$d", len(process))
		s.Process = process
	}
}

func (s *winServer) collectMetric(now time.Time) []*point.Point {
	s.lock.Lock()
	defer s.lock.Unlock()
	var pts []*point.Point
	// cpu
	perCPU, err := wmi.QueryPerfOSProcessor(s.conn)
	if err != nil {
		l.Errorf("wmi query percpu error: %s", err)
	} else {
		s.cpuPercent = float64(perCPU[0].PercentProcessorTime) / 100
		pts = append(pts, perCPU[0].ToPoint(s.host, &s.CPU))
	}
	// disk
	disks, err := wmi.QueryDisk(s.conn)
	if err != nil {
		l.Errorf("wmi query disk error: %s", err)
	} else {
		s.Disk = disks
		for _, disk := range disks {
			pts = append(pts, disk.ToPoint())
		}
	}
	// diskio
	disdios, err := wmi.QueryPerOSDisk(s.conn)
	diskReadPerSec, diskWritePerSec := int64(0), int64(0)
	for _, diskio := range disdios {
		diskReadPerSec += int64(diskio.DiskReadsPersec)
		diskWritePerSec += int64(diskio.DiskWritesPersec)
		pts = append(pts, diskio.ToPoint(s.host))
	}
	s.diskIOReadBytesPerSec = diskReadPerSec
	s.diskIOWriteBytesPerSec = diskWritePerSec
	// mem.
	system, err := wmi.QueryOperatingSystem(s.conn)
	if err != nil {
		l.Errorf("wmi query system error: %s", err)
	} else {
		pts = append(pts, system[0].ToMemPoint(s.host))
	}
	// net.
	netSendSec, netRecvSev := int64(0), int64(0)
	for _, name := range s.EnableNetwork {
		networkInterface, err := wmi.QueryNetworkInterface(s.conn, name)
		if err != nil {
			l.Errorf("wmi query network interface error: %s", err)
		} else {
			netSendSec += int64(networkInterface.BytesSentPersec)
			netRecvSev += int64(networkInterface.BytesReceivedPersec)
			pts = append(pts, networkInterface.ToPoint(s.host))
		}
	}

	s.netSendBytesPerSec = netSendSec
	s.netRecvBytesPerSec = netRecvSev
	// -- 设置统一的时间
	for _, pt := range pts {
		pt.SetTime(now)
		if len(s.tags) > 0 {
			var kvs point.KVs
			for k, v := range s.tags {
				kvs = kvs.AddTag(k, v)
			}
			pt.SetKVs(kvs...)
		}
	}
	l.Debugf("collect metric len=%d", len(pts))
	return pts
}

func (s *winServer) collectLog() []*point.Point {
	s.lock.Lock()
	defer s.lock.Unlock()
	var pts []*point.Point
	if s.lastLogTime.IsZero() {
		s.lastLogTime = time.Now().Add(-60 * time.Second)
	}
	events, err := wmi.QueryLogEvent(s.conn, s.lastLogTime)
	if err != nil {
		l.Errorf("wmi query log error: %s", err)
		return pts
	}
	s.lastLogTime = time.Now()
	for _, event := range events {
		pts = append(pts, event.ToPoint(s.tags))
	}
	return pts
}

func initWmiClient(ip, user, pw string) (conn *wmi.SWbemServices, err error) {
	s, err := wmi.InitializeSWbemServices(&wmi.Client{
		NonePtrZero:         false,
		PtrNil:              false,
		AllowMissingFields:  true,
		SWbemServicesClient: nil,
	}, []interface{}{ip, "root\\cimv2", user, pw})
	if err != nil {
		wmiConnCount.WithLabelValues("err").Add(1)
		return nil, err
	}
	wmiConnCount.WithLabelValues("ok").Add(1)
	return s, nil
}

func classifyIPs(ips []string) (ipv4s []string, ipv6s []string, invalidIPs []string) {
	for _, ip := range ips {
		parsedIP := net.ParseIP(ip)
		if parsedIP == nil {
			invalidIPs = append(invalidIPs, ip)
			continue
		}

		// 检查是IPv4还是IPv6
		if parsedIP.To4() != nil {
			ipv4s = append(ipv4s, ip)
		} else {
			ipv6s = append(ipv6s, ip)
		}
	}

	return
}
