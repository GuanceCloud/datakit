// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

//go:build windows
// +build windows

package wmi

import (
	"fmt"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs/windowsremote/wmi/class"
	"time"
)

func QueryProcessors(service *SWbemServices) (cpus []class.Win32_Processor, err error) {
	q := CreateQuery(&cpus, "")
	err = service.Query(q, &cpus)
	if len(cpus) > 0 {
		return
	}
	return cpus, fmt.Errorf("cpus is empty")
}

func QueryOperatingSystem(service *SWbemServices) (systems []class.Win32_OperatingSystem, err error) {
	q := CreateQuery(&systems, "")
	err = service.Query(q, &systems)
	if len(systems) > 0 {
		return
	}
	return systems, fmt.Errorf("systems is empty")
}

func QueryNetworkAdapterConfiguration(service *SWbemServices) (networks []class.Win32_NetworkAdapterConfiguration, err error) {
	q := CreateQuery(&networks, "WHERE IPEnabled = TRUE")
	err = service.Query(q, &networks)
	if len(networks) > 0 {
		return
	}
	return networks, fmt.Errorf("networks is empty")
}

func QueryNetworkInterface(
	service *SWbemServices, interfaceName string) (net class.Win32_PerfFormattedData_Tcpip_NetworkInterface, err error) {
	var dest []class.Win32_PerfFormattedData_Tcpip_NetworkInterface
	q := CreateQuery(&dest, fmt.Sprintf(" where Name = '%s'", interfaceName))
	err = service.Query(q, &dest)
	if len(dest) > 0 {
		return dest[0], nil
	}
	return net, err
}

func QueryDisk(service *SWbemServices) (disks []class.Win32_LogicalDisk, err error) {
	q := CreateQuery(&disks, "where DriveType=3") //3 磁盘
	err = service.Query(q, &disks)
	return
}

func QueryPerOSDisk(service *SWbemServices) (disks []class.Win32_PerfFormattedData_PerfDisk_PhysicalDisk, err error) {
	q := CreateQuery(&disks, "WHERE Name = '_Total'")
	err = service.Query(q, &disks)
	if len(disks) > 0 {
		return
	}
	return disks, fmt.Errorf("diskio _Total is empty")
}

func QueryPerfOSProcessor(service *SWbemServices) (process []class.Win32_PerfFormattedData_PerfOS_Processor, err error) {
	q := CreateQuery(&process, "WHERE Name = '_Total'")
	err = service.Query(q, &process)
	if len(process) > 0 {
		return
	}
	return process, fmt.Errorf("processor is empty")
}

func QueryProcess(service *SWbemServices) (ps []class.Process, err error) {
	var pros []class.Win32_Process
	q := CreateQuery(&pros, "")
	err = service.Query(q, &pros)

	var perProcess []class.Win32_PerfFormattedData_PerfProc_Process
	q2 := CreateQuery(&perProcess, "")
	err = service.Query(q2, &perProcess)

	if len(perProcess) == 0 {
		return ps, fmt.Errorf("PerfProc_Process is empty")
	}
	if len(pros) == 0 {
		return ps, fmt.Errorf("process is empty")
	}

	for i := 0; i < len(pros); i++ {
		for j := 0; j < len(perProcess); j++ {
			if perProcess[j].IDProcess == pros[i].ProcessId {
				ps = append(ps, class.Process{Win32_Process: pros[i],
					Win32_PerfFormattedData_PerfProc_Process: perProcess[j]})
			}
		}
	}
	return
}

func QueryLogEvent(service *SWbemServices, t time.Time) (ntLogs []class.Win32_NTLogEvent, err error) {
	var where = " WHERE "
	where += fmt.Sprintf(
		"TimeGenerated >= '%s'",
		t.UTC().Format("20060102150405.000000-070"),
	)
	q := CreateQuery(&ntLogs, where)
	err = service.Query(q, &ntLogs)
	if err != nil {
		return nil, err
	}

	return ntLogs, nil
}
