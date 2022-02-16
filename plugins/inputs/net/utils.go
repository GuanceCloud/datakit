package net

import (
	"fmt"
	"net"
	"regexp"

	psNet "github.com/shirou/gopsutil/net"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"
)

type NetProto func(protocols []string) ([]psNet.ProtoCountersStat, error)

type NetIO func() ([]psNet.IOCountersStat, error)

type NetVirtualIfaces func() (map[string]bool, error)

func NetIOCounters() ([]psNet.IOCountersStat, error) {
	return psNet.IOCounters(true)
}

func NetVirtualInterfaces() (map[string]bool, error) {
	return VirtualInterfaces()
}

type InterfaceFilter struct {
	filters []*regexp.Regexp
}

func (f *InterfaceFilter) Compile(exprs []string) error {
	f.filters = make([]*regexp.Regexp, 0) // clear
	for _, expr := range exprs {
		if filter, err := regexp.Compile(expr); err == nil {
			f.filters = append(f.filters, filter)
		} else {
			return err
		}
	}
	return nil
}

func (f *InterfaceFilter) Match(s string) bool {
	for _, filter := range f.filters {
		if filter.MatchString(s) {
			return true
		}
	}
	return false
}

// FilterInterface 通过正则过滤网络接口.
func FilterInterface(netio []psNet.IOCountersStat, interfaces []net.Interface,
	exprs []string, enableVirtual bool, netVIfaces NetVirtualIfaces,
) (map[string]psNet.IOCountersStat, error) {
	var err error
	interfacesByName := map[string]net.Interface{}
	for _, iface := range interfaces {
		interfacesByName[iface.Name] = iface
	}

	filteredInterfaces := make(map[string]psNet.IOCountersStat)

	netInterfacesArray := make([]string, 0)
	for _, itrfc := range netio {
		// 过滤非 up 状态的接口
		iface, ok := interfacesByName[itrfc.Name]
		if !(ok && (iface.Flags&net.FlagUp) == net.FlagUp) {
			continue
		}
		filteredInterfaces[itrfc.Name] = itrfc
		netInterfacesArray = append(netInterfacesArray, itrfc.Name)
	}
	// 正则表达式数组不为 nil 时过滤不匹配的 interface
	if len(exprs) > 0 {
		filter := &InterfaceFilter{}
		err = filter.Compile(exprs)
		if err != nil {
			return nil, fmt.Errorf("error compiling filter: %w", err)
		}
		for _, iName := range netInterfacesArray {
			if !filter.Match(iName) {
				delete(filteredInterfaces, iName)
			}
		}
	}
	//  移除 virtual interfaces
	if !enableVirtual {
		var virtualInterfaces map[string]bool
		if virtualInterfaces, err = netVIfaces(); err == nil {
			for iName := range virtualInterfaces {
				delete(filteredInterfaces, iName)
			}
		} else {
			return nil, err
		}
	}
	return filteredInterfaces, err
}

func NewFieldsInfoIByte(desc string) *inputs.FieldInfo {
	return &inputs.FieldInfo{
		Type:     inputs.Gauge,
		DataType: inputs.Int,
		Unit:     inputs.SizeByte,
		Desc:     desc,
	}
}

func NewFieldsInfoIBytePerSec(desc string) *inputs.FieldInfo {
	return &inputs.FieldInfo{
		Type:     inputs.Gauge,
		DataType: inputs.Int,
		Unit:     inputs.BytesPerSec,
		Desc:     desc,
	}
}

func NewFieldsInfoCount(desc string) *inputs.FieldInfo {
	return &inputs.FieldInfo{
		Type:     inputs.Gauge,
		DataType: inputs.Int,
		Unit:     inputs.NCount,
		Desc:     desc,
	}
}

func NewFieldsInfoCountPerSec(desc string) *inputs.FieldInfo {
	return &inputs.FieldInfo{
		Type:     inputs.Gauge,
		DataType: inputs.Int,
		Unit:     inputs.NCount,
		Desc:     desc,
	}
}

func NewFieldsInfoMS(desc string) *inputs.FieldInfo {
	return &inputs.FieldInfo{
		Type:     inputs.Gauge,
		DataType: inputs.Int,
		Unit:     inputs.DurationMS,
		Desc:     desc,
	}
}
