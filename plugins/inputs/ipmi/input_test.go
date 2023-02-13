// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2022-present Guance, Inc.

// Package ipmi collects host ipmi metrics.
package ipmi

import (
	"reflect"
	"testing"

	"github.com/GuanceCloud/cliutils"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"
)

func TestInput_getParameters(t *testing.T) {
	type fields struct {
		Interval         datakit.Duration
		Tags             map[string]string
		collectCache     []inputs.Measurement
		platform         string
		BinPath          string
		IpmiServers      []string
		IpmiInterfaces   []string
		IpmiUsers        []string
		IpmiPasswords    []string
		HexKeys          []string
		MetricVersions   []int
		RegexpCurrent    []string
		RegexpVoltage    []string
		RegexpPower      []string
		RegexpTemp       []string
		RegexpFanSpeed   []string
		RegexpUsage      []string
		RegexpCount      []string
		RegexpStatus     []string
		Timeout          datakit.Duration
		DropWarningDelay datakit.Duration
		servers          []ipmiServer
		semStop          *cliutils.Sem
		Election         bool
		pause            bool
		pauseCh          chan bool
	}
	type args struct {
		i int
	}
	tests := []struct {
		name              string
		fields            fields
		args              args
		wantOpts          []string
		wantMetricVersion int
		wantErr           bool
	}{
		{
			name: "ipmiserver_index=0",
			fields: fields{
				BinPath:        "/usr/bin/ipmitool",
				IpmiServers:    []string{"192.168.1.1"},
				IpmiInterfaces: []string{"lanplus"},
				IpmiUsers:      []string{"root"},
				IpmiPasswords:  []string{"calvin"},
				HexKeys:        []string{},
				MetricVersions: []int{2},
			},
			args: args{
				0,
			},
			wantOpts: []string{
				"-I", "lanplus",
				"-H", "192.168.1.1",
				"-U", "root",
				"-P", "calvin",
				"sdr", "elist",
			},
			wantMetricVersion: 2,
			wantErr:           false,
		},
		{
			name: "ipmiserver_index=1",
			fields: fields{
				BinPath:        "/usr/bin/ipmitool",
				IpmiServers:    []string{"192.168.1.1", "192.168.1.2"},
				IpmiInterfaces: []string{"lanplus"},
				IpmiUsers:      []string{"root"},
				IpmiPasswords:  []string{"calvin"},
				HexKeys:        []string{},
				MetricVersions: []int{2},
			},
			args: args{
				1,
			},
			wantOpts: []string{
				"-I", "lanplus",
				"-H", "192.168.1.2",
				"-U", "root",
				"-P", "calvin",
				"sdr", "elist",
			},
			wantMetricVersion: 2,
			wantErr:           false,
		},
		{
			name: "with hex_key",
			fields: fields{
				BinPath:        "/usr/bin/ipmitool",
				IpmiServers:    []string{"192.168.1.1"},
				IpmiInterfaces: []string{"lanplus"},
				IpmiUsers:      []string{"root"},
				IpmiPasswords:  []string{"calvin"},
				HexKeys:        []string{"50415353574F5244"},
				MetricVersions: []int{2},
			},
			args: args{
				0,
			},
			wantOpts: []string{
				"-I", "lanplus",
				"-H", "192.168.1.1",
				"-y", "50415353574F5244",
				"-U", "root",
				"-P", "calvin",
				"sdr", "elist",
			},
			wantMetricVersion: 2,
			wantErr:           false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ipt := &Input{
				Interval:         tt.fields.Interval,
				Tags:             tt.fields.Tags,
				collectCache:     tt.fields.collectCache,
				platform:         tt.fields.platform,
				BinPath:          tt.fields.BinPath,
				IpmiServers:      tt.fields.IpmiServers,
				IpmiInterfaces:   tt.fields.IpmiInterfaces,
				IpmiUsers:        tt.fields.IpmiUsers,
				IpmiPasswords:    tt.fields.IpmiPasswords,
				HexKeys:          tt.fields.HexKeys,
				MetricVersions:   tt.fields.MetricVersions,
				RegexpCurrent:    tt.fields.RegexpCurrent,
				RegexpVoltage:    tt.fields.RegexpVoltage,
				RegexpPower:      tt.fields.RegexpPower,
				RegexpTemp:       tt.fields.RegexpTemp,
				RegexpFanSpeed:   tt.fields.RegexpFanSpeed,
				RegexpUsage:      tt.fields.RegexpUsage,
				RegexpCount:      tt.fields.RegexpCount,
				RegexpStatus:     tt.fields.RegexpStatus,
				Timeout:          tt.fields.Timeout,
				DropWarningDelay: tt.fields.DropWarningDelay,
				servers:          tt.fields.servers,
				semStop:          tt.fields.semStop,
				Election:         tt.fields.Election,
				pause:            tt.fields.pause,
				pauseCh:          tt.fields.pauseCh,
			}
			gotOpts, gotMetricVersion, err := ipt.getParameters(tt.args.i)
			if (err != nil) != tt.wantErr {
				t.Errorf("Input.getParameters() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(gotOpts, tt.wantOpts) {
				t.Errorf("Input.getParameters() gotOpts = %v, want %v", gotOpts, tt.wantOpts)
			}
			if gotMetricVersion != tt.wantMetricVersion {
				t.Errorf("Input.getParameters() gotMetricVersion = %v, want %v", gotMetricVersion, tt.wantMetricVersion)
			}
		})
	}
}

func TestInput_convert(t *testing.T) {
	type fields struct {
		Interval         datakit.Duration
		Tags             map[string]string
		collectCache     []inputs.Measurement
		platform         string
		BinPath          string
		IpmiServers      []string
		IpmiInterfaces   []string
		IpmiUsers        []string
		IpmiPasswords    []string
		HexKeys          []string
		MetricVersions   []int
		RegexpCurrent    []string
		RegexpVoltage    []string
		RegexpPower      []string
		RegexpTemp       []string
		RegexpFanSpeed   []string
		RegexpUsage      []string
		RegexpCount      []string
		RegexpStatus     []string
		Timeout          datakit.Duration
		DropWarningDelay datakit.Duration
		servers          []ipmiServer
		semStop          *cliutils.Sem
		Election         bool
		pause            bool
		pauseCh          chan bool
	}
	type args struct {
		data          []byte
		metricVersion int
		server        string
	}
	tests := []struct {
		name             string
		fields           fields
		args             args
		wantCollectCache []inputs.Measurement
	}{
		{
			name: "R740V2",
			fields: fields{
				RegexpCurrent: []string{"current"},
				// RegexpVoltage:  []string{"voltage"},
				// RegexpPower:    []string{"pwr"},
				// RegexpTemp:     []string{"temp"},
				// RegexpFanSpeed: []string{"fan"},
				// RegexpUsage:    []string{"usage"},
				// RegexpCount:    []string{},
				// RegexpStatus:   []string{"fan", "slot", "drive"},
			},
			args: args{
				data:          []byte(dataR740V2),
				metricVersion: 2,
				server:        "192.168.1.2",
			},
			wantCollectCache: []inputs.Measurement{
				&ipmiMeasurement{
					name: inputName,
					tags: map[string]string{
						"host": "192.168.1.2",
						"unit": "current_1",
					},
					fields: map[string]interface{}{
						"current": float64(1),
					},
				},
				&ipmiMeasurement{
					name: inputName,
					tags: map[string]string{
						"host": "192.168.1.2",
						"unit": "current_2",
					},
					fields: map[string]interface{}{
						"current": float64(0.2),
					},
				},
			},
		},
		{
			name: "R740V2",
			fields: fields{
				// RegexpCurrent: []string{"current"},
				RegexpVoltage: []string{"voltage"},
				// RegexpPower:    []string{"pwr"},
				// RegexpTemp:     []string{"temp"},
				// RegexpFanSpeed: []string{"fan"},
				// RegexpUsage:    []string{"usage"},
				// RegexpCount:    []string{},
				// RegexpStatus:   []string{"fan", "slot", "drive"},
			},
			args: args{
				data:          []byte(dataR620V1),
				metricVersion: 1,
				server:        "192.168.1.1",
			},
			wantCollectCache: []inputs.Measurement{
				&ipmiMeasurement{
					name: inputName,
					tags: map[string]string{
						"host": "192.168.1.1",
						"unit": "voltage_1",
					},
					fields: map[string]interface{}{
						"voltage": float64(220),
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ipt := &Input{
				Interval:         tt.fields.Interval,
				Tags:             tt.fields.Tags,
				collectCache:     tt.fields.collectCache,
				platform:         tt.fields.platform,
				BinPath:          tt.fields.BinPath,
				IpmiServers:      tt.fields.IpmiServers,
				IpmiInterfaces:   tt.fields.IpmiInterfaces,
				IpmiUsers:        tt.fields.IpmiUsers,
				IpmiPasswords:    tt.fields.IpmiPasswords,
				HexKeys:          tt.fields.HexKeys,
				MetricVersions:   tt.fields.MetricVersions,
				RegexpCurrent:    tt.fields.RegexpCurrent,
				RegexpVoltage:    tt.fields.RegexpVoltage,
				RegexpPower:      tt.fields.RegexpPower,
				RegexpTemp:       tt.fields.RegexpTemp,
				RegexpFanSpeed:   tt.fields.RegexpFanSpeed,
				RegexpUsage:      tt.fields.RegexpUsage,
				RegexpCount:      tt.fields.RegexpCount,
				RegexpStatus:     tt.fields.RegexpStatus,
				Timeout:          tt.fields.Timeout,
				DropWarningDelay: tt.fields.DropWarningDelay,
				servers:          tt.fields.servers,
				semStop:          tt.fields.semStop,
				Election:         tt.fields.Election,
				pause:            tt.fields.pause,
				pauseCh:          tt.fields.pauseCh,
			}
			ipt.convert(tt.args.data, tt.args.metricVersion, tt.args.server)
			if !reflect.DeepEqual(ipt.collectCache, tt.wantCollectCache) {
				t.Errorf("Input.getParameters() wantCollectCache = %+v,  want %+v", ipt.collectCache, tt.wantCollectCache)
			}
		})
	}
}

func TestInput_Collect(t *testing.T) {
	type fields struct {
		Interval         datakit.Duration
		Tags             map[string]string
		collectCache     []inputs.Measurement
		platform         string
		BinPath          string
		IpmiServers      []string
		IpmiInterfaces   []string
		IpmiUsers        []string
		IpmiPasswords    []string
		HexKeys          []string
		MetricVersions   []int
		RegexpCurrent    []string
		RegexpVoltage    []string
		RegexpPower      []string
		RegexpTemp       []string
		RegexpFanSpeed   []string
		RegexpUsage      []string
		RegexpCount      []string
		RegexpStatus     []string
		Timeout          datakit.Duration
		DropWarningDelay datakit.Duration
		servers          []ipmiServer
		semStop          *cliutils.Sem
		Election         bool
		pause            bool
		pauseCh          chan bool
	}
	tests := []struct {
		name    string
		fields  fields
		wantErr bool
	}{
		{
			name: "error bin_path && serverip",
			fields: fields{
				BinPath:        "aaaa",
				IpmiServers:    []string{"192.168.1.1", "192.168.1.2"},
				IpmiUsers:      []string{"userdemo"},
				IpmiPasswords:  []string{"passworddemo"},
				IpmiInterfaces: []string{"lanplus"},
				MetricVersions: []int{2, 1},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ipt := &Input{
				Interval:         tt.fields.Interval,
				Tags:             tt.fields.Tags,
				collectCache:     tt.fields.collectCache,
				platform:         tt.fields.platform,
				BinPath:          tt.fields.BinPath,
				IpmiServers:      tt.fields.IpmiServers,
				IpmiInterfaces:   tt.fields.IpmiInterfaces,
				IpmiUsers:        tt.fields.IpmiUsers,
				IpmiPasswords:    tt.fields.IpmiPasswords,
				HexKeys:          tt.fields.HexKeys,
				MetricVersions:   tt.fields.MetricVersions,
				RegexpCurrent:    tt.fields.RegexpCurrent,
				RegexpVoltage:    tt.fields.RegexpVoltage,
				RegexpPower:      tt.fields.RegexpPower,
				RegexpTemp:       tt.fields.RegexpTemp,
				RegexpFanSpeed:   tt.fields.RegexpFanSpeed,
				RegexpUsage:      tt.fields.RegexpUsage,
				RegexpCount:      tt.fields.RegexpCount,
				RegexpStatus:     tt.fields.RegexpStatus,
				Timeout:          tt.fields.Timeout,
				DropWarningDelay: tt.fields.DropWarningDelay,
				servers:          tt.fields.servers,
				semStop:          tt.fields.semStop,
				Election:         tt.fields.Election,
				pause:            tt.fields.pause,
				pauseCh:          tt.fields.pauseCh,
			}
			if err := ipt.Collect(); (err != nil) != tt.wantErr {
				t.Errorf("Input.Collect() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

const (
	dataR620V1 = `SEL              | Not Readable      | ns
Intrusion        | 0x00              | ok
Fan1A RPM        | 2160 RPM          | ok
Fan2A RPM        | 2280 RPM          | ok
Fan3A RPM        | 2280 RPM          | ok
Fan4A RPM        | 2400 RPM          | ok
Fan5A RPM        | 2280 RPM          | ok
Fan6A RPM        | 2160 RPM          | ok
Inlet Temp       | 23 degrees C      | ok
Exhaust Temp     | 37 degrees C      | ok
Temp             | 47 degrees C      | ok
Temp             | 44 degrees C      | ok
OS Watchdog      | 0x00              | ok
VCORE PG         | 0x00              | ok
VCORE PG         | 0x00              | ok
3.3V PG          | 0x00              | ok
5V PG            | 0x00              | ok
USB Cable Pres   | 0x00              | ok
VGA Cable Pres   | 0x00              | ok
Dedicated NIC    | 0x00              | ok
Presence         | 0x00              | ok
Presence         | 0x00              | ok
Presence         | 0x00              | ok
PLL PG           | 0x00              | ok
PLL PG           | 0x00              | ok
1.1V PG          | 0x00              | ok
M23 VDDQ PG      | 0x00              | ok
M23 VTT PG       | 0x00              | ok
FETDRV PG        | 0x00              | ok
Presence         | 0x00              | ok
VSA PG           | 0x00              | ok
VSA PG           | 0x00              | ok
M01 VDDQ PG      | 0x00              | ok
M01 VDDQ PG      | 0x00              | ok
M23 VTT PG       | 0x00              | ok
M01 VTT PG       | 0x00              | ok
NDC PG           | 0x00              | ok
LCD Cable Pres   | 0x00              | ok
VTT PG           | 0x00              | ok
VTT PG           | 0x00              | ok
M23 VDDQ PG      | 0x00              | ok
Presence         | 0x00              | ok
Presence         | 0x00              | ok
Status           | 0x00              | ok
Status           | 0x00              | ok
Fan Redundancy   | 0x00              | ok
Riser Config Err | 0x00              | ok
Riser 3 Presence | 0x00              | ok
1.5V PG          | 0x00              | ok
PS2 PG Fail      | Not Readable      | ns
PS1 PG Fail      | Not Readable      | ns
BP1 5V PG        | 0x00              | ok
BP2 5V PG        | Not Readable      | ns
M01 VTT PG       | 0x00              | ok
Presence         | 0x00              | ok
PCIe Slot1       | Not Readable      | ns
PCIe Slot2       | Not Readable      | ns
PCIe Slot3       | Not Readable      | ns
PCIe Slot4       | Not Readable      | ns
PCIe Slot5       | Not Readable      | ns
PCIe Slot6       | Not Readable      | ns
PCIe Slot7       | Not Readable      | ns
vFlash           | 0x00              | ok
CMOS Battery     | 0x00              | ok
ROMB Battery     | 0x00              | ok
ROMB Battery     | Not Readable      | ns
Presence         | 0x00              | ok
Presence         | 0x00              | ok
Current 1        | 0.40 Amps         | ok
Current 2        | no reading        | ns
Voltage 1        | 220 Volts         | ok
Voltage 2        | no reading        | ns
PS Redundancy    | Not Readable      | ns
Status           | 0x00              | ok
Status           | Not Readable      | ns
Pwr Consumption  | 98 Watts          | ok
Power Optimized  | 0x00              | ok
SD1              | Not Readable      | ns
SD2              | Not Readable      | ns
Redundancy       | Not Readable      | ns
ECC Corr Err     | Not Readable      | ns
ECC Uncorr Err   | Not Readable      | ns
I/O Channel Chk  | Not Readable      | ns
PCI Parity Err   | Not Readable      | ns
PCI System Err   | Not Readable      | ns
SBE Log Disabled | Not Readable      | ns
Logging Disabled | Not Readable      | ns
Unknown          | Not Readable      | ns
CPU Protocol Err | Not Readable      | ns
CPU Bus PERR     | Not Readable      | ns
CPU Init Err     | Not Readable      | ns
CPU Machine Chk  | Not Readable      | ns
Memory Spared    | Not Readable      | ns
Memory Mirrored  | Not Readable      | ns
Memory RAID      | Not Readable      | ns
Memory Added     | Not Readable      | ns
Memory Removed   | Not Readable      | ns
Memory Cfg Err   | Not Readable      | ns
Mem Redun Gain   | Not Readable      | ns
PCIE Fatal Err   | Not Readable      | ns
Chipset Err      | Not Readable      | ns
Err Reg Pointer  | Not Readable      | ns
Mem ECC Warning  | Not Readable      | ns
Mem CRC Err      | Not Readable      | ns
USB Over-current | Not Readable      | ns
POST Err         | Not Readable      | ns
Hdwr version err | Not Readable      | ns
Mem Overtemp     | Not Readable      | ns
Mem Fatal SB CRC | Not Readable      | ns
Mem Fatal NB CRC | Not Readable      | ns
OS Watchdog Time | Not Readable      | ns
Non Fatal PCI Er | Not Readable      | ns
Fatal IO Error   | Not Readable      | ns
MSR Info Log     | Not Readable      | ns
TXT Status       | Not Readable      | ns
Drive 0          | 0x00              | ok
Cable SAS A      | 0x00              | ok
Cable SAS B      | 0x00              | ok
Cable SAS C      | Not Readable      | ns
Cable SAS D      | Not Readable      | ns
Cable SAS A      | Not Readable      | ns
Cable SAS B      | Not Readable      | ns
Cable SAS C      | Not Readable      | ns
Cable SAS D      | Not Readable      | ns
Power Cable      | 0x00              | ok
Signal Cable     | 0x00              | ok
Power Cable      | Not Readable      | ns
Signal Cable     | Not Readable      | ns
Fan7A RPM        | 2400 RPM          | ok
Fan1B RPM        | 2280 RPM          | ok
Fan2B RPM        | 2280 RPM          | ok
Fan3B RPM        | 2520 RPM          | ok
Fan4B RPM        | 2040 RPM          | ok
Fan5B RPM        | 2520 RPM          | ok
Fan6B RPM        | 2400 RPM          | ok
Fan7B RPM        | 2640 RPM          | ok
PFault Fail Safe | Not Readable      | ns
A                | 0x00              | ok
B                | 0x00              | ok
`

	dataR740V2 = `Temp             | 01h | ok  |  3.1 | 34 degrees C
Temp             | 02h | ok  |  3.2 | 45 degrees C
Inlet Temp       | 05h | ok  |  7.1 | 23 degrees C
DIMM PG          | 07h | ok  |  7.1 | State Deasserted
NDC PG           | 08h | ok  |  7.1 | State Deasserted
PS1 PG FAIL      | 09h | ok  |  7.1 | State Deasserted
PS2 PG FAIL      | 0Ah | ok  |  7.1 | State Deasserted
BP0 PG           | 0Bh | ok  |  7.1 | State Deasserted
BP1 PG           | 0Ch | ok  |  7.1 | State Deasserted
BP2 PG           | 0Dh | ok  |  7.1 | State Deasserted
1.8V SW PG       | 0Eh | ok  |  7.1 | State Deasserted
2.5V SW PG       | 0Fh | ok  |  7.1 | State Deasserted
5V SW PG         | 10h | ok  |  7.1 | State Deasserted
PVNN SW PG       | 11h | ok  |  7.1 | State Deasserted
VSB11 SW PG      | 12h | ok  |  7.1 | State Deasserted
VSBM SW PG       | 13h | ok  |  7.1 | State Deasserted
3.3V B PG        | 15h | ok  |  7.1 | State Deasserted
MEM012 VDDQ PG   | 24h | ok  |  3.1 | State Deasserted
MEM012 VPP PG    | 25h | ok  |  3.1 | State Deasserted
MEM012 VTT PG    | 26h | ok  |  3.1 | State Deasserted
MEM345 VDDQ PG   | 27h | ok  |  3.1 | State Deasserted
MEM345 VPP PG    | 28h | ok  |  3.1 | State Deasserted
MEM345 VTT PG    | 29h | ok  |  3.1 | State Deasserted
VCCIO PG         | 2Ah | ok  |  3.1 | State Deasserted
VCORE PG         | 2Bh | ok  |  3.1 | State Deasserted
FIVR PG          | 2Ch | ok  |  3.1 | State Deasserted
MEM012 VDDQ PG   | 2Eh | ok  |  3.2 | State Deasserted
MEM012 VPP PG    | 2Fh | ok  |  3.2 | State Deasserted
MEM012 VTT PG    | 30h | ok  |  3.2 | State Deasserted
MEM345 VDDQ PG   | 31h | ok  |  3.2 | State Deasserted
MEM345 VPP PG    | 32h | ok  |  3.2 | State Deasserted
MEM345 VTT PG    | 33h | ok  |  3.2 | State Deasserted
VCCIO PG         | 34h | ok  |  3.2 | State Deasserted
VCORE PG         | 35h | ok  |  3.2 | State Deasserted
FIVR PG          | 36h | ok  |  3.2 | State Deasserted
Fan1             | 38h | ok  |  7.1 | 3480 RPM
Fan2             | 39h | ok  |  7.1 | 3360 RPM
Fan3             | 3Ah | ok  |  7.1 | 3480 RPM
Fan4             | 3Bh | ok  |  7.1 | 3240 RPM
Fan5             | 3Ch | ok  |  7.1 | 3840 RPM
Fan6             | 3Dh | ok  |  7.1 | 3360 RPM
Presence         | 48h | ok  | 10.1 | Present
Presence         | 49h | ok  | 10.2 | Present
Presence         | 4Ah | ok  | 11.1 | Present
Presence         | 4Bh | ok  | 11.3 | Absent
Intrusion Cable  | 51h | ok  |  7.1 | Connected
VGA Cable Pres   | 52h | ok  |  7.1 | Connected
Presence         | 58h | ok  | 11.2 | Absent
BP0 Presence     | 59h | ok  | 26.1 | Absent
BP1 Presence     | 5Ah | ok  | 26.2 | Present
BP2 Presence     | 5Bh | ok  | 26.3 | Absent
Power Cable      | 5Ch | ns  | 26.1 | Disabled
Signal Cable     | 5Dh | ns  | 26.1 | Disabled
Power Cable      | 5Eh | ok  | 26.2 | Connected
Signal Cable     | 5Fh | ok  | 26.2 | Connected
Power Cable      | 60h | ns  | 26.3 | Disabled
Signal Cable     | 61h | ns  | 26.3 | Disabled
Presence         | 66h | ok  |  3.1 | Present
Presence         | 67h | ok  |  3.2 | Present
Current 1        | 6Bh | ok  | 10.1 | 1 Amps
Current 2        | 6Ch | ok  | 10.2 | 0.20 Amps
Voltage 1        | 6Dh | ok  | 10.1 | 220 Volts
Voltage 2        | 6Eh | ok  | 10.2 | 222 Volts
Riser Config Err | 6Fh | ok  |  7.1 | Connected
OS Watchdog      | 71h | ok  |  7.1 |
SEL              | 72h | ns  |  0.1 | No Reading
Intrusion        | 73h | ok  |  7.1 |
Power Optimized  | 75h | ok  |  7.1 |
Pwr Consumption  | 76h | ok  |  7.1 | 208 Watts
PS Redundancy    | 77h | ok  |  7.1 | Fully Redundant
Fan Redundancy   | 78h | ok  |  7.1 | Fully Redundant
Redundancy       | 79h | ns  | 11.3 | Disabled
SD1              | 7Ah | ns  | 11.3 | Disabled
SD2              | 7Bh | ns  | 11.3 | Disabled
SD               | 7Ch | ns  | 11.4 | Disabled
IO Usage         | 7Dh | ok  |  7.1 | 0 percent
MEM Usage        | 7Eh | ok  |  7.1 | 0 percent
SYS Usage        | 7Fh | ok  |  7.1 | 2 percent
CPU Usage        | 80h | ok  |  7.1 | 1 percent
Status           | 81h | ok  |  3.1 | Presence detected
Status           | 82h | ok  |  3.2 | Presence detected
Status           | 85h | ok  | 10.1 | Presence detected
Status           | 86h | ok  | 10.2 | Presence detected
CMOS Battery     | 87h | ok  |  7.1 | Presence Detected
ROMB Battery     | 88h | ns  | 11.2 | Disabled
PCIe Slot1       | 8Bh | ns  |  7.1 | Disabled
PCIe Slot2       | 8Ch | ns  |  7.1 | Disabled
PCIe Slot3       | 8Dh | ns  |  7.1 | Disabled
PCIe Slot4       | 8Eh | ok  |  7.1 |
PCIe Slot5       | 8Fh | ns  |  7.1 | Disabled
PCIe Slot6       | 90h | ns  |  7.1 | Disabled
PCIe Slot7       | 91h | ns  |  7.1 | Disabled
PCIe Slot8       | 92h | ns  |  7.1 | Disabled
Drive 0          | 98h | ok  |  7.1 | Drive Present
Drive 15         | A7h | ok  |  7.1 |
Drive 30         | B6h | ns  |  7.1 | Disabled
Cable PCIe A0    | BAh | ns  | 26.2 | Disabled
Cable PCIe B0    | BBh | ns  | 26.2 | Disabled
Cable PCIe A1    | BCh | ns  | 26.2 | Disabled
Cable PCIe B1    | BDh | ns  | 26.2 | Disabled
Cable PCIe A2    | BEh | ns  | 26.2 | Disabled
Cable PCIe B2    | BFh | ns  | 26.2 | Disabled
Cable SAS A0     | C0h | ok  | 26.2 | Connected
Cable SAS B0     | C1h | ok  | 26.2 | Connected
Cable SAS A1     | C2h | ns  | 26.2 | Disabled
Cable SAS B1     | C3h | ns  | 26.2 | Disabled
Cable SAS A2     | C4h | ns  | 26.2 | Disabled
Cable SAS B2     | C5h | ns  | 26.2 | Disabled
Cable PCIe A0    | C6h | ns  | 26.3 | Disabled
Cable PCIe B0    | C7h | ns  | 26.3 | Disabled
A                | CAh | ok  | 32.1 | Presence Detected
B                | D6h | ok  | 32.1 | Presence Detected
ECC Corr Err     | 01h | ns  | 34.1 | No Reading
ECC Uncorr Err   | 02h | ns  | 34.1 | No Reading
PCI Parity Err   | 04h | ns  | 34.1 | No Reading
PCI System Err   | 05h | ns  | 34.1 | No Reading
SBE Log Disabled | 06h | ns  | 34.1 | No Reading
Unknown          | 08h | ns  | 34.1 | No Reading
CPU Machine Chk  | 0Dh | ns  | 34.1 | No Reading
Memory Spared    | 11h | ns  | 34.1 | No Reading
Memory Mirrored  | 12h | ns  | 34.1 | No Reading
PCIE Fatal Err   | 18h | ns  | 34.1 | No Reading
Chipset Err      | 19h | ns  | 34.1 | No Reading
Err Reg Pointer  | 1Ah | ns  | 34.1 | No Reading
Mem ECC Warning  | 1Bh | ns  | 34.1 | No Reading
POST Err         | 1Eh | ns  | 34.1 | No Reading
Hdwr version err | 1Fh | ns  | 34.1 | No Reading
Non Fatal PCI Er | 26h | ns  | 34.1 | No Reading
Fatal IO Error   | 27h | ns  | 34.1 | No Reading
MSR Info Log     | 28h | ns  | 34.1 | No Reading
TXT Status       | 2Ah | ns  | 34.1 | No Reading
iDPT Mem Fail    | 2Bh | ns  | 34.1 | No Reading
Additional Info  | 2Eh | ns  | 34.1 | No Reading
CPU TDP          | 2Fh | ns  | 34.1 | No Reading
QPIRC Warning    | 30h | ns  | 34.1 | No Reading
QPIRC Warning    | 31h | ns  | 34.1 | No Reading
Link Warning     | 32h | ns  | 34.1 | No Reading
Link Warning     | 33h | ns  | 34.1 | No Reading
Link Error       | 34h | ns  | 34.1 | No Reading
MRC Warning      | 35h | ns  | 34.1 | No Reading
MRC Warning      | 36h | ns  | 34.1 | No Reading
Chassis Mismatch | 37h | ns  | 34.1 | No Reading
FatalPCIErrOnBus | 38h | ns  | 34.1 | No Reading
NonFatalPCIErBus | 39h | ns  | 34.1 | No Reading
Fatal PCI SSD Er | 3Ah | ns  | 34.1 | No Reading
NonFatalSSDEr    | 3Bh | ns  | 34.1 | No Reading
CPUMachineCheck  | 3Ch | ns  | 34.1 | No Reading
FatalPCIErARI    | 3Dh | ns  | 34.1 | No Reading
NonFatalPCIErARI | 3Eh | ns  | 34.1 | No Reading
FatalPCIExpEr    | 3Fh | ns  | 34.1 | No Reading
NonFatalPCIExpEr | 40h | ns  | 34.1 | No Reading
CP Left Pres     | 56h | ok  |  7.1 | Present
CP Right Pres    | 57h | ok  |  7.1 | Present
CPU Link Info    | 42h | ns  | 34.1 | No Reading
Chipset Info     | 43h | ns  | 34.1 | No Reading
Memory Config    | 44h | ns  | 34.1 | No Reading
QPI Link Err     | 29h | ns  | 34.1 | No Reading
LT/Flex Addr     | 25h | ns  | 34.1 | No Reading
OS Watchdog Time | 23h | ns  | 34.1 | No Reading
TPM Presence     | 41h | ns  | 34.1 | No Reading
VSA PG           | 2Dh | ok  |  3.1 | State Deasserted
VSA PG           | 37h | ok  |  3.2 | State Deasserted
TPM Presence     | 54h | ok  |  7.1 | Absent
Dedicated NIC    | 70h | ok  |  7.1 | Present
Status           | 53h | ns  | 11.6 | No Reading
Pfault Fail Safe | 74h | ns  |  7.1 | No Reading
Unresp sensor    | FEh | ok  |  7.1 |
Presence         | 4Ch | ok  | 11.4 | Absent
GPU1 Temp        | 89h | ns  |  7.1 | Disabled
GPU2 Temp        | 8Ah | ns  |  7.1 | Disabled
Riser 3 Presence | 4Fh | ok  |  7.1 |
Riser 2 Presence | 4Eh | ok  |  7.1 | Connected
Riser 1 Presence | 4Dh | ok  |  7.1 | Connected
GPU3 Temp        | 62h | ns  |  7.1 | Disabled
GPU4 Temp        | 63h | ns  |  7.1 | Disabled
GPU5 Temp        | 64h | ns  |  7.1 | Disabled
GPU6 Temp        | 65h | ns  |  7.1 | Disabled
GPU7 Temp        | FCh | ns  |  7.1 | Disabled
GPU8 Temp        | FDh | ns  |  7.1 | Disabled
Front LED Panel  | 55h | ok  |  7.1 |
NVDIMM Battery   | 6Ah | ns  |  7.1 | Disabled
NVDIMM Warning   | 46h | ns  | 34.1 | No Reading
NVDIMM Error     | 47h | ns  | 34.1 | No Reading
NVDIMM Info      | 48h | ns  | 34.1 | No Reading
POST Pkg Repair  | 45h | ns  | 34.1 | No Reading
Fan1 Status      | 3Eh | ok  |  7.1 |
Fan2 Status      | 3Fh | ok  |  7.1 |
Fan3 Status      | 40h | ok  |  7.1 |
Fan4 Status      | 41h | ok  |  7.1 |
Fan5 Status      | 42h | ok  |  7.1 |
Fan6 Status      | 43h | ok  |  7.1 |
Exhaust Temp     | 06h | ok  |  7.1 | 31 degrees C
ROMB Battery     | 83h | ns  | 11.5 | Disabled
Cable SAS A0     | C8h | ns  | 26.3 | Disabled
Cable SAS A0     | C9h | ns  | 26.1 | Disabled
MMIOChipset Info | 49h | ns  | 34.1 | No Reading
DIMM Media Info  | 4Ah | ns  | 34.1 | No Reading
DIMMThermal Info | 4Bh | ns  | 34.1 | No Reading
CPU Internal Err | 09h | ns  | 34.1 | No Reading
`
)
