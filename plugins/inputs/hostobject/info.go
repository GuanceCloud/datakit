package hostobject

import (
	"encoding/json"
	"fmt"
	"net"
	"strings"
	"time"

	cpuutil "github.com/shirou/gopsutil/cpu"
	diskutil "github.com/shirou/gopsutil/disk"
	hostutil "github.com/shirou/gopsutil/host"
	loadutil "github.com/shirou/gopsutil/load"
	memutil "github.com/shirou/gopsutil/mem"
	netutil "github.com/shirou/gopsutil/net"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"
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
		Addrs        []string `json:"-"`
	}

	DiskInfo struct {
		Device     string `json:"device"`
		Total      uint64 `json:"total"`
		Mountpoint string `json:"mountpoint"`
		Fstype     string `json:"fstype"`
		Opts       string `json:"-"`
	}

	HostInfo struct {
		HostMeta   *HostMetaInfo `json:"meta"`
		CPU        []*CPUInfo    `json:"cpu"`
		Mem        *MemInfo      `json:"mem"`
		Net        []*NetInfo    `json:"net"`
		Disk       []*DiskInfo   `json:"disk"`
		cpuPercent float64
		load5      float64
		cloudInfo  map[string]interface{}
	}

	HostObjectMessage struct {
		Host       *HostInfo          `json:"host"`
		Collectors []*CollectorStatus `json:"collectors,omitempty"`
	}

	CollectorStatus struct {
		Name        string `json:"name"`
		Count       int64  `json:"count"`
		LastTime    int64  `json:"last_time,omitempty"`
		LastErr     string `json:"last_err,omitempty"`
		LastErrTime int64  `json:"last_err_time,omitempty"`
	}
)

var (
	collectorStatHist []*CollectorStatus
)

func getHostMeta() *HostMetaInfo {
	info, err := hostutil.Info()
	if err != nil {
		l.Errorf("fail to get host info, %s", err)
		return nil
	}

	return &HostMetaInfo{
		//HostName:        info.Hostname,
		// 此处用户可能自定义 Hostname，如果用户不
		// 定义 Hostname，那么 datakit.Cfg.Hostname == info.Hostname
		HostName:        datakit.Cfg.Hostname,
		OS:              info.OS,
		BootTime:        info.BootTime,
		Platform:        info.Platform,
		PlatformFamily:  info.PlatformFamily,
		PlatformVersion: info.PlatformVersion,
		Kernel:          info.KernelVersion,
		Arch:            info.KernelArch,
	}
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
		l.Errorf("fail to get cpu info, %s", err)
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

func getMemInfo() *MemInfo {
	minfo, err := memutil.VirtualMemory()
	if err != nil {
		l.Error("fail to get memory toal, %s", err)
		return nil
	}

	vinfo, err := memutil.SwapMemory()
	if err != nil {
		l.Error("fail to get swap memory toal, %s", err)
	}

	return &MemInfo{
		MemoryTotal: minfo.Total,
		SwapTotal:   vinfo.Total,
		usedPercent: minfo.UsedPercent,
	}
}

func getNetInfo(enableVIfaces bool) []*NetInfo {
	ifs, err := netutil.Interfaces()
	if err != nil {
		l.Errorf("fail to get interfaces, %s", err)
		return nil
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
			} else if ip.To16() != nil {
				i.IP6 = ad.Addr
			}
		}
		infos = append(infos, i)
	}
	return infos
}

func getDiskInfo(ignoreFs []string) []*DiskInfo {

	ps, err := diskutil.Partitions(true)
	if err != nil {
		l.Errorf("fail to get disk info, %s", err)
		return nil
	}
	var infos []*DiskInfo

	fstypeExcludeSet, _ := DiskIgnoreFs(ignoreFs)

	for _, p := range ps {

		if _, ok := fstypeExcludeSet[p.Fstype]; ok {
			continue
		}

		info := &DiskInfo{
			Device:     p.Device,
			Mountpoint: p.Mountpoint,
			Fstype:     p.Fstype,
		}

		usage, err := diskutil.Usage(p.Mountpoint)
		if err == nil {
			info.Total = usage.Total
		}

		infos = append(infos, info)
	}

	return infos
}

func (c *Input) getEnabledInputs() (res []*CollectorStatus) {

	inputsStats, err := io.GetStats(c.IOTimeout.Duration) // get all inputs stats
	if err != nil {
		l.Warnf("fail to get inputs stats, %s", err)
		return
	}

	now := time.Now()
	for name, _ := range inputs.InputsInfo {
		if s, ok := inputsStats[name]; ok {

			ts := s.LastErrTS.Unix()
			if ts < 0 {
				ts = 0
			}

			lastErr := s.LastErr
			if ts > 0 && now.Sub(s.LastErrTS) > c.IgnoreInputsErrorsBefore.Duration { // ignore errors 30min ago
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
			})
		} else {
			res = append(res, &CollectorStatus{Count: 0, Name: name})
		}
	}

	return
}

func (c *Input) getHostObjectMessage() (*HostObjectMessage, error) {
	var msg HostObjectMessage

	stat := c.getEnabledInputs()

	// NOTE: 由于获取采集器的运行情况信息时，io 模块可能较忙，导致获取不到
	// 故此处缓存一下历史，以免在 message 字段中采集器信息字段(collectors)
	// 为空
	if len(stat) != 0 {
		collectorStatHist = stat
	}

	msg.Collectors = collectorStatHist
	if len(msg.Collectors) == 0 {
		// 此处也是为了避免采集器信息字段为空: 宁可丢弃当前这次对象采集，也不能导致采集器信息为空
		// 采集器信息为空（或缺失）的两种可能：
		//
		// 1: io 忙：不便于接收查询请求
		// 2: 具体的某个采集器，可能因为尚未来得及启动，就被要求查询运行信息，此时 io 模块肯定没有登记
		//
		// 故一般只有启动后第一次采集时会获取不到统计信息，后续基本都能获取到，即使拿不到，就用旧的统计信息替代
		return nil, fmt.Errorf("collector stats missing")
	}

	msg.Host = &HostInfo{
		HostMeta:   getHostMeta(),
		CPU:        getCPUInfo(),
		cpuPercent: getCPUPercent(),
		load5:      getLoad5(),
		Mem:        getMemInfo(),
		Net:        getNetInfo(c.EnableNetVirtualInterfaces),
		Disk:       getDiskInfo(c.IgnoreFS),
	}

	// sync cloud extra fields
	if v, ok := c.Tags["cloud_provider"]; ok {
		info, err := c.SyncCloudInfo(v)
		if err != nil {
			l.Warnf("sync cloud info failed: %v, ingored", err)
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

	return &msg, nil
}
