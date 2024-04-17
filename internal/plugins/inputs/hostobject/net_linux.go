// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

//go:build linux
// +build linux

package hostobject

import (
	"net"
	"syscall"

	"github.com/vishvananda/netlink"
)

const (
	FlagUp           net.Flags = 1 << iota // interface is up
	FlagBroadcast                          // interface supports broadcast access capability
	FlagLoopback                           // interface is a loopback interface
	FlagPointToPoint                       // interface belongs to a point-to-point link
	FlagMulticast                          // interface supports multicast access capability
	FlagMaster                             // interface is a master interface
	FlagSlave                              // interface is a slave interface
)

type InterfaceStat struct {
	Index        int             `json:"index"`
	MTU          int             `json:"mtu"`          // maximum transmission unit
	Name         string          `json:"name"`         // e.g., "en0", "lo0", "eth0.100"
	HardwareAddr string          `json:"hardwareaddr"` // IEEE MAC-48, EUI-48 and EUI-64 form
	Flags        []string        `json:"flags"`        // e.g., FlagUp, FlagLoopback, FlagMulticast
	Addrs        []InterfaceAddr `json:"addrs"`
}

// InterfaceAddr is designed for represent interface addresses.
type InterfaceAddr struct {
	Addr string `json:"addr"`
}

func interfaces() ([]InterfaceStat, error) {
	is, err := netlink.LinkList()
	if err != nil {
		return nil, err
	}
	ret := make([]InterfaceStat, 0, len(is))
	for _, ifi := range is {
		rawFlags := ifi.Attrs().RawFlags
		flag := linkFlags(rawFlags)

		var flags []string
		if flag&FlagUp != 0 {
			flags = append(flags, "up")
		}
		if flag&FlagBroadcast != 0 {
			flags = append(flags, "broadcast")
		}
		if flag&FlagLoopback != 0 {
			flags = append(flags, "loopback")
		}
		if flag&FlagPointToPoint != 0 {
			flags = append(flags, "pointtopoint")
		}
		if flag&FlagMaster != 0 {
			flags = append(flags, "master")
		}
		if flag&FlagSlave != 0 {
			flags = append(flags, "slave")
		}
		if flag&FlagMulticast != 0 {
			flags = append(flags, "multicast")
		}

		r := InterfaceStat{
			Index:        ifi.Attrs().Index,
			Name:         ifi.Attrs().Name,
			MTU:          ifi.Attrs().MTU,
			HardwareAddr: ifi.Attrs().HardwareAddr.String(),
			Flags:        flags,
		}
		addrs, err := netlink.AddrList(ifi, netlink.FAMILY_ALL)
		if err == nil {
			r.Addrs = make([]InterfaceAddr, 0, len(addrs))
			for _, addr := range addrs {
				r.Addrs = append(r.Addrs, InterfaceAddr{
					Addr: addr.IPNet.String(),
				})
			}
		}
		ret = append(ret, r)
	}
	return ret, nil
}

func linkFlags(rawFlags uint32) net.Flags {
	var f net.Flags
	if rawFlags&syscall.IFF_UP != 0 {
		f |= FlagUp
	}
	if rawFlags&syscall.IFF_BROADCAST != 0 {
		f |= FlagBroadcast
	}
	if rawFlags&syscall.IFF_LOOPBACK != 0 {
		f |= FlagLoopback
	}
	if rawFlags&syscall.IFF_POINTOPOINT != 0 {
		f |= FlagPointToPoint
	}
	if rawFlags&syscall.IFF_MASTER != 0 {
		f |= FlagMaster
	}
	if rawFlags&syscall.IFF_SLAVE != 0 {
		f |= FlagSlave
	}
	if rawFlags&syscall.IFF_MULTICAST != 0 {
		f |= FlagMulticast
	}
	return f
}
