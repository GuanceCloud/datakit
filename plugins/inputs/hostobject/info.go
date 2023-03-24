// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package hostobject

import (
	"encoding/json"
	"fmt"
	"net"
	"runtime"
	"strings"
	"sync/atomic"
	"time"

	cpuutil "github.com/shirou/gopsutil/cpu"
	diskutil "github.com/shirou/gopsutil/disk"
	hostutil "github.com/shirou/gopsutil/host"
	loadutil "github.com/shirou/gopsutil/load"
	memutil "github.com/shirou/gopsutil/mem"
	netutil "github.com/shirou/gopsutil/net"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/config"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/http"
	conntrackutil "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/hostutil/conntrack"
	filefdutil "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/hostutil/filefd"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io/election"
)

var (
	DiskUsed               uint64
	DiskFree               uint64
	DiskIOReadBytesPerSec  int64
	DiskIOWriteBytesPerSec int64
	NetRecvBytesPerSec     int64
	NetSendBytesPerSec     int64
)

type (
	HostMetaInfo struct {
		HostName        string `json:"host_name"`
		BootTime        uint64 `json:"boot_time"`
		OS              string `json:"os"`
		Platform        string `json:"platform"`         // ex: ubuntu, linuxmint
		PlatformFamily  string `json:"platform_family"`  // ex: debian, rhel
		PlatformVersion string `json:"platform_version"` // version of the complete OS
		Kernel          string `json:"kernel_release"`
		Arch            string `json:"arch"`
	}

	CPUInfo struct {
		Vendor    string  `json:"vendor_id"`
		Module    string  `json:"module_name"`
		Cores     int32   `json:"cores"`
		Mhz       float64 `json:"mhz"`
		CacheSize int32   `json:"cache_size"`
		Flags     string  `json:"-"`
	}

	MemInfo struct {
		MemoryTotal uint64 `json:"memory_total"`
		SwapTotal   uint64 `json:"swap_total"`
		usedPercent float64
	}

	NetInfo struct {
		Index        int      `json:"-"`
		MTU          int      `json:"mtu"`   // maximum transmission unit
		Name         string   `json:"name"`  // e.g., "en0", "lo0", "eth0.100"
		HardwareAddr string   `json:"mac"`   // IEEE MAC-48, EUI-48 and EUI-64 form
		Flags        []string `json:"flags"` // e.g., FlagUp, FlagLoopback, FlagMulticast
		IP4          string   `json:"ip4"`
		IP6          string   `json:"ip6"`
		IP4All       []string `json:"ip4_all"`
		IP6All       []string `json:"ip6_all"`
		Addrs        []string `json:"-"`
	}

	DiskInfo struct {
		Device string `json:"device"`
		Total  uint64 `json:"total"`
		Fstype string `json:"fstype"`
		Opts   string `json:"-"`
	}

	HostInfo struct {
		HostMeta               *HostMetaInfo       `json:"meta"`
		CPU                    []*CPUInfo          `json:"cpu"`
		Mem                    *MemInfo            `json:"mem"`
		Net                    []*NetInfo          `json:"net"`
		Disk                   []*DiskInfo         `json:"disk"`
		Conntrack              *conntrackutil.Info `json:"conntrack"`
		FileFd                 *filefdutil.Info    `json:"filefd"`
		Election               *ElectionInfo       `json:"election"`
		cpuPercent             float64
		load5                  float64
		cloudInfo              map[string]interface{}
		diskUsedPercent        float64
		diskIOReadBytesPerSec  int64
		diskIOWriteBytesPerSec int64
		netRecvBytesPerSec     int64
		netSendBytesPerSec     int64
		loggingLevel           string
	}
	ElectionInfo struct {
		Namespace string `json:"namespace"`
		Elected   string `json:"elected"`
	}
	HostConfig struct {
		IP         string          `json:"ip"`
		DCAConfig  *http.DCAConfig `json:"dca_config"`
		HTTPListen string          `json:"http_listen"`
	}

	HostObjectMessage struct {
		Host       *HostInfo          `json:"host"`
		Collectors []*CollectorStatus `json:"collectors,omitempty"`
		Config     *HostConfig        `json:"config"`
	}

	CollectorStatus struct {
		Name        string `json:"name"`
		Count       int64  `json:"count"`
		Version     string `json:"version,omitempty"`
		LastTime    int64  `json:"last_time,omitempty"`
		LastErr     string `json:"last_err,omitempty"`
		LastErrTime int64  `json:"last_err_time,omitempty"`
	}
)

var collectorStatHist []*CollectorStatus

func getHostMeta() (*HostMetaInfo, error) {
	info, err := hostutil.Info()
	if err != nil {
		l.Errorf("fail to get host info, %s", err)
		return nil, err
	}

	return &HostMetaInfo{
		HostName:        config.Cfg.Hostname,
		OS:              info.OS,
		BootTime:        info.BootTime,
		Platform:        info.Platform,
		PlatformFamily:  info.PlatformFamily,
		PlatformVersion: info.PlatformVersion,
		Kernel:          info.KernelVersion,
		Arch:            info.KernelArch,
	}, nil
}

func getCPUPercent() float64 {
	ps, err := cpuutil.Percent(0, false)
	if err != nil || len(ps) == 0 {
		l.Warnf("fail to get cpu percent: %s", err)
		return 0
	}
	return ps[0]
}

func getCPUInfo() []*CPUInfo {
	infos, err := cpuutil.Info()
	if err != nil {
		l.Warnf("fail to get cpu info, %s", err)
		return nil
	}

	var objs []*CPUInfo

	for _, info := range infos {
		objs = append(objs, &CPUInfo{
			Vendor:    info.VendorID,
			Module:    info.ModelName,
			Cores:     info.Cores,
			Mhz:       info.Mhz,
			CacheSize: info.CacheSize,
			Flags:     strings.Join(info.Flags, ","),
		})
	}

	return objs
}

func getLoad5() float64 {
	avgstat, err := loadutil.Avg()
	if err != nil {
		l.Errorf("fail to get load info, %s", err)
		return 0
	}

	return avgstat.Load5
}

func getMemInfo() (*MemInfo, error) {
	minfo, err := memutil.VirtualMemory()
	if err != nil {
		l.Error("fail to get memory toal, %s", err)
		return nil, err
	}

	vinfo, err := memutil.SwapMemory()
	if err != nil {
		l.Error("fail to get swap memory toal, %s", err)
		return nil, err
	}

	return &MemInfo{
		MemoryTotal: minfo.Total,
		SwapTotal:   vinfo.Total,
		usedPercent: minfo.UsedPercent,
	}, nil
}

func getNetInfo(enableVIfaces bool) ([]*NetInfo, error) {
	ifs, err := netutil.Interfaces()
	if err != nil {
		l.Errorf("fail to get interfaces, %s", err)
		return nil, err
	}
	var infos []*NetInfo

	netVIfaces := map[string]bool{}
	if !enableVIfaces {
		netVIfaces, _ = NetIgnoreIfaces()
	}

	for _, it := range ifs {
		if _, ok := netVIfaces[it.Name]; ok {
			continue
		}
		i := &NetInfo{
			Index:        it.Index,
			MTU:          it.MTU,
			Name:         it.Name,
			HardwareAddr: it.HardwareAddr,
			Flags:        it.Flags,
		}
		for _, ad := range it.Addrs {
			ip, _, _ := net.ParseCIDR(ad.Addr)
			if ip.IsLoopback() {
				continue
			}
			if ip.To4() != nil {
				i.IP4 = ad.Addr
				i.IP4All = append(i.IP4All, ad.Addr)
			} else if ip.To16() != nil {
				i.IP6 = ad.Addr
				i.IP6All = append(i.IP6All, ad.Addr)
			}
		}
		infos = append(infos, i)
	}
	return infos, nil
}

func getDiskInfo(excludeDevice []string, extraDevice []string, ignoreZeroBytesDisk, onlyPhysicalDevice bool) ([]*DiskInfo, error) {
	l.Debugf("get partitions(physical: %v)...", onlyPhysicalDevice)
	ps, err := diskutil.Partitions(!onlyPhysicalDevice)
	if err != nil {
		l.Errorf("fail to get disk info, %s", err)
		return nil, err
	}
	var infos []*DiskInfo

	excluded := func(x string, arr []string) bool {
		for _, fs := range arr {
			if strings.EqualFold(x, fs) {
				return true
			}
		}
		return false
	}

	for _, p := range ps {
		l.Debugf("hostobject---fstype:%s ,device:%s ,mountpoint:%s ", p.Fstype, p.Device, p.Mountpoint)

		// nolint
		if !strings.HasPrefix(p.Device, "/dev/") && runtime.GOOS != datakit.OSWindows && !excluded(p.Device, extraDevice) {
			continue
		}

		if excluded(p.Device, excludeDevice) {
			continue
		}

		// merge same device
		mergeFlag := false
		for _, cont := range infos {
			if cont.Device == p.Device {
				mergeFlag = true
				break
			}
		}

		if mergeFlag {
			continue
		}

		info := &DiskInfo{
			Device: p.Device,
			Fstype: p.Fstype,
		}

		usage, err := diskutil.Usage(p.Mountpoint)
		if err == nil {
			if ignoreZeroBytesDisk && usage.Total == 0 {
				continue // https://gitlab.jiagouyun.com/cloudcare-tools/datakit/-/issues/505
			}

			info.Total = usage.Total
		}

		l.Debugf("get disk %+#v", info)
		infos = append(infos, info)
	}

	return infos, nil
}

func (ipt *Input) getEnabledInputs() (res []*CollectorStatus) {
	inputsStats, err := io.GetInputsStats() // get all inputs stats
	if err != nil {
		l.Warnf("fail to get inputs stats, %s", err)
		return
	}

	now := time.Now()
	for name, s := range inputsStats {
		ts := s.LastErrTS.Unix()
		if ts < 0 {
			ts = 0
		}

		lastErr := s.LastErr
		if ts > 0 && now.Sub(s.LastErrTS) > ipt.IgnoreInputsErrorsBefore.Duration { // ignore errors 30s ago
			l.Debugf("ignore error %s(%v before)", s.LastErr, now.Sub(s.LastErrTS))
			lastErr = ""
			ts = 0
		}

		res = append(res, &CollectorStatus{
			Name:        name,
			Count:       s.Count,
			LastTime:    s.Last.Unix(),
			LastErr:     lastErr,
			LastErrTime: ts,
			Version:     s.Version,
		})
	}

	return res
}

func (ipt *Input) getHostObjectMessage() (*HostObjectMessage, error) {
	var msg HostObjectMessage

	if !ipt.isTestMode {
		stat := ipt.getEnabledInputs()

		// NOTE: Since the io module may be busy when obtaining the running status
		// information of the collector, it cannot be obtained.

		// Therefore, cache the history here to prevent the collector information
		// field (collectors) in the message field from being empty.
		if len(stat) != 0 {
			collectorStatHist = stat
		}

		msg.Collectors = collectorStatHist
		if len(msg.Collectors) == 0 {
			// This is also to prevent the collector information field from being empty:
			// it is better to discard the current object collection than to cause
			// the collector information to be empty
			//
			// There are two possibilities for the collector information to be empty (or missing):
			//
			// 1: io is busy: it is not convenient to receive query requests
			// 2: A specific collector may be asked to query the running information
			// because it has not had time to start, and the io module must not be registered at this time
			//
			// Therefore, generally only the first collection after startup will not be able to
			// obtain the statistical information, and the follow-up can basically be obtained.
			// Even if it cannot be obtained, the old statistical information will be used instead
			return nil, fmt.Errorf("collector stats missing")
		}
	}

	msg.Config = getHostConfig()

	fileFd, err := filefdutil.GetFileFdInfo()
	if err != nil {
		l.Warnf("filefdutil.GetFileFdInfo(): %s, ignored", err.Error())
	}

	l.Debugf("get host meta...")
	hostMeta, err := getHostMeta()
	if err != nil {
		return nil, err
	}

	l.Debugf("get CPU info...")
	cpuInfo := getCPUInfo()

	l.Debugf("get CPU percent...")
	cpuPercent := getCPUPercent()

	l.Debugf("get load5...")
	load5 := getLoad5()

	l.Debugf("get mem info...")
	mem, err := getMemInfo()
	if err != nil {
		return nil, err
	}

	l.Debugf("get net info...")
	net, err := getNetInfo(ipt.EnableNetVirtualInterfaces)
	if err != nil {
		return nil, err
	}

	l.Debugf("get disk info...")
	disk, err := getDiskInfo(ipt.ExcludeDevice, ipt.ExtraDevice, ipt.IgnoreZeroBytesDisk, ipt.OnlyPhysicalDevice)
	if err != nil {
		return nil, err
	}

	l.Debugf("get conntrack info...")
	conntrack := conntrackutil.GetConntrackInfo()

	l.Debugf("get election info...")
	election := getElectionInfo()

	var diskUsedPercent float64 = 0
	diskUsed := atomic.LoadUint64(&DiskUsed)
	diskFree := atomic.LoadUint64(&DiskFree)
	if diskUsed+diskFree > 0 {
		diskUsedPercent = float64(diskUsed) / (float64(diskUsed) + float64(diskFree)) * 100
	}

	msg.Host = &HostInfo{
		HostMeta:               hostMeta,
		CPU:                    cpuInfo,
		cpuPercent:             cpuPercent,
		load5:                  load5,
		Mem:                    mem,
		Net:                    net,
		Disk:                   disk,
		Conntrack:              conntrack,
		FileFd:                 fileFd,
		Election:               election,
		diskUsedPercent:        diskUsedPercent,
		diskIOReadBytesPerSec:  atomic.LoadInt64(&DiskIOReadBytesPerSec),
		diskIOWriteBytesPerSec: atomic.LoadInt64(&DiskIOWriteBytesPerSec),
		netRecvBytesPerSec:     atomic.LoadInt64(&NetRecvBytesPerSec),
		netSendBytesPerSec:     atomic.LoadInt64(&NetSendBytesPerSec),
		loggingLevel:           config.Cfg.Logging.Level,
	}

	// sync cloud extra fields
	if !ipt.DisableCloudProviderSync {
		_, has := ipt.Tags["cloud_provider"]
		if !has && time.Since(ipt.lastSync) > time.Hour*24 {
			if err := ipt.SetCloudProvider(); err != nil {
				l.Warn(err)
			} else {
				// set cloud provider tag successfully
				has = true
			}
			ipt.lastSync = time.Now()
		}
		if has {
			info, err := ipt.SyncCloudInfo(ipt.Tags["cloud_provider"])
			if err != nil {
				l.Warnf("sync cloud info failed: %v, ignored", err)
			} else {
				j, err := json.Marshal(info)
				if err != nil {
					l.Warn(err)
				} else {
					info["extra_cloud_meta"] = string(j)
				}
			}
			msg.Host.cloudInfo = info
		}
	}

	return &msg, nil
}

func getHostConfig() *HostConfig {
	hostConfig := &HostConfig{}

	ip, err := datakit.LocalIP()
	if err == nil {
		hostConfig.IP = ip
	}

	hostConfig.DCAConfig = config.Cfg.DCAConfig

	hostConfig.HTTPListen = config.Cfg.HTTPAPI.Listen

	return hostConfig
}

func getElectionInfo() *ElectionInfo {
	electionInfo := &ElectionInfo{}
	if config.Cfg.Election.Enable {
		elected, namespace, _ := election.Elected()
		electionInfo.Elected = elected
		electionInfo.Namespace = namespace
		return electionInfo
	}
	return nil
}
