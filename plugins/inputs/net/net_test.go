package net

import (
	"net"
	"runtime"
	"testing"
	"time"

	psNet "github.com/shirou/gopsutil/net"
)

// TODO
func TestCollect(t *testing.T) {
	i := &Input{
		netIO:            NetIOCounters4Test,
		netProto:         NetProtoCounters4Test,
		netVirtualIfaces: NetVirtualInterfaces4Test}
	i.Interfaces = nil
	i.EnableVirtualInterfaces = true
	i.IgnoreProtocolStats = true

	for x := 0; x < 6; x++ {
		time.Sleep(time.Second)
		if err := i.Collect(); err != nil {
			t.Errorf("Error collecting network statistics: %s", err)
		}
	}
	if len(i.collectCache) != 4 {
		t.Errorf("Need to clear collectCache")
	}

	i.Interfaces = []string{"eth.*", "wlp3s0", "docke[a-z]+\\d+"}

	// linux only, collect protocol stats test
	if runtime.GOOS == "linux" {
		i.EnableVirtualInterfaces = false
		i.IgnoreProtocolStats = false
		i.Collect()
		time.Sleep(time.Second)
		i.Collect()
	}
}

func TestFilterInterface(t *testing.T) {
	netio, _ := NetIOCounters4Test()
	ifaces := []net.Interface{
		{Index: 1, MTU: 65536, Name: "lo", HardwareAddr: []byte("f1:de:38:f2:a0:2f"), Flags: 5},
		{Index: 2, MTU: 1500, Name: "enp2s0", HardwareAddr: []byte("1c:fe:ab:f1:d0:2e"), Flags: 19},
		{Index: 3, MTU: 1500, Name: "wlp3s0", HardwareAddr: []byte("cc:2f:75:a6:b3:c5"), Flags: 19},
		{Index: 4, MTU: 1500, Name: "docker0", HardwareAddr: []byte("c2:42:b7:2d:e8:a5"), Flags: 19},
	}

	// 包含虚拟接口、使用正则
	// "docker0", "enp2s0", "wlp3s0",
	exprs := []string{
		"lp\\d+",
		"docker.*",
	}
	enableVirtual := true
	filtered, _ := FilterInterface(netio, ifaces, exprs, enableVirtual, NetVirtualInterfaces4Test)
	for _, iName := range []string{"docker0", "wlp3s0"} {
		if _, ok := filtered[iName]; !ok {
			t.Error("match failed")
		}
	}
	// 包含虚拟接口、不使用正则
	// "docker0", "enp2s0", "wlp3s0",
	exprs = []string{}
	enableVirtual = true
	filtered, _ = FilterInterface(netio, ifaces, exprs, enableVirtual, NetVirtualInterfaces4Test)
	for _, iName := range []string{"docker0", "enp2s0", "wlp3s0"} {
		if _, ok := filtered[iName]; !ok {
			t.Error("match failed")
		}
	}

	if runtime.GOOS == "linux" {
		// 不含虚拟接口、使用正则
		// "docker0", "enp2s0", "wlp3s0",
		exprs = []string{
			"lp\\d+",
			"docker.*",
		}
		enableVirtual = false
		filtered, _ = FilterInterface(netio, ifaces, exprs, enableVirtual, NetVirtualInterfaces4Test)
		for _, iName := range []string{"wlp3s0"} {
			if _, ok := filtered[iName]; !ok {
				t.Error("match failed")
			}
		}

		// 不含虚拟接口、不使用正则
		// "docker0", "enp2s0", "wlp3s0",
		exprs = []string{}
		enableVirtual = false
		filtered, _ = FilterInterface(netio, ifaces, exprs, enableVirtual, NetVirtualInterfaces4Test)
		for _, iName := range []string{"enp2s0", "wlp3s0"} {
			if _, ok := filtered[iName]; !ok {
				t.Error("match failed")
			}
		}
	}
}

func TestVirtualInterfaces(t *testing.T) {
	if runtime.GOOS == "linux" {
		if _, err := VirtualInterfaces(); err != nil {
			t.Error(err)
		}
		vIface, _ := VirtualInterfaces("abc\ndef\n")
		for iName := range map[string]bool{"abc": true, "def": true} {
			if _, ok := vIface[iName]; !ok {
				t.Error("error: get virtual interface")
			}
		}
	}
}

func NetIOCounters4Test() ([]psNet.IOCountersStat, error) {
	r := []psNet.IOCountersStat{
		{
			Name:        "lo",
			BytesSent:   1715387281,
			BytesRecv:   1715387281,
			PacketsSent: 3279790,
			PacketsRecv: 3279790,
			Errin:       0,
			Errout:      0,
			Dropin:      0,
			Dropout:     0,
			Fifoin:      0,
			Fifoout:     0,
		},
		{
			Name:        "enp2s0",
			BytesSent:   0,
			BytesRecv:   0,
			PacketsSent: 0,
			PacketsRecv: 0,
			Errin:       0,
			Errout:      0,
			Dropin:      0,
			Dropout:     0,
			Fifoin:      0,
			Fifoout:     0,
		},
		{
			Name:        "wlp3s0",
			BytesSent:   176812478,
			BytesRecv:   1037443863,
			PacketsSent: 645856,
			PacketsRecv: 1303474,
			Errin:       0,
			Errout:      0,
			Dropin:      0,
			Dropout:     0,
			Fifoin:      0,
			Fifoout:     0,
		},
		{
			Name:        "docker0",
			BytesSent:   0,
			BytesRecv:   0,
			PacketsSent: 0,
			PacketsRecv: 0,
			Errin:       0,
			Errout:      0,
			Dropin:      0,
			Dropout:     0,
			Fifoin:      0,
			Fifoout:     0,
		},
	}
	return r, nil
}

func NetProtoCounters4Test(protocols []string) ([]psNet.ProtoCountersStat, error) {
	r := []psNet.ProtoCountersStat{
		{
			Protocol: "tcp",
			Stats: map[string]int64{
				"ActiveOpens":  10949,
				"AttemptFails": 152,
				"CurrEstab":    38,
				"EstabResets":  4190,
				"InCsumErrors": 0,
				"InErrs":       45,
				"InSegs":       2544481,
				"MaxConn":      -1,
				"OutRsts":      5212,
				"OutSegs":      2503933,
				"PassiveOpens": 6331,
				"RetransSegs":  3604,
				"RtoAlgorithm": 1,
				"RtoMax":       120000,
				"RtoMin":       200,
			},
		},
		{
			Protocol: "udp",
			Stats: map[string]int64{
				"IgnoredMulti": 0,
				"InCsumErrors": 0,
				"InDatagrams":  187538,
				"InErrors":     0,
				"NoPorts":      40,
				"OutDatagrams": 27812,
				"RcvbufErrors": 0,
				"SndbufErrors": 0,
			},
		},
		// psNet.ProtoCountersStat{
		// 	Protocol: "udplite",
		// 	Stats: map[string]int64{
		// 		"IgnoredMulti": 0,
		// 		"InCsumErrors": 0,
		// 		"InDatagrams":  0,
		// 		"InErrors":     0,
		// 		"NoPorts":      0,
		// 		"OutDatagrams": 0,
		// 		"RcvbufErrors": 0,
		// 		"SndbufErrors": 0,
		// 	},
		// },
	}
	return r, nil
}

func NetVirtualInterfaces4Test() (map[string]bool, error) {
	return map[string]bool{"lo": true, "docker0": true}, nil
}
