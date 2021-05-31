# summary

计算机磁盘状态监测，支持 NVME 磁盘状态监测。使用 `smartctl` 命令抓取状态数据。

# prerequisite

`apt install smartmontools -y`
`apt install nvme-cli -y` // linux only

# raw data sample

## smartctl --scan

```
/dev/sda -d scsi # /dev/sda, SCSI device
/dev/sdb -d scsi # /dev/sdb, SCSI device
```

## smartctl --info --health --attributes --tolerance=verypermissive -n standby --format brief /dev/sda

```
=== START OF INFORMATION SECTION ===
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
                            |______ P prefailure warning
```

## smartctl --info --health --attributes --tolerance verypermissive -n standby --format=brief /dev/sdb

```
=== START OF INFORMATION SECTION ===
Model Family:     Seagate Surveillance
Device Model:     ST3000VX000-1ES166
Serial Number:    W5033JSY
LU WWN Device Id: 5 000c50 09def8171
Firmware Version: CV27
User Capacity:    3,000,592,982,016 bytes [3.00 TB]
Sector Sizes:     512 bytes logical, 4096 bytes physical
Rotation Rate:    7200 rpm
Form Factor:      3.5 inches
Device is:        In smartctl database [for details use: -P show]
ATA Version is:   ACS-2, ACS-3 T13/2161-D revision 3b
SATA Version is:  SATA 3.1, 6.0 Gb/s (current: 6.0 Gb/s)
Local Time is:    Fri May 28 17:35:13 2021 CST
SMART support is: Available - device has SMART capability.
SMART support is: Enabled
Power mode is:    ACTIVE or IDLE

=== START OF READ SMART DATA SECTION ===
SMART overall-health self-assessment test result: PASSED

SMART Attributes Data Structure revision number: 10
Vendor Specific SMART Attributes with Thresholds:
ID# ATTRIBUTE_NAME          FLAGS    VALUE WORST THRESH FAIL RAW_VALUE
  1 Raw_Read_Error_Rate     POSR--   119   099   006    -    216898424
  3 Spin_Up_Time            PO----   095   094   000    -    0
  4 Start_Stop_Count        -O--CK   100   100   020    -    62
  5 Reallocated_Sector_Ct   PO--CK   100   100   010    -    0
  7 Seek_Error_Rate         POSR--   081   060   030    -    149211294
  9 Power_On_Hours          -O--CK   078   078   000    -    19717
 10 Spin_Retry_Count        PO--C-   100   100   097    -    0
 12 Power_Cycle_Count       -O--CK   100   100   020    -    62
184 End-to-End_Error        -O--CK   100   100   099    -    0
187 Reported_Uncorrect      -O--CK   100   100   000    -    0
188 Command_Timeout         -O--CK   100   100   000    -    0
189 High_Fly_Writes         -O-RCK   001   001   000    -    1971
190 Airflow_Temperature_Cel -O---K   064   055   045    -    36 (Min/Max 23/37)
191 G-Sense_Error_Rate      -O--CK   100   100   000    -    0
192 Power-Off_Retract_Count -O--CK   100   100   000    -    41
193 Load_Cycle_Count        -O--CK   100   100   000    -    114
194 Temperature_Celsius     -O---K   036   045   000    -    36 (0 15 0 0 0)
197 Current_Pending_Sector  -O--C-   100   100   000    -    0
198 Offline_Uncorrectable   ----C-   100   100   000    -    0
199 UDMA_CRC_Error_Count    -OSRCK   200   200   000    -    0
                            ||||||_ K auto-keep
                            |||||__ C event count
                            ||||___ R error rate
                            |||____ S speed/performance
                            ||_____ O updated online
                            |______ P prefailure warning
```

# config sample

```
[[inputs.smart]]
	## The path to the smartctl executable
  # path_smartctl = "/usr/bin/smartctl"

  ## The path to the nvme-cli executable
  # path_nvme = "/usr/bin/nvme"

	## Gathering interval
	# interval = "10s"

  ## Timeout for the cli command to complete.
  # timeout = "30s"

  ## Optionally specify if vendor specific attributes should be propagated for NVMe disk case
  ## ["auto-on"] - automatically find and enable additional vendor specific disk info
  ## ["vendor1", "vendor2", ...] - e.g. "Intel" enable additional Intel specific disk info
  # enable_extensions = ["auto-on"]

  ## On most platforms used cli utilities requires root access.
  ## Setting 'use_sudo' to true will make use of sudo to run smartctl or nvme-cli.
  ## Sudo must be configured to allow the telegraf user to run smartctl or nvme-cli
  ## without a password.
  # use_sudo = false

  ## Skip checking disks in this power mode. Defaults to "standby" to not wake up disks that have stopped rotating.
  ## See --nocheck in the man pages for smartctl.
  ## smartctl version 5.41 and 5.42 have faulty detection of power mode and might require changing this value to "never" depending on your disks.
  # no_check = "standby"

  ## Optionally specify devices to exclude from reporting if disks auto-discovery is performed.
  # excludes = [ "/dev/pass6" ]

  ## Optionally specify devices and device type, if unset a scan (smartctl --scan and smartctl --scan -d nvme) for S.M.A.R.T. devices will be done
  ## and all found will be included except for the excluded in excludes.
  # devices = [ "/dev/ada0 -d atacam", "/dev/nvme0"]

	## Customer tags, if set will be seen with every metric.
	[inputs.smart.tags]
		# "key1" = "value1"
		# "key2" = "value2"
```

# metrics

## smart

| 标签名      | 描述                                             |
| ----------- | ------------------------------------------------ |
| capacity    | disk capacity                                    |
| device      | device mount name                                |
| enabled     | is SMART supported                               |
| exit_status | command process status                           |
| health_ok   | SMART overall-health self-assessment test result |
| host        | host name                                        |
| model       | device model                                     |
| serial_no   | device serial number                             |
| wwn         | WWN Device Id                                    |

| 指标                               | 类型 | 指标源 | 单位        | 描述                                                                                                           |
| ---------------------------------- | ---- | ------ | ----------- | -------------------------------------------------------------------------------------------------------------- |
| airflow_temperature_cel_raw_value" | Int  | Gauge  | Celsius     | The raw value of air celsius temperature read from device record.                                              |
| airflow_temperature_cel_threshold" | Int  | Gauge  | Celsius     | The threshold of air celsius temperature read from device record.                                              |
| airflow_temperature_cel_value"     | Int  | Gauge  | Celsius     | The value of air celsius temperature read from device record.                                                  |
| airflow_temperature_cel_worst"     | Int  | Gauge  | Celsius     | The worst value of air celsius temperature read from device record.                                            |
| avg_write/erase_count_raw_value"   | Int  | Gauge  | NCount      | The raw value of average write/ease count.                                                                     |
| avg_write/erase_count_value"       | Int  | Gauge  | NCount      | The value of average write/ease count.                                                                         |
| avg_write/erase_count_worst"       | Int  | Gauge  | NCount      | The worst value of average write/ease count.                                                                   |
| command_timeout_raw_value"         | Int  | Gauge  | NCount      | The raw value of command timeout.                                                                              |
| command_timeout_threshold"         | Int  | Gauge  | NCount      | The threshold of command timeout.                                                                              |
| command_timeout_value"             | Int  | Gauge  | NCount      | The value of command timeout.                                                                                  |
| command_timeout_worst"             | Int  | Gauge  | NCount      | The worst value of command timeout.                                                                            |
| current_pending_sector_raw_value"  | Int  | Gauge  | NCount      | The raw value of current pending sector.                                                                       |
| current_pending_sector_threshold"  | Int  | Gauge  | NCount      | The threshold of current pending sector.                                                                       |
| current_pending_sector_value"      | Int  | Gauge  | NCount      | The value of current pending sector.                                                                           |
| current_pending_sector_worst"      | Int  | Gauge  | NCount      | The worst value of current pending sector.                                                                     |
| end-to-end_error_raw_value"        | Int  | Gauge  | NCount      | The raw value of bad data that loaded into cache and then written to the driver have had a different parity.   |
| end-to-end_error_threshold"        | Int  | Gauge  | NCount      | The threshold of bad data that loaded into cache and then written to the driver have had a different parity.   |
| end-to-end_error_value"            | Int  | Gauge  | NCount      | The value of bad data that loaded into cache and then written to the driver have had a different parity.       |
| end-to-end_error_worst"            | Int  | Gauge  | NCount      | The worst value of bad data that loaded into cache and then written to the driver have had a different parity. |
| erase_fail_count_raw_value"        | Int  | Gauge  | NCount      | The raw value of erase failed count.                                                                           |
| erase_fail_count_value"            | Int  | Gauge  | NCount      | The value of erase failed count.                                                                               |
| erase_fail_count_worst"            | Int  | Gauge  | NCount      | The worst value of erase failed count.                                                                         |
| fail"                              | Bool | Gauge  | NCount      | Read attribute failed.                                                                                         |
| flags"                             | Int  | Gauge  | NCount      | Attribute falgs.                                                                                               |
| g-sense_error_rate_raw_value"      | Int  | Gauge  | NCount      | The raw value of                                                                                               |
| g-sense_error_rate_threshold"      | Int  | Gauge  | NCount      | The threshold value of g-sensor error rate.                                                                    |
| g-sense_error_rate_value"          | Int  | Gauge  | NCount      | The value of g-sensor error rate.                                                                              |
| g-sense_error_rate_worst"          | Int  | Gauge  | NCount      | The worst value of g-sensor error rate.                                                                        |
| high_fly_writes_raw_value"         | Int  | Gauge  | NCount      | The raw value of Fly Height Monitor.                                                                           |
| high_fly_writes_threshold"         | Int  | Gauge  | NCount      | The threshold value of Fly Height Monitor.                                                                     |
| high_fly_writes_value"             | Int  | Gauge  | NCount      | The value of Fly Height Monitor.                                                                               |
| high_fly_writes_worst"             | Int  | Gauge  | NCount      | The worst value of Fly Height Monitor.                                                                         |
| load_cycle_count_raw_value"        | Int  | Gauge  | NCount      | The raw value of load cycle count.                                                                             |
| load_cycle_count_threshold"        | Int  | Gauge  | NCount      | The threshold value of load cycle count.                                                                       |
| load_cycle_count_value"            | Int  | Gauge  | NCount      | The value of load cycle count.                                                                                 |
| load_cycle_count_worst"            | Int  | Gauge  | NCount      | The worst value of load cycle count.                                                                           |
| maximum_erase_cycle_raw_value"     | Int  | Gauge  | NCount      | The raw value of maximum erase cycle count.                                                                    |
| maximum_erase_cycle_value"         | Int  | Gauge  | NCount      | The raw value of maximum erase cycle count.                                                                    |
| maximum_erase_cycle_worst"         | Int  | Gauge  | NCount      | The worst value of maximum erase cycle count.                                                                  |
| min_bad_block/die_raw_value"       | Int  | Gauge  | NCount      | The raw value of min bad block.                                                                                |
| min_bad_block/die_value"           | Int  | Gauge  | NCount      | The value of min bad block.                                                                                    |
| min_bad_block/die_worst"           | Int  | Gauge  | NCount      | The worst value of min bad block.                                                                              |
| min_w/e_cycle_raw_value"           | Int  | Gauge  | NCount      | The raw value of min write/erase cycle count.                                                                  |
| min_w/e_cycle_value"               | Int  | Gauge  | NCount      | The value of min write/erase cycle count.                                                                      |
| min_w/e_cycle_worst"               | Int  | Gauge  | NCount      | The worst value of min write/erase cycle count.                                                                |
| offline_uncorrectable_raw_value"   | Int  | Gauge  | NCount      | The raw value of offline uncorrectable.                                                                        |
| offline_uncorrectable_threshold"   | Int  | Gauge  | NCount      | The threshold value of offline uncorrectable.                                                                  |
| offline_uncorrectable_value"       | Int  | Gauge  | NCount      | The value of offline uncorrectable.                                                                            |
| offline_uncorrectable_worst"       | Int  | Gauge  | NCount      | The worst value of offline uncorrectable.                                                                      |
| perc_avail_resrvd_space_raw_value" | Int  | Gauge  | NCount      | The raw value of available percentage of reserved space.                                                       |
| perc_avail_resrvd_space_threshold" | Int  | Gauge  | NCount      | The threshold value of available percentage of reserved space.                                                 |
| perc_avail_resrvd_space_value"     | Int  | Gauge  | NCount      | The value of available reserved space.                                                                         |
| perc_avail_resrvd_space_worst"     | Int  | Gauge  | NCount      | The worst value of available reserved space.                                                                   |
| perc_write/erase_count_raw_value"  | Int  | Gauge  | NCount      | The raw value of write/erase count.                                                                            |
| perc_write/erase_count_value"      | Int  | Gauge  | NCount      | The value of of write/erase count.                                                                             |
| perc_write/erase_count_worst"      | Int  | Gauge  | NCount      | The worst value of of write/erase count.                                                                       |
| perc_write/erase_ct_bc_raw_value"  | Int  | Gauge  | NCount      | The raw value of write/erase count.                                                                            |
| perc_write/erase_ct_bc_value"      | Int  | Gauge  | NCount      | The value of write/erase count.                                                                                |
| perc_write/erase_ct_bc_worst"      | Int  | Gauge  | NCount      | The worst value of write/erase count.                                                                          |
| power_cycle_count_raw_value"       | Int  | Gauge  | NCount      | The raw value of power cycle count.                                                                            |
| power_cycle_count_threshold"       | Int  | Gauge  | NCount      | The threshold value of power cycle count.                                                                      |
| power_cycle_count_value"           | Int  | Gauge  | NCount      | The value of power cycle count.                                                                                |
| power_cycle_count_worst"           | Int  | Gauge  | NCount      | The worst value of power cycle count.                                                                          |
| power_on_hours_raw_value"          | Int  | Gauge  | NCount      | The raw value of power on hours.                                                                               |
| power_on_hours_threshold"          | Int  | Gauge  | NCount      | The threshold value of power on hours.                                                                         |
| power_on_hours_value"              | Int  | Gauge  | NCount      | The value of power on hours.                                                                                   |
| power_on_hours_worst"              | Int  | Gauge  | NCount      | The worst value of power on hours.                                                                             |
| power-off_retract_count_raw_value" | Int  | Gauge  | NCount      | The raw value of power-off retract count.                                                                      |
| power-off_retract_count_threshold" | Int  | Gauge  | NCount      | The threshold value of power-off retract count.                                                                |
| power-off_retract_count_value"     | Int  | Gauge  | NCount      | The value of power-off retract count.                                                                          |
| power-off_retract_count_worst"     | Int  | Gauge  | NCount      | The worst value of power-off retract count.                                                                    |
| program_fail_count_raw_value"      | Int  | Gauge  | NCount      | The raw value of program fail count.                                                                           |
| program_fail_count_value"          | Int  | Gauge  | NCount      | The value of program fail count.                                                                               |
| program_fail_count_worst"          | Int  | Gauge  | NCount      | The worst value of program fail count.                                                                         |
| raw_read_error_rate_raw_value"     | Int  | Gauge  | NCount      | The raw value of raw read error rate.                                                                          |
| raw_read_error_rate_threshold"     | Int  | Gauge  | NCount      | The threshold value of raw read error rate.                                                                    |
| raw_read_error_rate_value"         | Int  | Gauge  | NCount      | The value of raw read error rate.                                                                              |
| raw_read_error_rate_worst"         | Int  | Gauge  | NCount      | The worst value of raw read error rate.                                                                        |
| read_error_rate"                   | Int  | Gauge  | NCount      | The read error rate.                                                                                           |
| reallocated_sector_ct_raw_value"   | Int  | Gauge  | NCount      | The raw value of reallocated sector count.                                                                     |
| reallocated_sector_ct_threshold"   | Int  | Gauge  | NCount      | The threshold value of reallocated sector count.                                                               |
| reallocated_sector_ct_value"       | Int  | Gauge  | NCount      | The value of reallocated sector count.                                                                         |
| reallocated_sector_ct_worst"       | Int  | Gauge  | NCount      | The worst value of reallocated sector count.                                                                   |
| reported_uncorrect_raw_value"      | Int  | Gauge  | NCount      | The raw value of reported uncorrect.                                                                           |
| reported_uncorrect_threshold"      | Int  | Gauge  | NCount      | The threshold value of reported uncorrect.                                                                     |
| reported_uncorrect_value"          | Int  | Gauge  | NCount      | The value of reported uncorrect.                                                                               |
| reported_uncorrect_worst"          | Int  | Gauge  | NCount      | The worst value of reported uncorrect.                                                                         |
| sata_crc_error_raw_value"          | Int  | Gauge  | NCount      | The raw value of S-ATA cyclic redundancy check error.                                                          |
| sata_crc_error_value"              | Int  | Gauge  | NCount      | The value of S-ATA cyclic redundancy check error.                                                              |
| sata_crc_error_worst"              | Int  | Gauge  | NCount      | The worst value of S-ATA cyclic redundancy check error.                                                        |
| seek_error_rate_raw_value"         | Int  | Gauge  | NCount      | The raw value of seek error rate.                                                                              |
| seek_error_rate_threshold"         | Int  | Gauge  | NCount      | The threshold value of seek error rate.                                                                        |
| seek_error_rate_value"             | Int  | Gauge  | NCount      | The value of seek error rate.                                                                                  |
| seek_error_rate_worst"             | Int  | Gauge  | NCount      | The worst value of seek error rate.                                                                            |
| seek_error_rate"                   | Int  | Gauge  | NCount      | Seek error rate.                                                                                               |
| spin_retry_count_raw_value"        | Int  | Gauge  | NCount      | The raw value of spin retry count.                                                                             |
| spin_retry_count_threshold"        | Int  | Gauge  | NCount      | The threshold value of spin retry count.                                                                       |
| spin_retry_count_value"            | Int  | Gauge  | NCount      | The value of spin retry count.                                                                                 |
| spin_retry_count_worst"            | Int  | Gauge  | NCount      | The worst value of spin retry count.                                                                           |
| spin_up_time_raw_value"            | Int  | Gauge  | NCount      | The raw value of spin up time.                                                                                 |
| spin_up_time_threshold"            | Int  | Gauge  | NCount      | The threshold value of spin up time.                                                                           |
| spin_up_time_value"                | Int  | Gauge  | NCount      | The value of spin up time.                                                                                     |
| spin_up_time_worst"                | Int  | Gauge  | NCount      | The worst value of spin up time.                                                                               |
| start_stop_count_raw_value"        | Int  | Gauge  | NCount      | The raw value of start and stop count.                                                                         |
| start_stop_count_threshold"        | Int  | Gauge  | NCount      | The threshold value of start and stop count.                                                                   |
| start_stop_count_value"            | Int  | Gauge  | NCount      | The value of start and stop count.                                                                             |
| start_stop_count_worst"            | Int  | Gauge  | NCount      | The worst value of start and stop count.                                                                       |
| temp_c"                            | Int  | Gauge  | Celsius     | Device temperature.                                                                                            |
| temperature_celsius_raw_value"     | Int  | Gauge  | Celsius     | The raw value of temperature.                                                                                  |
| temperature_celsius_threshold"     | Int  | Gauge  | Celsius     | The threshold value of themperature.                                                                           |
| temperature_celsius_value"         | Int  | Gauge  | Celsius     | The value of temperature.                                                                                      |
| temperature_celsius_worst"         | Int  | Gauge  | Celsius     | The worst value of temperature.                                                                                |
| thermal_throttle_raw_value"        | Int  | Gauge  | NCount      | The raw value of thermal throttle.                                                                             |
| thermal_throttle_value"            | Int  | Gauge  | NCount      | The value of thermal throttle.                                                                                 |
| thermal_throttle_worst"            | Int  | Gauge  | NCount      | The worst value of thermal throttle.                                                                           |
| total_bad_block_raw_value"         | Int  | Gauge  | NCount      | The raw value of total bad block.                                                                              |
| total_bad_block_value"             | Int  | Gauge  | NCount      | The value of total bad block.                                                                                  |
| total_bad_block_worst"             | Int  | Gauge  | NCount      | The worst value of total bad block.                                                                            |
| total_nand_writes_gib_raw_value"   | Int  | Gauge  | NCount      | The raw value of total NAND flush writes.                                                                      |
| total_nand_writes_gib_value"       | Int  | Gauge  | NCount      | The value of total NAND flush writes.                                                                          |
| total_nand_writes_gib_worst"       | Int  | Gauge  | NCount      | The worst value of total NAND flush writes.                                                                    |
| total_reads_gib_raw_value"         | Int  | Gauge  | NCount      | The raw value of total read.                                                                                   |
| total_reads_gib_value"             | Int  | Gauge  | NCount      | The value of total read.                                                                                       |
| total_reads_gib_worst"             | Int  | Gauge  | NCount      | The worst value of total read                                                                                  |
| total_write/erase_count_raw_value" | Int  | Gauge  | NCount      | The raw value of total write/erase count.                                                                      |
| total_write/erase_count_value"     | Int  | Gauge  | NCount      | The value of total write/erase count.                                                                          |
| total_write/erase_count_worst"     | Int  | Gauge  | NCount      | The worst value of total write/erase count.                                                                    |
| total_writes_gib_raw_value"        | Int  | Gauge  | NCount      | The raw value of total write.                                                                                  |
| total_writes_gib_value"            | Int  | Gauge  | NCount      | The value of total write.                                                                                      |
| total_writes_gib_worst"            | Int  | Gauge  | NCount      | The worst value of total write.                                                                                |
| udma_crc_error_count_raw_value"    | Int  | Gauge  | NCount      | The raw value of ultra direct memory access cyclic redundancy check error count.                               |
| udma_crc_error_count_threshold"    | Int  | Gauge  | NCount      | The threshold value of ultra direct memory access cyclic redundancy check error count.                         |
| udma_crc_error_count_value"        | Int  | Gauge  | NCount      | The value of ultra direct memory access cyclic redundancy check error count.                                   |
| udma_crc_error_count_worst"        | Int  | Gauge  | NCount      | The worst value of ultra direct memory access cyclic redundancy check error count.                             |
| udma_crc_errors"                   | Int  | Gauge  | NCount      | Ultra direct memory access cyclic redundancy check error count.                                                |
| unexpect_power_loss_ct_raw_value"  | Int  | Gauge  | NCount      | The raw value of unexpected power loss count.                                                                  |
| unexpect_power_loss_ct_value"      | Int  | Gauge  | NCount      | The value of unexpected power loss count.                                                                      |
| unexpect_power_loss_ct_worst"      | Int  | Gauge  | NCount      | The worst value of unexpected power loss count.                                                                |
| unknown_attribute_raw_value"       | Int  | Gauge  | UnknownUnit | The raw value of nknow attribute.                                                                              |
| unknown_attribute_value"           | Int  | Gauge  | UnknownUnit | The value of unknow attribute.                                                                                 |
| unknown_attribute_worst"           | Int  | Gauge  | UnknownUnit | The worst value of unknow attribute.                                                                           |
