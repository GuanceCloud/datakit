package hostobject

import (
	"net"
	"strings"

	cpuutil "github.com/shirou/gopsutil/v3/cpu"
	diskutil "github.com/shirou/gopsutil/v3/disk"
	hostutil "github.com/shirou/gopsutil/v3/host"
	memutil "github.com/shirou/gopsutil/v3/mem"
	netutil "github.com/shirou/gopsutil/v3/net"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"
	tgi "gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs/telegraf_inputs"
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
	}

	NetInfo struct {
		Index        int      `json:"-"`
		MTU          int      `json:"mtu"`   // maximum transmission unit
		Name         string   `json:"name"`  // e.g., "en0", "lo0", "eth0.100"
		HardwareAddr string   `json:"mac"`   // IEEE MAC-48, EUI-48 and EUI-64 form
		Flags        []string `json:"flags"` // e.g., FlagUp, FlagLoopback, FlagMulticast
		IP4          string   `json:"ip4"`
		IP6          string   `json:"ip6"`
		Addrs        []string `json:"addrs"`
	}

	DiskInfo struct {
		Device     string `json:"device"`
		Total      uint64 `json:"total"`
		Mountpoint string `json:"mountpoint"`
		Fstype     string `json:"fstype"`
		Opts       string `json:"-"`
	}

	HostInfo struct {
		HostMeta *HostMetaInfo `json:"meta"`
		CPU      []*CPUInfo    `json:"cpu"`
		Mem      *MemInfo      `json:"mem"`
		Net      []*NetInfo    `json:"net"`
		Disk     []*DiskInfo   `json:"disk"`
	}

	HostObjectMessage struct {
		Host       *HostInfo `json:"host"`
		Collectors []string  `json:"collectors"`
	}
)

func getHostMeta() *HostMetaInfo {
	info, err := hostutil.Info()
	if err != nil {
		moduleLogger.Errorf("fail to get host info, %s", err)
		return nil
	}

	moduleLogger.Infof("raw host meta info: %s", info)

	return &HostMetaInfo{
		HostName:        info.Hostname,
		OS:              info.OS,
		BootTime:        info.BootTime,
		Platform:        info.Platform,
		PlatformFamily:  info.PlatformFamily,
		PlatformVersion: info.PlatformVersion,
		Kernel:          info.KernelVersion,
		Arch:            info.KernelArch,
	}
}

func getCPUInfo() []*CPUInfo {
	infos, err := cpuutil.Info()
	if err != nil {
		moduleLogger.Errorf("fail to get cpu info, %s", err)
		return nil
	}

	moduleLogger.Infof("raw cpu info: %s", infos)

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

func getMemInfo() *MemInfo {
	minfo, err := memutil.VirtualMemory()
	if err != nil {
		moduleLogger.Error("fail to get memory toal, %s", err)
		return nil
	}

	vinfo, err := memutil.SwapMemory()
	if err != nil {
		moduleLogger.Error("fail to get swap memory toal, %s", err)
	}

	return &MemInfo{
		MemoryTotal: minfo.Total,
		SwapTotal:   vinfo.Total,
	}
}

func getNetInfo() []*NetInfo {
	ifs, err := netutil.Interfaces()
	if err != nil {
		moduleLogger.Errorf("fail to get interfaces, %s", err)
		return nil
	}
	var infos []*NetInfo
	for _, it := range ifs {
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

func getDiskInfo() []*DiskInfo {

	ps, err := diskutil.Partitions(false)
	if err != nil {
		moduleLogger.Errorf("%s", err)
	}
	var infos []*DiskInfo
	for _, p := range ps {
		info := &DiskInfo{
			Device:     p.Device,
			Mountpoint: p.Mountpoint,
			Fstype:     p.Fstype,
			Opts:       strings.Join(p.Opts, ","),
		}

		usage, err := diskutil.Usage(p.Mountpoint)
		if err == nil {
			info.Total = usage.Total
		}

		infos = append(infos, info)
	}

	return infos
}

func getEnabledInputs() []string {

	var names []string

	for k := range inputs.Inputs {
		n, _ := inputs.InputEnabled(k)
		if n > 0 {
			names = append(names, k)
		}
	}

	for k := range tgi.TelegrafInputs {
		n, _ := inputs.InputEnabled(k)
		if n > 0 {
			names = append(names, k)
		}
	}

	return names
}

func getHostObjectMessage() *HostObjectMessage {
	var msg HostObjectMessage

	msg.Collectors = getEnabledInputs()
	msg.Host = &HostInfo{
		HostMeta: getHostMeta(),
		CPU:      getCPUInfo(),
		Mem:      getMemInfo(),
		Net:      getNetInfo(),
		Disk:     getDiskInfo(),
	}

	return &msg
}
