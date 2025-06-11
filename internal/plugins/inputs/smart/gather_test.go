// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package smart

import (
	T "testing"

	"github.com/stretchr/testify/assert"
)

type mockGetter struct {
	sample []byte
}

var (
	macSample = []byte(`smartctl 7.5 2025-04-30 r5714 [Darwin 23.6.0 arm64] (local build)
Copyright (C) 2002-25, Bruce Allen, Christian Franke, www.smartmontools.org

=== START OF INFORMATION SECTION ===
Model Number:                       APPLE SSD AP0512R
Serial Number:                      0ba0160a2204883e
Firmware Version:                   373.140.
PCI Vendor/Subsystem ID:            0x106b
IEEE OUI Identifier:                0x000000
Controller ID:                      0
NVMe Version:                       <1.2
Number of Namespaces:               3
Local Time is:                      Fri May 23 12:12:27 2025 CST

=== START OF SMART DATA SECTION ===
SMART overall-health self-assessment test result: PASSED

SMART/Health Information (NVMe Log 0x02, NSID 0xffffffff)
Critical Warning:                   0x00
Temperature:                        35 Celsius
Available Spare:                    100%
Available Spare Threshold:          99%
Percentage Used:                    5%
Data Units Read:                    336,372,478 [172 TB]
Data Units Written:                 293,070,898 [150 TB]
Host Read Commands:                 11,327,132,499
Host Write Commands:                4,134,253,645
Controller Busy Time:               0
Power Cycles:                       143
Power On Hours:                     2,845
Unsafe Shutdowns:                   14
Media and Data Integrity Errors:    0
Error Information Log Entries:      0`)

	linuxSample = []byte(`=== START OF INFORMATION SECTION ===
Model Family:     Marvell based SanDisk SSDs
Device Model:     SanDisk Ultra II 240GB
Serial Number:    170634802358
LU WWN Device Id: 5 001b44 8b459e841
Firmware Version: X41200RL
User Capacity:    240,057,409,536 bytes [240 GB]
Sector Size:      512 bytes logical/physical
Rotation Rate:    Solid State Device
Form Factor:      2.5 inches
Device is:        In smartctl database [for details use: -P show]
ATA Version is:   ACS-2 T13/2015-D revision 3
SATA Version is:  SATA 3.2, 6.0 Gb/s (current: 6.0 Gb/s)
Local Time is:    Mon May 31 11:44:32 2021 CST
SMART support is: Available - device has SMART capability.
SMART support is: Enabled
Power mode is:    ACTIVE or IDLE

=== START OF READ SMART DATA SECTION ===
SMART overall-health self-assessment test result: PASSED

SMART Attributes Data Structure revision number: 4
Vendor Specific SMART Attributes with Thresholds:
ID# ATTRIBUTE_NAME          FLAGS    VALUE WORST THRESH FAIL RAW_VALUE
  5 Reallocated_Sector_Ct   -O--CK   100   100   ---    -    0
  9 Power_On_Hours          -O--CK   100   100   ---    -    14491
 12 Power_Cycle_Count       -O--CK   100   100   ---    -    58
165 Total_Write/Erase_Count -O--CK   100   100   ---    -    13812053225
166 Min_W/E_Cycle           -O--CK   100   100   ---    -    7
167 Min_Bad_Block/Die       -O--CK   100   100   ---    -    41
168 Maximum_Erase_Cycle     -O--CK   100   100   ---    -    68
169 Total_Bad_Block         -O--CK   100   100   ---    -    329
170 Unknown_Attribute       -O--CK   100   100   ---    -    0
171 Program_Fail_Count      -O--CK   100   100   ---    -    0
172 Erase_Fail_Count        -O--CK   100   100   ---    -    0
173 Avg_Write/Erase_Count   -O--CK   100   100   ---    -    31
174 Unexpect_Power_Loss_Ct  -O--CK   100   100   ---    -    31
184 End-to-End_Error        -O--CK   100   100   ---    -    0
187 Reported_Uncorrect      -O--CK   100   100   ---    -    0
188 Command_Timeout         -O--CK   100   100   ---    -    0
194 Temperature_Celsius     -O---K   064   042   ---    -    36 (Min/Max 13/42)
199 SATA_CRC_Error          -O--CK   100   100   ---    -    0
230 Perc_Write/Erase_Count  -O--CK   100   100   ---    -    9798 1066 9798
232 Perc_Avail_Resrvd_Space PO--CK   100   100   004    -    100
233 Total_NAND_Writes_GiB   -O--CK   100   100   ---    -    7510
234 Perc_Write/Erase_Ct_BC  -O--CK   100   100   ---    -    138778
241 Total_Writes_GiB        ----CK   253   253   ---    -    65774
242 Total_Reads_GiB         ----CK   253   253   ---    -    991
244 Thermal_Throttle        -O--CK   000   100   ---    -    0
                            ||||||_ K auto-keep
                            |||||__ C event count
                            ||||___ R error rate
                            |||____ S speed/performance
                            ||_____ O updated online
                            |______ P prefailure warning`)
)

func (x *mockGetter) Get([]string) []byte {
	return x.sample
}

func (*mockGetter) ExitStatus() int {
	return 0
}

func Test_gatherDisk(t *T.T) {
	t.Run(`basic-linux`, func(t *T.T) {
		ipt := defaultInput()
		ipt.getter = &mockGetter{sample: linuxSample}
		pt, err := ipt.gatherDisk("some-device")
		assert.NoError(t, err)

		t.Logf("point: %s", pt.Pretty())
	})

	t.Run(`mac-linux`, func(t *T.T) {
		ipt := defaultInput()
		ipt.getter = &mockGetter{sample: macSample}
		pt, err := ipt.gatherDisk("some-device")
		assert.NoError(t, err)

		t.Logf("point: %s", pt.Pretty())
	})
}
