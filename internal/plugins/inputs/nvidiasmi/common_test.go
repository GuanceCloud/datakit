// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2022-present Guance, Inc.

// Package nvidiasmi collects host nvidiasmi metrics.
package nvidiasmi

import (
	"reflect"
	"sort"
	"testing"

	"github.com/GuanceCloud/cliutils"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/io/point"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs"
)

func TestInput_convert(t *testing.T) {
	type fields struct {
		Interval            datakit.Duration
		Tags                map[string]string
		collectCache        []inputs.Measurement
		pts                 []*point.Point
		platform            string
		BinPaths            []string
		Timeout             datakit.Duration
		ProcessInfoMaxLen   int
		GPUDropWarningDelay datakit.Duration
		Envs                []string
		SSHServers          datakit.SSHServers
		GPUs                []GPUInfo
		semStop             *cliutils.Sem
		Election            bool
	}
	type args struct {
		data   []byte
		server string
	}
	tests := []struct {
		name             string
		fields           fields
		args             args
		wantCollectCache []inputs.Measurement
	}{
		// TODO: Add test cases.
		{
			name: "local single card",
			fields: fields{
				collectCache:      []inputs.Measurement{},
				pts:               []*point.Point{},
				GPUs:              []GPUInfo{},
				ProcessInfoMaxLen: 0,
			},
			args: args{
				data:   gtx10501,
				server: "",
			},
			wantCollectCache: []inputs.Measurement{
				&nvidiaSmiMeasurement{
					name:   "gpu_smi",
					tags:   map[string]string{"pstate": "P3", "name": "NVIDIA GeForce GTX 1050", "uuid": "GPU-06e04616-0ed5-4069-5ebc-345349a0d4f3", "compute_mode": "Default", "pci_bus_id": "00000000:01:00.0", "driver_version": "515.65.01", "cuda_version": "11.7"},
					fields: map[string]interface{}{"memory_total": 4096, "utilization_encoder": 0, "encoder_stats_session_count": 0, "fbc_stats_session_count": 0, "fbc_stats_average_latency": 0, "clocks_current_sm": 683, "clocks_current_memory": 2504, "fbc_stats_average_fps": 0, "temperature_gpu": 44, "utilization_memory": 3, "pcie_link_gen_current": 3, "clocks_current_graphics": 683, "clocks_current_video": 620, "memory_used": 65, "utilization_gpu": 7, "utilization_decoder": 0, "pcie_link_width_current": 8, "encoder_stats_average_fps": 0, "encoder_stats_average_latency": 0},
				},
			},
		},
		{
			name: "remote 8 card",
			fields: fields{
				collectCache:      []inputs.Measurement{},
				pts:               []*point.Point{},
				GPUs:              []GPUInfo{},
				ProcessInfoMaxLen: 0,
			},
			args: args{
				data:   gtx8card,
				server: "1.1.1.1:22",
			},
			wantCollectCache: []inputs.Measurement{
				&nvidiaSmiMeasurement{
					name:   "gpu_smi",
					tags:   map[string]string{"host": "1.1.1.1:22", "pstate": "P0", "name": "Tesla V100S-PCIE-32GB", "uuid": "GPU-1d955959-c95b-4f40-2cac-100d3006ffe2", "compute_mode": "Default", "pci_bus_id": "00000000:00:14.0", "driver_version": "460.73.01", "cuda_version": "11.2"},
					fields: map[string]interface{}{"memory_used": 29501, "utilization_gpu": 99, "encoder_stats_average_latency": 0, "fbc_stats_average_latency": 0, "clocks_current_sm": 1575, "clocks_current_video": 1417, "power_draw": 61.47, "memory_total": 32510, "encoder_stats_session_count": 0, "fbc_stats_session_count": 0, "clocks_current_memory": 1107, "utilization_memory": 4, "utilization_encoder": 0, "encoder_stats_average_fps": 0, "clocks_current_graphics": 1575, "temperature_gpu": 57, "utilization_decoder": 0, "pcie_link_gen_current": 3, "pcie_link_width_current": 16, "fbc_stats_average_fps": 0},
				},
				&nvidiaSmiMeasurement{
					name:   "gpu_smi",
					tags:   map[string]string{"host": "1.1.1.1:22", "pstate": "P0", "name": "Tesla V100S-PCIE-32GB", "uuid": "GPU-2aa87d54-d4ff-2900-812b-1a4b7e119583", "compute_mode": "Default", "pci_bus_id": "00000000:00:11.0", "driver_version": "460.73.01", "cuda_version": "11.2"},
					fields: map[string]interface{}{"temperature_gpu": 59, "utilization_memory": 55, "pcie_link_gen_current": 3, "fbc_stats_average_fps": 0, "clocks_current_graphics": 1342, "clocks_current_memory": 1107, "power_draw": 117.99, "memory_total": 32510, "utilization_encoder": 0, "encoder_stats_average_fps": 0, "clocks_current_video": 1207, "utilization_gpu": 98, "pcie_link_width_current": 16, "encoder_stats_session_count": 0, "encoder_stats_average_latency": 0, "fbc_stats_session_count": 0, "fbc_stats_average_latency": 0, "clocks_current_sm": 1342, "memory_used": 29517, "utilization_decoder": 0},
				},
				&nvidiaSmiMeasurement{
					name:   "gpu_smi",
					tags:   map[string]string{"host": "1.1.1.1:22", "pstate": "P0", "name": "Tesla V100S-PCIE-32GB", "uuid": "GPU-3181eb56-505c-8214-ba05-c132f68f89b8", "compute_mode": "Default", "pci_bus_id": "00000000:00:10.0", "driver_version": "460.73.01", "cuda_version": "11.2"},
					fields: map[string]interface{}{"temperature_gpu": 61, "utilization_memory": 12, "utilization_decoder": 0, "clocks_current_video": 1282, "memory_total": 32510, "memory_used": 32099, "utilization_gpu": 23, "pcie_link_gen_current": 3, "encoder_stats_session_count": 0, "encoder_stats_average_fps": 0, "fbc_stats_average_fps": 0, "fbc_stats_average_latency": 0, "clocks_current_sm": 1432, "power_draw": 198.41, "encoder_stats_average_latency": 0, "fbc_stats_session_count": 0, "utilization_encoder": 0, "pcie_link_width_current": 16, "clocks_current_graphics": 1432, "clocks_current_memory": 1107},
				},
				&nvidiaSmiMeasurement{
					name:   "gpu_smi",
					tags:   map[string]string{"host": "1.1.1.1:22", "pstate": "P0", "name": "Tesla V100S-PCIE-32GB", "uuid": "GPU-39c3a168-9034-0b5a-223a-1042f30a9242", "compute_mode": "Default", "pci_bus_id": "00000000:00:12.0", "driver_version": "460.73.01", "cuda_version": "11.2"},
					fields: map[string]interface{}{"pcie_link_width_current": 16, "encoder_stats_session_count": 0, "fbc_stats_average_fps": 0, "memory_total": 32510, "temperature_gpu": 64, "utilization_decoder": 0, "encoder_stats_average_fps": 0, "encoder_stats_average_latency": 0, "fbc_stats_session_count": 0, "clocks_current_graphics": 1590, "clocks_current_sm": 1590, "power_draw": 85.8, "utilization_gpu": 98, "clocks_current_memory": 1107, "clocks_current_video": 1425, "memory_used": 32141, "utilization_memory": 50, "utilization_encoder": 0, "pcie_link_gen_current": 3, "fbc_stats_average_latency": 0},
				},
				&nvidiaSmiMeasurement{
					name:   "gpu_smi",
					tags:   map[string]string{"host": "1.1.1.1:22", "pstate": "P0", "name": "Tesla V100S-PCIE-32GB", "uuid": "GPU-76098dc1-d3bc-9a1b-7ea6-f7b48837829e", "compute_mode": "Default", "pci_bus_id": "00000000:00:0D.0", "driver_version": "460.73.01", "cuda_version": "11.2"},
					fields: map[string]interface{}{"memory_total": 32510, "temperature_gpu": 66, "utilization_gpu": 97, "utilization_encoder": 0, "pcie_link_width_current": 16, "fbc_stats_average_latency": 0, "clocks_current_sm": 1410, "utilization_memory": 52, "pcie_link_gen_current": 3, "encoder_stats_average_fps": 0, "clocks_current_memory": 1107, "clocks_current_video": 1275, "encoder_stats_average_latency": 0, "fbc_stats_average_fps": 0, "clocks_current_graphics": 1410, "power_draw": 232.8, "memory_used": 31689, "utilization_decoder": 0, "encoder_stats_session_count": 0, "fbc_stats_session_count": 0},
				},
				&nvidiaSmiMeasurement{
					name:   "gpu_smi",
					tags:   map[string]string{"host": "1.1.1.1:22", "pstate": "P0", "name": "Tesla V100S-PCIE-32GB", "uuid": "GPU-8f75e71c-23dd-a060-d997-4caf592954db", "compute_mode": "Default", "pci_bus_id": "00000000:00:0F.0", "driver_version": "460.73.01", "cuda_version": "11.2"},
					fields: map[string]interface{}{"memory_total": 32510, "temperature_gpu": 57, "encoder_stats_average_fps": 0, "fbc_stats_average_fps": 0, "fbc_stats_average_latency": 0, "power_draw": 67.57, "utilization_gpu": 80, "encoder_stats_session_count": 0, "encoder_stats_average_latency": 0, "fbc_stats_session_count": 0, "clocks_current_graphics": 1590, "clocks_current_memory": 1107, "clocks_current_video": 1425, "memory_used": 29539, "utilization_decoder": 0, "clocks_current_sm": 1590, "utilization_memory": 42, "utilization_encoder": 0, "pcie_link_gen_current": 3, "pcie_link_width_current": 16},
				},
				&nvidiaSmiMeasurement{
					name:   "gpu_smi",
					tags:   map[string]string{"host": "1.1.1.1:22", "pstate": "P0", "name": "Tesla V100S-PCIE-32GB", "uuid": "GPU-bbf7a241-bf9a-eb90-b2fb-d213294faf48", "compute_mode": "Default", "pci_bus_id": "00000000:00:0E.0", "driver_version": "460.73.01", "cuda_version": "11.2"},
					fields: map[string]interface{}{"temperature_gpu": 58, "clocks_current_graphics": 1432, "clocks_current_sm": 1432, "power_draw": 152.03, "utilization_gpu": 75, "encoder_stats_average_fps": 0, "memory_total": 32510, "memory_used": 29545, "utilization_memory": 41, "pcie_link_gen_current": 3, "pcie_link_width_current": 16, "encoder_stats_session_count": 0, "encoder_stats_average_latency": 0, "fbc_stats_session_count": 0, "clocks_current_video": 1282, "utilization_encoder": 0, "utilization_decoder": 0, "fbc_stats_average_fps": 0, "fbc_stats_average_latency": 0, "clocks_current_memory": 1107},
				},
				&nvidiaSmiMeasurement{
					name:   "gpu_smi",
					tags:   map[string]string{"host": "1.1.1.1:22", "pstate": "P0", "name": "Tesla V100S-PCIE-32GB", "uuid": "GPU-d2446f91-171e-1b54-515d-9d9b03484ddf", "compute_mode": "Default", "pci_bus_id": "00000000:00:13.0", "driver_version": "460.73.01", "cuda_version": "11.2"},
					fields: map[string]interface{}{"memory_total": 32510, "pcie_link_width_current": 16, "encoder_stats_session_count": 0, "fbc_stats_average_latency": 0, "clocks_current_sm": 1597, "clocks_current_video": 1432, "utilization_memory": 22, "utilization_encoder": 0, "fbc_stats_average_fps": 0, "utilization_gpu": 91, "fbc_stats_session_count": 0, "clocks_current_graphics": 1597, "memory_used": 32157, "temperature_gpu": 61, "utilization_decoder": 0, "pcie_link_gen_current": 3, "encoder_stats_average_fps": 0, "encoder_stats_average_latency": 0, "clocks_current_memory": 1107, "power_draw": 68.14},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ipt := &Input{
				Interval:            tt.fields.Interval,
				Tags:                tt.fields.Tags,
				collectCache:        tt.fields.collectCache,
				pts:                 tt.fields.pts,
				platform:            tt.fields.platform,
				BinPaths:            tt.fields.BinPaths,
				Timeout:             tt.fields.Timeout,
				ProcessInfoMaxLen:   tt.fields.ProcessInfoMaxLen,
				GPUDropWarningDelay: tt.fields.GPUDropWarningDelay,
				Envs:                tt.fields.Envs,
				SSHServers:          tt.fields.SSHServers,
				GPUs:                tt.fields.GPUs,
				semStop:             tt.fields.semStop,
				Election:            tt.fields.Election,
			}

			err := ipt.convert(tt.args.data, tt.args.server)
			if err != nil {
				t.Errorf("testing error = %v,", err)
			}

			sort.SliceStable(ipt.collectCache, func(i, j int) bool {
				return ipt.collectCache[i].(*nvidiaSmiMeasurement).tags["uuid"] < ipt.collectCache[j].(*nvidiaSmiMeasurement).tags["uuid"]
			})

			sort.SliceStable(tt.wantCollectCache, func(i, j int) bool {
				return tt.wantCollectCache[i].(*nvidiaSmiMeasurement).tags["uuid"] < tt.wantCollectCache[j].(*nvidiaSmiMeasurement).tags["uuid"]
			})

			if !reflect.DeepEqual(ipt.collectCache, tt.wantCollectCache) {
				t.Errorf(" collectCache = %+v,  want %+v", ipt.collectCache, tt.wantCollectCache)
			}
		})
	}
}

var gtx10501 []byte = []byte(`<nvidia_smi_log>
<timestamp>Thu Sep  1 15:20:22 2022</timestamp>
<driver_version>515.65.01</driver_version>
<cuda_version>11.7</cuda_version>
<attached_gpus>1</attached_gpus>
<gpu id="00000000:01:00.0">
	<product_name>NVIDIA GeForce GTX 1050</product_name>
	<product_brand>GeForce</product_brand>
	<product_architecture>Pascal</product_architecture>
	<display_mode>Enabled</display_mode>
	<display_active>Enabled</display_active>
	<persistence_mode>Enabled</persistence_mode>
	<mig_mode>
		<current_mig>N/A</current_mig>
		<pending_mig>N/A</pending_mig>
	</mig_mode>
	<mig_devices>
		None
	</mig_devices>
	<accounting_mode>Disabled</accounting_mode>
	<accounting_mode_buffer_size>4000</accounting_mode_buffer_size>
	<driver_model>
		<current_dm>N/A</current_dm>
		<pending_dm>N/A</pending_dm>
	</driver_model>
	<serial>N/A</serial>
	<uuid>GPU-06e04616-0ed5-4069-5ebc-345349a0d4f3</uuid>
	<minor_number>0</minor_number>
	<vbios_version>86.07.67.00.14</vbios_version>
	<multigpu_board>No</multigpu_board>
	<board_id>0x100</board_id>
	<gpu_part_number>N/A</gpu_part_number>
	<gpu_module_id>0</gpu_module_id>
	<inforom_version>
		<img_version>N/A</img_version>
		<oem_object>N/A</oem_object>
		<ecc_object>N/A</ecc_object>
		<pwr_object>N/A</pwr_object>
	</inforom_version>
	<gpu_operation_mode>
		<current_gom>N/A</current_gom>
		<pending_gom>N/A</pending_gom>
	</gpu_operation_mode>
	<gsp_firmware_version>N/A</gsp_firmware_version>
	<gpu_virtualization_mode>
		<virtualization_mode>None</virtualization_mode>
		<host_vgpu_mode>N/A</host_vgpu_mode>
	</gpu_virtualization_mode>
	<ibmnpu>
		<relaxed_ordering_mode>N/A</relaxed_ordering_mode>
	</ibmnpu>
	<pci>
		<pci_bus>01</pci_bus>
		<pci_device>00</pci_device>
		<pci_domain>0000</pci_domain>
		<pci_device_id>1C8D10DE</pci_device_id>
		<pci_bus_id>00000000:01:00.0</pci_bus_id>
		<pci_sub_system_id>860A103C</pci_sub_system_id>
		<pci_gpu_link_info>
			<pcie_gen>
				<max_link_gen>3</max_link_gen>
				<current_link_gen>3</current_link_gen>
			</pcie_gen>
			<link_widths>
				<max_link_width>16x</max_link_width>
				<current_link_width>8x</current_link_width>
			</link_widths>
		</pci_gpu_link_info>
		<pci_bridge_chip>
			<bridge_chip_type>N/A</bridge_chip_type>
			<bridge_chip_fw>N/A</bridge_chip_fw>
		</pci_bridge_chip>
		<replay_counter>0</replay_counter>
		<replay_rollover_counter>0</replay_rollover_counter>
		<tx_util>50000 KB/s</tx_util>
		<rx_util>0 KB/s</rx_util>
	</pci>
	<fan_speed>N/A</fan_speed>
	<performance_state>P3</performance_state>
	<clocks_throttle_reasons>
		<clocks_throttle_reason_gpu_idle>Active</clocks_throttle_reason_gpu_idle>
		<clocks_throttle_reason_applications_clocks_setting>Not Active</clocks_throttle_reason_applications_clocks_setting>
		<clocks_throttle_reason_sw_power_cap>Not Active</clocks_throttle_reason_sw_power_cap>
		<clocks_throttle_reason_hw_slowdown>Not Active</clocks_throttle_reason_hw_slowdown>
		<clocks_throttle_reason_hw_thermal_slowdown>Not Active</clocks_throttle_reason_hw_thermal_slowdown>
		<clocks_throttle_reason_hw_power_brake_slowdown>Not Active</clocks_throttle_reason_hw_power_brake_slowdown>
		<clocks_throttle_reason_sync_boost>Not Active</clocks_throttle_reason_sync_boost>
		<clocks_throttle_reason_sw_thermal_slowdown>Not Active</clocks_throttle_reason_sw_thermal_slowdown>
		<clocks_throttle_reason_display_clocks_setting>Not Active</clocks_throttle_reason_display_clocks_setting>
	</clocks_throttle_reasons>
	<fb_memory_usage>
		<total>4096 MiB</total>
		<reserved>55 MiB</reserved>
		<used>65 MiB</used>
		<free>3974 MiB</free>
	</fb_memory_usage>
	<bar1_memory_usage>
		<total>256 MiB</total>
		<used>5 MiB</used>
		<free>251 MiB</free>
	</bar1_memory_usage>
	<compute_mode>Default</compute_mode>
	<utilization>
		<gpu_util>7 %</gpu_util>
		<memory_util>3 %</memory_util>
		<encoder_util>0 %</encoder_util>
		<decoder_util>0 %</decoder_util>
	</utilization>
	<encoder_stats>
		<session_count>0</session_count>
		<average_fps>0</average_fps>
		<average_latency>0</average_latency>
	</encoder_stats>
	<fbc_stats>
		<session_count>0</session_count>
		<average_fps>0</average_fps>
		<average_latency>0</average_latency>
	</fbc_stats>
	<ecc_mode>
		<current_ecc>N/A</current_ecc>
		<pending_ecc>N/A</pending_ecc>
	</ecc_mode>
	<ecc_errors>
		<volatile>
			<single_bit>
				<device_memory>N/A</device_memory>
				<register_file>N/A</register_file>
				<l1_cache>N/A</l1_cache>
				<l2_cache>N/A</l2_cache>
				<texture_memory>N/A</texture_memory>
				<texture_shm>N/A</texture_shm>
				<cbu>N/A</cbu>
				<total>N/A</total>
			</single_bit>
			<double_bit>
				<device_memory>N/A</device_memory>
				<register_file>N/A</register_file>
				<l1_cache>N/A</l1_cache>
				<l2_cache>N/A</l2_cache>
				<texture_memory>N/A</texture_memory>
				<texture_shm>N/A</texture_shm>
				<cbu>N/A</cbu>
				<total>N/A</total>
			</double_bit>
		</volatile>
		<aggregate>
			<single_bit>
				<device_memory>N/A</device_memory>
				<register_file>N/A</register_file>
				<l1_cache>N/A</l1_cache>
				<l2_cache>N/A</l2_cache>
				<texture_memory>N/A</texture_memory>
				<texture_shm>N/A</texture_shm>
				<cbu>N/A</cbu>
				<total>N/A</total>
			</single_bit>
			<double_bit>
				<device_memory>N/A</device_memory>
				<register_file>N/A</register_file>
				<l1_cache>N/A</l1_cache>
				<l2_cache>N/A</l2_cache>
				<texture_memory>N/A</texture_memory>
				<texture_shm>N/A</texture_shm>
				<cbu>N/A</cbu>
				<total>N/A</total>
			</double_bit>
		</aggregate>
	</ecc_errors>
	<retired_pages>
		<multiple_single_bit_retirement>
			<retired_count>N/A</retired_count>
			<retired_pagelist>N/A</retired_pagelist>
		</multiple_single_bit_retirement>
		<double_bit_retirement>
			<retired_count>N/A</retired_count>
			<retired_pagelist>N/A</retired_pagelist>
		</double_bit_retirement>
		<pending_blacklist>N/A</pending_blacklist>
		<pending_retirement>N/A</pending_retirement>
	</retired_pages>
	<remapped_rows>N/A</remapped_rows>
	<temperature>
		<gpu_temp>44 C</gpu_temp>
		<gpu_temp_max_threshold>102 C</gpu_temp_max_threshold>
		<gpu_temp_slow_threshold>97 C</gpu_temp_slow_threshold>
		<gpu_temp_max_gpu_threshold>94 C</gpu_temp_max_gpu_threshold>
		<gpu_target_temperature>N/A</gpu_target_temperature>
		<memory_temp>N/A</memory_temp>
		<gpu_temp_max_mem_threshold>N/A</gpu_temp_max_mem_threshold>
	</temperature>
	<supported_gpu_target_temp>
		<gpu_target_temp_min>N/A</gpu_target_temp_min>
		<gpu_target_temp_max>N/A</gpu_target_temp_max>
	</supported_gpu_target_temp>
	<power_readings>
		<power_state>P3</power_state>
		<power_management>N/A</power_management>
		<power_draw>N/A</power_draw>
		<power_limit>N/A</power_limit>
		<default_power_limit>N/A</default_power_limit>
		<enforced_power_limit>N/A</enforced_power_limit>
		<min_power_limit>N/A</min_power_limit>
		<max_power_limit>N/A</max_power_limit>
	</power_readings>
	<clocks>
		<graphics_clock>683 MHz</graphics_clock>
		<sm_clock>683 MHz</sm_clock>
		<mem_clock>2504 MHz</mem_clock>
		<video_clock>620 MHz</video_clock>
	</clocks>
	<applications_clocks>
		<graphics_clock>N/A</graphics_clock>
		<mem_clock>N/A</mem_clock>
	</applications_clocks>
	<default_applications_clocks>
		<graphics_clock>N/A</graphics_clock>
		<mem_clock>N/A</mem_clock>
	</default_applications_clocks>
	<max_clocks>
		<graphics_clock>1911 MHz</graphics_clock>
		<sm_clock>1911 MHz</sm_clock>
		<mem_clock>3504 MHz</mem_clock>
		<video_clock>1708 MHz</video_clock>
	</max_clocks>
	<max_customer_boost_clocks>
		<graphics_clock>N/A</graphics_clock>
	</max_customer_boost_clocks>
	<clock_policy>
		<auto_boost>N/A</auto_boost>
		<auto_boost_default>N/A</auto_boost_default>
	</clock_policy>
	<voltage>
		<graphics_volt>N/A</graphics_volt>
	</voltage>
	<supported_clocks>
		<supported_mem_clock>
			<value>3504 MHz</value>
			<supported_graphics_clock>1911 MHz</supported_graphics_clock>
			<supported_graphics_clock>1898 MHz</supported_graphics_clock>
			<supported_graphics_clock>1885 MHz</supported_graphics_clock>
			<supported_graphics_clock>1873 MHz</supported_graphics_clock>
			<supported_graphics_clock>1860 MHz</supported_graphics_clock>
			<supported_graphics_clock>1847 MHz</supported_graphics_clock>
			<supported_graphics_clock>1835 MHz</supported_graphics_clock>
			<supported_graphics_clock>1822 MHz</supported_graphics_clock>
			<supported_graphics_clock>1809 MHz</supported_graphics_clock>
			<supported_graphics_clock>1797 MHz</supported_graphics_clock>
			<supported_graphics_clock>1784 MHz</supported_graphics_clock>
			<supported_graphics_clock>1771 MHz</supported_graphics_clock>
			<supported_graphics_clock>1759 MHz</supported_graphics_clock>
			<supported_graphics_clock>1746 MHz</supported_graphics_clock>
			<supported_graphics_clock>1733 MHz</supported_graphics_clock>
			<supported_graphics_clock>1721 MHz</supported_graphics_clock>
			<supported_graphics_clock>1708 MHz</supported_graphics_clock>
			<supported_graphics_clock>1695 MHz</supported_graphics_clock>
			<supported_graphics_clock>1683 MHz</supported_graphics_clock>
			<supported_graphics_clock>1670 MHz</supported_graphics_clock>
			<supported_graphics_clock>1657 MHz</supported_graphics_clock>
			<supported_graphics_clock>1645 MHz</supported_graphics_clock>
			<supported_graphics_clock>1632 MHz</supported_graphics_clock>
			<supported_graphics_clock>1620 MHz</supported_graphics_clock>
			<supported_graphics_clock>1607 MHz</supported_graphics_clock>
			<supported_graphics_clock>1594 MHz</supported_graphics_clock>
			<supported_graphics_clock>1582 MHz</supported_graphics_clock>
			<supported_graphics_clock>1569 MHz</supported_graphics_clock>
			<supported_graphics_clock>1556 MHz</supported_graphics_clock>
			<supported_graphics_clock>1544 MHz</supported_graphics_clock>
			<supported_graphics_clock>1531 MHz</supported_graphics_clock>
			<supported_graphics_clock>1518 MHz</supported_graphics_clock>
			<supported_graphics_clock>1506 MHz</supported_graphics_clock>
			<supported_graphics_clock>1493 MHz</supported_graphics_clock>
			<supported_graphics_clock>1480 MHz</supported_graphics_clock>
			<supported_graphics_clock>1468 MHz</supported_graphics_clock>
			<supported_graphics_clock>1455 MHz</supported_graphics_clock>
			<supported_graphics_clock>1442 MHz</supported_graphics_clock>
			<supported_graphics_clock>1430 MHz</supported_graphics_clock>
			<supported_graphics_clock>1417 MHz</supported_graphics_clock>
			<supported_graphics_clock>1404 MHz</supported_graphics_clock>
			<supported_graphics_clock>1392 MHz</supported_graphics_clock>
			<supported_graphics_clock>1379 MHz</supported_graphics_clock>
			<supported_graphics_clock>1366 MHz</supported_graphics_clock>
			<supported_graphics_clock>1354 MHz</supported_graphics_clock>
			<supported_graphics_clock>1341 MHz</supported_graphics_clock>
			<supported_graphics_clock>1328 MHz</supported_graphics_clock>
			<supported_graphics_clock>1316 MHz</supported_graphics_clock>
			<supported_graphics_clock>1303 MHz</supported_graphics_clock>
			<supported_graphics_clock>1290 MHz</supported_graphics_clock>
			<supported_graphics_clock>1278 MHz</supported_graphics_clock>
			<supported_graphics_clock>1265 MHz</supported_graphics_clock>
			<supported_graphics_clock>1252 MHz</supported_graphics_clock>
			<supported_graphics_clock>1240 MHz</supported_graphics_clock>
			<supported_graphics_clock>1227 MHz</supported_graphics_clock>
			<supported_graphics_clock>1215 MHz</supported_graphics_clock>
			<supported_graphics_clock>1202 MHz</supported_graphics_clock>
			<supported_graphics_clock>1189 MHz</supported_graphics_clock>
			<supported_graphics_clock>1177 MHz</supported_graphics_clock>
			<supported_graphics_clock>1164 MHz</supported_graphics_clock>
			<supported_graphics_clock>1151 MHz</supported_graphics_clock>
			<supported_graphics_clock>1139 MHz</supported_graphics_clock>
			<supported_graphics_clock>1126 MHz</supported_graphics_clock>
			<supported_graphics_clock>1113 MHz</supported_graphics_clock>
			<supported_graphics_clock>1101 MHz</supported_graphics_clock>
			<supported_graphics_clock>1088 MHz</supported_graphics_clock>
			<supported_graphics_clock>1075 MHz</supported_graphics_clock>
			<supported_graphics_clock>1063 MHz</supported_graphics_clock>
			<supported_graphics_clock>1050 MHz</supported_graphics_clock>
			<supported_graphics_clock>1037 MHz</supported_graphics_clock>
			<supported_graphics_clock>1025 MHz</supported_graphics_clock>
			<supported_graphics_clock>1012 MHz</supported_graphics_clock>
			<supported_graphics_clock>999 MHz</supported_graphics_clock>
			<supported_graphics_clock>987 MHz</supported_graphics_clock>
			<supported_graphics_clock>974 MHz</supported_graphics_clock>
			<supported_graphics_clock>961 MHz</supported_graphics_clock>
			<supported_graphics_clock>949 MHz</supported_graphics_clock>
			<supported_graphics_clock>936 MHz</supported_graphics_clock>
			<supported_graphics_clock>923 MHz</supported_graphics_clock>
			<supported_graphics_clock>911 MHz</supported_graphics_clock>
			<supported_graphics_clock>898 MHz</supported_graphics_clock>
			<supported_graphics_clock>885 MHz</supported_graphics_clock>
			<supported_graphics_clock>873 MHz</supported_graphics_clock>
			<supported_graphics_clock>860 MHz</supported_graphics_clock>
			<supported_graphics_clock>847 MHz</supported_graphics_clock>
			<supported_graphics_clock>835 MHz</supported_graphics_clock>
			<supported_graphics_clock>822 MHz</supported_graphics_clock>
			<supported_graphics_clock>810 MHz</supported_graphics_clock>
			<supported_graphics_clock>797 MHz</supported_graphics_clock>
			<supported_graphics_clock>784 MHz</supported_graphics_clock>
			<supported_graphics_clock>772 MHz</supported_graphics_clock>
			<supported_graphics_clock>759 MHz</supported_graphics_clock>
			<supported_graphics_clock>746 MHz</supported_graphics_clock>
			<supported_graphics_clock>734 MHz</supported_graphics_clock>
			<supported_graphics_clock>721 MHz</supported_graphics_clock>
			<supported_graphics_clock>708 MHz</supported_graphics_clock>
			<supported_graphics_clock>696 MHz</supported_graphics_clock>
			<supported_graphics_clock>683 MHz</supported_graphics_clock>
			<supported_graphics_clock>670 MHz</supported_graphics_clock>
			<supported_graphics_clock>658 MHz</supported_graphics_clock>
			<supported_graphics_clock>645 MHz</supported_graphics_clock>
			<supported_graphics_clock>632 MHz</supported_graphics_clock>
			<supported_graphics_clock>620 MHz</supported_graphics_clock>
			<supported_graphics_clock>607 MHz</supported_graphics_clock>
			<supported_graphics_clock>594 MHz</supported_graphics_clock>
			<supported_graphics_clock>582 MHz</supported_graphics_clock>
			<supported_graphics_clock>569 MHz</supported_graphics_clock>
			<supported_graphics_clock>556 MHz</supported_graphics_clock>
			<supported_graphics_clock>544 MHz</supported_graphics_clock>
			<supported_graphics_clock>531 MHz</supported_graphics_clock>
			<supported_graphics_clock>518 MHz</supported_graphics_clock>
			<supported_graphics_clock>506 MHz</supported_graphics_clock>
			<supported_graphics_clock>493 MHz</supported_graphics_clock>
			<supported_graphics_clock>480 MHz</supported_graphics_clock>
			<supported_graphics_clock>468 MHz</supported_graphics_clock>
			<supported_graphics_clock>455 MHz</supported_graphics_clock>
			<supported_graphics_clock>442 MHz</supported_graphics_clock>
			<supported_graphics_clock>430 MHz</supported_graphics_clock>
			<supported_graphics_clock>417 MHz</supported_graphics_clock>
			<supported_graphics_clock>405 MHz</supported_graphics_clock>
			<supported_graphics_clock>392 MHz</supported_graphics_clock>
			<supported_graphics_clock>379 MHz</supported_graphics_clock>
			<supported_graphics_clock>367 MHz</supported_graphics_clock>
			<supported_graphics_clock>354 MHz</supported_graphics_clock>
			<supported_graphics_clock>341 MHz</supported_graphics_clock>
			<supported_graphics_clock>329 MHz</supported_graphics_clock>
			<supported_graphics_clock>316 MHz</supported_graphics_clock>
			<supported_graphics_clock>303 MHz</supported_graphics_clock>
			<supported_graphics_clock>291 MHz</supported_graphics_clock>
			<supported_graphics_clock>278 MHz</supported_graphics_clock>
			<supported_graphics_clock>265 MHz</supported_graphics_clock>
			<supported_graphics_clock>253 MHz</supported_graphics_clock>
			<supported_graphics_clock>240 MHz</supported_graphics_clock>
			<supported_graphics_clock>227 MHz</supported_graphics_clock>
			<supported_graphics_clock>215 MHz</supported_graphics_clock>
			<supported_graphics_clock>202 MHz</supported_graphics_clock>
			<supported_graphics_clock>189 MHz</supported_graphics_clock>
			<supported_graphics_clock>177 MHz</supported_graphics_clock>
			<supported_graphics_clock>164 MHz</supported_graphics_clock>
			<supported_graphics_clock>151 MHz</supported_graphics_clock>
			<supported_graphics_clock>139 MHz</supported_graphics_clock>
		</supported_mem_clock>
		<supported_mem_clock>
			<value>2505 MHz</value>
			<supported_graphics_clock>1911 MHz</supported_graphics_clock>
			<supported_graphics_clock>1898 MHz</supported_graphics_clock>
			<supported_graphics_clock>1885 MHz</supported_graphics_clock>
			<supported_graphics_clock>1873 MHz</supported_graphics_clock>
			<supported_graphics_clock>1860 MHz</supported_graphics_clock>
			<supported_graphics_clock>1847 MHz</supported_graphics_clock>
			<supported_graphics_clock>1835 MHz</supported_graphics_clock>
			<supported_graphics_clock>1822 MHz</supported_graphics_clock>
			<supported_graphics_clock>1809 MHz</supported_graphics_clock>
			<supported_graphics_clock>1797 MHz</supported_graphics_clock>
			<supported_graphics_clock>1784 MHz</supported_graphics_clock>
			<supported_graphics_clock>1771 MHz</supported_graphics_clock>
			<supported_graphics_clock>1759 MHz</supported_graphics_clock>
			<supported_graphics_clock>1746 MHz</supported_graphics_clock>
			<supported_graphics_clock>1733 MHz</supported_graphics_clock>
			<supported_graphics_clock>1721 MHz</supported_graphics_clock>
			<supported_graphics_clock>1708 MHz</supported_graphics_clock>
			<supported_graphics_clock>1695 MHz</supported_graphics_clock>
			<supported_graphics_clock>1683 MHz</supported_graphics_clock>
			<supported_graphics_clock>1670 MHz</supported_graphics_clock>
			<supported_graphics_clock>1657 MHz</supported_graphics_clock>
			<supported_graphics_clock>1645 MHz</supported_graphics_clock>
			<supported_graphics_clock>1632 MHz</supported_graphics_clock>
			<supported_graphics_clock>1620 MHz</supported_graphics_clock>
			<supported_graphics_clock>1607 MHz</supported_graphics_clock>
			<supported_graphics_clock>1594 MHz</supported_graphics_clock>
			<supported_graphics_clock>1582 MHz</supported_graphics_clock>
			<supported_graphics_clock>1569 MHz</supported_graphics_clock>
			<supported_graphics_clock>1556 MHz</supported_graphics_clock>
			<supported_graphics_clock>1544 MHz</supported_graphics_clock>
			<supported_graphics_clock>1531 MHz</supported_graphics_clock>
			<supported_graphics_clock>1518 MHz</supported_graphics_clock>
			<supported_graphics_clock>1506 MHz</supported_graphics_clock>
			<supported_graphics_clock>1493 MHz</supported_graphics_clock>
			<supported_graphics_clock>1480 MHz</supported_graphics_clock>
			<supported_graphics_clock>1468 MHz</supported_graphics_clock>
			<supported_graphics_clock>1455 MHz</supported_graphics_clock>
			<supported_graphics_clock>1442 MHz</supported_graphics_clock>
			<supported_graphics_clock>1430 MHz</supported_graphics_clock>
			<supported_graphics_clock>1417 MHz</supported_graphics_clock>
			<supported_graphics_clock>1404 MHz</supported_graphics_clock>
			<supported_graphics_clock>1392 MHz</supported_graphics_clock>
			<supported_graphics_clock>1379 MHz</supported_graphics_clock>
			<supported_graphics_clock>1366 MHz</supported_graphics_clock>
			<supported_graphics_clock>1354 MHz</supported_graphics_clock>
			<supported_graphics_clock>1341 MHz</supported_graphics_clock>
			<supported_graphics_clock>1328 MHz</supported_graphics_clock>
			<supported_graphics_clock>1316 MHz</supported_graphics_clock>
			<supported_graphics_clock>1303 MHz</supported_graphics_clock>
			<supported_graphics_clock>1290 MHz</supported_graphics_clock>
			<supported_graphics_clock>1278 MHz</supported_graphics_clock>
			<supported_graphics_clock>1265 MHz</supported_graphics_clock>
			<supported_graphics_clock>1252 MHz</supported_graphics_clock>
			<supported_graphics_clock>1240 MHz</supported_graphics_clock>
			<supported_graphics_clock>1227 MHz</supported_graphics_clock>
			<supported_graphics_clock>1215 MHz</supported_graphics_clock>
			<supported_graphics_clock>1202 MHz</supported_graphics_clock>
			<supported_graphics_clock>1189 MHz</supported_graphics_clock>
			<supported_graphics_clock>1177 MHz</supported_graphics_clock>
			<supported_graphics_clock>1164 MHz</supported_graphics_clock>
			<supported_graphics_clock>1151 MHz</supported_graphics_clock>
			<supported_graphics_clock>1139 MHz</supported_graphics_clock>
			<supported_graphics_clock>1126 MHz</supported_graphics_clock>
			<supported_graphics_clock>1113 MHz</supported_graphics_clock>
			<supported_graphics_clock>1101 MHz</supported_graphics_clock>
			<supported_graphics_clock>1088 MHz</supported_graphics_clock>
			<supported_graphics_clock>1075 MHz</supported_graphics_clock>
			<supported_graphics_clock>1063 MHz</supported_graphics_clock>
			<supported_graphics_clock>1050 MHz</supported_graphics_clock>
			<supported_graphics_clock>1037 MHz</supported_graphics_clock>
			<supported_graphics_clock>1025 MHz</supported_graphics_clock>
			<supported_graphics_clock>1012 MHz</supported_graphics_clock>
			<supported_graphics_clock>999 MHz</supported_graphics_clock>
			<supported_graphics_clock>987 MHz</supported_graphics_clock>
			<supported_graphics_clock>974 MHz</supported_graphics_clock>
			<supported_graphics_clock>961 MHz</supported_graphics_clock>
			<supported_graphics_clock>949 MHz</supported_graphics_clock>
			<supported_graphics_clock>936 MHz</supported_graphics_clock>
			<supported_graphics_clock>923 MHz</supported_graphics_clock>
			<supported_graphics_clock>911 MHz</supported_graphics_clock>
			<supported_graphics_clock>898 MHz</supported_graphics_clock>
			<supported_graphics_clock>885 MHz</supported_graphics_clock>
			<supported_graphics_clock>873 MHz</supported_graphics_clock>
			<supported_graphics_clock>860 MHz</supported_graphics_clock>
			<supported_graphics_clock>847 MHz</supported_graphics_clock>
			<supported_graphics_clock>835 MHz</supported_graphics_clock>
			<supported_graphics_clock>822 MHz</supported_graphics_clock>
			<supported_graphics_clock>810 MHz</supported_graphics_clock>
			<supported_graphics_clock>797 MHz</supported_graphics_clock>
			<supported_graphics_clock>784 MHz</supported_graphics_clock>
			<supported_graphics_clock>772 MHz</supported_graphics_clock>
			<supported_graphics_clock>759 MHz</supported_graphics_clock>
			<supported_graphics_clock>746 MHz</supported_graphics_clock>
			<supported_graphics_clock>734 MHz</supported_graphics_clock>
			<supported_graphics_clock>721 MHz</supported_graphics_clock>
			<supported_graphics_clock>708 MHz</supported_graphics_clock>
			<supported_graphics_clock>696 MHz</supported_graphics_clock>
			<supported_graphics_clock>683 MHz</supported_graphics_clock>
			<supported_graphics_clock>670 MHz</supported_graphics_clock>
			<supported_graphics_clock>658 MHz</supported_graphics_clock>
			<supported_graphics_clock>645 MHz</supported_graphics_clock>
			<supported_graphics_clock>632 MHz</supported_graphics_clock>
			<supported_graphics_clock>620 MHz</supported_graphics_clock>
			<supported_graphics_clock>607 MHz</supported_graphics_clock>
			<supported_graphics_clock>594 MHz</supported_graphics_clock>
			<supported_graphics_clock>582 MHz</supported_graphics_clock>
			<supported_graphics_clock>569 MHz</supported_graphics_clock>
			<supported_graphics_clock>556 MHz</supported_graphics_clock>
			<supported_graphics_clock>544 MHz</supported_graphics_clock>
			<supported_graphics_clock>531 MHz</supported_graphics_clock>
			<supported_graphics_clock>518 MHz</supported_graphics_clock>
			<supported_graphics_clock>506 MHz</supported_graphics_clock>
			<supported_graphics_clock>493 MHz</supported_graphics_clock>
			<supported_graphics_clock>480 MHz</supported_graphics_clock>
			<supported_graphics_clock>468 MHz</supported_graphics_clock>
			<supported_graphics_clock>455 MHz</supported_graphics_clock>
			<supported_graphics_clock>442 MHz</supported_graphics_clock>
			<supported_graphics_clock>430 MHz</supported_graphics_clock>
			<supported_graphics_clock>417 MHz</supported_graphics_clock>
			<supported_graphics_clock>405 MHz</supported_graphics_clock>
			<supported_graphics_clock>392 MHz</supported_graphics_clock>
			<supported_graphics_clock>379 MHz</supported_graphics_clock>
			<supported_graphics_clock>367 MHz</supported_graphics_clock>
			<supported_graphics_clock>354 MHz</supported_graphics_clock>
			<supported_graphics_clock>341 MHz</supported_graphics_clock>
			<supported_graphics_clock>329 MHz</supported_graphics_clock>
			<supported_graphics_clock>316 MHz</supported_graphics_clock>
			<supported_graphics_clock>303 MHz</supported_graphics_clock>
			<supported_graphics_clock>291 MHz</supported_graphics_clock>
			<supported_graphics_clock>278 MHz</supported_graphics_clock>
			<supported_graphics_clock>265 MHz</supported_graphics_clock>
			<supported_graphics_clock>253 MHz</supported_graphics_clock>
			<supported_graphics_clock>240 MHz</supported_graphics_clock>
			<supported_graphics_clock>227 MHz</supported_graphics_clock>
			<supported_graphics_clock>215 MHz</supported_graphics_clock>
			<supported_graphics_clock>202 MHz</supported_graphics_clock>
			<supported_graphics_clock>189 MHz</supported_graphics_clock>
			<supported_graphics_clock>177 MHz</supported_graphics_clock>
			<supported_graphics_clock>164 MHz</supported_graphics_clock>
			<supported_graphics_clock>151 MHz</supported_graphics_clock>
			<supported_graphics_clock>139 MHz</supported_graphics_clock>
		</supported_mem_clock>
		<supported_mem_clock>
			<value>810 MHz</value>
			<supported_graphics_clock>1911 MHz</supported_graphics_clock>
			<supported_graphics_clock>1898 MHz</supported_graphics_clock>
			<supported_graphics_clock>1885 MHz</supported_graphics_clock>
			<supported_graphics_clock>1873 MHz</supported_graphics_clock>
			<supported_graphics_clock>1860 MHz</supported_graphics_clock>
			<supported_graphics_clock>1847 MHz</supported_graphics_clock>
			<supported_graphics_clock>1835 MHz</supported_graphics_clock>
			<supported_graphics_clock>1822 MHz</supported_graphics_clock>
			<supported_graphics_clock>1809 MHz</supported_graphics_clock>
			<supported_graphics_clock>1797 MHz</supported_graphics_clock>
			<supported_graphics_clock>1784 MHz</supported_graphics_clock>
			<supported_graphics_clock>1771 MHz</supported_graphics_clock>
			<supported_graphics_clock>1759 MHz</supported_graphics_clock>
			<supported_graphics_clock>1746 MHz</supported_graphics_clock>
			<supported_graphics_clock>1733 MHz</supported_graphics_clock>
			<supported_graphics_clock>1721 MHz</supported_graphics_clock>
			<supported_graphics_clock>1708 MHz</supported_graphics_clock>
			<supported_graphics_clock>1695 MHz</supported_graphics_clock>
			<supported_graphics_clock>1683 MHz</supported_graphics_clock>
			<supported_graphics_clock>1670 MHz</supported_graphics_clock>
			<supported_graphics_clock>1657 MHz</supported_graphics_clock>
			<supported_graphics_clock>1645 MHz</supported_graphics_clock>
			<supported_graphics_clock>1632 MHz</supported_graphics_clock>
			<supported_graphics_clock>1620 MHz</supported_graphics_clock>
			<supported_graphics_clock>1607 MHz</supported_graphics_clock>
			<supported_graphics_clock>1594 MHz</supported_graphics_clock>
			<supported_graphics_clock>1582 MHz</supported_graphics_clock>
			<supported_graphics_clock>1569 MHz</supported_graphics_clock>
			<supported_graphics_clock>1556 MHz</supported_graphics_clock>
			<supported_graphics_clock>1544 MHz</supported_graphics_clock>
			<supported_graphics_clock>1531 MHz</supported_graphics_clock>
			<supported_graphics_clock>1518 MHz</supported_graphics_clock>
			<supported_graphics_clock>1506 MHz</supported_graphics_clock>
			<supported_graphics_clock>1493 MHz</supported_graphics_clock>
			<supported_graphics_clock>1480 MHz</supported_graphics_clock>
			<supported_graphics_clock>1468 MHz</supported_graphics_clock>
			<supported_graphics_clock>1455 MHz</supported_graphics_clock>
			<supported_graphics_clock>1442 MHz</supported_graphics_clock>
			<supported_graphics_clock>1430 MHz</supported_graphics_clock>
			<supported_graphics_clock>1417 MHz</supported_graphics_clock>
			<supported_graphics_clock>1404 MHz</supported_graphics_clock>
			<supported_graphics_clock>1392 MHz</supported_graphics_clock>
			<supported_graphics_clock>1379 MHz</supported_graphics_clock>
			<supported_graphics_clock>1366 MHz</supported_graphics_clock>
			<supported_graphics_clock>1354 MHz</supported_graphics_clock>
			<supported_graphics_clock>1341 MHz</supported_graphics_clock>
			<supported_graphics_clock>1328 MHz</supported_graphics_clock>
			<supported_graphics_clock>1316 MHz</supported_graphics_clock>
			<supported_graphics_clock>1303 MHz</supported_graphics_clock>
			<supported_graphics_clock>1290 MHz</supported_graphics_clock>
			<supported_graphics_clock>1278 MHz</supported_graphics_clock>
			<supported_graphics_clock>1265 MHz</supported_graphics_clock>
			<supported_graphics_clock>1252 MHz</supported_graphics_clock>
			<supported_graphics_clock>1240 MHz</supported_graphics_clock>
			<supported_graphics_clock>1227 MHz</supported_graphics_clock>
			<supported_graphics_clock>1215 MHz</supported_graphics_clock>
			<supported_graphics_clock>1202 MHz</supported_graphics_clock>
			<supported_graphics_clock>1189 MHz</supported_graphics_clock>
			<supported_graphics_clock>1177 MHz</supported_graphics_clock>
			<supported_graphics_clock>1164 MHz</supported_graphics_clock>
			<supported_graphics_clock>1151 MHz</supported_graphics_clock>
			<supported_graphics_clock>1139 MHz</supported_graphics_clock>
			<supported_graphics_clock>1126 MHz</supported_graphics_clock>
			<supported_graphics_clock>1113 MHz</supported_graphics_clock>
			<supported_graphics_clock>1101 MHz</supported_graphics_clock>
			<supported_graphics_clock>1088 MHz</supported_graphics_clock>
			<supported_graphics_clock>1075 MHz</supported_graphics_clock>
			<supported_graphics_clock>1063 MHz</supported_graphics_clock>
			<supported_graphics_clock>1050 MHz</supported_graphics_clock>
			<supported_graphics_clock>1037 MHz</supported_graphics_clock>
			<supported_graphics_clock>1025 MHz</supported_graphics_clock>
			<supported_graphics_clock>1012 MHz</supported_graphics_clock>
			<supported_graphics_clock>999 MHz</supported_graphics_clock>
			<supported_graphics_clock>987 MHz</supported_graphics_clock>
			<supported_graphics_clock>974 MHz</supported_graphics_clock>
			<supported_graphics_clock>961 MHz</supported_graphics_clock>
			<supported_graphics_clock>949 MHz</supported_graphics_clock>
			<supported_graphics_clock>936 MHz</supported_graphics_clock>
			<supported_graphics_clock>923 MHz</supported_graphics_clock>
			<supported_graphics_clock>911 MHz</supported_graphics_clock>
			<supported_graphics_clock>898 MHz</supported_graphics_clock>
			<supported_graphics_clock>885 MHz</supported_graphics_clock>
			<supported_graphics_clock>873 MHz</supported_graphics_clock>
			<supported_graphics_clock>860 MHz</supported_graphics_clock>
			<supported_graphics_clock>847 MHz</supported_graphics_clock>
			<supported_graphics_clock>835 MHz</supported_graphics_clock>
			<supported_graphics_clock>822 MHz</supported_graphics_clock>
			<supported_graphics_clock>810 MHz</supported_graphics_clock>
			<supported_graphics_clock>797 MHz</supported_graphics_clock>
			<supported_graphics_clock>784 MHz</supported_graphics_clock>
			<supported_graphics_clock>772 MHz</supported_graphics_clock>
			<supported_graphics_clock>759 MHz</supported_graphics_clock>
			<supported_graphics_clock>746 MHz</supported_graphics_clock>
			<supported_graphics_clock>734 MHz</supported_graphics_clock>
			<supported_graphics_clock>721 MHz</supported_graphics_clock>
			<supported_graphics_clock>708 MHz</supported_graphics_clock>
			<supported_graphics_clock>696 MHz</supported_graphics_clock>
			<supported_graphics_clock>683 MHz</supported_graphics_clock>
			<supported_graphics_clock>670 MHz</supported_graphics_clock>
			<supported_graphics_clock>658 MHz</supported_graphics_clock>
			<supported_graphics_clock>645 MHz</supported_graphics_clock>
			<supported_graphics_clock>632 MHz</supported_graphics_clock>
			<supported_graphics_clock>620 MHz</supported_graphics_clock>
			<supported_graphics_clock>607 MHz</supported_graphics_clock>
			<supported_graphics_clock>594 MHz</supported_graphics_clock>
			<supported_graphics_clock>582 MHz</supported_graphics_clock>
			<supported_graphics_clock>569 MHz</supported_graphics_clock>
			<supported_graphics_clock>556 MHz</supported_graphics_clock>
			<supported_graphics_clock>544 MHz</supported_graphics_clock>
			<supported_graphics_clock>531 MHz</supported_graphics_clock>
			<supported_graphics_clock>518 MHz</supported_graphics_clock>
			<supported_graphics_clock>506 MHz</supported_graphics_clock>
			<supported_graphics_clock>493 MHz</supported_graphics_clock>
			<supported_graphics_clock>480 MHz</supported_graphics_clock>
			<supported_graphics_clock>468 MHz</supported_graphics_clock>
			<supported_graphics_clock>455 MHz</supported_graphics_clock>
			<supported_graphics_clock>442 MHz</supported_graphics_clock>
			<supported_graphics_clock>430 MHz</supported_graphics_clock>
			<supported_graphics_clock>417 MHz</supported_graphics_clock>
			<supported_graphics_clock>405 MHz</supported_graphics_clock>
			<supported_graphics_clock>392 MHz</supported_graphics_clock>
			<supported_graphics_clock>379 MHz</supported_graphics_clock>
			<supported_graphics_clock>367 MHz</supported_graphics_clock>
			<supported_graphics_clock>354 MHz</supported_graphics_clock>
			<supported_graphics_clock>341 MHz</supported_graphics_clock>
			<supported_graphics_clock>329 MHz</supported_graphics_clock>
			<supported_graphics_clock>316 MHz</supported_graphics_clock>
			<supported_graphics_clock>303 MHz</supported_graphics_clock>
			<supported_graphics_clock>291 MHz</supported_graphics_clock>
			<supported_graphics_clock>278 MHz</supported_graphics_clock>
			<supported_graphics_clock>265 MHz</supported_graphics_clock>
			<supported_graphics_clock>253 MHz</supported_graphics_clock>
			<supported_graphics_clock>240 MHz</supported_graphics_clock>
			<supported_graphics_clock>227 MHz</supported_graphics_clock>
			<supported_graphics_clock>215 MHz</supported_graphics_clock>
			<supported_graphics_clock>202 MHz</supported_graphics_clock>
			<supported_graphics_clock>189 MHz</supported_graphics_clock>
			<supported_graphics_clock>177 MHz</supported_graphics_clock>
			<supported_graphics_clock>164 MHz</supported_graphics_clock>
			<supported_graphics_clock>151 MHz</supported_graphics_clock>
			<supported_graphics_clock>139 MHz</supported_graphics_clock>
		</supported_mem_clock>
		<supported_mem_clock>
			<value>405 MHz</value>
			<supported_graphics_clock>607 MHz</supported_graphics_clock>
			<supported_graphics_clock>594 MHz</supported_graphics_clock>
			<supported_graphics_clock>582 MHz</supported_graphics_clock>
			<supported_graphics_clock>569 MHz</supported_graphics_clock>
			<supported_graphics_clock>556 MHz</supported_graphics_clock>
			<supported_graphics_clock>544 MHz</supported_graphics_clock>
			<supported_graphics_clock>531 MHz</supported_graphics_clock>
			<supported_graphics_clock>518 MHz</supported_graphics_clock>
			<supported_graphics_clock>506 MHz</supported_graphics_clock>
			<supported_graphics_clock>493 MHz</supported_graphics_clock>
			<supported_graphics_clock>480 MHz</supported_graphics_clock>
			<supported_graphics_clock>468 MHz</supported_graphics_clock>
			<supported_graphics_clock>455 MHz</supported_graphics_clock>
			<supported_graphics_clock>442 MHz</supported_graphics_clock>
			<supported_graphics_clock>430 MHz</supported_graphics_clock>
			<supported_graphics_clock>417 MHz</supported_graphics_clock>
			<supported_graphics_clock>405 MHz</supported_graphics_clock>
			<supported_graphics_clock>392 MHz</supported_graphics_clock>
			<supported_graphics_clock>379 MHz</supported_graphics_clock>
			<supported_graphics_clock>367 MHz</supported_graphics_clock>
			<supported_graphics_clock>354 MHz</supported_graphics_clock>
			<supported_graphics_clock>341 MHz</supported_graphics_clock>
			<supported_graphics_clock>329 MHz</supported_graphics_clock>
			<supported_graphics_clock>316 MHz</supported_graphics_clock>
			<supported_graphics_clock>303 MHz</supported_graphics_clock>
			<supported_graphics_clock>291 MHz</supported_graphics_clock>
			<supported_graphics_clock>278 MHz</supported_graphics_clock>
			<supported_graphics_clock>265 MHz</supported_graphics_clock>
			<supported_graphics_clock>253 MHz</supported_graphics_clock>
			<supported_graphics_clock>240 MHz</supported_graphics_clock>
			<supported_graphics_clock>227 MHz</supported_graphics_clock>
			<supported_graphics_clock>215 MHz</supported_graphics_clock>
			<supported_graphics_clock>202 MHz</supported_graphics_clock>
			<supported_graphics_clock>189 MHz</supported_graphics_clock>
			<supported_graphics_clock>177 MHz</supported_graphics_clock>
			<supported_graphics_clock>164 MHz</supported_graphics_clock>
			<supported_graphics_clock>151 MHz</supported_graphics_clock>
			<supported_graphics_clock>139 MHz</supported_graphics_clock>
		</supported_mem_clock>
	</supported_clocks>
	<processes>
		<process_info>
			<gpu_instance_id>N/A</gpu_instance_id>
			<compute_instance_id>N/A</compute_instance_id>
			<pid>1200</pid>
			<type>G</type>
			<process_name>/usr/lib/xorg/Xorg</process_name>
			<used_memory>61 MiB</used_memory>
		</process_info>
		<process_info>
			<gpu_instance_id>N/A</gpu_instance_id>
			<compute_instance_id>N/A</compute_instance_id>
			<pid>30091</pid>
			<type>G</type>
			<process_name>/snap/firefox/1775/usr/lib/firefox/firefox</process_name>
			<used_memory>1 MiB</used_memory>
		</process_info>
	</processes>
	<accounted_processes>
	</accounted_processes>
</gpu>

</nvidia_smi_log>
`)

var gtx8card []byte = []byte(`<?xml version="1.0" ?>
<!DOCTYPE nvidia_smi_log SYSTEM "nvsmi_device_v11.dtd">
<nvidia_smi_log>
    <timestamp>Thu Sep 15 13:55:33 2022</timestamp>
    <driver_version>460.73.01</driver_version>
    <cuda_version>11.2</cuda_version>
    <attached_gpus>8</attached_gpus>
    <gpu id="00000000:00:0D.0">
        <product_name>Tesla V100S-PCIE-32GB</product_name>
        <product_brand>Tesla</product_brand>
        <display_mode>Enabled</display_mode>
        <display_active>Disabled</display_active>
        <persistence_mode>Disabled</persistence_mode>
        <mig_mode>
            <current_mig>N/A</current_mig>
            <pending_mig>N/A</pending_mig>
        </mig_mode>
        <mig_devices>
            None
        </mig_devices>
        <accounting_mode>Enabled</accounting_mode>
        <accounting_mode_buffer_size>4000</accounting_mode_buffer_size>
        <driver_model>
            <current_dm>N/A</current_dm>
            <pending_dm>N/A</pending_dm>
        </driver_model>
        <serial>1562920006088</serial>
        <uuid>GPU-76098dc1-d3bc-9a1b-7ea6-f7b48837829e</uuid>
        <minor_number>0</minor_number>
        <vbios_version>88.00.98.00.01</vbios_version>
        <multigpu_board>No</multigpu_board>
        <board_id>0xd</board_id>
        <gpu_part_number>900-2G500-0040-000</gpu_part_number>
        <inforom_version>
            <img_version>G500.0212.00.02</img_version>
            <oem_object>1.1</oem_object>
            <ecc_object>5.0</ecc_object>
            <pwr_object>N/A</pwr_object>
        </inforom_version>
        <gpu_operation_mode>
            <current_gom>N/A</current_gom>
            <pending_gom>N/A</pending_gom>
        </gpu_operation_mode>
        <gpu_virtualization_mode>
            <virtualization_mode>Pass-Through</virtualization_mode>
            <host_vgpu_mode>N/A</host_vgpu_mode>
        </gpu_virtualization_mode>
        <ibmnpu>
            <relaxed_ordering_mode>N/A</relaxed_ordering_mode>
        </ibmnpu>
        <pci>
            <pci_bus>00</pci_bus>
            <pci_device>0D</pci_device>
            <pci_domain>0000</pci_domain>
            <pci_device_id>1DF610DE</pci_device_id>
            <pci_bus_id>00000000:00:0D.0</pci_bus_id>
            <pci_sub_system_id>13D610DE</pci_sub_system_id>
            <pci_gpu_link_info>
                <pcie_gen>
                    <max_link_gen>3</max_link_gen>
                    <current_link_gen>3</current_link_gen>
                </pcie_gen>
                <link_widths>
                    <max_link_width>16x</max_link_width>
                    <current_link_width>16x</current_link_width>
                </link_widths>
            </pci_gpu_link_info>
            <pci_bridge_chip>
                <bridge_chip_type>N/A</bridge_chip_type>
                <bridge_chip_fw>N/A</bridge_chip_fw>
            </pci_bridge_chip>
            <replay_counter>0</replay_counter>
            <replay_rollover_counter>0</replay_rollover_counter>
            <tx_util>22000 KB/s</tx_util>
            <rx_util>19000 KB/s</rx_util>
        </pci>
        <fan_speed>N/A</fan_speed>
        <performance_state>P0</performance_state>
        <clocks_throttle_reasons>
            <clocks_throttle_reason_gpu_idle>Not Active</clocks_throttle_reason_gpu_idle>
            <clocks_throttle_reason_applications_clocks_setting>Not Active</clocks_throttle_reason_applications_clocks_setting>
            <clocks_throttle_reason_sw_power_cap>Not Active</clocks_throttle_reason_sw_power_cap>
            <clocks_throttle_reason_hw_slowdown>Not Active</clocks_throttle_reason_hw_slowdown>
            <clocks_throttle_reason_hw_thermal_slowdown>Not Active</clocks_throttle_reason_hw_thermal_slowdown>
            <clocks_throttle_reason_hw_power_brake_slowdown>Not Active</clocks_throttle_reason_hw_power_brake_slowdown>
            <clocks_throttle_reason_sync_boost>Not Active</clocks_throttle_reason_sync_boost>
            <clocks_throttle_reason_sw_thermal_slowdown>Not Active</clocks_throttle_reason_sw_thermal_slowdown>
            <clocks_throttle_reason_display_clocks_setting>Not Active</clocks_throttle_reason_display_clocks_setting>
        </clocks_throttle_reasons>
        <fb_memory_usage>
            <total>32510 MiB</total>
            <used>31689 MiB</used>
            <free>821 MiB</free>
        </fb_memory_usage>
        <bar1_memory_usage>
            <total>32768 MiB</total>
            <used>32 MiB</used>
            <free>32736 MiB</free>
        </bar1_memory_usage>
        <compute_mode>Default</compute_mode>
        <utilization>
            <gpu_util>97 %</gpu_util>
            <memory_util>52 %</memory_util>
            <encoder_util>0 %</encoder_util>
            <decoder_util>0 %</decoder_util>
        </utilization>
        <encoder_stats>
            <session_count>0</session_count>
            <average_fps>0</average_fps>
            <average_latency>0</average_latency>
        </encoder_stats>
        <fbc_stats>
            <session_count>0</session_count>
            <average_fps>0</average_fps>
            <average_latency>0</average_latency>
        </fbc_stats>
        <ecc_mode>
            <current_ecc>Enabled</current_ecc>
            <pending_ecc>Enabled</pending_ecc>
        </ecc_mode>
        <ecc_errors>
            <volatile>
                <single_bit>
                    <device_memory>0</device_memory>
                    <register_file>0</register_file>
                    <l1_cache>0</l1_cache>
                    <l2_cache>0</l2_cache>
                    <texture_memory>N/A</texture_memory>
                    <texture_shm>N/A</texture_shm>
                    <cbu>N/A</cbu>
                    <total>0</total>
                </single_bit>
                <double_bit>
                    <device_memory>0</device_memory>
                    <register_file>0</register_file>
                    <l1_cache>0</l1_cache>
                    <l2_cache>0</l2_cache>
                    <texture_memory>N/A</texture_memory>
                    <texture_shm>N/A</texture_shm>
                    <cbu>0</cbu>
                    <total>0</total>
                </double_bit>
            </volatile>
            <aggregate>
                <single_bit>
                    <device_memory>0</device_memory>
                    <register_file>0</register_file>
                    <l1_cache>0</l1_cache>
                    <l2_cache>0</l2_cache>
                    <texture_memory>N/A</texture_memory>
                    <texture_shm>N/A</texture_shm>
                    <cbu>N/A</cbu>
                    <total>0</total>
                </single_bit>
                <double_bit>
                    <device_memory>0</device_memory>
                    <register_file>0</register_file>
                    <l1_cache>0</l1_cache>
                    <l2_cache>0</l2_cache>
                    <texture_memory>N/A</texture_memory>
                    <texture_shm>N/A</texture_shm>
                    <cbu>0</cbu>
                    <total>0</total>
                </double_bit>
            </aggregate>
        </ecc_errors>
        <retired_pages>
            <multiple_single_bit_retirement>
                <retired_count>0</retired_count>
                <retired_pagelist>
                </retired_pagelist>
            </multiple_single_bit_retirement>
            <double_bit_retirement>
                <retired_count>0</retired_count>
                <retired_pagelist>
                </retired_pagelist>
            </double_bit_retirement>
            <pending_blacklist>No</pending_blacklist>
            <pending_retirement>No</pending_retirement>
        </retired_pages>
        <remapped_rows>N/A</remapped_rows>
        <temperature>
            <gpu_temp>66 C</gpu_temp>
            <gpu_temp_max_threshold>90 C</gpu_temp_max_threshold>
            <gpu_temp_slow_threshold>87 C</gpu_temp_slow_threshold>
            <gpu_temp_max_gpu_threshold>83 C</gpu_temp_max_gpu_threshold>
            <gpu_target_temperature>N/A</gpu_target_temperature>
            <memory_temp>63 C</memory_temp>
            <gpu_temp_max_mem_threshold>85 C</gpu_temp_max_mem_threshold>
        </temperature>
        <supported_gpu_target_temp>
            <gpu_target_temp_min>N/A</gpu_target_temp_min>
            <gpu_target_temp_max>N/A</gpu_target_temp_max>
        </supported_gpu_target_temp>
        <power_readings>
            <power_state>P0</power_state>
            <power_management>Supported</power_management>
            <power_draw>232.80 W</power_draw>
            <power_limit>250.00 W</power_limit>
            <default_power_limit>250.00 W</default_power_limit>
            <enforced_power_limit>250.00 W</enforced_power_limit>
            <min_power_limit>100.00 W</min_power_limit>
            <max_power_limit>250.00 W</max_power_limit>
        </power_readings>
        <clocks>
            <graphics_clock>1410 MHz</graphics_clock>
            <sm_clock>1410 MHz</sm_clock>
            <mem_clock>1107 MHz</mem_clock>
            <video_clock>1275 MHz</video_clock>
        </clocks>
        <applications_clocks>
            <graphics_clock>1245 MHz</graphics_clock>
            <mem_clock>1107 MHz</mem_clock>
        </applications_clocks>
        <default_applications_clocks>
            <graphics_clock>1245 MHz</graphics_clock>
            <mem_clock>1107 MHz</mem_clock>
        </default_applications_clocks>
        <max_clocks>
            <graphics_clock>1597 MHz</graphics_clock>
            <sm_clock>1597 MHz</sm_clock>
            <mem_clock>1107 MHz</mem_clock>
            <video_clock>1432 MHz</video_clock>
        </max_clocks>
        <max_customer_boost_clocks>
            <graphics_clock>1597 MHz</graphics_clock>
        </max_customer_boost_clocks>
        <clock_policy>
            <auto_boost>N/A</auto_boost>
            <auto_boost_default>N/A</auto_boost_default>
        </clock_policy>
        <supported_clocks>
            <supported_mem_clock>
                <value>1107 MHz</value>
                <supported_graphics_clock>1597 MHz</supported_graphics_clock>
                <supported_graphics_clock>1590 MHz</supported_graphics_clock>
                <supported_graphics_clock>1582 MHz</supported_graphics_clock>
                <supported_graphics_clock>1575 MHz</supported_graphics_clock>
                <supported_graphics_clock>1567 MHz</supported_graphics_clock>
                <supported_graphics_clock>1560 MHz</supported_graphics_clock>
                <supported_graphics_clock>1552 MHz</supported_graphics_clock>
                <supported_graphics_clock>1545 MHz</supported_graphics_clock>
                <supported_graphics_clock>1537 MHz</supported_graphics_clock>
                <supported_graphics_clock>1530 MHz</supported_graphics_clock>
                <supported_graphics_clock>1522 MHz</supported_graphics_clock>
                <supported_graphics_clock>1515 MHz</supported_graphics_clock>
                <supported_graphics_clock>1507 MHz</supported_graphics_clock>
                <supported_graphics_clock>1500 MHz</supported_graphics_clock>
                <supported_graphics_clock>1492 MHz</supported_graphics_clock>
                <supported_graphics_clock>1485 MHz</supported_graphics_clock>
                <supported_graphics_clock>1477 MHz</supported_graphics_clock>
                <supported_graphics_clock>1470 MHz</supported_graphics_clock>
                <supported_graphics_clock>1462 MHz</supported_graphics_clock>
                <supported_graphics_clock>1455 MHz</supported_graphics_clock>
                <supported_graphics_clock>1447 MHz</supported_graphics_clock>
                <supported_graphics_clock>1440 MHz</supported_graphics_clock>
                <supported_graphics_clock>1432 MHz</supported_graphics_clock>
                <supported_graphics_clock>1425 MHz</supported_graphics_clock>
                <supported_graphics_clock>1417 MHz</supported_graphics_clock>
                <supported_graphics_clock>1410 MHz</supported_graphics_clock>
                <supported_graphics_clock>1402 MHz</supported_graphics_clock>
                <supported_graphics_clock>1395 MHz</supported_graphics_clock>
                <supported_graphics_clock>1387 MHz</supported_graphics_clock>
                <supported_graphics_clock>1380 MHz</supported_graphics_clock>
                <supported_graphics_clock>1372 MHz</supported_graphics_clock>
                <supported_graphics_clock>1365 MHz</supported_graphics_clock>
                <supported_graphics_clock>1357 MHz</supported_graphics_clock>
                <supported_graphics_clock>1350 MHz</supported_graphics_clock>
                <supported_graphics_clock>1342 MHz</supported_graphics_clock>
                <supported_graphics_clock>1335 MHz</supported_graphics_clock>
                <supported_graphics_clock>1327 MHz</supported_graphics_clock>
                <supported_graphics_clock>1320 MHz</supported_graphics_clock>
                <supported_graphics_clock>1312 MHz</supported_graphics_clock>
                <supported_graphics_clock>1305 MHz</supported_graphics_clock>
                <supported_graphics_clock>1297 MHz</supported_graphics_clock>
                <supported_graphics_clock>1290 MHz</supported_graphics_clock>
                <supported_graphics_clock>1282 MHz</supported_graphics_clock>
                <supported_graphics_clock>1275 MHz</supported_graphics_clock>
                <supported_graphics_clock>1267 MHz</supported_graphics_clock>
                <supported_graphics_clock>1260 MHz</supported_graphics_clock>
                <supported_graphics_clock>1252 MHz</supported_graphics_clock>
                <supported_graphics_clock>1245 MHz</supported_graphics_clock>
                <supported_graphics_clock>1237 MHz</supported_graphics_clock>
                <supported_graphics_clock>1230 MHz</supported_graphics_clock>
                <supported_graphics_clock>1222 MHz</supported_graphics_clock>
                <supported_graphics_clock>1215 MHz</supported_graphics_clock>
                <supported_graphics_clock>1207 MHz</supported_graphics_clock>
                <supported_graphics_clock>1200 MHz</supported_graphics_clock>
                <supported_graphics_clock>1192 MHz</supported_graphics_clock>
                <supported_graphics_clock>1185 MHz</supported_graphics_clock>
                <supported_graphics_clock>1177 MHz</supported_graphics_clock>
                <supported_graphics_clock>1170 MHz</supported_graphics_clock>
                <supported_graphics_clock>1162 MHz</supported_graphics_clock>
                <supported_graphics_clock>1155 MHz</supported_graphics_clock>
                <supported_graphics_clock>1147 MHz</supported_graphics_clock>
                <supported_graphics_clock>1140 MHz</supported_graphics_clock>
                <supported_graphics_clock>1132 MHz</supported_graphics_clock>
                <supported_graphics_clock>1125 MHz</supported_graphics_clock>
                <supported_graphics_clock>1117 MHz</supported_graphics_clock>
                <supported_graphics_clock>1110 MHz</supported_graphics_clock>
                <supported_graphics_clock>1102 MHz</supported_graphics_clock>
                <supported_graphics_clock>1095 MHz</supported_graphics_clock>
                <supported_graphics_clock>1087 MHz</supported_graphics_clock>
                <supported_graphics_clock>1080 MHz</supported_graphics_clock>
                <supported_graphics_clock>1072 MHz</supported_graphics_clock>
                <supported_graphics_clock>1065 MHz</supported_graphics_clock>
                <supported_graphics_clock>1057 MHz</supported_graphics_clock>
                <supported_graphics_clock>1050 MHz</supported_graphics_clock>
                <supported_graphics_clock>1042 MHz</supported_graphics_clock>
                <supported_graphics_clock>1035 MHz</supported_graphics_clock>
                <supported_graphics_clock>1027 MHz</supported_graphics_clock>
                <supported_graphics_clock>1020 MHz</supported_graphics_clock>
                <supported_graphics_clock>1012 MHz</supported_graphics_clock>
                <supported_graphics_clock>1005 MHz</supported_graphics_clock>
                <supported_graphics_clock>997 MHz</supported_graphics_clock>
                <supported_graphics_clock>990 MHz</supported_graphics_clock>
                <supported_graphics_clock>982 MHz</supported_graphics_clock>
                <supported_graphics_clock>975 MHz</supported_graphics_clock>
                <supported_graphics_clock>967 MHz</supported_graphics_clock>
                <supported_graphics_clock>960 MHz</supported_graphics_clock>
                <supported_graphics_clock>952 MHz</supported_graphics_clock>
                <supported_graphics_clock>945 MHz</supported_graphics_clock>
                <supported_graphics_clock>937 MHz</supported_graphics_clock>
                <supported_graphics_clock>930 MHz</supported_graphics_clock>
                <supported_graphics_clock>922 MHz</supported_graphics_clock>
                <supported_graphics_clock>915 MHz</supported_graphics_clock>
                <supported_graphics_clock>907 MHz</supported_graphics_clock>
                <supported_graphics_clock>900 MHz</supported_graphics_clock>
                <supported_graphics_clock>892 MHz</supported_graphics_clock>
                <supported_graphics_clock>885 MHz</supported_graphics_clock>
                <supported_graphics_clock>877 MHz</supported_graphics_clock>
                <supported_graphics_clock>870 MHz</supported_graphics_clock>
                <supported_graphics_clock>862 MHz</supported_graphics_clock>
                <supported_graphics_clock>855 MHz</supported_graphics_clock>
                <supported_graphics_clock>847 MHz</supported_graphics_clock>
                <supported_graphics_clock>840 MHz</supported_graphics_clock>
                <supported_graphics_clock>832 MHz</supported_graphics_clock>
                <supported_graphics_clock>825 MHz</supported_graphics_clock>
                <supported_graphics_clock>817 MHz</supported_graphics_clock>
                <supported_graphics_clock>810 MHz</supported_graphics_clock>
                <supported_graphics_clock>802 MHz</supported_graphics_clock>
                <supported_graphics_clock>795 MHz</supported_graphics_clock>
                <supported_graphics_clock>787 MHz</supported_graphics_clock>
                <supported_graphics_clock>780 MHz</supported_graphics_clock>
                <supported_graphics_clock>772 MHz</supported_graphics_clock>
                <supported_graphics_clock>765 MHz</supported_graphics_clock>
                <supported_graphics_clock>757 MHz</supported_graphics_clock>
                <supported_graphics_clock>750 MHz</supported_graphics_clock>
                <supported_graphics_clock>742 MHz</supported_graphics_clock>
                <supported_graphics_clock>735 MHz</supported_graphics_clock>
                <supported_graphics_clock>727 MHz</supported_graphics_clock>
                <supported_graphics_clock>720 MHz</supported_graphics_clock>
                <supported_graphics_clock>712 MHz</supported_graphics_clock>
                <supported_graphics_clock>705 MHz</supported_graphics_clock>
                <supported_graphics_clock>697 MHz</supported_graphics_clock>
                <supported_graphics_clock>690 MHz</supported_graphics_clock>
                <supported_graphics_clock>682 MHz</supported_graphics_clock>
                <supported_graphics_clock>675 MHz</supported_graphics_clock>
                <supported_graphics_clock>667 MHz</supported_graphics_clock>
                <supported_graphics_clock>660 MHz</supported_graphics_clock>
                <supported_graphics_clock>652 MHz</supported_graphics_clock>
                <supported_graphics_clock>645 MHz</supported_graphics_clock>
                <supported_graphics_clock>637 MHz</supported_graphics_clock>
                <supported_graphics_clock>630 MHz</supported_graphics_clock>
                <supported_graphics_clock>622 MHz</supported_graphics_clock>
                <supported_graphics_clock>615 MHz</supported_graphics_clock>
                <supported_graphics_clock>607 MHz</supported_graphics_clock>
                <supported_graphics_clock>600 MHz</supported_graphics_clock>
                <supported_graphics_clock>592 MHz</supported_graphics_clock>
                <supported_graphics_clock>585 MHz</supported_graphics_clock>
                <supported_graphics_clock>577 MHz</supported_graphics_clock>
                <supported_graphics_clock>570 MHz</supported_graphics_clock>
                <supported_graphics_clock>562 MHz</supported_graphics_clock>
                <supported_graphics_clock>555 MHz</supported_graphics_clock>
                <supported_graphics_clock>547 MHz</supported_graphics_clock>
                <supported_graphics_clock>540 MHz</supported_graphics_clock>
                <supported_graphics_clock>532 MHz</supported_graphics_clock>
                <supported_graphics_clock>525 MHz</supported_graphics_clock>
                <supported_graphics_clock>517 MHz</supported_graphics_clock>
                <supported_graphics_clock>510 MHz</supported_graphics_clock>
                <supported_graphics_clock>502 MHz</supported_graphics_clock>
                <supported_graphics_clock>495 MHz</supported_graphics_clock>
                <supported_graphics_clock>487 MHz</supported_graphics_clock>
                <supported_graphics_clock>480 MHz</supported_graphics_clock>
                <supported_graphics_clock>472 MHz</supported_graphics_clock>
                <supported_graphics_clock>465 MHz</supported_graphics_clock>
                <supported_graphics_clock>457 MHz</supported_graphics_clock>
                <supported_graphics_clock>450 MHz</supported_graphics_clock>
                <supported_graphics_clock>442 MHz</supported_graphics_clock>
                <supported_graphics_clock>435 MHz</supported_graphics_clock>
                <supported_graphics_clock>427 MHz</supported_graphics_clock>
                <supported_graphics_clock>420 MHz</supported_graphics_clock>
                <supported_graphics_clock>412 MHz</supported_graphics_clock>
                <supported_graphics_clock>405 MHz</supported_graphics_clock>
                <supported_graphics_clock>397 MHz</supported_graphics_clock>
                <supported_graphics_clock>390 MHz</supported_graphics_clock>
                <supported_graphics_clock>382 MHz</supported_graphics_clock>
                <supported_graphics_clock>375 MHz</supported_graphics_clock>
                <supported_graphics_clock>367 MHz</supported_graphics_clock>
                <supported_graphics_clock>360 MHz</supported_graphics_clock>
                <supported_graphics_clock>352 MHz</supported_graphics_clock>
                <supported_graphics_clock>345 MHz</supported_graphics_clock>
                <supported_graphics_clock>337 MHz</supported_graphics_clock>
                <supported_graphics_clock>330 MHz</supported_graphics_clock>
                <supported_graphics_clock>322 MHz</supported_graphics_clock>
                <supported_graphics_clock>315 MHz</supported_graphics_clock>
                <supported_graphics_clock>307 MHz</supported_graphics_clock>
                <supported_graphics_clock>300 MHz</supported_graphics_clock>
                <supported_graphics_clock>292 MHz</supported_graphics_clock>
                <supported_graphics_clock>285 MHz</supported_graphics_clock>
                <supported_graphics_clock>277 MHz</supported_graphics_clock>
                <supported_graphics_clock>270 MHz</supported_graphics_clock>
                <supported_graphics_clock>262 MHz</supported_graphics_clock>
                <supported_graphics_clock>255 MHz</supported_graphics_clock>
                <supported_graphics_clock>247 MHz</supported_graphics_clock>
                <supported_graphics_clock>240 MHz</supported_graphics_clock>
                <supported_graphics_clock>232 MHz</supported_graphics_clock>
                <supported_graphics_clock>225 MHz</supported_graphics_clock>
                <supported_graphics_clock>217 MHz</supported_graphics_clock>
                <supported_graphics_clock>210 MHz</supported_graphics_clock>
                <supported_graphics_clock>202 MHz</supported_graphics_clock>
                <supported_graphics_clock>195 MHz</supported_graphics_clock>
                <supported_graphics_clock>187 MHz</supported_graphics_clock>
                <supported_graphics_clock>180 MHz</supported_graphics_clock>
                <supported_graphics_clock>172 MHz</supported_graphics_clock>
                <supported_graphics_clock>165 MHz</supported_graphics_clock>
                <supported_graphics_clock>157 MHz</supported_graphics_clock>
                <supported_graphics_clock>150 MHz</supported_graphics_clock>
                <supported_graphics_clock>142 MHz</supported_graphics_clock>
                <supported_graphics_clock>135 MHz</supported_graphics_clock>
            </supported_mem_clock>
        </supported_clocks>
        <processes>
            <process_info>
                <gpu_instance_id>N/A</gpu_instance_id>
                <compute_instance_id>N/A</compute_instance_id>
                <pid>37305</pid>
                <type>C</type>
                <process_name>python</process_name>
                <used_memory>31683 MiB</used_memory>
            </process_info>
        </processes>
        <accounted_processes>
            <accounted_process_info>
                <pid>54796</pid>
                <gpu_util>0 %</gpu_util>
                <memory_util>0 %</memory_util>
                <max_memory_usage>843 MiB</max_memory_usage>
                <time>18228 ms</time>
                <is_running>0</is_running>
            </accounted_process_info>
            <accounted_process_info>
                <pid>38011</pid>
                <gpu_util>1 %</gpu_util>
                <memory_util>0 %</memory_util>
                <max_memory_usage>3161 MiB</max_memory_usage>
                <time>19326 ms</time>
                <is_running>0</is_running>
            </accounted_process_info>
            <accounted_process_info>
                <pid>39318</pid>
                <gpu_util>3 %</gpu_util>
                <memory_util>0 %</memory_util>
                <max_memory_usage>3161 MiB</max_memory_usage>
                <time>9857 ms</time>
                <is_running>0</is_running>
            </accounted_process_info>
            <accounted_process_info>
                <pid>39527</pid>
                <gpu_util>0 %</gpu_util>
                <memory_util>0 %</memory_util>
                <max_memory_usage>3161 MiB</max_memory_usage>
                <time>359731 ms</time>
                <is_running>0</is_running>
            </accounted_process_info>
            <accounted_process_info>
                <pid>40224</pid>
                <gpu_util>0 %</gpu_util>
                <memory_util>0 %</memory_util>
                <max_memory_usage>3161 MiB</max_memory_usage>
                <time>55206 ms</time>
                <is_running>0</is_running>
            </accounted_process_info>
            <accounted_process_info>
                <pid>40632</pid>
                <gpu_util>0 %</gpu_util>
                <memory_util>0 %</memory_util>
                <max_memory_usage>3161 MiB</max_memory_usage>
                <time>1451154 ms</time>
                <is_running>0</is_running>
            </accounted_process_info>
            <accounted_process_info>
                <pid>42661</pid>
                <gpu_util>0 %</gpu_util>
                <memory_util>0 %</memory_util>
                <max_memory_usage>3161 MiB</max_memory_usage>
                <time>1291538 ms</time>
                <is_running>0</is_running>
            </accounted_process_info>
            <accounted_process_info>
                <pid>44457</pid>
                <gpu_util>0 %</gpu_util>
                <memory_util>0 %</memory_util>
                <max_memory_usage>3161 MiB</max_memory_usage>
                <time>1170512 ms</time>
                <is_running>0</is_running>
            </accounted_process_info>
            <accounted_process_info>
                <pid>46059</pid>
                <gpu_util>0 %</gpu_util>
                <memory_util>0 %</memory_util>
                <max_memory_usage>3161 MiB</max_memory_usage>
                <time>1452616 ms</time>
                <is_running>0</is_running>
            </accounted_process_info>
            <accounted_process_info>
                <pid>48437</pid>
                <gpu_util>0 %</gpu_util>
                <memory_util>0 %</memory_util>
                <max_memory_usage>3161 MiB</max_memory_usage>
                <time>326744 ms</time>
                <is_running>0</is_running>
            </accounted_process_info>
            <accounted_process_info>
                <pid>49523</pid>
                <gpu_util>0 %</gpu_util>
                <memory_util>0 %</memory_util>
                <max_memory_usage>3161 MiB</max_memory_usage>
                <time>1007466 ms</time>
                <is_running>0</is_running>
            </accounted_process_info>
            <accounted_process_info>
                <pid>51442</pid>
                <gpu_util>0 %</gpu_util>
                <memory_util>0 %</memory_util>
                <max_memory_usage>3161 MiB</max_memory_usage>
                <time>1452339 ms</time>
                <is_running>0</is_running>
            </accounted_process_info>
            <accounted_process_info>
                <pid>56076</pid>
                <gpu_util>2 %</gpu_util>
                <memory_util>0 %</memory_util>
                <max_memory_usage>3161 MiB</max_memory_usage>
                <time>11221 ms</time>
                <is_running>0</is_running>
            </accounted_process_info>
            <accounted_process_info>
                <pid>57006</pid>
                <gpu_util>1 %</gpu_util>
                <memory_util>0 %</memory_util>
                <max_memory_usage>1505 MiB</max_memory_usage>
                <time>8256 ms</time>
                <is_running>0</is_running>
            </accounted_process_info>
            <accounted_process_info>
                <pid>57134</pid>
                <gpu_util>0 %</gpu_util>
                <memory_util>0 %</memory_util>
                <max_memory_usage>32501 MiB</max_memory_usage>
                <time>4908860 ms</time>
                <is_running>0</is_running>
            </accounted_process_info>
            <accounted_process_info>
                <pid>64749</pid>
                <gpu_util>2 %</gpu_util>
                <memory_util>0 %</memory_util>
                <max_memory_usage>32503 MiB</max_memory_usage>
                <time>188931 ms</time>
                <is_running>0</is_running>
            </accounted_process_info>
            <accounted_process_info>
                <pid>2956</pid>
                <gpu_util>76 %</gpu_util>
                <memory_util>19 %</memory_util>
                <max_memory_usage>29057 MiB</max_memory_usage>
                <time>313868569 ms</time>
                <is_running>0</is_running>
            </accounted_process_info>
            <accounted_process_info>
                <pid>45030</pid>
                <gpu_util>0 %</gpu_util>
                <memory_util>0 %</memory_util>
                <max_memory_usage>32503 MiB</max_memory_usage>
                <time>16829455 ms</time>
                <is_running>0</is_running>
            </accounted_process_info>
            <accounted_process_info>
                <pid>9380</pid>
                <gpu_util>2 %</gpu_util>
                <memory_util>0 %</memory_util>
                <max_memory_usage>2453 MiB</max_memory_usage>
                <time>11027 ms</time>
                <is_running>0</is_running>
            </accounted_process_info>
            <accounted_process_info>
                <pid>9782</pid>
                <gpu_util>3 %</gpu_util>
                <memory_util>0 %</memory_util>
                <max_memory_usage>3667 MiB</max_memory_usage>
                <time>12918 ms</time>
                <is_running>0</is_running>
            </accounted_process_info>
            <accounted_process_info>
                <pid>10459</pid>
                <gpu_util>3 %</gpu_util>
                <memory_util>0 %</memory_util>
                <max_memory_usage>3667 MiB</max_memory_usage>
                <time>12586 ms</time>
                <is_running>0</is_running>
            </accounted_process_info>
            <accounted_process_info>
                <pid>10733</pid>
                <gpu_util>3 %</gpu_util>
                <memory_util>0 %</memory_util>
                <max_memory_usage>3119 MiB</max_memory_usage>
                <time>11670 ms</time>
                <is_running>0</is_running>
            </accounted_process_info>
            <accounted_process_info>
                <pid>11207</pid>
                <gpu_util>2 %</gpu_util>
                <memory_util>0 %</memory_util>
                <max_memory_usage>3119 MiB</max_memory_usage>
                <time>11388 ms</time>
                <is_running>0</is_running>
            </accounted_process_info>
            <accounted_process_info>
                <pid>11447</pid>
                <gpu_util>3 %</gpu_util>
                <memory_util>0 %</memory_util>
                <max_memory_usage>3119 MiB</max_memory_usage>
                <time>11460 ms</time>
                <is_running>0</is_running>
            </accounted_process_info>
            <accounted_process_info>
                <pid>11954</pid>
                <gpu_util>3 %</gpu_util>
                <memory_util>0 %</memory_util>
                <max_memory_usage>3119 MiB</max_memory_usage>
                <time>11398 ms</time>
                <is_running>0</is_running>
            </accounted_process_info>
            <accounted_process_info>
                <pid>13016</pid>
                <gpu_util>3 %</gpu_util>
                <memory_util>0 %</memory_util>
                <max_memory_usage>3119 MiB</max_memory_usage>
                <time>11422 ms</time>
                <is_running>0</is_running>
            </accounted_process_info>
            <accounted_process_info>
                <pid>37305</pid>
                <gpu_util>77 %</gpu_util>
                <memory_util>22 %</memory_util>
                <max_memory_usage>31683 MiB</max_memory_usage>
                <time>0 ms</time>
                <is_running>1</is_running>
            </accounted_process_info>
        </accounted_processes>
    </gpu>

    <gpu id="00000000:00:0E.0">
        <product_name>Tesla V100S-PCIE-32GB</product_name>
        <product_brand>Tesla</product_brand>
        <display_mode>Enabled</display_mode>
        <display_active>Disabled</display_active>
        <persistence_mode>Disabled</persistence_mode>
        <mig_mode>
            <current_mig>N/A</current_mig>
            <pending_mig>N/A</pending_mig>
        </mig_mode>
        <mig_devices>
            None
        </mig_devices>
        <accounting_mode>Enabled</accounting_mode>
        <accounting_mode_buffer_size>4000</accounting_mode_buffer_size>
        <driver_model>
            <current_dm>N/A</current_dm>
            <pending_dm>N/A</pending_dm>
        </driver_model>
        <serial>1562920005886</serial>
        <uuid>GPU-bbf7a241-bf9a-eb90-b2fb-d213294faf48</uuid>
        <minor_number>1</minor_number>
        <vbios_version>88.00.98.00.01</vbios_version>
        <multigpu_board>No</multigpu_board>
        <board_id>0xe</board_id>
        <gpu_part_number>900-2G500-0040-000</gpu_part_number>
        <inforom_version>
            <img_version>G500.0212.00.02</img_version>
            <oem_object>1.1</oem_object>
            <ecc_object>5.0</ecc_object>
            <pwr_object>N/A</pwr_object>
        </inforom_version>
        <gpu_operation_mode>
            <current_gom>N/A</current_gom>
            <pending_gom>N/A</pending_gom>
        </gpu_operation_mode>
        <gpu_virtualization_mode>
            <virtualization_mode>Pass-Through</virtualization_mode>
            <host_vgpu_mode>N/A</host_vgpu_mode>
        </gpu_virtualization_mode>
        <ibmnpu>
            <relaxed_ordering_mode>N/A</relaxed_ordering_mode>
        </ibmnpu>
        <pci>
            <pci_bus>00</pci_bus>
            <pci_device>0E</pci_device>
            <pci_domain>0000</pci_domain>
            <pci_device_id>1DF610DE</pci_device_id>
            <pci_bus_id>00000000:00:0E.0</pci_bus_id>
            <pci_sub_system_id>13D610DE</pci_sub_system_id>
            <pci_gpu_link_info>
                <pcie_gen>
                    <max_link_gen>3</max_link_gen>
                    <current_link_gen>3</current_link_gen>
                </pcie_gen>
                <link_widths>
                    <max_link_width>16x</max_link_width>
                    <current_link_width>16x</current_link_width>
                </link_widths>
            </pci_gpu_link_info>
            <pci_bridge_chip>
                <bridge_chip_type>N/A</bridge_chip_type>
                <bridge_chip_fw>N/A</bridge_chip_fw>
            </pci_bridge_chip>
            <replay_counter>0</replay_counter>
            <replay_rollover_counter>0</replay_rollover_counter>
            <tx_util>9000 KB/s</tx_util>
            <rx_util>39000 KB/s</rx_util>
        </pci>
        <fan_speed>N/A</fan_speed>
        <performance_state>P0</performance_state>
        <clocks_throttle_reasons>
            <clocks_throttle_reason_gpu_idle>Not Active</clocks_throttle_reason_gpu_idle>
            <clocks_throttle_reason_applications_clocks_setting>Not Active</clocks_throttle_reason_applications_clocks_setting>
            <clocks_throttle_reason_sw_power_cap>Not Active</clocks_throttle_reason_sw_power_cap>
            <clocks_throttle_reason_hw_slowdown>Not Active</clocks_throttle_reason_hw_slowdown>
            <clocks_throttle_reason_hw_thermal_slowdown>Not Active</clocks_throttle_reason_hw_thermal_slowdown>
            <clocks_throttle_reason_hw_power_brake_slowdown>Not Active</clocks_throttle_reason_hw_power_brake_slowdown>
            <clocks_throttle_reason_sync_boost>Not Active</clocks_throttle_reason_sync_boost>
            <clocks_throttle_reason_sw_thermal_slowdown>Not Active</clocks_throttle_reason_sw_thermal_slowdown>
            <clocks_throttle_reason_display_clocks_setting>Not Active</clocks_throttle_reason_display_clocks_setting>
        </clocks_throttle_reasons>
        <fb_memory_usage>
            <total>32510 MiB</total>
            <used>29545 MiB</used>
            <free>2965 MiB</free>
        </fb_memory_usage>
        <bar1_memory_usage>
            <total>32768 MiB</total>
            <used>32 MiB</used>
            <free>32736 MiB</free>
        </bar1_memory_usage>
        <compute_mode>Default</compute_mode>
        <utilization>
            <gpu_util>75 %</gpu_util>
            <memory_util>41 %</memory_util>
            <encoder_util>0 %</encoder_util>
            <decoder_util>0 %</decoder_util>
        </utilization>
        <encoder_stats>
            <session_count>0</session_count>
            <average_fps>0</average_fps>
            <average_latency>0</average_latency>
        </encoder_stats>
        <fbc_stats>
            <session_count>0</session_count>
            <average_fps>0</average_fps>
            <average_latency>0</average_latency>
        </fbc_stats>
        <ecc_mode>
            <current_ecc>Enabled</current_ecc>
            <pending_ecc>Enabled</pending_ecc>
        </ecc_mode>
        <ecc_errors>
            <volatile>
                <single_bit>
                    <device_memory>0</device_memory>
                    <register_file>0</register_file>
                    <l1_cache>0</l1_cache>
                    <l2_cache>0</l2_cache>
                    <texture_memory>N/A</texture_memory>
                    <texture_shm>N/A</texture_shm>
                    <cbu>N/A</cbu>
                    <total>0</total>
                </single_bit>
                <double_bit>
                    <device_memory>0</device_memory>
                    <register_file>0</register_file>
                    <l1_cache>0</l1_cache>
                    <l2_cache>0</l2_cache>
                    <texture_memory>N/A</texture_memory>
                    <texture_shm>N/A</texture_shm>
                    <cbu>0</cbu>
                    <total>0</total>
                </double_bit>
            </volatile>
            <aggregate>
                <single_bit>
                    <device_memory>0</device_memory>
                    <register_file>0</register_file>
                    <l1_cache>0</l1_cache>
                    <l2_cache>0</l2_cache>
                    <texture_memory>N/A</texture_memory>
                    <texture_shm>N/A</texture_shm>
                    <cbu>N/A</cbu>
                    <total>0</total>
                </single_bit>
                <double_bit>
                    <device_memory>0</device_memory>
                    <register_file>0</register_file>
                    <l1_cache>0</l1_cache>
                    <l2_cache>0</l2_cache>
                    <texture_memory>N/A</texture_memory>
                    <texture_shm>N/A</texture_shm>
                    <cbu>0</cbu>
                    <total>0</total>
                </double_bit>
            </aggregate>
        </ecc_errors>
        <retired_pages>
            <multiple_single_bit_retirement>
                <retired_count>0</retired_count>
                <retired_pagelist>
                </retired_pagelist>
            </multiple_single_bit_retirement>
            <double_bit_retirement>
                <retired_count>0</retired_count>
                <retired_pagelist>
                </retired_pagelist>
            </double_bit_retirement>
            <pending_blacklist>No</pending_blacklist>
            <pending_retirement>No</pending_retirement>
        </retired_pages>
        <remapped_rows>N/A</remapped_rows>
        <temperature>
            <gpu_temp>58 C</gpu_temp>
            <gpu_temp_max_threshold>90 C</gpu_temp_max_threshold>
            <gpu_temp_slow_threshold>87 C</gpu_temp_slow_threshold>
            <gpu_temp_max_gpu_threshold>83 C</gpu_temp_max_gpu_threshold>
            <gpu_target_temperature>N/A</gpu_target_temperature>
            <memory_temp>56 C</memory_temp>
            <gpu_temp_max_mem_threshold>85 C</gpu_temp_max_mem_threshold>
        </temperature>
        <supported_gpu_target_temp>
            <gpu_target_temp_min>N/A</gpu_target_temp_min>
            <gpu_target_temp_max>N/A</gpu_target_temp_max>
        </supported_gpu_target_temp>
        <power_readings>
            <power_state>P0</power_state>
            <power_management>Supported</power_management>
            <power_draw>152.03 W</power_draw>
            <power_limit>250.00 W</power_limit>
            <default_power_limit>250.00 W</default_power_limit>
            <enforced_power_limit>250.00 W</enforced_power_limit>
            <min_power_limit>100.00 W</min_power_limit>
            <max_power_limit>250.00 W</max_power_limit>
        </power_readings>
        <clocks>
            <graphics_clock>1432 MHz</graphics_clock>
            <sm_clock>1432 MHz</sm_clock>
            <mem_clock>1107 MHz</mem_clock>
            <video_clock>1282 MHz</video_clock>
        </clocks>
        <applications_clocks>
            <graphics_clock>1245 MHz</graphics_clock>
            <mem_clock>1107 MHz</mem_clock>
        </applications_clocks>
        <default_applications_clocks>
            <graphics_clock>1245 MHz</graphics_clock>
            <mem_clock>1107 MHz</mem_clock>
        </default_applications_clocks>
        <max_clocks>
            <graphics_clock>1597 MHz</graphics_clock>
            <sm_clock>1597 MHz</sm_clock>
            <mem_clock>1107 MHz</mem_clock>
            <video_clock>1432 MHz</video_clock>
        </max_clocks>
        <max_customer_boost_clocks>
            <graphics_clock>1597 MHz</graphics_clock>
        </max_customer_boost_clocks>
        <clock_policy>
            <auto_boost>N/A</auto_boost>
            <auto_boost_default>N/A</auto_boost_default>
        </clock_policy>
        <supported_clocks>
            <supported_mem_clock>
                <value>1107 MHz</value>
                <supported_graphics_clock>1597 MHz</supported_graphics_clock>
                <supported_graphics_clock>1590 MHz</supported_graphics_clock>
                <supported_graphics_clock>1582 MHz</supported_graphics_clock>
                <supported_graphics_clock>1575 MHz</supported_graphics_clock>
                <supported_graphics_clock>1567 MHz</supported_graphics_clock>
                <supported_graphics_clock>1560 MHz</supported_graphics_clock>
                <supported_graphics_clock>1552 MHz</supported_graphics_clock>
                <supported_graphics_clock>1545 MHz</supported_graphics_clock>
                <supported_graphics_clock>1537 MHz</supported_graphics_clock>
                <supported_graphics_clock>1530 MHz</supported_graphics_clock>
                <supported_graphics_clock>1522 MHz</supported_graphics_clock>
                <supported_graphics_clock>1515 MHz</supported_graphics_clock>
                <supported_graphics_clock>1507 MHz</supported_graphics_clock>
                <supported_graphics_clock>1500 MHz</supported_graphics_clock>
                <supported_graphics_clock>1492 MHz</supported_graphics_clock>
                <supported_graphics_clock>1485 MHz</supported_graphics_clock>
                <supported_graphics_clock>1477 MHz</supported_graphics_clock>
                <supported_graphics_clock>1470 MHz</supported_graphics_clock>
                <supported_graphics_clock>1462 MHz</supported_graphics_clock>
                <supported_graphics_clock>1455 MHz</supported_graphics_clock>
                <supported_graphics_clock>1447 MHz</supported_graphics_clock>
                <supported_graphics_clock>1440 MHz</supported_graphics_clock>
                <supported_graphics_clock>1432 MHz</supported_graphics_clock>
                <supported_graphics_clock>1425 MHz</supported_graphics_clock>
                <supported_graphics_clock>1417 MHz</supported_graphics_clock>
                <supported_graphics_clock>1410 MHz</supported_graphics_clock>
                <supported_graphics_clock>1402 MHz</supported_graphics_clock>
                <supported_graphics_clock>1395 MHz</supported_graphics_clock>
                <supported_graphics_clock>1387 MHz</supported_graphics_clock>
                <supported_graphics_clock>1380 MHz</supported_graphics_clock>
                <supported_graphics_clock>1372 MHz</supported_graphics_clock>
                <supported_graphics_clock>1365 MHz</supported_graphics_clock>
                <supported_graphics_clock>1357 MHz</supported_graphics_clock>
                <supported_graphics_clock>1350 MHz</supported_graphics_clock>
                <supported_graphics_clock>1342 MHz</supported_graphics_clock>
                <supported_graphics_clock>1335 MHz</supported_graphics_clock>
                <supported_graphics_clock>1327 MHz</supported_graphics_clock>
                <supported_graphics_clock>1320 MHz</supported_graphics_clock>
                <supported_graphics_clock>1312 MHz</supported_graphics_clock>
                <supported_graphics_clock>1305 MHz</supported_graphics_clock>
                <supported_graphics_clock>1297 MHz</supported_graphics_clock>
                <supported_graphics_clock>1290 MHz</supported_graphics_clock>
                <supported_graphics_clock>1282 MHz</supported_graphics_clock>
                <supported_graphics_clock>1275 MHz</supported_graphics_clock>
                <supported_graphics_clock>1267 MHz</supported_graphics_clock>
                <supported_graphics_clock>1260 MHz</supported_graphics_clock>
                <supported_graphics_clock>1252 MHz</supported_graphics_clock>
                <supported_graphics_clock>1245 MHz</supported_graphics_clock>
                <supported_graphics_clock>1237 MHz</supported_graphics_clock>
                <supported_graphics_clock>1230 MHz</supported_graphics_clock>
                <supported_graphics_clock>1222 MHz</supported_graphics_clock>
                <supported_graphics_clock>1215 MHz</supported_graphics_clock>
                <supported_graphics_clock>1207 MHz</supported_graphics_clock>
                <supported_graphics_clock>1200 MHz</supported_graphics_clock>
                <supported_graphics_clock>1192 MHz</supported_graphics_clock>
                <supported_graphics_clock>1185 MHz</supported_graphics_clock>
                <supported_graphics_clock>1177 MHz</supported_graphics_clock>
                <supported_graphics_clock>1170 MHz</supported_graphics_clock>
                <supported_graphics_clock>1162 MHz</supported_graphics_clock>
                <supported_graphics_clock>1155 MHz</supported_graphics_clock>
                <supported_graphics_clock>1147 MHz</supported_graphics_clock>
                <supported_graphics_clock>1140 MHz</supported_graphics_clock>
                <supported_graphics_clock>1132 MHz</supported_graphics_clock>
                <supported_graphics_clock>1125 MHz</supported_graphics_clock>
                <supported_graphics_clock>1117 MHz</supported_graphics_clock>
                <supported_graphics_clock>1110 MHz</supported_graphics_clock>
                <supported_graphics_clock>1102 MHz</supported_graphics_clock>
                <supported_graphics_clock>1095 MHz</supported_graphics_clock>
                <supported_graphics_clock>1087 MHz</supported_graphics_clock>
                <supported_graphics_clock>1080 MHz</supported_graphics_clock>
                <supported_graphics_clock>1072 MHz</supported_graphics_clock>
                <supported_graphics_clock>1065 MHz</supported_graphics_clock>
                <supported_graphics_clock>1057 MHz</supported_graphics_clock>
                <supported_graphics_clock>1050 MHz</supported_graphics_clock>
                <supported_graphics_clock>1042 MHz</supported_graphics_clock>
                <supported_graphics_clock>1035 MHz</supported_graphics_clock>
                <supported_graphics_clock>1027 MHz</supported_graphics_clock>
                <supported_graphics_clock>1020 MHz</supported_graphics_clock>
                <supported_graphics_clock>1012 MHz</supported_graphics_clock>
                <supported_graphics_clock>1005 MHz</supported_graphics_clock>
                <supported_graphics_clock>997 MHz</supported_graphics_clock>
                <supported_graphics_clock>990 MHz</supported_graphics_clock>
                <supported_graphics_clock>982 MHz</supported_graphics_clock>
                <supported_graphics_clock>975 MHz</supported_graphics_clock>
                <supported_graphics_clock>967 MHz</supported_graphics_clock>
                <supported_graphics_clock>960 MHz</supported_graphics_clock>
                <supported_graphics_clock>952 MHz</supported_graphics_clock>
                <supported_graphics_clock>945 MHz</supported_graphics_clock>
                <supported_graphics_clock>937 MHz</supported_graphics_clock>
                <supported_graphics_clock>930 MHz</supported_graphics_clock>
                <supported_graphics_clock>922 MHz</supported_graphics_clock>
                <supported_graphics_clock>915 MHz</supported_graphics_clock>
                <supported_graphics_clock>907 MHz</supported_graphics_clock>
                <supported_graphics_clock>900 MHz</supported_graphics_clock>
                <supported_graphics_clock>892 MHz</supported_graphics_clock>
                <supported_graphics_clock>885 MHz</supported_graphics_clock>
                <supported_graphics_clock>877 MHz</supported_graphics_clock>
                <supported_graphics_clock>870 MHz</supported_graphics_clock>
                <supported_graphics_clock>862 MHz</supported_graphics_clock>
                <supported_graphics_clock>855 MHz</supported_graphics_clock>
                <supported_graphics_clock>847 MHz</supported_graphics_clock>
                <supported_graphics_clock>840 MHz</supported_graphics_clock>
                <supported_graphics_clock>832 MHz</supported_graphics_clock>
                <supported_graphics_clock>825 MHz</supported_graphics_clock>
                <supported_graphics_clock>817 MHz</supported_graphics_clock>
                <supported_graphics_clock>810 MHz</supported_graphics_clock>
                <supported_graphics_clock>802 MHz</supported_graphics_clock>
                <supported_graphics_clock>795 MHz</supported_graphics_clock>
                <supported_graphics_clock>787 MHz</supported_graphics_clock>
                <supported_graphics_clock>780 MHz</supported_graphics_clock>
                <supported_graphics_clock>772 MHz</supported_graphics_clock>
                <supported_graphics_clock>765 MHz</supported_graphics_clock>
                <supported_graphics_clock>757 MHz</supported_graphics_clock>
                <supported_graphics_clock>750 MHz</supported_graphics_clock>
                <supported_graphics_clock>742 MHz</supported_graphics_clock>
                <supported_graphics_clock>735 MHz</supported_graphics_clock>
                <supported_graphics_clock>727 MHz</supported_graphics_clock>
                <supported_graphics_clock>720 MHz</supported_graphics_clock>
                <supported_graphics_clock>712 MHz</supported_graphics_clock>
                <supported_graphics_clock>705 MHz</supported_graphics_clock>
                <supported_graphics_clock>697 MHz</supported_graphics_clock>
                <supported_graphics_clock>690 MHz</supported_graphics_clock>
                <supported_graphics_clock>682 MHz</supported_graphics_clock>
                <supported_graphics_clock>675 MHz</supported_graphics_clock>
                <supported_graphics_clock>667 MHz</supported_graphics_clock>
                <supported_graphics_clock>660 MHz</supported_graphics_clock>
                <supported_graphics_clock>652 MHz</supported_graphics_clock>
                <supported_graphics_clock>645 MHz</supported_graphics_clock>
                <supported_graphics_clock>637 MHz</supported_graphics_clock>
                <supported_graphics_clock>630 MHz</supported_graphics_clock>
                <supported_graphics_clock>622 MHz</supported_graphics_clock>
                <supported_graphics_clock>615 MHz</supported_graphics_clock>
                <supported_graphics_clock>607 MHz</supported_graphics_clock>
                <supported_graphics_clock>600 MHz</supported_graphics_clock>
                <supported_graphics_clock>592 MHz</supported_graphics_clock>
                <supported_graphics_clock>585 MHz</supported_graphics_clock>
                <supported_graphics_clock>577 MHz</supported_graphics_clock>
                <supported_graphics_clock>570 MHz</supported_graphics_clock>
                <supported_graphics_clock>562 MHz</supported_graphics_clock>
                <supported_graphics_clock>555 MHz</supported_graphics_clock>
                <supported_graphics_clock>547 MHz</supported_graphics_clock>
                <supported_graphics_clock>540 MHz</supported_graphics_clock>
                <supported_graphics_clock>532 MHz</supported_graphics_clock>
                <supported_graphics_clock>525 MHz</supported_graphics_clock>
                <supported_graphics_clock>517 MHz</supported_graphics_clock>
                <supported_graphics_clock>510 MHz</supported_graphics_clock>
                <supported_graphics_clock>502 MHz</supported_graphics_clock>
                <supported_graphics_clock>495 MHz</supported_graphics_clock>
                <supported_graphics_clock>487 MHz</supported_graphics_clock>
                <supported_graphics_clock>480 MHz</supported_graphics_clock>
                <supported_graphics_clock>472 MHz</supported_graphics_clock>
                <supported_graphics_clock>465 MHz</supported_graphics_clock>
                <supported_graphics_clock>457 MHz</supported_graphics_clock>
                <supported_graphics_clock>450 MHz</supported_graphics_clock>
                <supported_graphics_clock>442 MHz</supported_graphics_clock>
                <supported_graphics_clock>435 MHz</supported_graphics_clock>
                <supported_graphics_clock>427 MHz</supported_graphics_clock>
                <supported_graphics_clock>420 MHz</supported_graphics_clock>
                <supported_graphics_clock>412 MHz</supported_graphics_clock>
                <supported_graphics_clock>405 MHz</supported_graphics_clock>
                <supported_graphics_clock>397 MHz</supported_graphics_clock>
                <supported_graphics_clock>390 MHz</supported_graphics_clock>
                <supported_graphics_clock>382 MHz</supported_graphics_clock>
                <supported_graphics_clock>375 MHz</supported_graphics_clock>
                <supported_graphics_clock>367 MHz</supported_graphics_clock>
                <supported_graphics_clock>360 MHz</supported_graphics_clock>
                <supported_graphics_clock>352 MHz</supported_graphics_clock>
                <supported_graphics_clock>345 MHz</supported_graphics_clock>
                <supported_graphics_clock>337 MHz</supported_graphics_clock>
                <supported_graphics_clock>330 MHz</supported_graphics_clock>
                <supported_graphics_clock>322 MHz</supported_graphics_clock>
                <supported_graphics_clock>315 MHz</supported_graphics_clock>
                <supported_graphics_clock>307 MHz</supported_graphics_clock>
                <supported_graphics_clock>300 MHz</supported_graphics_clock>
                <supported_graphics_clock>292 MHz</supported_graphics_clock>
                <supported_graphics_clock>285 MHz</supported_graphics_clock>
                <supported_graphics_clock>277 MHz</supported_graphics_clock>
                <supported_graphics_clock>270 MHz</supported_graphics_clock>
                <supported_graphics_clock>262 MHz</supported_graphics_clock>
                <supported_graphics_clock>255 MHz</supported_graphics_clock>
                <supported_graphics_clock>247 MHz</supported_graphics_clock>
                <supported_graphics_clock>240 MHz</supported_graphics_clock>
                <supported_graphics_clock>232 MHz</supported_graphics_clock>
                <supported_graphics_clock>225 MHz</supported_graphics_clock>
                <supported_graphics_clock>217 MHz</supported_graphics_clock>
                <supported_graphics_clock>210 MHz</supported_graphics_clock>
                <supported_graphics_clock>202 MHz</supported_graphics_clock>
                <supported_graphics_clock>195 MHz</supported_graphics_clock>
                <supported_graphics_clock>187 MHz</supported_graphics_clock>
                <supported_graphics_clock>180 MHz</supported_graphics_clock>
                <supported_graphics_clock>172 MHz</supported_graphics_clock>
                <supported_graphics_clock>165 MHz</supported_graphics_clock>
                <supported_graphics_clock>157 MHz</supported_graphics_clock>
                <supported_graphics_clock>150 MHz</supported_graphics_clock>
                <supported_graphics_clock>142 MHz</supported_graphics_clock>
                <supported_graphics_clock>135 MHz</supported_graphics_clock>
            </supported_mem_clock>
        </supported_clocks>
        <processes>
            <process_info>
                <gpu_instance_id>N/A</gpu_instance_id>
                <compute_instance_id>N/A</compute_instance_id>
                <pid>37305</pid>
                <type>C</type>
                <process_name>python</process_name>
                <used_memory>29539 MiB</used_memory>
            </process_info>
        </processes>
        <accounted_processes>
            <accounted_process_info>
                <pid>54796</pid>
                <gpu_util>0 %</gpu_util>
                <memory_util>0 %</memory_util>
                <max_memory_usage>2723 MiB</max_memory_usage>
                <time>18195 ms</time>
                <is_running>0</is_running>
            </accounted_process_info>
            <accounted_process_info>
                <pid>57134</pid>
                <gpu_util>0 %</gpu_util>
                <memory_util>0 %</memory_util>
                <max_memory_usage>31903 MiB</max_memory_usage>
                <time>4908676 ms</time>
                <is_running>0</is_running>
            </accounted_process_info>
            <accounted_process_info>
                <pid>64749</pid>
                <gpu_util>1 %</gpu_util>
                <memory_util>0 %</memory_util>
                <max_memory_usage>31875 MiB</max_memory_usage>
                <time>188768 ms</time>
                <is_running>0</is_running>
            </accounted_process_info>
            <accounted_process_info>
                <pid>2956</pid>
                <gpu_util>48 %</gpu_util>
                <memory_util>18 %</memory_util>
                <max_memory_usage>30485 MiB</max_memory_usage>
                <time>313868447 ms</time>
                <is_running>0</is_running>
            </accounted_process_info>
            <accounted_process_info>
                <pid>45030</pid>
                <gpu_util>0 %</gpu_util>
                <memory_util>0 %</memory_util>
                <max_memory_usage>31463 MiB</max_memory_usage>
                <time>16829254 ms</time>
                <is_running>0</is_running>
            </accounted_process_info>
            <accounted_process_info>
                <pid>37305</pid>
                <gpu_util>53 %</gpu_util>
                <memory_util>22 %</memory_util>
                <max_memory_usage>30627 MiB</max_memory_usage>
                <time>0 ms</time>
                <is_running>1</is_running>
            </accounted_process_info>
        </accounted_processes>
    </gpu>

    <gpu id="00000000:00:0F.0">
        <product_name>Tesla V100S-PCIE-32GB</product_name>
        <product_brand>Tesla</product_brand>
        <display_mode>Enabled</display_mode>
        <display_active>Disabled</display_active>
        <persistence_mode>Disabled</persistence_mode>
        <mig_mode>
            <current_mig>N/A</current_mig>
            <pending_mig>N/A</pending_mig>
        </mig_mode>
        <mig_devices>
            None
        </mig_devices>
        <accounting_mode>Enabled</accounting_mode>
        <accounting_mode_buffer_size>4000</accounting_mode_buffer_size>
        <driver_model>
            <current_dm>N/A</current_dm>
            <pending_dm>N/A</pending_dm>
        </driver_model>
        <serial>1563320001045</serial>
        <uuid>GPU-8f75e71c-23dd-a060-d997-4caf592954db</uuid>
        <minor_number>2</minor_number>
        <vbios_version>88.00.98.00.01</vbios_version>
        <multigpu_board>No</multigpu_board>
        <board_id>0xf</board_id>
        <gpu_part_number>900-2G500-0040-000</gpu_part_number>
        <inforom_version>
            <img_version>G500.0212.00.02</img_version>
            <oem_object>1.1</oem_object>
            <ecc_object>5.0</ecc_object>
            <pwr_object>N/A</pwr_object>
        </inforom_version>
        <gpu_operation_mode>
            <current_gom>N/A</current_gom>
            <pending_gom>N/A</pending_gom>
        </gpu_operation_mode>
        <gpu_virtualization_mode>
            <virtualization_mode>Pass-Through</virtualization_mode>
            <host_vgpu_mode>N/A</host_vgpu_mode>
        </gpu_virtualization_mode>
        <ibmnpu>
            <relaxed_ordering_mode>N/A</relaxed_ordering_mode>
        </ibmnpu>
        <pci>
            <pci_bus>00</pci_bus>
            <pci_device>0F</pci_device>
            <pci_domain>0000</pci_domain>
            <pci_device_id>1DF610DE</pci_device_id>
            <pci_bus_id>00000000:00:0F.0</pci_bus_id>
            <pci_sub_system_id>13D610DE</pci_sub_system_id>
            <pci_gpu_link_info>
                <pcie_gen>
                    <max_link_gen>3</max_link_gen>
                    <current_link_gen>3</current_link_gen>
                </pcie_gen>
                <link_widths>
                    <max_link_width>16x</max_link_width>
                    <current_link_width>16x</current_link_width>
                </link_widths>
            </pci_gpu_link_info>
            <pci_bridge_chip>
                <bridge_chip_type>N/A</bridge_chip_type>
                <bridge_chip_fw>N/A</bridge_chip_fw>
            </pci_bridge_chip>
            <replay_counter>0</replay_counter>
            <replay_rollover_counter>0</replay_rollover_counter>
            <tx_util>10000 KB/s</tx_util>
            <rx_util>41000 KB/s</rx_util>
        </pci>
        <fan_speed>N/A</fan_speed>
        <performance_state>P0</performance_state>
        <clocks_throttle_reasons>
            <clocks_throttle_reason_gpu_idle>Not Active</clocks_throttle_reason_gpu_idle>
            <clocks_throttle_reason_applications_clocks_setting>Not Active</clocks_throttle_reason_applications_clocks_setting>
            <clocks_throttle_reason_sw_power_cap>Not Active</clocks_throttle_reason_sw_power_cap>
            <clocks_throttle_reason_hw_slowdown>Not Active</clocks_throttle_reason_hw_slowdown>
            <clocks_throttle_reason_hw_thermal_slowdown>Not Active</clocks_throttle_reason_hw_thermal_slowdown>
            <clocks_throttle_reason_hw_power_brake_slowdown>Not Active</clocks_throttle_reason_hw_power_brake_slowdown>
            <clocks_throttle_reason_sync_boost>Not Active</clocks_throttle_reason_sync_boost>
            <clocks_throttle_reason_sw_thermal_slowdown>Not Active</clocks_throttle_reason_sw_thermal_slowdown>
            <clocks_throttle_reason_display_clocks_setting>Not Active</clocks_throttle_reason_display_clocks_setting>
        </clocks_throttle_reasons>
        <fb_memory_usage>
            <total>32510 MiB</total>
            <used>29539 MiB</used>
            <free>2971 MiB</free>
        </fb_memory_usage>
        <bar1_memory_usage>
            <total>32768 MiB</total>
            <used>32 MiB</used>
            <free>32736 MiB</free>
        </bar1_memory_usage>
        <compute_mode>Default</compute_mode>
        <utilization>
            <gpu_util>80 %</gpu_util>
            <memory_util>42 %</memory_util>
            <encoder_util>0 %</encoder_util>
            <decoder_util>0 %</decoder_util>
        </utilization>
        <encoder_stats>
            <session_count>0</session_count>
            <average_fps>0</average_fps>
            <average_latency>0</average_latency>
        </encoder_stats>
        <fbc_stats>
            <session_count>0</session_count>
            <average_fps>0</average_fps>
            <average_latency>0</average_latency>
        </fbc_stats>
        <ecc_mode>
            <current_ecc>Enabled</current_ecc>
            <pending_ecc>Enabled</pending_ecc>
        </ecc_mode>
        <ecc_errors>
            <volatile>
                <single_bit>
                    <device_memory>0</device_memory>
                    <register_file>0</register_file>
                    <l1_cache>0</l1_cache>
                    <l2_cache>0</l2_cache>
                    <texture_memory>N/A</texture_memory>
                    <texture_shm>N/A</texture_shm>
                    <cbu>N/A</cbu>
                    <total>0</total>
                </single_bit>
                <double_bit>
                    <device_memory>0</device_memory>
                    <register_file>0</register_file>
                    <l1_cache>0</l1_cache>
                    <l2_cache>0</l2_cache>
                    <texture_memory>N/A</texture_memory>
                    <texture_shm>N/A</texture_shm>
                    <cbu>0</cbu>
                    <total>0</total>
                </double_bit>
            </volatile>
            <aggregate>
                <single_bit>
                    <device_memory>0</device_memory>
                    <register_file>0</register_file>
                    <l1_cache>0</l1_cache>
                    <l2_cache>0</l2_cache>
                    <texture_memory>N/A</texture_memory>
                    <texture_shm>N/A</texture_shm>
                    <cbu>N/A</cbu>
                    <total>0</total>
                </single_bit>
                <double_bit>
                    <device_memory>0</device_memory>
                    <register_file>0</register_file>
                    <l1_cache>0</l1_cache>
                    <l2_cache>0</l2_cache>
                    <texture_memory>N/A</texture_memory>
                    <texture_shm>N/A</texture_shm>
                    <cbu>0</cbu>
                    <total>0</total>
                </double_bit>
            </aggregate>
        </ecc_errors>
        <retired_pages>
            <multiple_single_bit_retirement>
                <retired_count>0</retired_count>
                <retired_pagelist>
                </retired_pagelist>
            </multiple_single_bit_retirement>
            <double_bit_retirement>
                <retired_count>0</retired_count>
                <retired_pagelist>
                </retired_pagelist>
            </double_bit_retirement>
            <pending_blacklist>No</pending_blacklist>
            <pending_retirement>No</pending_retirement>
        </retired_pages>
        <remapped_rows>N/A</remapped_rows>
        <temperature>
            <gpu_temp>57 C</gpu_temp>
            <gpu_temp_max_threshold>90 C</gpu_temp_max_threshold>
            <gpu_temp_slow_threshold>87 C</gpu_temp_slow_threshold>
            <gpu_temp_max_gpu_threshold>83 C</gpu_temp_max_gpu_threshold>
            <gpu_target_temperature>N/A</gpu_target_temperature>
            <memory_temp>55 C</memory_temp>
            <gpu_temp_max_mem_threshold>85 C</gpu_temp_max_mem_threshold>
        </temperature>
        <supported_gpu_target_temp>
            <gpu_target_temp_min>N/A</gpu_target_temp_min>
            <gpu_target_temp_max>N/A</gpu_target_temp_max>
        </supported_gpu_target_temp>
        <power_readings>
            <power_state>P0</power_state>
            <power_management>Supported</power_management>
            <power_draw>67.57 W</power_draw>
            <power_limit>250.00 W</power_limit>
            <default_power_limit>250.00 W</default_power_limit>
            <enforced_power_limit>250.00 W</enforced_power_limit>
            <min_power_limit>100.00 W</min_power_limit>
            <max_power_limit>250.00 W</max_power_limit>
        </power_readings>
        <clocks>
            <graphics_clock>1590 MHz</graphics_clock>
            <sm_clock>1590 MHz</sm_clock>
            <mem_clock>1107 MHz</mem_clock>
            <video_clock>1425 MHz</video_clock>
        </clocks>
        <applications_clocks>
            <graphics_clock>1245 MHz</graphics_clock>
            <mem_clock>1107 MHz</mem_clock>
        </applications_clocks>
        <default_applications_clocks>
            <graphics_clock>1245 MHz</graphics_clock>
            <mem_clock>1107 MHz</mem_clock>
        </default_applications_clocks>
        <max_clocks>
            <graphics_clock>1597 MHz</graphics_clock>
            <sm_clock>1597 MHz</sm_clock>
            <mem_clock>1107 MHz</mem_clock>
            <video_clock>1432 MHz</video_clock>
        </max_clocks>
        <max_customer_boost_clocks>
            <graphics_clock>1597 MHz</graphics_clock>
        </max_customer_boost_clocks>
        <clock_policy>
            <auto_boost>N/A</auto_boost>
            <auto_boost_default>N/A</auto_boost_default>
        </clock_policy>
        <supported_clocks>
            <supported_mem_clock>
                <value>1107 MHz</value>
                <supported_graphics_clock>1597 MHz</supported_graphics_clock>
                <supported_graphics_clock>1590 MHz</supported_graphics_clock>
                <supported_graphics_clock>1582 MHz</supported_graphics_clock>
                <supported_graphics_clock>1575 MHz</supported_graphics_clock>
                <supported_graphics_clock>1567 MHz</supported_graphics_clock>
                <supported_graphics_clock>1560 MHz</supported_graphics_clock>
                <supported_graphics_clock>1552 MHz</supported_graphics_clock>
                <supported_graphics_clock>1545 MHz</supported_graphics_clock>
                <supported_graphics_clock>1537 MHz</supported_graphics_clock>
                <supported_graphics_clock>1530 MHz</supported_graphics_clock>
                <supported_graphics_clock>1522 MHz</supported_graphics_clock>
                <supported_graphics_clock>1515 MHz</supported_graphics_clock>
                <supported_graphics_clock>1507 MHz</supported_graphics_clock>
                <supported_graphics_clock>1500 MHz</supported_graphics_clock>
                <supported_graphics_clock>1492 MHz</supported_graphics_clock>
                <supported_graphics_clock>1485 MHz</supported_graphics_clock>
                <supported_graphics_clock>1477 MHz</supported_graphics_clock>
                <supported_graphics_clock>1470 MHz</supported_graphics_clock>
                <supported_graphics_clock>1462 MHz</supported_graphics_clock>
                <supported_graphics_clock>1455 MHz</supported_graphics_clock>
                <supported_graphics_clock>1447 MHz</supported_graphics_clock>
                <supported_graphics_clock>1440 MHz</supported_graphics_clock>
                <supported_graphics_clock>1432 MHz</supported_graphics_clock>
                <supported_graphics_clock>1425 MHz</supported_graphics_clock>
                <supported_graphics_clock>1417 MHz</supported_graphics_clock>
                <supported_graphics_clock>1410 MHz</supported_graphics_clock>
                <supported_graphics_clock>1402 MHz</supported_graphics_clock>
                <supported_graphics_clock>1395 MHz</supported_graphics_clock>
                <supported_graphics_clock>1387 MHz</supported_graphics_clock>
                <supported_graphics_clock>1380 MHz</supported_graphics_clock>
                <supported_graphics_clock>1372 MHz</supported_graphics_clock>
                <supported_graphics_clock>1365 MHz</supported_graphics_clock>
                <supported_graphics_clock>1357 MHz</supported_graphics_clock>
                <supported_graphics_clock>1350 MHz</supported_graphics_clock>
                <supported_graphics_clock>1342 MHz</supported_graphics_clock>
                <supported_graphics_clock>1335 MHz</supported_graphics_clock>
                <supported_graphics_clock>1327 MHz</supported_graphics_clock>
                <supported_graphics_clock>1320 MHz</supported_graphics_clock>
                <supported_graphics_clock>1312 MHz</supported_graphics_clock>
                <supported_graphics_clock>1305 MHz</supported_graphics_clock>
                <supported_graphics_clock>1297 MHz</supported_graphics_clock>
                <supported_graphics_clock>1290 MHz</supported_graphics_clock>
                <supported_graphics_clock>1282 MHz</supported_graphics_clock>
                <supported_graphics_clock>1275 MHz</supported_graphics_clock>
                <supported_graphics_clock>1267 MHz</supported_graphics_clock>
                <supported_graphics_clock>1260 MHz</supported_graphics_clock>
                <supported_graphics_clock>1252 MHz</supported_graphics_clock>
                <supported_graphics_clock>1245 MHz</supported_graphics_clock>
                <supported_graphics_clock>1237 MHz</supported_graphics_clock>
                <supported_graphics_clock>1230 MHz</supported_graphics_clock>
                <supported_graphics_clock>1222 MHz</supported_graphics_clock>
                <supported_graphics_clock>1215 MHz</supported_graphics_clock>
                <supported_graphics_clock>1207 MHz</supported_graphics_clock>
                <supported_graphics_clock>1200 MHz</supported_graphics_clock>
                <supported_graphics_clock>1192 MHz</supported_graphics_clock>
                <supported_graphics_clock>1185 MHz</supported_graphics_clock>
                <supported_graphics_clock>1177 MHz</supported_graphics_clock>
                <supported_graphics_clock>1170 MHz</supported_graphics_clock>
                <supported_graphics_clock>1162 MHz</supported_graphics_clock>
                <supported_graphics_clock>1155 MHz</supported_graphics_clock>
                <supported_graphics_clock>1147 MHz</supported_graphics_clock>
                <supported_graphics_clock>1140 MHz</supported_graphics_clock>
                <supported_graphics_clock>1132 MHz</supported_graphics_clock>
                <supported_graphics_clock>1125 MHz</supported_graphics_clock>
                <supported_graphics_clock>1117 MHz</supported_graphics_clock>
                <supported_graphics_clock>1110 MHz</supported_graphics_clock>
                <supported_graphics_clock>1102 MHz</supported_graphics_clock>
                <supported_graphics_clock>1095 MHz</supported_graphics_clock>
                <supported_graphics_clock>1087 MHz</supported_graphics_clock>
                <supported_graphics_clock>1080 MHz</supported_graphics_clock>
                <supported_graphics_clock>1072 MHz</supported_graphics_clock>
                <supported_graphics_clock>1065 MHz</supported_graphics_clock>
                <supported_graphics_clock>1057 MHz</supported_graphics_clock>
                <supported_graphics_clock>1050 MHz</supported_graphics_clock>
                <supported_graphics_clock>1042 MHz</supported_graphics_clock>
                <supported_graphics_clock>1035 MHz</supported_graphics_clock>
                <supported_graphics_clock>1027 MHz</supported_graphics_clock>
                <supported_graphics_clock>1020 MHz</supported_graphics_clock>
                <supported_graphics_clock>1012 MHz</supported_graphics_clock>
                <supported_graphics_clock>1005 MHz</supported_graphics_clock>
                <supported_graphics_clock>997 MHz</supported_graphics_clock>
                <supported_graphics_clock>990 MHz</supported_graphics_clock>
                <supported_graphics_clock>982 MHz</supported_graphics_clock>
                <supported_graphics_clock>975 MHz</supported_graphics_clock>
                <supported_graphics_clock>967 MHz</supported_graphics_clock>
                <supported_graphics_clock>960 MHz</supported_graphics_clock>
                <supported_graphics_clock>952 MHz</supported_graphics_clock>
                <supported_graphics_clock>945 MHz</supported_graphics_clock>
                <supported_graphics_clock>937 MHz</supported_graphics_clock>
                <supported_graphics_clock>930 MHz</supported_graphics_clock>
                <supported_graphics_clock>922 MHz</supported_graphics_clock>
                <supported_graphics_clock>915 MHz</supported_graphics_clock>
                <supported_graphics_clock>907 MHz</supported_graphics_clock>
                <supported_graphics_clock>900 MHz</supported_graphics_clock>
                <supported_graphics_clock>892 MHz</supported_graphics_clock>
                <supported_graphics_clock>885 MHz</supported_graphics_clock>
                <supported_graphics_clock>877 MHz</supported_graphics_clock>
                <supported_graphics_clock>870 MHz</supported_graphics_clock>
                <supported_graphics_clock>862 MHz</supported_graphics_clock>
                <supported_graphics_clock>855 MHz</supported_graphics_clock>
                <supported_graphics_clock>847 MHz</supported_graphics_clock>
                <supported_graphics_clock>840 MHz</supported_graphics_clock>
                <supported_graphics_clock>832 MHz</supported_graphics_clock>
                <supported_graphics_clock>825 MHz</supported_graphics_clock>
                <supported_graphics_clock>817 MHz</supported_graphics_clock>
                <supported_graphics_clock>810 MHz</supported_graphics_clock>
                <supported_graphics_clock>802 MHz</supported_graphics_clock>
                <supported_graphics_clock>795 MHz</supported_graphics_clock>
                <supported_graphics_clock>787 MHz</supported_graphics_clock>
                <supported_graphics_clock>780 MHz</supported_graphics_clock>
                <supported_graphics_clock>772 MHz</supported_graphics_clock>
                <supported_graphics_clock>765 MHz</supported_graphics_clock>
                <supported_graphics_clock>757 MHz</supported_graphics_clock>
                <supported_graphics_clock>750 MHz</supported_graphics_clock>
                <supported_graphics_clock>742 MHz</supported_graphics_clock>
                <supported_graphics_clock>735 MHz</supported_graphics_clock>
                <supported_graphics_clock>727 MHz</supported_graphics_clock>
                <supported_graphics_clock>720 MHz</supported_graphics_clock>
                <supported_graphics_clock>712 MHz</supported_graphics_clock>
                <supported_graphics_clock>705 MHz</supported_graphics_clock>
                <supported_graphics_clock>697 MHz</supported_graphics_clock>
                <supported_graphics_clock>690 MHz</supported_graphics_clock>
                <supported_graphics_clock>682 MHz</supported_graphics_clock>
                <supported_graphics_clock>675 MHz</supported_graphics_clock>
                <supported_graphics_clock>667 MHz</supported_graphics_clock>
                <supported_graphics_clock>660 MHz</supported_graphics_clock>
                <supported_graphics_clock>652 MHz</supported_graphics_clock>
                <supported_graphics_clock>645 MHz</supported_graphics_clock>
                <supported_graphics_clock>637 MHz</supported_graphics_clock>
                <supported_graphics_clock>630 MHz</supported_graphics_clock>
                <supported_graphics_clock>622 MHz</supported_graphics_clock>
                <supported_graphics_clock>615 MHz</supported_graphics_clock>
                <supported_graphics_clock>607 MHz</supported_graphics_clock>
                <supported_graphics_clock>600 MHz</supported_graphics_clock>
                <supported_graphics_clock>592 MHz</supported_graphics_clock>
                <supported_graphics_clock>585 MHz</supported_graphics_clock>
                <supported_graphics_clock>577 MHz</supported_graphics_clock>
                <supported_graphics_clock>570 MHz</supported_graphics_clock>
                <supported_graphics_clock>562 MHz</supported_graphics_clock>
                <supported_graphics_clock>555 MHz</supported_graphics_clock>
                <supported_graphics_clock>547 MHz</supported_graphics_clock>
                <supported_graphics_clock>540 MHz</supported_graphics_clock>
                <supported_graphics_clock>532 MHz</supported_graphics_clock>
                <supported_graphics_clock>525 MHz</supported_graphics_clock>
                <supported_graphics_clock>517 MHz</supported_graphics_clock>
                <supported_graphics_clock>510 MHz</supported_graphics_clock>
                <supported_graphics_clock>502 MHz</supported_graphics_clock>
                <supported_graphics_clock>495 MHz</supported_graphics_clock>
                <supported_graphics_clock>487 MHz</supported_graphics_clock>
                <supported_graphics_clock>480 MHz</supported_graphics_clock>
                <supported_graphics_clock>472 MHz</supported_graphics_clock>
                <supported_graphics_clock>465 MHz</supported_graphics_clock>
                <supported_graphics_clock>457 MHz</supported_graphics_clock>
                <supported_graphics_clock>450 MHz</supported_graphics_clock>
                <supported_graphics_clock>442 MHz</supported_graphics_clock>
                <supported_graphics_clock>435 MHz</supported_graphics_clock>
                <supported_graphics_clock>427 MHz</supported_graphics_clock>
                <supported_graphics_clock>420 MHz</supported_graphics_clock>
                <supported_graphics_clock>412 MHz</supported_graphics_clock>
                <supported_graphics_clock>405 MHz</supported_graphics_clock>
                <supported_graphics_clock>397 MHz</supported_graphics_clock>
                <supported_graphics_clock>390 MHz</supported_graphics_clock>
                <supported_graphics_clock>382 MHz</supported_graphics_clock>
                <supported_graphics_clock>375 MHz</supported_graphics_clock>
                <supported_graphics_clock>367 MHz</supported_graphics_clock>
                <supported_graphics_clock>360 MHz</supported_graphics_clock>
                <supported_graphics_clock>352 MHz</supported_graphics_clock>
                <supported_graphics_clock>345 MHz</supported_graphics_clock>
                <supported_graphics_clock>337 MHz</supported_graphics_clock>
                <supported_graphics_clock>330 MHz</supported_graphics_clock>
                <supported_graphics_clock>322 MHz</supported_graphics_clock>
                <supported_graphics_clock>315 MHz</supported_graphics_clock>
                <supported_graphics_clock>307 MHz</supported_graphics_clock>
                <supported_graphics_clock>300 MHz</supported_graphics_clock>
                <supported_graphics_clock>292 MHz</supported_graphics_clock>
                <supported_graphics_clock>285 MHz</supported_graphics_clock>
                <supported_graphics_clock>277 MHz</supported_graphics_clock>
                <supported_graphics_clock>270 MHz</supported_graphics_clock>
                <supported_graphics_clock>262 MHz</supported_graphics_clock>
                <supported_graphics_clock>255 MHz</supported_graphics_clock>
                <supported_graphics_clock>247 MHz</supported_graphics_clock>
                <supported_graphics_clock>240 MHz</supported_graphics_clock>
                <supported_graphics_clock>232 MHz</supported_graphics_clock>
                <supported_graphics_clock>225 MHz</supported_graphics_clock>
                <supported_graphics_clock>217 MHz</supported_graphics_clock>
                <supported_graphics_clock>210 MHz</supported_graphics_clock>
                <supported_graphics_clock>202 MHz</supported_graphics_clock>
                <supported_graphics_clock>195 MHz</supported_graphics_clock>
                <supported_graphics_clock>187 MHz</supported_graphics_clock>
                <supported_graphics_clock>180 MHz</supported_graphics_clock>
                <supported_graphics_clock>172 MHz</supported_graphics_clock>
                <supported_graphics_clock>165 MHz</supported_graphics_clock>
                <supported_graphics_clock>157 MHz</supported_graphics_clock>
                <supported_graphics_clock>150 MHz</supported_graphics_clock>
                <supported_graphics_clock>142 MHz</supported_graphics_clock>
                <supported_graphics_clock>135 MHz</supported_graphics_clock>
            </supported_mem_clock>
        </supported_clocks>
        <processes>
            <process_info>
                <gpu_instance_id>N/A</gpu_instance_id>
                <compute_instance_id>N/A</compute_instance_id>
                <pid>37305</pid>
                <type>C</type>
                <process_name>python</process_name>
                <used_memory>29533 MiB</used_memory>
            </process_info>
        </processes>
        <accounted_processes>
            <accounted_process_info>
                <pid>54796</pid>
                <gpu_util>0 %</gpu_util>
                <memory_util>0 %</memory_util>
                <max_memory_usage>307 MiB</max_memory_usage>
                <time>18172 ms</time>
                <is_running>0</is_running>
            </accounted_process_info>
            <accounted_process_info>
                <pid>57134</pid>
                <gpu_util>0 %</gpu_util>
                <memory_util>0 %</memory_util>
                <max_memory_usage>31981 MiB</max_memory_usage>
                <time>4908534 ms</time>
                <is_running>0</is_running>
            </accounted_process_info>
            <accounted_process_info>
                <pid>64749</pid>
                <gpu_util>1 %</gpu_util>
                <memory_util>0 %</memory_util>
                <max_memory_usage>31935 MiB</max_memory_usage>
                <time>188638 ms</time>
                <is_running>0</is_running>
            </accounted_process_info>
            <accounted_process_info>
                <pid>2956</pid>
                <gpu_util>48 %</gpu_util>
                <memory_util>18 %</memory_util>
                <max_memory_usage>27981 MiB</max_memory_usage>
                <time>313868366 ms</time>
                <is_running>0</is_running>
            </accounted_process_info>
            <accounted_process_info>
                <pid>45030</pid>
                <gpu_util>0 %</gpu_util>
                <memory_util>0 %</memory_util>
                <max_memory_usage>31535 MiB</max_memory_usage>
                <time>16829107 ms</time>
                <is_running>0</is_running>
            </accounted_process_info>
            <accounted_process_info>
                <pid>37305</pid>
                <gpu_util>53 %</gpu_util>
                <memory_util>22 %</memory_util>
                <max_memory_usage>32133 MiB</max_memory_usage>
                <time>0 ms</time>
                <is_running>1</is_running>
            </accounted_process_info>
        </accounted_processes>
    </gpu>

    <gpu id="00000000:00:10.0">
        <product_name>Tesla V100S-PCIE-32GB</product_name>
        <product_brand>Tesla</product_brand>
        <display_mode>Enabled</display_mode>
        <display_active>Disabled</display_active>
        <persistence_mode>Disabled</persistence_mode>
        <mig_mode>
            <current_mig>N/A</current_mig>
            <pending_mig>N/A</pending_mig>
        </mig_mode>
        <mig_devices>
            None
        </mig_devices>
        <accounting_mode>Enabled</accounting_mode>
        <accounting_mode_buffer_size>4000</accounting_mode_buffer_size>
        <driver_model>
            <current_dm>N/A</current_dm>
            <pending_dm>N/A</pending_dm>
        </driver_model>
        <serial>1562920007325</serial>
        <uuid>GPU-3181eb56-505c-8214-ba05-c132f68f89b8</uuid>
        <minor_number>3</minor_number>
        <vbios_version>88.00.98.00.01</vbios_version>
        <multigpu_board>No</multigpu_board>
        <board_id>0x10</board_id>
        <gpu_part_number>900-2G500-0040-000</gpu_part_number>
        <inforom_version>
            <img_version>G500.0212.00.02</img_version>
            <oem_object>1.1</oem_object>
            <ecc_object>5.0</ecc_object>
            <pwr_object>N/A</pwr_object>
        </inforom_version>
        <gpu_operation_mode>
            <current_gom>N/A</current_gom>
            <pending_gom>N/A</pending_gom>
        </gpu_operation_mode>
        <gpu_virtualization_mode>
            <virtualization_mode>Pass-Through</virtualization_mode>
            <host_vgpu_mode>N/A</host_vgpu_mode>
        </gpu_virtualization_mode>
        <ibmnpu>
            <relaxed_ordering_mode>N/A</relaxed_ordering_mode>
        </ibmnpu>
        <pci>
            <pci_bus>00</pci_bus>
            <pci_device>10</pci_device>
            <pci_domain>0000</pci_domain>
            <pci_device_id>1DF610DE</pci_device_id>
            <pci_bus_id>00000000:00:10.0</pci_bus_id>
            <pci_sub_system_id>13D610DE</pci_sub_system_id>
            <pci_gpu_link_info>
                <pcie_gen>
                    <max_link_gen>3</max_link_gen>
                    <current_link_gen>3</current_link_gen>
                </pcie_gen>
                <link_widths>
                    <max_link_width>16x</max_link_width>
                    <current_link_width>16x</current_link_width>
                </link_widths>
            </pci_gpu_link_info>
            <pci_bridge_chip>
                <bridge_chip_type>N/A</bridge_chip_type>
                <bridge_chip_fw>N/A</bridge_chip_fw>
            </pci_bridge_chip>
            <replay_counter>0</replay_counter>
            <replay_rollover_counter>0</replay_rollover_counter>
            <tx_util>11000 KB/s</tx_util>
            <rx_util>38000 KB/s</rx_util>
        </pci>
        <fan_speed>N/A</fan_speed>
        <performance_state>P0</performance_state>
        <clocks_throttle_reasons>
            <clocks_throttle_reason_gpu_idle>Not Active</clocks_throttle_reason_gpu_idle>
            <clocks_throttle_reason_applications_clocks_setting>Not Active</clocks_throttle_reason_applications_clocks_setting>
            <clocks_throttle_reason_sw_power_cap>Active</clocks_throttle_reason_sw_power_cap>
            <clocks_throttle_reason_hw_slowdown>Not Active</clocks_throttle_reason_hw_slowdown>
            <clocks_throttle_reason_hw_thermal_slowdown>Not Active</clocks_throttle_reason_hw_thermal_slowdown>
            <clocks_throttle_reason_hw_power_brake_slowdown>Not Active</clocks_throttle_reason_hw_power_brake_slowdown>
            <clocks_throttle_reason_sync_boost>Not Active</clocks_throttle_reason_sync_boost>
            <clocks_throttle_reason_sw_thermal_slowdown>Not Active</clocks_throttle_reason_sw_thermal_slowdown>
            <clocks_throttle_reason_display_clocks_setting>Not Active</clocks_throttle_reason_display_clocks_setting>
        </clocks_throttle_reasons>
        <fb_memory_usage>
            <total>32510 MiB</total>
            <used>32099 MiB</used>
            <free>411 MiB</free>
        </fb_memory_usage>
        <bar1_memory_usage>
            <total>32768 MiB</total>
            <used>32 MiB</used>
            <free>32736 MiB</free>
        </bar1_memory_usage>
        <compute_mode>Default</compute_mode>
        <utilization>
            <gpu_util>23 %</gpu_util>
            <memory_util>12 %</memory_util>
            <encoder_util>0 %</encoder_util>
            <decoder_util>0 %</decoder_util>
        </utilization>
        <encoder_stats>
            <session_count>0</session_count>
            <average_fps>0</average_fps>
            <average_latency>0</average_latency>
        </encoder_stats>
        <fbc_stats>
            <session_count>0</session_count>
            <average_fps>0</average_fps>
            <average_latency>0</average_latency>
        </fbc_stats>
        <ecc_mode>
            <current_ecc>Enabled</current_ecc>
            <pending_ecc>Enabled</pending_ecc>
        </ecc_mode>
        <ecc_errors>
            <volatile>
                <single_bit>
                    <device_memory>0</device_memory>
                    <register_file>0</register_file>
                    <l1_cache>0</l1_cache>
                    <l2_cache>0</l2_cache>
                    <texture_memory>N/A</texture_memory>
                    <texture_shm>N/A</texture_shm>
                    <cbu>N/A</cbu>
                    <total>0</total>
                </single_bit>
                <double_bit>
                    <device_memory>0</device_memory>
                    <register_file>0</register_file>
                    <l1_cache>0</l1_cache>
                    <l2_cache>0</l2_cache>
                    <texture_memory>N/A</texture_memory>
                    <texture_shm>N/A</texture_shm>
                    <cbu>0</cbu>
                    <total>0</total>
                </double_bit>
            </volatile>
            <aggregate>
                <single_bit>
                    <device_memory>0</device_memory>
                    <register_file>0</register_file>
                    <l1_cache>0</l1_cache>
                    <l2_cache>0</l2_cache>
                    <texture_memory>N/A</texture_memory>
                    <texture_shm>N/A</texture_shm>
                    <cbu>N/A</cbu>
                    <total>0</total>
                </single_bit>
                <double_bit>
                    <device_memory>0</device_memory>
                    <register_file>0</register_file>
                    <l1_cache>0</l1_cache>
                    <l2_cache>0</l2_cache>
                    <texture_memory>N/A</texture_memory>
                    <texture_shm>N/A</texture_shm>
                    <cbu>0</cbu>
                    <total>0</total>
                </double_bit>
            </aggregate>
        </ecc_errors>
        <retired_pages>
            <multiple_single_bit_retirement>
                <retired_count>0</retired_count>
                <retired_pagelist>
                </retired_pagelist>
            </multiple_single_bit_retirement>
            <double_bit_retirement>
                <retired_count>0</retired_count>
                <retired_pagelist>
                </retired_pagelist>
            </double_bit_retirement>
            <pending_blacklist>No</pending_blacklist>
            <pending_retirement>No</pending_retirement>
        </retired_pages>
        <remapped_rows>N/A</remapped_rows>
        <temperature>
            <gpu_temp>61 C</gpu_temp>
            <gpu_temp_max_threshold>90 C</gpu_temp_max_threshold>
            <gpu_temp_slow_threshold>87 C</gpu_temp_slow_threshold>
            <gpu_temp_max_gpu_threshold>83 C</gpu_temp_max_gpu_threshold>
            <gpu_target_temperature>N/A</gpu_target_temperature>
            <memory_temp>59 C</memory_temp>
            <gpu_temp_max_mem_threshold>85 C</gpu_temp_max_mem_threshold>
        </temperature>
        <supported_gpu_target_temp>
            <gpu_target_temp_min>N/A</gpu_target_temp_min>
            <gpu_target_temp_max>N/A</gpu_target_temp_max>
        </supported_gpu_target_temp>
        <power_readings>
            <power_state>P0</power_state>
            <power_management>Supported</power_management>
            <power_draw>198.41 W</power_draw>
            <power_limit>250.00 W</power_limit>
            <default_power_limit>250.00 W</default_power_limit>
            <enforced_power_limit>250.00 W</enforced_power_limit>
            <min_power_limit>100.00 W</min_power_limit>
            <max_power_limit>250.00 W</max_power_limit>
        </power_readings>
        <clocks>
            <graphics_clock>1432 MHz</graphics_clock>
            <sm_clock>1432 MHz</sm_clock>
            <mem_clock>1107 MHz</mem_clock>
            <video_clock>1282 MHz</video_clock>
        </clocks>
        <applications_clocks>
            <graphics_clock>1245 MHz</graphics_clock>
            <mem_clock>1107 MHz</mem_clock>
        </applications_clocks>
        <default_applications_clocks>
            <graphics_clock>1245 MHz</graphics_clock>
            <mem_clock>1107 MHz</mem_clock>
        </default_applications_clocks>
        <max_clocks>
            <graphics_clock>1597 MHz</graphics_clock>
            <sm_clock>1597 MHz</sm_clock>
            <mem_clock>1107 MHz</mem_clock>
            <video_clock>1432 MHz</video_clock>
        </max_clocks>
        <max_customer_boost_clocks>
            <graphics_clock>1597 MHz</graphics_clock>
        </max_customer_boost_clocks>
        <clock_policy>
            <auto_boost>N/A</auto_boost>
            <auto_boost_default>N/A</auto_boost_default>
        </clock_policy>
        <supported_clocks>
            <supported_mem_clock>
                <value>1107 MHz</value>
                <supported_graphics_clock>1597 MHz</supported_graphics_clock>
                <supported_graphics_clock>1590 MHz</supported_graphics_clock>
                <supported_graphics_clock>1582 MHz</supported_graphics_clock>
                <supported_graphics_clock>1575 MHz</supported_graphics_clock>
                <supported_graphics_clock>1567 MHz</supported_graphics_clock>
                <supported_graphics_clock>1560 MHz</supported_graphics_clock>
                <supported_graphics_clock>1552 MHz</supported_graphics_clock>
                <supported_graphics_clock>1545 MHz</supported_graphics_clock>
                <supported_graphics_clock>1537 MHz</supported_graphics_clock>
                <supported_graphics_clock>1530 MHz</supported_graphics_clock>
                <supported_graphics_clock>1522 MHz</supported_graphics_clock>
                <supported_graphics_clock>1515 MHz</supported_graphics_clock>
                <supported_graphics_clock>1507 MHz</supported_graphics_clock>
                <supported_graphics_clock>1500 MHz</supported_graphics_clock>
                <supported_graphics_clock>1492 MHz</supported_graphics_clock>
                <supported_graphics_clock>1485 MHz</supported_graphics_clock>
                <supported_graphics_clock>1477 MHz</supported_graphics_clock>
                <supported_graphics_clock>1470 MHz</supported_graphics_clock>
                <supported_graphics_clock>1462 MHz</supported_graphics_clock>
                <supported_graphics_clock>1455 MHz</supported_graphics_clock>
                <supported_graphics_clock>1447 MHz</supported_graphics_clock>
                <supported_graphics_clock>1440 MHz</supported_graphics_clock>
                <supported_graphics_clock>1432 MHz</supported_graphics_clock>
                <supported_graphics_clock>1425 MHz</supported_graphics_clock>
                <supported_graphics_clock>1417 MHz</supported_graphics_clock>
                <supported_graphics_clock>1410 MHz</supported_graphics_clock>
                <supported_graphics_clock>1402 MHz</supported_graphics_clock>
                <supported_graphics_clock>1395 MHz</supported_graphics_clock>
                <supported_graphics_clock>1387 MHz</supported_graphics_clock>
                <supported_graphics_clock>1380 MHz</supported_graphics_clock>
                <supported_graphics_clock>1372 MHz</supported_graphics_clock>
                <supported_graphics_clock>1365 MHz</supported_graphics_clock>
                <supported_graphics_clock>1357 MHz</supported_graphics_clock>
                <supported_graphics_clock>1350 MHz</supported_graphics_clock>
                <supported_graphics_clock>1342 MHz</supported_graphics_clock>
                <supported_graphics_clock>1335 MHz</supported_graphics_clock>
                <supported_graphics_clock>1327 MHz</supported_graphics_clock>
                <supported_graphics_clock>1320 MHz</supported_graphics_clock>
                <supported_graphics_clock>1312 MHz</supported_graphics_clock>
                <supported_graphics_clock>1305 MHz</supported_graphics_clock>
                <supported_graphics_clock>1297 MHz</supported_graphics_clock>
                <supported_graphics_clock>1290 MHz</supported_graphics_clock>
                <supported_graphics_clock>1282 MHz</supported_graphics_clock>
                <supported_graphics_clock>1275 MHz</supported_graphics_clock>
                <supported_graphics_clock>1267 MHz</supported_graphics_clock>
                <supported_graphics_clock>1260 MHz</supported_graphics_clock>
                <supported_graphics_clock>1252 MHz</supported_graphics_clock>
                <supported_graphics_clock>1245 MHz</supported_graphics_clock>
                <supported_graphics_clock>1237 MHz</supported_graphics_clock>
                <supported_graphics_clock>1230 MHz</supported_graphics_clock>
                <supported_graphics_clock>1222 MHz</supported_graphics_clock>
                <supported_graphics_clock>1215 MHz</supported_graphics_clock>
                <supported_graphics_clock>1207 MHz</supported_graphics_clock>
                <supported_graphics_clock>1200 MHz</supported_graphics_clock>
                <supported_graphics_clock>1192 MHz</supported_graphics_clock>
                <supported_graphics_clock>1185 MHz</supported_graphics_clock>
                <supported_graphics_clock>1177 MHz</supported_graphics_clock>
                <supported_graphics_clock>1170 MHz</supported_graphics_clock>
                <supported_graphics_clock>1162 MHz</supported_graphics_clock>
                <supported_graphics_clock>1155 MHz</supported_graphics_clock>
                <supported_graphics_clock>1147 MHz</supported_graphics_clock>
                <supported_graphics_clock>1140 MHz</supported_graphics_clock>
                <supported_graphics_clock>1132 MHz</supported_graphics_clock>
                <supported_graphics_clock>1125 MHz</supported_graphics_clock>
                <supported_graphics_clock>1117 MHz</supported_graphics_clock>
                <supported_graphics_clock>1110 MHz</supported_graphics_clock>
                <supported_graphics_clock>1102 MHz</supported_graphics_clock>
                <supported_graphics_clock>1095 MHz</supported_graphics_clock>
                <supported_graphics_clock>1087 MHz</supported_graphics_clock>
                <supported_graphics_clock>1080 MHz</supported_graphics_clock>
                <supported_graphics_clock>1072 MHz</supported_graphics_clock>
                <supported_graphics_clock>1065 MHz</supported_graphics_clock>
                <supported_graphics_clock>1057 MHz</supported_graphics_clock>
                <supported_graphics_clock>1050 MHz</supported_graphics_clock>
                <supported_graphics_clock>1042 MHz</supported_graphics_clock>
                <supported_graphics_clock>1035 MHz</supported_graphics_clock>
                <supported_graphics_clock>1027 MHz</supported_graphics_clock>
                <supported_graphics_clock>1020 MHz</supported_graphics_clock>
                <supported_graphics_clock>1012 MHz</supported_graphics_clock>
                <supported_graphics_clock>1005 MHz</supported_graphics_clock>
                <supported_graphics_clock>997 MHz</supported_graphics_clock>
                <supported_graphics_clock>990 MHz</supported_graphics_clock>
                <supported_graphics_clock>982 MHz</supported_graphics_clock>
                <supported_graphics_clock>975 MHz</supported_graphics_clock>
                <supported_graphics_clock>967 MHz</supported_graphics_clock>
                <supported_graphics_clock>960 MHz</supported_graphics_clock>
                <supported_graphics_clock>952 MHz</supported_graphics_clock>
                <supported_graphics_clock>945 MHz</supported_graphics_clock>
                <supported_graphics_clock>937 MHz</supported_graphics_clock>
                <supported_graphics_clock>930 MHz</supported_graphics_clock>
                <supported_graphics_clock>922 MHz</supported_graphics_clock>
                <supported_graphics_clock>915 MHz</supported_graphics_clock>
                <supported_graphics_clock>907 MHz</supported_graphics_clock>
                <supported_graphics_clock>900 MHz</supported_graphics_clock>
                <supported_graphics_clock>892 MHz</supported_graphics_clock>
                <supported_graphics_clock>885 MHz</supported_graphics_clock>
                <supported_graphics_clock>877 MHz</supported_graphics_clock>
                <supported_graphics_clock>870 MHz</supported_graphics_clock>
                <supported_graphics_clock>862 MHz</supported_graphics_clock>
                <supported_graphics_clock>855 MHz</supported_graphics_clock>
                <supported_graphics_clock>847 MHz</supported_graphics_clock>
                <supported_graphics_clock>840 MHz</supported_graphics_clock>
                <supported_graphics_clock>832 MHz</supported_graphics_clock>
                <supported_graphics_clock>825 MHz</supported_graphics_clock>
                <supported_graphics_clock>817 MHz</supported_graphics_clock>
                <supported_graphics_clock>810 MHz</supported_graphics_clock>
                <supported_graphics_clock>802 MHz</supported_graphics_clock>
                <supported_graphics_clock>795 MHz</supported_graphics_clock>
                <supported_graphics_clock>787 MHz</supported_graphics_clock>
                <supported_graphics_clock>780 MHz</supported_graphics_clock>
                <supported_graphics_clock>772 MHz</supported_graphics_clock>
                <supported_graphics_clock>765 MHz</supported_graphics_clock>
                <supported_graphics_clock>757 MHz</supported_graphics_clock>
                <supported_graphics_clock>750 MHz</supported_graphics_clock>
                <supported_graphics_clock>742 MHz</supported_graphics_clock>
                <supported_graphics_clock>735 MHz</supported_graphics_clock>
                <supported_graphics_clock>727 MHz</supported_graphics_clock>
                <supported_graphics_clock>720 MHz</supported_graphics_clock>
                <supported_graphics_clock>712 MHz</supported_graphics_clock>
                <supported_graphics_clock>705 MHz</supported_graphics_clock>
                <supported_graphics_clock>697 MHz</supported_graphics_clock>
                <supported_graphics_clock>690 MHz</supported_graphics_clock>
                <supported_graphics_clock>682 MHz</supported_graphics_clock>
                <supported_graphics_clock>675 MHz</supported_graphics_clock>
                <supported_graphics_clock>667 MHz</supported_graphics_clock>
                <supported_graphics_clock>660 MHz</supported_graphics_clock>
                <supported_graphics_clock>652 MHz</supported_graphics_clock>
                <supported_graphics_clock>645 MHz</supported_graphics_clock>
                <supported_graphics_clock>637 MHz</supported_graphics_clock>
                <supported_graphics_clock>630 MHz</supported_graphics_clock>
                <supported_graphics_clock>622 MHz</supported_graphics_clock>
                <supported_graphics_clock>615 MHz</supported_graphics_clock>
                <supported_graphics_clock>607 MHz</supported_graphics_clock>
                <supported_graphics_clock>600 MHz</supported_graphics_clock>
                <supported_graphics_clock>592 MHz</supported_graphics_clock>
                <supported_graphics_clock>585 MHz</supported_graphics_clock>
                <supported_graphics_clock>577 MHz</supported_graphics_clock>
                <supported_graphics_clock>570 MHz</supported_graphics_clock>
                <supported_graphics_clock>562 MHz</supported_graphics_clock>
                <supported_graphics_clock>555 MHz</supported_graphics_clock>
                <supported_graphics_clock>547 MHz</supported_graphics_clock>
                <supported_graphics_clock>540 MHz</supported_graphics_clock>
                <supported_graphics_clock>532 MHz</supported_graphics_clock>
                <supported_graphics_clock>525 MHz</supported_graphics_clock>
                <supported_graphics_clock>517 MHz</supported_graphics_clock>
                <supported_graphics_clock>510 MHz</supported_graphics_clock>
                <supported_graphics_clock>502 MHz</supported_graphics_clock>
                <supported_graphics_clock>495 MHz</supported_graphics_clock>
                <supported_graphics_clock>487 MHz</supported_graphics_clock>
                <supported_graphics_clock>480 MHz</supported_graphics_clock>
                <supported_graphics_clock>472 MHz</supported_graphics_clock>
                <supported_graphics_clock>465 MHz</supported_graphics_clock>
                <supported_graphics_clock>457 MHz</supported_graphics_clock>
                <supported_graphics_clock>450 MHz</supported_graphics_clock>
                <supported_graphics_clock>442 MHz</supported_graphics_clock>
                <supported_graphics_clock>435 MHz</supported_graphics_clock>
                <supported_graphics_clock>427 MHz</supported_graphics_clock>
                <supported_graphics_clock>420 MHz</supported_graphics_clock>
                <supported_graphics_clock>412 MHz</supported_graphics_clock>
                <supported_graphics_clock>405 MHz</supported_graphics_clock>
                <supported_graphics_clock>397 MHz</supported_graphics_clock>
                <supported_graphics_clock>390 MHz</supported_graphics_clock>
                <supported_graphics_clock>382 MHz</supported_graphics_clock>
                <supported_graphics_clock>375 MHz</supported_graphics_clock>
                <supported_graphics_clock>367 MHz</supported_graphics_clock>
                <supported_graphics_clock>360 MHz</supported_graphics_clock>
                <supported_graphics_clock>352 MHz</supported_graphics_clock>
                <supported_graphics_clock>345 MHz</supported_graphics_clock>
                <supported_graphics_clock>337 MHz</supported_graphics_clock>
                <supported_graphics_clock>330 MHz</supported_graphics_clock>
                <supported_graphics_clock>322 MHz</supported_graphics_clock>
                <supported_graphics_clock>315 MHz</supported_graphics_clock>
                <supported_graphics_clock>307 MHz</supported_graphics_clock>
                <supported_graphics_clock>300 MHz</supported_graphics_clock>
                <supported_graphics_clock>292 MHz</supported_graphics_clock>
                <supported_graphics_clock>285 MHz</supported_graphics_clock>
                <supported_graphics_clock>277 MHz</supported_graphics_clock>
                <supported_graphics_clock>270 MHz</supported_graphics_clock>
                <supported_graphics_clock>262 MHz</supported_graphics_clock>
                <supported_graphics_clock>255 MHz</supported_graphics_clock>
                <supported_graphics_clock>247 MHz</supported_graphics_clock>
                <supported_graphics_clock>240 MHz</supported_graphics_clock>
                <supported_graphics_clock>232 MHz</supported_graphics_clock>
                <supported_graphics_clock>225 MHz</supported_graphics_clock>
                <supported_graphics_clock>217 MHz</supported_graphics_clock>
                <supported_graphics_clock>210 MHz</supported_graphics_clock>
                <supported_graphics_clock>202 MHz</supported_graphics_clock>
                <supported_graphics_clock>195 MHz</supported_graphics_clock>
                <supported_graphics_clock>187 MHz</supported_graphics_clock>
                <supported_graphics_clock>180 MHz</supported_graphics_clock>
                <supported_graphics_clock>172 MHz</supported_graphics_clock>
                <supported_graphics_clock>165 MHz</supported_graphics_clock>
                <supported_graphics_clock>157 MHz</supported_graphics_clock>
                <supported_graphics_clock>150 MHz</supported_graphics_clock>
                <supported_graphics_clock>142 MHz</supported_graphics_clock>
                <supported_graphics_clock>135 MHz</supported_graphics_clock>
            </supported_mem_clock>
        </supported_clocks>
        <processes>
            <process_info>
                <gpu_instance_id>N/A</gpu_instance_id>
                <compute_instance_id>N/A</compute_instance_id>
                <pid>37305</pid>
                <type>C</type>
                <process_name>python</process_name>
                <used_memory>32093 MiB</used_memory>
            </process_info>
        </processes>
        <accounted_processes>
            <accounted_process_info>
                <pid>54796</pid>
                <gpu_util>0 %</gpu_util>
                <memory_util>0 %</memory_util>
                <max_memory_usage>307 MiB</max_memory_usage>
                <time>18149 ms</time>
                <is_running>0</is_running>
            </accounted_process_info>
            <accounted_process_info>
                <pid>57134</pid>
                <gpu_util>0 %</gpu_util>
                <memory_util>0 %</memory_util>
                <max_memory_usage>31893 MiB</max_memory_usage>
                <time>4908391 ms</time>
                <is_running>0</is_running>
            </accounted_process_info>
            <accounted_process_info>
                <pid>64749</pid>
                <gpu_util>1 %</gpu_util>
                <memory_util>0 %</memory_util>
                <max_memory_usage>31879 MiB</max_memory_usage>
                <time>188507 ms</time>
                <is_running>0</is_running>
            </accounted_process_info>
            <accounted_process_info>
                <pid>2956</pid>
                <gpu_util>48 %</gpu_util>
                <memory_util>18 %</memory_util>
                <max_memory_usage>30315 MiB</max_memory_usage>
                <time>313868272 ms</time>
                <is_running>0</is_running>
            </accounted_process_info>
            <accounted_process_info>
                <pid>45030</pid>
                <gpu_util>0 %</gpu_util>
                <memory_util>0 %</memory_util>
                <max_memory_usage>31507 MiB</max_memory_usage>
                <time>16828957 ms</time>
                <is_running>0</is_running>
            </accounted_process_info>
            <accounted_process_info>
                <pid>37305</pid>
                <gpu_util>53 %</gpu_util>
                <memory_util>22 %</memory_util>
                <max_memory_usage>32093 MiB</max_memory_usage>
                <time>0 ms</time>
                <is_running>1</is_running>
            </accounted_process_info>
        </accounted_processes>
    </gpu>

    <gpu id="00000000:00:11.0">
        <product_name>Tesla V100S-PCIE-32GB</product_name>
        <product_brand>Tesla</product_brand>
        <display_mode>Enabled</display_mode>
        <display_active>Disabled</display_active>
        <persistence_mode>Disabled</persistence_mode>
        <mig_mode>
            <current_mig>N/A</current_mig>
            <pending_mig>N/A</pending_mig>
        </mig_mode>
        <mig_devices>
            None
        </mig_devices>
        <accounting_mode>Enabled</accounting_mode>
        <accounting_mode_buffer_size>4000</accounting_mode_buffer_size>
        <driver_model>
            <current_dm>N/A</current_dm>
            <pending_dm>N/A</pending_dm>
        </driver_model>
        <serial>1562920005822</serial>
        <uuid>GPU-2aa87d54-d4ff-2900-812b-1a4b7e119583</uuid>
        <minor_number>4</minor_number>
        <vbios_version>88.00.98.00.01</vbios_version>
        <multigpu_board>No</multigpu_board>
        <board_id>0x11</board_id>
        <gpu_part_number>900-2G500-0040-000</gpu_part_number>
        <inforom_version>
            <img_version>G500.0212.00.02</img_version>
            <oem_object>1.1</oem_object>
            <ecc_object>5.0</ecc_object>
            <pwr_object>N/A</pwr_object>
        </inforom_version>
        <gpu_operation_mode>
            <current_gom>N/A</current_gom>
            <pending_gom>N/A</pending_gom>
        </gpu_operation_mode>
        <gpu_virtualization_mode>
            <virtualization_mode>Pass-Through</virtualization_mode>
            <host_vgpu_mode>N/A</host_vgpu_mode>
        </gpu_virtualization_mode>
        <ibmnpu>
            <relaxed_ordering_mode>N/A</relaxed_ordering_mode>
        </ibmnpu>
        <pci>
            <pci_bus>00</pci_bus>
            <pci_device>11</pci_device>
            <pci_domain>0000</pci_domain>
            <pci_device_id>1DF610DE</pci_device_id>
            <pci_bus_id>00000000:00:11.0</pci_bus_id>
            <pci_sub_system_id>13D610DE</pci_sub_system_id>
            <pci_gpu_link_info>
                <pcie_gen>
                    <max_link_gen>3</max_link_gen>
                    <current_link_gen>3</current_link_gen>
                </pcie_gen>
                <link_widths>
                    <max_link_width>16x</max_link_width>
                    <current_link_width>16x</current_link_width>
                </link_widths>
            </pci_gpu_link_info>
            <pci_bridge_chip>
                <bridge_chip_type>N/A</bridge_chip_type>
                <bridge_chip_fw>N/A</bridge_chip_fw>
            </pci_bridge_chip>
            <replay_counter>0</replay_counter>
            <replay_rollover_counter>0</replay_rollover_counter>
            <tx_util>602000 KB/s</tx_util>
            <rx_util>2797000 KB/s</rx_util>
        </pci>
        <fan_speed>N/A</fan_speed>
        <performance_state>P0</performance_state>
        <clocks_throttle_reasons>
            <clocks_throttle_reason_gpu_idle>Not Active</clocks_throttle_reason_gpu_idle>
            <clocks_throttle_reason_applications_clocks_setting>Not Active</clocks_throttle_reason_applications_clocks_setting>
            <clocks_throttle_reason_sw_power_cap>Not Active</clocks_throttle_reason_sw_power_cap>
            <clocks_throttle_reason_hw_slowdown>Not Active</clocks_throttle_reason_hw_slowdown>
            <clocks_throttle_reason_hw_thermal_slowdown>Not Active</clocks_throttle_reason_hw_thermal_slowdown>
            <clocks_throttle_reason_hw_power_brake_slowdown>Not Active</clocks_throttle_reason_hw_power_brake_slowdown>
            <clocks_throttle_reason_sync_boost>Not Active</clocks_throttle_reason_sync_boost>
            <clocks_throttle_reason_sw_thermal_slowdown>Not Active</clocks_throttle_reason_sw_thermal_slowdown>
            <clocks_throttle_reason_display_clocks_setting>Not Active</clocks_throttle_reason_display_clocks_setting>
        </clocks_throttle_reasons>
        <fb_memory_usage>
            <total>32510 MiB</total>
            <used>29517 MiB</used>
            <free>2993 MiB</free>
        </fb_memory_usage>
        <bar1_memory_usage>
            <total>32768 MiB</total>
            <used>32 MiB</used>
            <free>32736 MiB</free>
        </bar1_memory_usage>
        <compute_mode>Default</compute_mode>
        <utilization>
            <gpu_util>98 %</gpu_util>
            <memory_util>55 %</memory_util>
            <encoder_util>0 %</encoder_util>
            <decoder_util>0 %</decoder_util>
        </utilization>
        <encoder_stats>
            <session_count>0</session_count>
            <average_fps>0</average_fps>
            <average_latency>0</average_latency>
        </encoder_stats>
        <fbc_stats>
            <session_count>0</session_count>
            <average_fps>0</average_fps>
            <average_latency>0</average_latency>
        </fbc_stats>
        <ecc_mode>
            <current_ecc>Enabled</current_ecc>
            <pending_ecc>Enabled</pending_ecc>
        </ecc_mode>
        <ecc_errors>
            <volatile>
                <single_bit>
                    <device_memory>0</device_memory>
                    <register_file>0</register_file>
                    <l1_cache>0</l1_cache>
                    <l2_cache>0</l2_cache>
                    <texture_memory>N/A</texture_memory>
                    <texture_shm>N/A</texture_shm>
                    <cbu>N/A</cbu>
                    <total>0</total>
                </single_bit>
                <double_bit>
                    <device_memory>0</device_memory>
                    <register_file>0</register_file>
                    <l1_cache>0</l1_cache>
                    <l2_cache>0</l2_cache>
                    <texture_memory>N/A</texture_memory>
                    <texture_shm>N/A</texture_shm>
                    <cbu>0</cbu>
                    <total>0</total>
                </double_bit>
            </volatile>
            <aggregate>
                <single_bit>
                    <device_memory>0</device_memory>
                    <register_file>0</register_file>
                    <l1_cache>0</l1_cache>
                    <l2_cache>0</l2_cache>
                    <texture_memory>N/A</texture_memory>
                    <texture_shm>N/A</texture_shm>
                    <cbu>N/A</cbu>
                    <total>0</total>
                </single_bit>
                <double_bit>
                    <device_memory>0</device_memory>
                    <register_file>0</register_file>
                    <l1_cache>0</l1_cache>
                    <l2_cache>0</l2_cache>
                    <texture_memory>N/A</texture_memory>
                    <texture_shm>N/A</texture_shm>
                    <cbu>0</cbu>
                    <total>0</total>
                </double_bit>
            </aggregate>
        </ecc_errors>
        <retired_pages>
            <multiple_single_bit_retirement>
                <retired_count>0</retired_count>
                <retired_pagelist>
                </retired_pagelist>
            </multiple_single_bit_retirement>
            <double_bit_retirement>
                <retired_count>0</retired_count>
                <retired_pagelist>
                </retired_pagelist>
            </double_bit_retirement>
            <pending_blacklist>No</pending_blacklist>
            <pending_retirement>No</pending_retirement>
        </retired_pages>
        <remapped_rows>N/A</remapped_rows>
        <temperature>
            <gpu_temp>59 C</gpu_temp>
            <gpu_temp_max_threshold>90 C</gpu_temp_max_threshold>
            <gpu_temp_slow_threshold>87 C</gpu_temp_slow_threshold>
            <gpu_temp_max_gpu_threshold>83 C</gpu_temp_max_gpu_threshold>
            <gpu_target_temperature>N/A</gpu_target_temperature>
            <memory_temp>58 C</memory_temp>
            <gpu_temp_max_mem_threshold>85 C</gpu_temp_max_mem_threshold>
        </temperature>
        <supported_gpu_target_temp>
            <gpu_target_temp_min>N/A</gpu_target_temp_min>
            <gpu_target_temp_max>N/A</gpu_target_temp_max>
        </supported_gpu_target_temp>
        <power_readings>
            <power_state>P0</power_state>
            <power_management>Supported</power_management>
            <power_draw>117.99 W</power_draw>
            <power_limit>250.00 W</power_limit>
            <default_power_limit>250.00 W</default_power_limit>
            <enforced_power_limit>250.00 W</enforced_power_limit>
            <min_power_limit>100.00 W</min_power_limit>
            <max_power_limit>250.00 W</max_power_limit>
        </power_readings>
        <clocks>
            <graphics_clock>1342 MHz</graphics_clock>
            <sm_clock>1342 MHz</sm_clock>
            <mem_clock>1107 MHz</mem_clock>
            <video_clock>1207 MHz</video_clock>
        </clocks>
        <applications_clocks>
            <graphics_clock>1245 MHz</graphics_clock>
            <mem_clock>1107 MHz</mem_clock>
        </applications_clocks>
        <default_applications_clocks>
            <graphics_clock>1245 MHz</graphics_clock>
            <mem_clock>1107 MHz</mem_clock>
        </default_applications_clocks>
        <max_clocks>
            <graphics_clock>1597 MHz</graphics_clock>
            <sm_clock>1597 MHz</sm_clock>
            <mem_clock>1107 MHz</mem_clock>
            <video_clock>1432 MHz</video_clock>
        </max_clocks>
        <max_customer_boost_clocks>
            <graphics_clock>1597 MHz</graphics_clock>
        </max_customer_boost_clocks>
        <clock_policy>
            <auto_boost>N/A</auto_boost>
            <auto_boost_default>N/A</auto_boost_default>
        </clock_policy>
        <supported_clocks>
            <supported_mem_clock>
                <value>1107 MHz</value>
                <supported_graphics_clock>1597 MHz</supported_graphics_clock>
                <supported_graphics_clock>1590 MHz</supported_graphics_clock>
                <supported_graphics_clock>1582 MHz</supported_graphics_clock>
                <supported_graphics_clock>1575 MHz</supported_graphics_clock>
                <supported_graphics_clock>1567 MHz</supported_graphics_clock>
                <supported_graphics_clock>1560 MHz</supported_graphics_clock>
                <supported_graphics_clock>1552 MHz</supported_graphics_clock>
                <supported_graphics_clock>1545 MHz</supported_graphics_clock>
                <supported_graphics_clock>1537 MHz</supported_graphics_clock>
                <supported_graphics_clock>1530 MHz</supported_graphics_clock>
                <supported_graphics_clock>1522 MHz</supported_graphics_clock>
                <supported_graphics_clock>1515 MHz</supported_graphics_clock>
                <supported_graphics_clock>1507 MHz</supported_graphics_clock>
                <supported_graphics_clock>1500 MHz</supported_graphics_clock>
                <supported_graphics_clock>1492 MHz</supported_graphics_clock>
                <supported_graphics_clock>1485 MHz</supported_graphics_clock>
                <supported_graphics_clock>1477 MHz</supported_graphics_clock>
                <supported_graphics_clock>1470 MHz</supported_graphics_clock>
                <supported_graphics_clock>1462 MHz</supported_graphics_clock>
                <supported_graphics_clock>1455 MHz</supported_graphics_clock>
                <supported_graphics_clock>1447 MHz</supported_graphics_clock>
                <supported_graphics_clock>1440 MHz</supported_graphics_clock>
                <supported_graphics_clock>1432 MHz</supported_graphics_clock>
                <supported_graphics_clock>1425 MHz</supported_graphics_clock>
                <supported_graphics_clock>1417 MHz</supported_graphics_clock>
                <supported_graphics_clock>1410 MHz</supported_graphics_clock>
                <supported_graphics_clock>1402 MHz</supported_graphics_clock>
                <supported_graphics_clock>1395 MHz</supported_graphics_clock>
                <supported_graphics_clock>1387 MHz</supported_graphics_clock>
                <supported_graphics_clock>1380 MHz</supported_graphics_clock>
                <supported_graphics_clock>1372 MHz</supported_graphics_clock>
                <supported_graphics_clock>1365 MHz</supported_graphics_clock>
                <supported_graphics_clock>1357 MHz</supported_graphics_clock>
                <supported_graphics_clock>1350 MHz</supported_graphics_clock>
                <supported_graphics_clock>1342 MHz</supported_graphics_clock>
                <supported_graphics_clock>1335 MHz</supported_graphics_clock>
                <supported_graphics_clock>1327 MHz</supported_graphics_clock>
                <supported_graphics_clock>1320 MHz</supported_graphics_clock>
                <supported_graphics_clock>1312 MHz</supported_graphics_clock>
                <supported_graphics_clock>1305 MHz</supported_graphics_clock>
                <supported_graphics_clock>1297 MHz</supported_graphics_clock>
                <supported_graphics_clock>1290 MHz</supported_graphics_clock>
                <supported_graphics_clock>1282 MHz</supported_graphics_clock>
                <supported_graphics_clock>1275 MHz</supported_graphics_clock>
                <supported_graphics_clock>1267 MHz</supported_graphics_clock>
                <supported_graphics_clock>1260 MHz</supported_graphics_clock>
                <supported_graphics_clock>1252 MHz</supported_graphics_clock>
                <supported_graphics_clock>1245 MHz</supported_graphics_clock>
                <supported_graphics_clock>1237 MHz</supported_graphics_clock>
                <supported_graphics_clock>1230 MHz</supported_graphics_clock>
                <supported_graphics_clock>1222 MHz</supported_graphics_clock>
                <supported_graphics_clock>1215 MHz</supported_graphics_clock>
                <supported_graphics_clock>1207 MHz</supported_graphics_clock>
                <supported_graphics_clock>1200 MHz</supported_graphics_clock>
                <supported_graphics_clock>1192 MHz</supported_graphics_clock>
                <supported_graphics_clock>1185 MHz</supported_graphics_clock>
                <supported_graphics_clock>1177 MHz</supported_graphics_clock>
                <supported_graphics_clock>1170 MHz</supported_graphics_clock>
                <supported_graphics_clock>1162 MHz</supported_graphics_clock>
                <supported_graphics_clock>1155 MHz</supported_graphics_clock>
                <supported_graphics_clock>1147 MHz</supported_graphics_clock>
                <supported_graphics_clock>1140 MHz</supported_graphics_clock>
                <supported_graphics_clock>1132 MHz</supported_graphics_clock>
                <supported_graphics_clock>1125 MHz</supported_graphics_clock>
                <supported_graphics_clock>1117 MHz</supported_graphics_clock>
                <supported_graphics_clock>1110 MHz</supported_graphics_clock>
                <supported_graphics_clock>1102 MHz</supported_graphics_clock>
                <supported_graphics_clock>1095 MHz</supported_graphics_clock>
                <supported_graphics_clock>1087 MHz</supported_graphics_clock>
                <supported_graphics_clock>1080 MHz</supported_graphics_clock>
                <supported_graphics_clock>1072 MHz</supported_graphics_clock>
                <supported_graphics_clock>1065 MHz</supported_graphics_clock>
                <supported_graphics_clock>1057 MHz</supported_graphics_clock>
                <supported_graphics_clock>1050 MHz</supported_graphics_clock>
                <supported_graphics_clock>1042 MHz</supported_graphics_clock>
                <supported_graphics_clock>1035 MHz</supported_graphics_clock>
                <supported_graphics_clock>1027 MHz</supported_graphics_clock>
                <supported_graphics_clock>1020 MHz</supported_graphics_clock>
                <supported_graphics_clock>1012 MHz</supported_graphics_clock>
                <supported_graphics_clock>1005 MHz</supported_graphics_clock>
                <supported_graphics_clock>997 MHz</supported_graphics_clock>
                <supported_graphics_clock>990 MHz</supported_graphics_clock>
                <supported_graphics_clock>982 MHz</supported_graphics_clock>
                <supported_graphics_clock>975 MHz</supported_graphics_clock>
                <supported_graphics_clock>967 MHz</supported_graphics_clock>
                <supported_graphics_clock>960 MHz</supported_graphics_clock>
                <supported_graphics_clock>952 MHz</supported_graphics_clock>
                <supported_graphics_clock>945 MHz</supported_graphics_clock>
                <supported_graphics_clock>937 MHz</supported_graphics_clock>
                <supported_graphics_clock>930 MHz</supported_graphics_clock>
                <supported_graphics_clock>922 MHz</supported_graphics_clock>
                <supported_graphics_clock>915 MHz</supported_graphics_clock>
                <supported_graphics_clock>907 MHz</supported_graphics_clock>
                <supported_graphics_clock>900 MHz</supported_graphics_clock>
                <supported_graphics_clock>892 MHz</supported_graphics_clock>
                <supported_graphics_clock>885 MHz</supported_graphics_clock>
                <supported_graphics_clock>877 MHz</supported_graphics_clock>
                <supported_graphics_clock>870 MHz</supported_graphics_clock>
                <supported_graphics_clock>862 MHz</supported_graphics_clock>
                <supported_graphics_clock>855 MHz</supported_graphics_clock>
                <supported_graphics_clock>847 MHz</supported_graphics_clock>
                <supported_graphics_clock>840 MHz</supported_graphics_clock>
                <supported_graphics_clock>832 MHz</supported_graphics_clock>
                <supported_graphics_clock>825 MHz</supported_graphics_clock>
                <supported_graphics_clock>817 MHz</supported_graphics_clock>
                <supported_graphics_clock>810 MHz</supported_graphics_clock>
                <supported_graphics_clock>802 MHz</supported_graphics_clock>
                <supported_graphics_clock>795 MHz</supported_graphics_clock>
                <supported_graphics_clock>787 MHz</supported_graphics_clock>
                <supported_graphics_clock>780 MHz</supported_graphics_clock>
                <supported_graphics_clock>772 MHz</supported_graphics_clock>
                <supported_graphics_clock>765 MHz</supported_graphics_clock>
                <supported_graphics_clock>757 MHz</supported_graphics_clock>
                <supported_graphics_clock>750 MHz</supported_graphics_clock>
                <supported_graphics_clock>742 MHz</supported_graphics_clock>
                <supported_graphics_clock>735 MHz</supported_graphics_clock>
                <supported_graphics_clock>727 MHz</supported_graphics_clock>
                <supported_graphics_clock>720 MHz</supported_graphics_clock>
                <supported_graphics_clock>712 MHz</supported_graphics_clock>
                <supported_graphics_clock>705 MHz</supported_graphics_clock>
                <supported_graphics_clock>697 MHz</supported_graphics_clock>
                <supported_graphics_clock>690 MHz</supported_graphics_clock>
                <supported_graphics_clock>682 MHz</supported_graphics_clock>
                <supported_graphics_clock>675 MHz</supported_graphics_clock>
                <supported_graphics_clock>667 MHz</supported_graphics_clock>
                <supported_graphics_clock>660 MHz</supported_graphics_clock>
                <supported_graphics_clock>652 MHz</supported_graphics_clock>
                <supported_graphics_clock>645 MHz</supported_graphics_clock>
                <supported_graphics_clock>637 MHz</supported_graphics_clock>
                <supported_graphics_clock>630 MHz</supported_graphics_clock>
                <supported_graphics_clock>622 MHz</supported_graphics_clock>
                <supported_graphics_clock>615 MHz</supported_graphics_clock>
                <supported_graphics_clock>607 MHz</supported_graphics_clock>
                <supported_graphics_clock>600 MHz</supported_graphics_clock>
                <supported_graphics_clock>592 MHz</supported_graphics_clock>
                <supported_graphics_clock>585 MHz</supported_graphics_clock>
                <supported_graphics_clock>577 MHz</supported_graphics_clock>
                <supported_graphics_clock>570 MHz</supported_graphics_clock>
                <supported_graphics_clock>562 MHz</supported_graphics_clock>
                <supported_graphics_clock>555 MHz</supported_graphics_clock>
                <supported_graphics_clock>547 MHz</supported_graphics_clock>
                <supported_graphics_clock>540 MHz</supported_graphics_clock>
                <supported_graphics_clock>532 MHz</supported_graphics_clock>
                <supported_graphics_clock>525 MHz</supported_graphics_clock>
                <supported_graphics_clock>517 MHz</supported_graphics_clock>
                <supported_graphics_clock>510 MHz</supported_graphics_clock>
                <supported_graphics_clock>502 MHz</supported_graphics_clock>
                <supported_graphics_clock>495 MHz</supported_graphics_clock>
                <supported_graphics_clock>487 MHz</supported_graphics_clock>
                <supported_graphics_clock>480 MHz</supported_graphics_clock>
                <supported_graphics_clock>472 MHz</supported_graphics_clock>
                <supported_graphics_clock>465 MHz</supported_graphics_clock>
                <supported_graphics_clock>457 MHz</supported_graphics_clock>
                <supported_graphics_clock>450 MHz</supported_graphics_clock>
                <supported_graphics_clock>442 MHz</supported_graphics_clock>
                <supported_graphics_clock>435 MHz</supported_graphics_clock>
                <supported_graphics_clock>427 MHz</supported_graphics_clock>
                <supported_graphics_clock>420 MHz</supported_graphics_clock>
                <supported_graphics_clock>412 MHz</supported_graphics_clock>
                <supported_graphics_clock>405 MHz</supported_graphics_clock>
                <supported_graphics_clock>397 MHz</supported_graphics_clock>
                <supported_graphics_clock>390 MHz</supported_graphics_clock>
                <supported_graphics_clock>382 MHz</supported_graphics_clock>
                <supported_graphics_clock>375 MHz</supported_graphics_clock>
                <supported_graphics_clock>367 MHz</supported_graphics_clock>
                <supported_graphics_clock>360 MHz</supported_graphics_clock>
                <supported_graphics_clock>352 MHz</supported_graphics_clock>
                <supported_graphics_clock>345 MHz</supported_graphics_clock>
                <supported_graphics_clock>337 MHz</supported_graphics_clock>
                <supported_graphics_clock>330 MHz</supported_graphics_clock>
                <supported_graphics_clock>322 MHz</supported_graphics_clock>
                <supported_graphics_clock>315 MHz</supported_graphics_clock>
                <supported_graphics_clock>307 MHz</supported_graphics_clock>
                <supported_graphics_clock>300 MHz</supported_graphics_clock>
                <supported_graphics_clock>292 MHz</supported_graphics_clock>
                <supported_graphics_clock>285 MHz</supported_graphics_clock>
                <supported_graphics_clock>277 MHz</supported_graphics_clock>
                <supported_graphics_clock>270 MHz</supported_graphics_clock>
                <supported_graphics_clock>262 MHz</supported_graphics_clock>
                <supported_graphics_clock>255 MHz</supported_graphics_clock>
                <supported_graphics_clock>247 MHz</supported_graphics_clock>
                <supported_graphics_clock>240 MHz</supported_graphics_clock>
                <supported_graphics_clock>232 MHz</supported_graphics_clock>
                <supported_graphics_clock>225 MHz</supported_graphics_clock>
                <supported_graphics_clock>217 MHz</supported_graphics_clock>
                <supported_graphics_clock>210 MHz</supported_graphics_clock>
                <supported_graphics_clock>202 MHz</supported_graphics_clock>
                <supported_graphics_clock>195 MHz</supported_graphics_clock>
                <supported_graphics_clock>187 MHz</supported_graphics_clock>
                <supported_graphics_clock>180 MHz</supported_graphics_clock>
                <supported_graphics_clock>172 MHz</supported_graphics_clock>
                <supported_graphics_clock>165 MHz</supported_graphics_clock>
                <supported_graphics_clock>157 MHz</supported_graphics_clock>
                <supported_graphics_clock>150 MHz</supported_graphics_clock>
                <supported_graphics_clock>142 MHz</supported_graphics_clock>
                <supported_graphics_clock>135 MHz</supported_graphics_clock>
            </supported_mem_clock>
        </supported_clocks>
        <processes>
            <process_info>
                <gpu_instance_id>N/A</gpu_instance_id>
                <compute_instance_id>N/A</compute_instance_id>
                <pid>37305</pid>
                <type>C</type>
                <process_name>python</process_name>
                <used_memory>29511 MiB</used_memory>
            </process_info>
        </processes>
        <accounted_processes>
            <accounted_process_info>
                <pid>54796</pid>
                <gpu_util>0 %</gpu_util>
                <memory_util>0 %</memory_util>
                <max_memory_usage>307 MiB</max_memory_usage>
                <time>18092 ms</time>
                <is_running>0</is_running>
            </accounted_process_info>
            <accounted_process_info>
                <pid>57134</pid>
                <gpu_util>0 %</gpu_util>
                <memory_util>0 %</memory_util>
                <max_memory_usage>31913 MiB</max_memory_usage>
                <time>4908259 ms</time>
                <is_running>0</is_running>
            </accounted_process_info>
            <accounted_process_info>
                <pid>64749</pid>
                <gpu_util>1 %</gpu_util>
                <memory_util>0 %</memory_util>
                <max_memory_usage>31887 MiB</max_memory_usage>
                <time>188375 ms</time>
                <is_running>0</is_running>
            </accounted_process_info>
            <accounted_process_info>
                <pid>2956</pid>
                <gpu_util>53 %</gpu_util>
                <memory_util>18 %</memory_util>
                <max_memory_usage>30481 MiB</max_memory_usage>
                <time>313868193 ms</time>
                <is_running>0</is_running>
            </accounted_process_info>
            <accounted_process_info>
                <pid>45030</pid>
                <gpu_util>0 %</gpu_util>
                <memory_util>0 %</memory_util>
                <max_memory_usage>31521 MiB</max_memory_usage>
                <time>16828801 ms</time>
                <is_running>0</is_running>
            </accounted_process_info>
            <accounted_process_info>
                <pid>37305</pid>
                <gpu_util>58 %</gpu_util>
                <memory_util>22 %</memory_util>
                <max_memory_usage>30591 MiB</max_memory_usage>
                <time>0 ms</time>
                <is_running>1</is_running>
            </accounted_process_info>
        </accounted_processes>
    </gpu>

    <gpu id="00000000:00:12.0">
        <product_name>Tesla V100S-PCIE-32GB</product_name>
        <product_brand>Tesla</product_brand>
        <display_mode>Enabled</display_mode>
        <display_active>Disabled</display_active>
        <persistence_mode>Disabled</persistence_mode>
        <mig_mode>
            <current_mig>N/A</current_mig>
            <pending_mig>N/A</pending_mig>
        </mig_mode>
        <mig_devices>
            None
        </mig_devices>
        <accounting_mode>Enabled</accounting_mode>
        <accounting_mode_buffer_size>4000</accounting_mode_buffer_size>
        <driver_model>
            <current_dm>N/A</current_dm>
            <pending_dm>N/A</pending_dm>
        </driver_model>
        <serial>1562920005539</serial>
        <uuid>GPU-39c3a168-9034-0b5a-223a-1042f30a9242</uuid>
        <minor_number>5</minor_number>
        <vbios_version>88.00.98.00.01</vbios_version>
        <multigpu_board>No</multigpu_board>
        <board_id>0x12</board_id>
        <gpu_part_number>900-2G500-0040-000</gpu_part_number>
        <inforom_version>
            <img_version>G500.0212.00.02</img_version>
            <oem_object>1.1</oem_object>
            <ecc_object>5.0</ecc_object>
            <pwr_object>N/A</pwr_object>
        </inforom_version>
        <gpu_operation_mode>
            <current_gom>N/A</current_gom>
            <pending_gom>N/A</pending_gom>
        </gpu_operation_mode>
        <gpu_virtualization_mode>
            <virtualization_mode>Pass-Through</virtualization_mode>
            <host_vgpu_mode>N/A</host_vgpu_mode>
        </gpu_virtualization_mode>
        <ibmnpu>
            <relaxed_ordering_mode>N/A</relaxed_ordering_mode>
        </ibmnpu>
        <pci>
            <pci_bus>00</pci_bus>
            <pci_device>12</pci_device>
            <pci_domain>0000</pci_domain>
            <pci_device_id>1DF610DE</pci_device_id>
            <pci_bus_id>00000000:00:12.0</pci_bus_id>
            <pci_sub_system_id>13D610DE</pci_sub_system_id>
            <pci_gpu_link_info>
                <pcie_gen>
                    <max_link_gen>3</max_link_gen>
                    <current_link_gen>3</current_link_gen>
                </pcie_gen>
                <link_widths>
                    <max_link_width>16x</max_link_width>
                    <current_link_width>16x</current_link_width>
                </link_widths>
            </pci_gpu_link_info>
            <pci_bridge_chip>
                <bridge_chip_type>N/A</bridge_chip_type>
                <bridge_chip_fw>N/A</bridge_chip_fw>
            </pci_bridge_chip>
            <replay_counter>0</replay_counter>
            <replay_rollover_counter>0</replay_rollover_counter>
            <tx_util>7151000 KB/s</tx_util>
            <rx_util>8083000 KB/s</rx_util>
        </pci>
        <fan_speed>N/A</fan_speed>
        <performance_state>P0</performance_state>
        <clocks_throttle_reasons>
            <clocks_throttle_reason_gpu_idle>Not Active</clocks_throttle_reason_gpu_idle>
            <clocks_throttle_reason_applications_clocks_setting>Not Active</clocks_throttle_reason_applications_clocks_setting>
            <clocks_throttle_reason_sw_power_cap>Not Active</clocks_throttle_reason_sw_power_cap>
            <clocks_throttle_reason_hw_slowdown>Not Active</clocks_throttle_reason_hw_slowdown>
            <clocks_throttle_reason_hw_thermal_slowdown>Not Active</clocks_throttle_reason_hw_thermal_slowdown>
            <clocks_throttle_reason_hw_power_brake_slowdown>Not Active</clocks_throttle_reason_hw_power_brake_slowdown>
            <clocks_throttle_reason_sync_boost>Not Active</clocks_throttle_reason_sync_boost>
            <clocks_throttle_reason_sw_thermal_slowdown>Not Active</clocks_throttle_reason_sw_thermal_slowdown>
            <clocks_throttle_reason_display_clocks_setting>Not Active</clocks_throttle_reason_display_clocks_setting>
        </clocks_throttle_reasons>
        <fb_memory_usage>
            <total>32510 MiB</total>
            <used>32141 MiB</used>
            <free>369 MiB</free>
        </fb_memory_usage>
        <bar1_memory_usage>
            <total>32768 MiB</total>
            <used>32 MiB</used>
            <free>32736 MiB</free>
        </bar1_memory_usage>
        <compute_mode>Default</compute_mode>
        <utilization>
            <gpu_util>98 %</gpu_util>
            <memory_util>50 %</memory_util>
            <encoder_util>0 %</encoder_util>
            <decoder_util>0 %</decoder_util>
        </utilization>
        <encoder_stats>
            <session_count>0</session_count>
            <average_fps>0</average_fps>
            <average_latency>0</average_latency>
        </encoder_stats>
        <fbc_stats>
            <session_count>0</session_count>
            <average_fps>0</average_fps>
            <average_latency>0</average_latency>
        </fbc_stats>
        <ecc_mode>
            <current_ecc>Enabled</current_ecc>
            <pending_ecc>Enabled</pending_ecc>
        </ecc_mode>
        <ecc_errors>
            <volatile>
                <single_bit>
                    <device_memory>0</device_memory>
                    <register_file>0</register_file>
                    <l1_cache>0</l1_cache>
                    <l2_cache>0</l2_cache>
                    <texture_memory>N/A</texture_memory>
                    <texture_shm>N/A</texture_shm>
                    <cbu>N/A</cbu>
                    <total>0</total>
                </single_bit>
                <double_bit>
                    <device_memory>0</device_memory>
                    <register_file>0</register_file>
                    <l1_cache>0</l1_cache>
                    <l2_cache>0</l2_cache>
                    <texture_memory>N/A</texture_memory>
                    <texture_shm>N/A</texture_shm>
                    <cbu>0</cbu>
                    <total>0</total>
                </double_bit>
            </volatile>
            <aggregate>
                <single_bit>
                    <device_memory>0</device_memory>
                    <register_file>0</register_file>
                    <l1_cache>0</l1_cache>
                    <l2_cache>0</l2_cache>
                    <texture_memory>N/A</texture_memory>
                    <texture_shm>N/A</texture_shm>
                    <cbu>N/A</cbu>
                    <total>0</total>
                </single_bit>
                <double_bit>
                    <device_memory>0</device_memory>
                    <register_file>0</register_file>
                    <l1_cache>0</l1_cache>
                    <l2_cache>0</l2_cache>
                    <texture_memory>N/A</texture_memory>
                    <texture_shm>N/A</texture_shm>
                    <cbu>0</cbu>
                    <total>0</total>
                </double_bit>
            </aggregate>
        </ecc_errors>
        <retired_pages>
            <multiple_single_bit_retirement>
                <retired_count>0</retired_count>
                <retired_pagelist>
                </retired_pagelist>
            </multiple_single_bit_retirement>
            <double_bit_retirement>
                <retired_count>0</retired_count>
                <retired_pagelist>
                </retired_pagelist>
            </double_bit_retirement>
            <pending_blacklist>No</pending_blacklist>
            <pending_retirement>No</pending_retirement>
        </retired_pages>
        <remapped_rows>N/A</remapped_rows>
        <temperature>
            <gpu_temp>64 C</gpu_temp>
            <gpu_temp_max_threshold>90 C</gpu_temp_max_threshold>
            <gpu_temp_slow_threshold>87 C</gpu_temp_slow_threshold>
            <gpu_temp_max_gpu_threshold>83 C</gpu_temp_max_gpu_threshold>
            <gpu_target_temperature>N/A</gpu_target_temperature>
            <memory_temp>62 C</memory_temp>
            <gpu_temp_max_mem_threshold>85 C</gpu_temp_max_mem_threshold>
        </temperature>
        <supported_gpu_target_temp>
            <gpu_target_temp_min>N/A</gpu_target_temp_min>
            <gpu_target_temp_max>N/A</gpu_target_temp_max>
        </supported_gpu_target_temp>
        <power_readings>
            <power_state>P0</power_state>
            <power_management>Supported</power_management>
            <power_draw>85.80 W</power_draw>
            <power_limit>250.00 W</power_limit>
            <default_power_limit>250.00 W</default_power_limit>
            <enforced_power_limit>250.00 W</enforced_power_limit>
            <min_power_limit>100.00 W</min_power_limit>
            <max_power_limit>250.00 W</max_power_limit>
        </power_readings>
        <clocks>
            <graphics_clock>1590 MHz</graphics_clock>
            <sm_clock>1590 MHz</sm_clock>
            <mem_clock>1107 MHz</mem_clock>
            <video_clock>1425 MHz</video_clock>
        </clocks>
        <applications_clocks>
            <graphics_clock>1245 MHz</graphics_clock>
            <mem_clock>1107 MHz</mem_clock>
        </applications_clocks>
        <default_applications_clocks>
            <graphics_clock>1245 MHz</graphics_clock>
            <mem_clock>1107 MHz</mem_clock>
        </default_applications_clocks>
        <max_clocks>
            <graphics_clock>1597 MHz</graphics_clock>
            <sm_clock>1597 MHz</sm_clock>
            <mem_clock>1107 MHz</mem_clock>
            <video_clock>1432 MHz</video_clock>
        </max_clocks>
        <max_customer_boost_clocks>
            <graphics_clock>1597 MHz</graphics_clock>
        </max_customer_boost_clocks>
        <clock_policy>
            <auto_boost>N/A</auto_boost>
            <auto_boost_default>N/A</auto_boost_default>
        </clock_policy>
        <supported_clocks>
            <supported_mem_clock>
                <value>1107 MHz</value>
                <supported_graphics_clock>1597 MHz</supported_graphics_clock>
                <supported_graphics_clock>1590 MHz</supported_graphics_clock>
                <supported_graphics_clock>1582 MHz</supported_graphics_clock>
                <supported_graphics_clock>1575 MHz</supported_graphics_clock>
                <supported_graphics_clock>1567 MHz</supported_graphics_clock>
                <supported_graphics_clock>1560 MHz</supported_graphics_clock>
                <supported_graphics_clock>1552 MHz</supported_graphics_clock>
                <supported_graphics_clock>1545 MHz</supported_graphics_clock>
                <supported_graphics_clock>1537 MHz</supported_graphics_clock>
                <supported_graphics_clock>1530 MHz</supported_graphics_clock>
                <supported_graphics_clock>1522 MHz</supported_graphics_clock>
                <supported_graphics_clock>1515 MHz</supported_graphics_clock>
                <supported_graphics_clock>1507 MHz</supported_graphics_clock>
                <supported_graphics_clock>1500 MHz</supported_graphics_clock>
                <supported_graphics_clock>1492 MHz</supported_graphics_clock>
                <supported_graphics_clock>1485 MHz</supported_graphics_clock>
                <supported_graphics_clock>1477 MHz</supported_graphics_clock>
                <supported_graphics_clock>1470 MHz</supported_graphics_clock>
                <supported_graphics_clock>1462 MHz</supported_graphics_clock>
                <supported_graphics_clock>1455 MHz</supported_graphics_clock>
                <supported_graphics_clock>1447 MHz</supported_graphics_clock>
                <supported_graphics_clock>1440 MHz</supported_graphics_clock>
                <supported_graphics_clock>1432 MHz</supported_graphics_clock>
                <supported_graphics_clock>1425 MHz</supported_graphics_clock>
                <supported_graphics_clock>1417 MHz</supported_graphics_clock>
                <supported_graphics_clock>1410 MHz</supported_graphics_clock>
                <supported_graphics_clock>1402 MHz</supported_graphics_clock>
                <supported_graphics_clock>1395 MHz</supported_graphics_clock>
                <supported_graphics_clock>1387 MHz</supported_graphics_clock>
                <supported_graphics_clock>1380 MHz</supported_graphics_clock>
                <supported_graphics_clock>1372 MHz</supported_graphics_clock>
                <supported_graphics_clock>1365 MHz</supported_graphics_clock>
                <supported_graphics_clock>1357 MHz</supported_graphics_clock>
                <supported_graphics_clock>1350 MHz</supported_graphics_clock>
                <supported_graphics_clock>1342 MHz</supported_graphics_clock>
                <supported_graphics_clock>1335 MHz</supported_graphics_clock>
                <supported_graphics_clock>1327 MHz</supported_graphics_clock>
                <supported_graphics_clock>1320 MHz</supported_graphics_clock>
                <supported_graphics_clock>1312 MHz</supported_graphics_clock>
                <supported_graphics_clock>1305 MHz</supported_graphics_clock>
                <supported_graphics_clock>1297 MHz</supported_graphics_clock>
                <supported_graphics_clock>1290 MHz</supported_graphics_clock>
                <supported_graphics_clock>1282 MHz</supported_graphics_clock>
                <supported_graphics_clock>1275 MHz</supported_graphics_clock>
                <supported_graphics_clock>1267 MHz</supported_graphics_clock>
                <supported_graphics_clock>1260 MHz</supported_graphics_clock>
                <supported_graphics_clock>1252 MHz</supported_graphics_clock>
                <supported_graphics_clock>1245 MHz</supported_graphics_clock>
                <supported_graphics_clock>1237 MHz</supported_graphics_clock>
                <supported_graphics_clock>1230 MHz</supported_graphics_clock>
                <supported_graphics_clock>1222 MHz</supported_graphics_clock>
                <supported_graphics_clock>1215 MHz</supported_graphics_clock>
                <supported_graphics_clock>1207 MHz</supported_graphics_clock>
                <supported_graphics_clock>1200 MHz</supported_graphics_clock>
                <supported_graphics_clock>1192 MHz</supported_graphics_clock>
                <supported_graphics_clock>1185 MHz</supported_graphics_clock>
                <supported_graphics_clock>1177 MHz</supported_graphics_clock>
                <supported_graphics_clock>1170 MHz</supported_graphics_clock>
                <supported_graphics_clock>1162 MHz</supported_graphics_clock>
                <supported_graphics_clock>1155 MHz</supported_graphics_clock>
                <supported_graphics_clock>1147 MHz</supported_graphics_clock>
                <supported_graphics_clock>1140 MHz</supported_graphics_clock>
                <supported_graphics_clock>1132 MHz</supported_graphics_clock>
                <supported_graphics_clock>1125 MHz</supported_graphics_clock>
                <supported_graphics_clock>1117 MHz</supported_graphics_clock>
                <supported_graphics_clock>1110 MHz</supported_graphics_clock>
                <supported_graphics_clock>1102 MHz</supported_graphics_clock>
                <supported_graphics_clock>1095 MHz</supported_graphics_clock>
                <supported_graphics_clock>1087 MHz</supported_graphics_clock>
                <supported_graphics_clock>1080 MHz</supported_graphics_clock>
                <supported_graphics_clock>1072 MHz</supported_graphics_clock>
                <supported_graphics_clock>1065 MHz</supported_graphics_clock>
                <supported_graphics_clock>1057 MHz</supported_graphics_clock>
                <supported_graphics_clock>1050 MHz</supported_graphics_clock>
                <supported_graphics_clock>1042 MHz</supported_graphics_clock>
                <supported_graphics_clock>1035 MHz</supported_graphics_clock>
                <supported_graphics_clock>1027 MHz</supported_graphics_clock>
                <supported_graphics_clock>1020 MHz</supported_graphics_clock>
                <supported_graphics_clock>1012 MHz</supported_graphics_clock>
                <supported_graphics_clock>1005 MHz</supported_graphics_clock>
                <supported_graphics_clock>997 MHz</supported_graphics_clock>
                <supported_graphics_clock>990 MHz</supported_graphics_clock>
                <supported_graphics_clock>982 MHz</supported_graphics_clock>
                <supported_graphics_clock>975 MHz</supported_graphics_clock>
                <supported_graphics_clock>967 MHz</supported_graphics_clock>
                <supported_graphics_clock>960 MHz</supported_graphics_clock>
                <supported_graphics_clock>952 MHz</supported_graphics_clock>
                <supported_graphics_clock>945 MHz</supported_graphics_clock>
                <supported_graphics_clock>937 MHz</supported_graphics_clock>
                <supported_graphics_clock>930 MHz</supported_graphics_clock>
                <supported_graphics_clock>922 MHz</supported_graphics_clock>
                <supported_graphics_clock>915 MHz</supported_graphics_clock>
                <supported_graphics_clock>907 MHz</supported_graphics_clock>
                <supported_graphics_clock>900 MHz</supported_graphics_clock>
                <supported_graphics_clock>892 MHz</supported_graphics_clock>
                <supported_graphics_clock>885 MHz</supported_graphics_clock>
                <supported_graphics_clock>877 MHz</supported_graphics_clock>
                <supported_graphics_clock>870 MHz</supported_graphics_clock>
                <supported_graphics_clock>862 MHz</supported_graphics_clock>
                <supported_graphics_clock>855 MHz</supported_graphics_clock>
                <supported_graphics_clock>847 MHz</supported_graphics_clock>
                <supported_graphics_clock>840 MHz</supported_graphics_clock>
                <supported_graphics_clock>832 MHz</supported_graphics_clock>
                <supported_graphics_clock>825 MHz</supported_graphics_clock>
                <supported_graphics_clock>817 MHz</supported_graphics_clock>
                <supported_graphics_clock>810 MHz</supported_graphics_clock>
                <supported_graphics_clock>802 MHz</supported_graphics_clock>
                <supported_graphics_clock>795 MHz</supported_graphics_clock>
                <supported_graphics_clock>787 MHz</supported_graphics_clock>
                <supported_graphics_clock>780 MHz</supported_graphics_clock>
                <supported_graphics_clock>772 MHz</supported_graphics_clock>
                <supported_graphics_clock>765 MHz</supported_graphics_clock>
                <supported_graphics_clock>757 MHz</supported_graphics_clock>
                <supported_graphics_clock>750 MHz</supported_graphics_clock>
                <supported_graphics_clock>742 MHz</supported_graphics_clock>
                <supported_graphics_clock>735 MHz</supported_graphics_clock>
                <supported_graphics_clock>727 MHz</supported_graphics_clock>
                <supported_graphics_clock>720 MHz</supported_graphics_clock>
                <supported_graphics_clock>712 MHz</supported_graphics_clock>
                <supported_graphics_clock>705 MHz</supported_graphics_clock>
                <supported_graphics_clock>697 MHz</supported_graphics_clock>
                <supported_graphics_clock>690 MHz</supported_graphics_clock>
                <supported_graphics_clock>682 MHz</supported_graphics_clock>
                <supported_graphics_clock>675 MHz</supported_graphics_clock>
                <supported_graphics_clock>667 MHz</supported_graphics_clock>
                <supported_graphics_clock>660 MHz</supported_graphics_clock>
                <supported_graphics_clock>652 MHz</supported_graphics_clock>
                <supported_graphics_clock>645 MHz</supported_graphics_clock>
                <supported_graphics_clock>637 MHz</supported_graphics_clock>
                <supported_graphics_clock>630 MHz</supported_graphics_clock>
                <supported_graphics_clock>622 MHz</supported_graphics_clock>
                <supported_graphics_clock>615 MHz</supported_graphics_clock>
                <supported_graphics_clock>607 MHz</supported_graphics_clock>
                <supported_graphics_clock>600 MHz</supported_graphics_clock>
                <supported_graphics_clock>592 MHz</supported_graphics_clock>
                <supported_graphics_clock>585 MHz</supported_graphics_clock>
                <supported_graphics_clock>577 MHz</supported_graphics_clock>
                <supported_graphics_clock>570 MHz</supported_graphics_clock>
                <supported_graphics_clock>562 MHz</supported_graphics_clock>
                <supported_graphics_clock>555 MHz</supported_graphics_clock>
                <supported_graphics_clock>547 MHz</supported_graphics_clock>
                <supported_graphics_clock>540 MHz</supported_graphics_clock>
                <supported_graphics_clock>532 MHz</supported_graphics_clock>
                <supported_graphics_clock>525 MHz</supported_graphics_clock>
                <supported_graphics_clock>517 MHz</supported_graphics_clock>
                <supported_graphics_clock>510 MHz</supported_graphics_clock>
                <supported_graphics_clock>502 MHz</supported_graphics_clock>
                <supported_graphics_clock>495 MHz</supported_graphics_clock>
                <supported_graphics_clock>487 MHz</supported_graphics_clock>
                <supported_graphics_clock>480 MHz</supported_graphics_clock>
                <supported_graphics_clock>472 MHz</supported_graphics_clock>
                <supported_graphics_clock>465 MHz</supported_graphics_clock>
                <supported_graphics_clock>457 MHz</supported_graphics_clock>
                <supported_graphics_clock>450 MHz</supported_graphics_clock>
                <supported_graphics_clock>442 MHz</supported_graphics_clock>
                <supported_graphics_clock>435 MHz</supported_graphics_clock>
                <supported_graphics_clock>427 MHz</supported_graphics_clock>
                <supported_graphics_clock>420 MHz</supported_graphics_clock>
                <supported_graphics_clock>412 MHz</supported_graphics_clock>
                <supported_graphics_clock>405 MHz</supported_graphics_clock>
                <supported_graphics_clock>397 MHz</supported_graphics_clock>
                <supported_graphics_clock>390 MHz</supported_graphics_clock>
                <supported_graphics_clock>382 MHz</supported_graphics_clock>
                <supported_graphics_clock>375 MHz</supported_graphics_clock>
                <supported_graphics_clock>367 MHz</supported_graphics_clock>
                <supported_graphics_clock>360 MHz</supported_graphics_clock>
                <supported_graphics_clock>352 MHz</supported_graphics_clock>
                <supported_graphics_clock>345 MHz</supported_graphics_clock>
                <supported_graphics_clock>337 MHz</supported_graphics_clock>
                <supported_graphics_clock>330 MHz</supported_graphics_clock>
                <supported_graphics_clock>322 MHz</supported_graphics_clock>
                <supported_graphics_clock>315 MHz</supported_graphics_clock>
                <supported_graphics_clock>307 MHz</supported_graphics_clock>
                <supported_graphics_clock>300 MHz</supported_graphics_clock>
                <supported_graphics_clock>292 MHz</supported_graphics_clock>
                <supported_graphics_clock>285 MHz</supported_graphics_clock>
                <supported_graphics_clock>277 MHz</supported_graphics_clock>
                <supported_graphics_clock>270 MHz</supported_graphics_clock>
                <supported_graphics_clock>262 MHz</supported_graphics_clock>
                <supported_graphics_clock>255 MHz</supported_graphics_clock>
                <supported_graphics_clock>247 MHz</supported_graphics_clock>
                <supported_graphics_clock>240 MHz</supported_graphics_clock>
                <supported_graphics_clock>232 MHz</supported_graphics_clock>
                <supported_graphics_clock>225 MHz</supported_graphics_clock>
                <supported_graphics_clock>217 MHz</supported_graphics_clock>
                <supported_graphics_clock>210 MHz</supported_graphics_clock>
                <supported_graphics_clock>202 MHz</supported_graphics_clock>
                <supported_graphics_clock>195 MHz</supported_graphics_clock>
                <supported_graphics_clock>187 MHz</supported_graphics_clock>
                <supported_graphics_clock>180 MHz</supported_graphics_clock>
                <supported_graphics_clock>172 MHz</supported_graphics_clock>
                <supported_graphics_clock>165 MHz</supported_graphics_clock>
                <supported_graphics_clock>157 MHz</supported_graphics_clock>
                <supported_graphics_clock>150 MHz</supported_graphics_clock>
                <supported_graphics_clock>142 MHz</supported_graphics_clock>
                <supported_graphics_clock>135 MHz</supported_graphics_clock>
            </supported_mem_clock>
        </supported_clocks>
        <processes>
            <process_info>
                <gpu_instance_id>N/A</gpu_instance_id>
                <compute_instance_id>N/A</compute_instance_id>
                <pid>37305</pid>
                <type>C</type>
                <process_name>python</process_name>
                <used_memory>32135 MiB</used_memory>
            </process_info>
        </processes>
        <accounted_processes>
            <accounted_process_info>
                <pid>54796</pid>
                <gpu_util>0 %</gpu_util>
                <memory_util>0 %</memory_util>
                <max_memory_usage>307 MiB</max_memory_usage>
                <time>18050 ms</time>
                <is_running>0</is_running>
            </accounted_process_info>
            <accounted_process_info>
                <pid>57134</pid>
                <gpu_util>0 %</gpu_util>
                <memory_util>0 %</memory_util>
                <max_memory_usage>32015 MiB</max_memory_usage>
                <time>4908121 ms</time>
                <is_running>0</is_running>
            </accounted_process_info>
            <accounted_process_info>
                <pid>64749</pid>
                <gpu_util>1 %</gpu_util>
                <memory_util>0 %</memory_util>
                <max_memory_usage>31973 MiB</max_memory_usage>
                <time>188206 ms</time>
                <is_running>0</is_running>
            </accounted_process_info>
            <accounted_process_info>
                <pid>2956</pid>
                <gpu_util>54 %</gpu_util>
                <memory_util>18 %</memory_util>
                <max_memory_usage>28999 MiB</max_memory_usage>
                <time>313868111 ms</time>
                <is_running>0</is_running>
            </accounted_process_info>
            <accounted_process_info>
                <pid>45030</pid>
                <gpu_util>0 %</gpu_util>
                <memory_util>0 %</memory_util>
                <max_memory_usage>31641 MiB</max_memory_usage>
                <time>16828657 ms</time>
                <is_running>0</is_running>
            </accounted_process_info>
            <accounted_process_info>
                <pid>37305</pid>
                <gpu_util>58 %</gpu_util>
                <memory_util>22 %</memory_util>
                <max_memory_usage>32175 MiB</max_memory_usage>
                <time>0 ms</time>
                <is_running>1</is_running>
            </accounted_process_info>
        </accounted_processes>
    </gpu>

    <gpu id="00000000:00:13.0">
        <product_name>Tesla V100S-PCIE-32GB</product_name>
        <product_brand>Tesla</product_brand>
        <display_mode>Enabled</display_mode>
        <display_active>Disabled</display_active>
        <persistence_mode>Disabled</persistence_mode>
        <mig_mode>
            <current_mig>N/A</current_mig>
            <pending_mig>N/A</pending_mig>
        </mig_mode>
        <mig_devices>
            None
        </mig_devices>
        <accounting_mode>Enabled</accounting_mode>
        <accounting_mode_buffer_size>4000</accounting_mode_buffer_size>
        <driver_model>
            <current_dm>N/A</current_dm>
            <pending_dm>N/A</pending_dm>
        </driver_model>
        <serial>1562920007403</serial>
        <uuid>GPU-d2446f91-171e-1b54-515d-9d9b03484ddf</uuid>
        <minor_number>6</minor_number>
        <vbios_version>88.00.98.00.01</vbios_version>
        <multigpu_board>No</multigpu_board>
        <board_id>0x13</board_id>
        <gpu_part_number>900-2G500-0040-000</gpu_part_number>
        <inforom_version>
            <img_version>G500.0212.00.02</img_version>
            <oem_object>1.1</oem_object>
            <ecc_object>5.0</ecc_object>
            <pwr_object>N/A</pwr_object>
        </inforom_version>
        <gpu_operation_mode>
            <current_gom>N/A</current_gom>
            <pending_gom>N/A</pending_gom>
        </gpu_operation_mode>
        <gpu_virtualization_mode>
            <virtualization_mode>Pass-Through</virtualization_mode>
            <host_vgpu_mode>N/A</host_vgpu_mode>
        </gpu_virtualization_mode>
        <ibmnpu>
            <relaxed_ordering_mode>N/A</relaxed_ordering_mode>
        </ibmnpu>
        <pci>
            <pci_bus>00</pci_bus>
            <pci_device>13</pci_device>
            <pci_domain>0000</pci_domain>
            <pci_device_id>1DF610DE</pci_device_id>
            <pci_bus_id>00000000:00:13.0</pci_bus_id>
            <pci_sub_system_id>13D610DE</pci_sub_system_id>
            <pci_gpu_link_info>
                <pcie_gen>
                    <max_link_gen>3</max_link_gen>
                    <current_link_gen>3</current_link_gen>
                </pcie_gen>
                <link_widths>
                    <max_link_width>16x</max_link_width>
                    <current_link_width>16x</current_link_width>
                </link_widths>
            </pci_gpu_link_info>
            <pci_bridge_chip>
                <bridge_chip_type>N/A</bridge_chip_type>
                <bridge_chip_fw>N/A</bridge_chip_fw>
            </pci_bridge_chip>
            <replay_counter>0</replay_counter>
            <replay_rollover_counter>0</replay_rollover_counter>
            <tx_util>384000 KB/s</tx_util>
            <rx_util>4695000 KB/s</rx_util>
        </pci>
        <fan_speed>N/A</fan_speed>
        <performance_state>P0</performance_state>
        <clocks_throttle_reasons>
            <clocks_throttle_reason_gpu_idle>Not Active</clocks_throttle_reason_gpu_idle>
            <clocks_throttle_reason_applications_clocks_setting>Not Active</clocks_throttle_reason_applications_clocks_setting>
            <clocks_throttle_reason_sw_power_cap>Active</clocks_throttle_reason_sw_power_cap>
            <clocks_throttle_reason_hw_slowdown>Not Active</clocks_throttle_reason_hw_slowdown>
            <clocks_throttle_reason_hw_thermal_slowdown>Not Active</clocks_throttle_reason_hw_thermal_slowdown>
            <clocks_throttle_reason_hw_power_brake_slowdown>Not Active</clocks_throttle_reason_hw_power_brake_slowdown>
            <clocks_throttle_reason_sync_boost>Not Active</clocks_throttle_reason_sync_boost>
            <clocks_throttle_reason_sw_thermal_slowdown>Not Active</clocks_throttle_reason_sw_thermal_slowdown>
            <clocks_throttle_reason_display_clocks_setting>Not Active</clocks_throttle_reason_display_clocks_setting>
        </clocks_throttle_reasons>
        <fb_memory_usage>
            <total>32510 MiB</total>
            <used>32157 MiB</used>
            <free>353 MiB</free>
        </fb_memory_usage>
        <bar1_memory_usage>
            <total>32768 MiB</total>
            <used>32 MiB</used>
            <free>32736 MiB</free>
        </bar1_memory_usage>
        <compute_mode>Default</compute_mode>
        <utilization>
            <gpu_util>91 %</gpu_util>
            <memory_util>22 %</memory_util>
            <encoder_util>0 %</encoder_util>
            <decoder_util>0 %</decoder_util>
        </utilization>
        <encoder_stats>
            <session_count>0</session_count>
            <average_fps>0</average_fps>
            <average_latency>0</average_latency>
        </encoder_stats>
        <fbc_stats>
            <session_count>0</session_count>
            <average_fps>0</average_fps>
            <average_latency>0</average_latency>
        </fbc_stats>
        <ecc_mode>
            <current_ecc>Enabled</current_ecc>
            <pending_ecc>Enabled</pending_ecc>
        </ecc_mode>
        <ecc_errors>
            <volatile>
                <single_bit>
                    <device_memory>0</device_memory>
                    <register_file>0</register_file>
                    <l1_cache>0</l1_cache>
                    <l2_cache>0</l2_cache>
                    <texture_memory>N/A</texture_memory>
                    <texture_shm>N/A</texture_shm>
                    <cbu>N/A</cbu>
                    <total>0</total>
                </single_bit>
                <double_bit>
                    <device_memory>0</device_memory>
                    <register_file>0</register_file>
                    <l1_cache>0</l1_cache>
                    <l2_cache>0</l2_cache>
                    <texture_memory>N/A</texture_memory>
                    <texture_shm>N/A</texture_shm>
                    <cbu>0</cbu>
                    <total>0</total>
                </double_bit>
            </volatile>
            <aggregate>
                <single_bit>
                    <device_memory>0</device_memory>
                    <register_file>0</register_file>
                    <l1_cache>0</l1_cache>
                    <l2_cache>0</l2_cache>
                    <texture_memory>N/A</texture_memory>
                    <texture_shm>N/A</texture_shm>
                    <cbu>N/A</cbu>
                    <total>0</total>
                </single_bit>
                <double_bit>
                    <device_memory>0</device_memory>
                    <register_file>0</register_file>
                    <l1_cache>0</l1_cache>
                    <l2_cache>0</l2_cache>
                    <texture_memory>N/A</texture_memory>
                    <texture_shm>N/A</texture_shm>
                    <cbu>0</cbu>
                    <total>0</total>
                </double_bit>
            </aggregate>
        </ecc_errors>
        <retired_pages>
            <multiple_single_bit_retirement>
                <retired_count>0</retired_count>
                <retired_pagelist>
                </retired_pagelist>
            </multiple_single_bit_retirement>
            <double_bit_retirement>
                <retired_count>0</retired_count>
                <retired_pagelist>
                </retired_pagelist>
            </double_bit_retirement>
            <pending_blacklist>No</pending_blacklist>
            <pending_retirement>No</pending_retirement>
        </retired_pages>
        <remapped_rows>N/A</remapped_rows>
        <temperature>
            <gpu_temp>61 C</gpu_temp>
            <gpu_temp_max_threshold>90 C</gpu_temp_max_threshold>
            <gpu_temp_slow_threshold>87 C</gpu_temp_slow_threshold>
            <gpu_temp_max_gpu_threshold>83 C</gpu_temp_max_gpu_threshold>
            <gpu_target_temperature>N/A</gpu_target_temperature>
            <memory_temp>56 C</memory_temp>
            <gpu_temp_max_mem_threshold>85 C</gpu_temp_max_mem_threshold>
        </temperature>
        <supported_gpu_target_temp>
            <gpu_target_temp_min>N/A</gpu_target_temp_min>
            <gpu_target_temp_max>N/A</gpu_target_temp_max>
        </supported_gpu_target_temp>
        <power_readings>
            <power_state>P0</power_state>
            <power_management>Supported</power_management>
            <power_draw>68.14 W</power_draw>
            <power_limit>250.00 W</power_limit>
            <default_power_limit>250.00 W</default_power_limit>
            <enforced_power_limit>250.00 W</enforced_power_limit>
            <min_power_limit>100.00 W</min_power_limit>
            <max_power_limit>250.00 W</max_power_limit>
        </power_readings>
        <clocks>
            <graphics_clock>1597 MHz</graphics_clock>
            <sm_clock>1597 MHz</sm_clock>
            <mem_clock>1107 MHz</mem_clock>
            <video_clock>1432 MHz</video_clock>
        </clocks>
        <applications_clocks>
            <graphics_clock>1245 MHz</graphics_clock>
            <mem_clock>1107 MHz</mem_clock>
        </applications_clocks>
        <default_applications_clocks>
            <graphics_clock>1245 MHz</graphics_clock>
            <mem_clock>1107 MHz</mem_clock>
        </default_applications_clocks>
        <max_clocks>
            <graphics_clock>1597 MHz</graphics_clock>
            <sm_clock>1597 MHz</sm_clock>
            <mem_clock>1107 MHz</mem_clock>
            <video_clock>1432 MHz</video_clock>
        </max_clocks>
        <max_customer_boost_clocks>
            <graphics_clock>1597 MHz</graphics_clock>
        </max_customer_boost_clocks>
        <clock_policy>
            <auto_boost>N/A</auto_boost>
            <auto_boost_default>N/A</auto_boost_default>
        </clock_policy>
        <supported_clocks>
            <supported_mem_clock>
                <value>1107 MHz</value>
                <supported_graphics_clock>1597 MHz</supported_graphics_clock>
                <supported_graphics_clock>1590 MHz</supported_graphics_clock>
                <supported_graphics_clock>1582 MHz</supported_graphics_clock>
                <supported_graphics_clock>1575 MHz</supported_graphics_clock>
                <supported_graphics_clock>1567 MHz</supported_graphics_clock>
                <supported_graphics_clock>1560 MHz</supported_graphics_clock>
                <supported_graphics_clock>1552 MHz</supported_graphics_clock>
                <supported_graphics_clock>1545 MHz</supported_graphics_clock>
                <supported_graphics_clock>1537 MHz</supported_graphics_clock>
                <supported_graphics_clock>1530 MHz</supported_graphics_clock>
                <supported_graphics_clock>1522 MHz</supported_graphics_clock>
                <supported_graphics_clock>1515 MHz</supported_graphics_clock>
                <supported_graphics_clock>1507 MHz</supported_graphics_clock>
                <supported_graphics_clock>1500 MHz</supported_graphics_clock>
                <supported_graphics_clock>1492 MHz</supported_graphics_clock>
                <supported_graphics_clock>1485 MHz</supported_graphics_clock>
                <supported_graphics_clock>1477 MHz</supported_graphics_clock>
                <supported_graphics_clock>1470 MHz</supported_graphics_clock>
                <supported_graphics_clock>1462 MHz</supported_graphics_clock>
                <supported_graphics_clock>1455 MHz</supported_graphics_clock>
                <supported_graphics_clock>1447 MHz</supported_graphics_clock>
                <supported_graphics_clock>1440 MHz</supported_graphics_clock>
                <supported_graphics_clock>1432 MHz</supported_graphics_clock>
                <supported_graphics_clock>1425 MHz</supported_graphics_clock>
                <supported_graphics_clock>1417 MHz</supported_graphics_clock>
                <supported_graphics_clock>1410 MHz</supported_graphics_clock>
                <supported_graphics_clock>1402 MHz</supported_graphics_clock>
                <supported_graphics_clock>1395 MHz</supported_graphics_clock>
                <supported_graphics_clock>1387 MHz</supported_graphics_clock>
                <supported_graphics_clock>1380 MHz</supported_graphics_clock>
                <supported_graphics_clock>1372 MHz</supported_graphics_clock>
                <supported_graphics_clock>1365 MHz</supported_graphics_clock>
                <supported_graphics_clock>1357 MHz</supported_graphics_clock>
                <supported_graphics_clock>1350 MHz</supported_graphics_clock>
                <supported_graphics_clock>1342 MHz</supported_graphics_clock>
                <supported_graphics_clock>1335 MHz</supported_graphics_clock>
                <supported_graphics_clock>1327 MHz</supported_graphics_clock>
                <supported_graphics_clock>1320 MHz</supported_graphics_clock>
                <supported_graphics_clock>1312 MHz</supported_graphics_clock>
                <supported_graphics_clock>1305 MHz</supported_graphics_clock>
                <supported_graphics_clock>1297 MHz</supported_graphics_clock>
                <supported_graphics_clock>1290 MHz</supported_graphics_clock>
                <supported_graphics_clock>1282 MHz</supported_graphics_clock>
                <supported_graphics_clock>1275 MHz</supported_graphics_clock>
                <supported_graphics_clock>1267 MHz</supported_graphics_clock>
                <supported_graphics_clock>1260 MHz</supported_graphics_clock>
                <supported_graphics_clock>1252 MHz</supported_graphics_clock>
                <supported_graphics_clock>1245 MHz</supported_graphics_clock>
                <supported_graphics_clock>1237 MHz</supported_graphics_clock>
                <supported_graphics_clock>1230 MHz</supported_graphics_clock>
                <supported_graphics_clock>1222 MHz</supported_graphics_clock>
                <supported_graphics_clock>1215 MHz</supported_graphics_clock>
                <supported_graphics_clock>1207 MHz</supported_graphics_clock>
                <supported_graphics_clock>1200 MHz</supported_graphics_clock>
                <supported_graphics_clock>1192 MHz</supported_graphics_clock>
                <supported_graphics_clock>1185 MHz</supported_graphics_clock>
                <supported_graphics_clock>1177 MHz</supported_graphics_clock>
                <supported_graphics_clock>1170 MHz</supported_graphics_clock>
                <supported_graphics_clock>1162 MHz</supported_graphics_clock>
                <supported_graphics_clock>1155 MHz</supported_graphics_clock>
                <supported_graphics_clock>1147 MHz</supported_graphics_clock>
                <supported_graphics_clock>1140 MHz</supported_graphics_clock>
                <supported_graphics_clock>1132 MHz</supported_graphics_clock>
                <supported_graphics_clock>1125 MHz</supported_graphics_clock>
                <supported_graphics_clock>1117 MHz</supported_graphics_clock>
                <supported_graphics_clock>1110 MHz</supported_graphics_clock>
                <supported_graphics_clock>1102 MHz</supported_graphics_clock>
                <supported_graphics_clock>1095 MHz</supported_graphics_clock>
                <supported_graphics_clock>1087 MHz</supported_graphics_clock>
                <supported_graphics_clock>1080 MHz</supported_graphics_clock>
                <supported_graphics_clock>1072 MHz</supported_graphics_clock>
                <supported_graphics_clock>1065 MHz</supported_graphics_clock>
                <supported_graphics_clock>1057 MHz</supported_graphics_clock>
                <supported_graphics_clock>1050 MHz</supported_graphics_clock>
                <supported_graphics_clock>1042 MHz</supported_graphics_clock>
                <supported_graphics_clock>1035 MHz</supported_graphics_clock>
                <supported_graphics_clock>1027 MHz</supported_graphics_clock>
                <supported_graphics_clock>1020 MHz</supported_graphics_clock>
                <supported_graphics_clock>1012 MHz</supported_graphics_clock>
                <supported_graphics_clock>1005 MHz</supported_graphics_clock>
                <supported_graphics_clock>997 MHz</supported_graphics_clock>
                <supported_graphics_clock>990 MHz</supported_graphics_clock>
                <supported_graphics_clock>982 MHz</supported_graphics_clock>
                <supported_graphics_clock>975 MHz</supported_graphics_clock>
                <supported_graphics_clock>967 MHz</supported_graphics_clock>
                <supported_graphics_clock>960 MHz</supported_graphics_clock>
                <supported_graphics_clock>952 MHz</supported_graphics_clock>
                <supported_graphics_clock>945 MHz</supported_graphics_clock>
                <supported_graphics_clock>937 MHz</supported_graphics_clock>
                <supported_graphics_clock>930 MHz</supported_graphics_clock>
                <supported_graphics_clock>922 MHz</supported_graphics_clock>
                <supported_graphics_clock>915 MHz</supported_graphics_clock>
                <supported_graphics_clock>907 MHz</supported_graphics_clock>
                <supported_graphics_clock>900 MHz</supported_graphics_clock>
                <supported_graphics_clock>892 MHz</supported_graphics_clock>
                <supported_graphics_clock>885 MHz</supported_graphics_clock>
                <supported_graphics_clock>877 MHz</supported_graphics_clock>
                <supported_graphics_clock>870 MHz</supported_graphics_clock>
                <supported_graphics_clock>862 MHz</supported_graphics_clock>
                <supported_graphics_clock>855 MHz</supported_graphics_clock>
                <supported_graphics_clock>847 MHz</supported_graphics_clock>
                <supported_graphics_clock>840 MHz</supported_graphics_clock>
                <supported_graphics_clock>832 MHz</supported_graphics_clock>
                <supported_graphics_clock>825 MHz</supported_graphics_clock>
                <supported_graphics_clock>817 MHz</supported_graphics_clock>
                <supported_graphics_clock>810 MHz</supported_graphics_clock>
                <supported_graphics_clock>802 MHz</supported_graphics_clock>
                <supported_graphics_clock>795 MHz</supported_graphics_clock>
                <supported_graphics_clock>787 MHz</supported_graphics_clock>
                <supported_graphics_clock>780 MHz</supported_graphics_clock>
                <supported_graphics_clock>772 MHz</supported_graphics_clock>
                <supported_graphics_clock>765 MHz</supported_graphics_clock>
                <supported_graphics_clock>757 MHz</supported_graphics_clock>
                <supported_graphics_clock>750 MHz</supported_graphics_clock>
                <supported_graphics_clock>742 MHz</supported_graphics_clock>
                <supported_graphics_clock>735 MHz</supported_graphics_clock>
                <supported_graphics_clock>727 MHz</supported_graphics_clock>
                <supported_graphics_clock>720 MHz</supported_graphics_clock>
                <supported_graphics_clock>712 MHz</supported_graphics_clock>
                <supported_graphics_clock>705 MHz</supported_graphics_clock>
                <supported_graphics_clock>697 MHz</supported_graphics_clock>
                <supported_graphics_clock>690 MHz</supported_graphics_clock>
                <supported_graphics_clock>682 MHz</supported_graphics_clock>
                <supported_graphics_clock>675 MHz</supported_graphics_clock>
                <supported_graphics_clock>667 MHz</supported_graphics_clock>
                <supported_graphics_clock>660 MHz</supported_graphics_clock>
                <supported_graphics_clock>652 MHz</supported_graphics_clock>
                <supported_graphics_clock>645 MHz</supported_graphics_clock>
                <supported_graphics_clock>637 MHz</supported_graphics_clock>
                <supported_graphics_clock>630 MHz</supported_graphics_clock>
                <supported_graphics_clock>622 MHz</supported_graphics_clock>
                <supported_graphics_clock>615 MHz</supported_graphics_clock>
                <supported_graphics_clock>607 MHz</supported_graphics_clock>
                <supported_graphics_clock>600 MHz</supported_graphics_clock>
                <supported_graphics_clock>592 MHz</supported_graphics_clock>
                <supported_graphics_clock>585 MHz</supported_graphics_clock>
                <supported_graphics_clock>577 MHz</supported_graphics_clock>
                <supported_graphics_clock>570 MHz</supported_graphics_clock>
                <supported_graphics_clock>562 MHz</supported_graphics_clock>
                <supported_graphics_clock>555 MHz</supported_graphics_clock>
                <supported_graphics_clock>547 MHz</supported_graphics_clock>
                <supported_graphics_clock>540 MHz</supported_graphics_clock>
                <supported_graphics_clock>532 MHz</supported_graphics_clock>
                <supported_graphics_clock>525 MHz</supported_graphics_clock>
                <supported_graphics_clock>517 MHz</supported_graphics_clock>
                <supported_graphics_clock>510 MHz</supported_graphics_clock>
                <supported_graphics_clock>502 MHz</supported_graphics_clock>
                <supported_graphics_clock>495 MHz</supported_graphics_clock>
                <supported_graphics_clock>487 MHz</supported_graphics_clock>
                <supported_graphics_clock>480 MHz</supported_graphics_clock>
                <supported_graphics_clock>472 MHz</supported_graphics_clock>
                <supported_graphics_clock>465 MHz</supported_graphics_clock>
                <supported_graphics_clock>457 MHz</supported_graphics_clock>
                <supported_graphics_clock>450 MHz</supported_graphics_clock>
                <supported_graphics_clock>442 MHz</supported_graphics_clock>
                <supported_graphics_clock>435 MHz</supported_graphics_clock>
                <supported_graphics_clock>427 MHz</supported_graphics_clock>
                <supported_graphics_clock>420 MHz</supported_graphics_clock>
                <supported_graphics_clock>412 MHz</supported_graphics_clock>
                <supported_graphics_clock>405 MHz</supported_graphics_clock>
                <supported_graphics_clock>397 MHz</supported_graphics_clock>
                <supported_graphics_clock>390 MHz</supported_graphics_clock>
                <supported_graphics_clock>382 MHz</supported_graphics_clock>
                <supported_graphics_clock>375 MHz</supported_graphics_clock>
                <supported_graphics_clock>367 MHz</supported_graphics_clock>
                <supported_graphics_clock>360 MHz</supported_graphics_clock>
                <supported_graphics_clock>352 MHz</supported_graphics_clock>
                <supported_graphics_clock>345 MHz</supported_graphics_clock>
                <supported_graphics_clock>337 MHz</supported_graphics_clock>
                <supported_graphics_clock>330 MHz</supported_graphics_clock>
                <supported_graphics_clock>322 MHz</supported_graphics_clock>
                <supported_graphics_clock>315 MHz</supported_graphics_clock>
                <supported_graphics_clock>307 MHz</supported_graphics_clock>
                <supported_graphics_clock>300 MHz</supported_graphics_clock>
                <supported_graphics_clock>292 MHz</supported_graphics_clock>
                <supported_graphics_clock>285 MHz</supported_graphics_clock>
                <supported_graphics_clock>277 MHz</supported_graphics_clock>
                <supported_graphics_clock>270 MHz</supported_graphics_clock>
                <supported_graphics_clock>262 MHz</supported_graphics_clock>
                <supported_graphics_clock>255 MHz</supported_graphics_clock>
                <supported_graphics_clock>247 MHz</supported_graphics_clock>
                <supported_graphics_clock>240 MHz</supported_graphics_clock>
                <supported_graphics_clock>232 MHz</supported_graphics_clock>
                <supported_graphics_clock>225 MHz</supported_graphics_clock>
                <supported_graphics_clock>217 MHz</supported_graphics_clock>
                <supported_graphics_clock>210 MHz</supported_graphics_clock>
                <supported_graphics_clock>202 MHz</supported_graphics_clock>
                <supported_graphics_clock>195 MHz</supported_graphics_clock>
                <supported_graphics_clock>187 MHz</supported_graphics_clock>
                <supported_graphics_clock>180 MHz</supported_graphics_clock>
                <supported_graphics_clock>172 MHz</supported_graphics_clock>
                <supported_graphics_clock>165 MHz</supported_graphics_clock>
                <supported_graphics_clock>157 MHz</supported_graphics_clock>
                <supported_graphics_clock>150 MHz</supported_graphics_clock>
                <supported_graphics_clock>142 MHz</supported_graphics_clock>
                <supported_graphics_clock>135 MHz</supported_graphics_clock>
            </supported_mem_clock>
        </supported_clocks>
        <processes>
            <process_info>
                <gpu_instance_id>N/A</gpu_instance_id>
                <compute_instance_id>N/A</compute_instance_id>
                <pid>37305</pid>
                <type>C</type>
                <process_name>python</process_name>
                <used_memory>32151 MiB</used_memory>
            </process_info>
        </processes>
        <accounted_processes>
            <accounted_process_info>
                <pid>54796</pid>
                <gpu_util>0 %</gpu_util>
                <memory_util>0 %</memory_util>
                <max_memory_usage>307 MiB</max_memory_usage>
                <time>17997 ms</time>
                <is_running>0</is_running>
            </accounted_process_info>
            <accounted_process_info>
                <pid>57134</pid>
                <gpu_util>0 %</gpu_util>
                <memory_util>0 %</memory_util>
                <max_memory_usage>32031 MiB</max_memory_usage>
                <time>4907990 ms</time>
                <is_running>0</is_running>
            </accounted_process_info>
            <accounted_process_info>
                <pid>64749</pid>
                <gpu_util>1 %</gpu_util>
                <memory_util>0 %</memory_util>
                <max_memory_usage>31959 MiB</max_memory_usage>
                <time>188084 ms</time>
                <is_running>0</is_running>
            </accounted_process_info>
            <accounted_process_info>
                <pid>2956</pid>
                <gpu_util>54 %</gpu_util>
                <memory_util>18 %</memory_util>
                <max_memory_usage>28315 MiB</max_memory_usage>
                <time>313867999 ms</time>
                <is_running>0</is_running>
            </accounted_process_info>
            <accounted_process_info>
                <pid>45030</pid>
                <gpu_util>0 %</gpu_util>
                <memory_util>0 %</memory_util>
                <max_memory_usage>31599 MiB</max_memory_usage>
                <time>16828510 ms</time>
                <is_running>0</is_running>
            </accounted_process_info>
            <accounted_process_info>
                <pid>37305</pid>
                <gpu_util>58 %</gpu_util>
                <memory_util>22 %</memory_util>
                <max_memory_usage>32151 MiB</max_memory_usage>
                <time>0 ms</time>
                <is_running>1</is_running>
            </accounted_process_info>
        </accounted_processes>
    </gpu>

    <gpu id="00000000:00:14.0">
        <product_name>Tesla V100S-PCIE-32GB</product_name>
        <product_brand>Tesla</product_brand>
        <display_mode>Enabled</display_mode>
        <display_active>Disabled</display_active>
        <persistence_mode>Disabled</persistence_mode>
        <mig_mode>
            <current_mig>N/A</current_mig>
            <pending_mig>N/A</pending_mig>
        </mig_mode>
        <mig_devices>
            None
        </mig_devices>
        <accounting_mode>Enabled</accounting_mode>
        <accounting_mode_buffer_size>4000</accounting_mode_buffer_size>
        <driver_model>
            <current_dm>N/A</current_dm>
            <pending_dm>N/A</pending_dm>
        </driver_model>
        <serial>1562920006372</serial>
        <uuid>GPU-1d955959-c95b-4f40-2cac-100d3006ffe2</uuid>
        <minor_number>7</minor_number>
        <vbios_version>88.00.98.00.01</vbios_version>
        <multigpu_board>No</multigpu_board>
        <board_id>0x14</board_id>
        <gpu_part_number>900-2G500-0040-000</gpu_part_number>
        <inforom_version>
            <img_version>G500.0212.00.02</img_version>
            <oem_object>1.1</oem_object>
            <ecc_object>5.0</ecc_object>
            <pwr_object>N/A</pwr_object>
        </inforom_version>
        <gpu_operation_mode>
            <current_gom>N/A</current_gom>
            <pending_gom>N/A</pending_gom>
        </gpu_operation_mode>
        <gpu_virtualization_mode>
            <virtualization_mode>Pass-Through</virtualization_mode>
            <host_vgpu_mode>N/A</host_vgpu_mode>
        </gpu_virtualization_mode>
        <ibmnpu>
            <relaxed_ordering_mode>N/A</relaxed_ordering_mode>
        </ibmnpu>
        <pci>
            <pci_bus>00</pci_bus>
            <pci_device>14</pci_device>
            <pci_domain>0000</pci_domain>
            <pci_device_id>1DF610DE</pci_device_id>
            <pci_bus_id>00000000:00:14.0</pci_bus_id>
            <pci_sub_system_id>13D610DE</pci_sub_system_id>
            <pci_gpu_link_info>
                <pcie_gen>
                    <max_link_gen>3</max_link_gen>
                    <current_link_gen>3</current_link_gen>
                </pcie_gen>
                <link_widths>
                    <max_link_width>16x</max_link_width>
                    <current_link_width>16x</current_link_width>
                </link_widths>
            </pci_gpu_link_info>
            <pci_bridge_chip>
                <bridge_chip_type>N/A</bridge_chip_type>
                <bridge_chip_fw>N/A</bridge_chip_fw>
            </pci_bridge_chip>
            <replay_counter>0</replay_counter>
            <replay_rollover_counter>0</replay_rollover_counter>
            <tx_util>0 KB/s</tx_util>
            <rx_util>3000 KB/s</rx_util>
        </pci>
        <fan_speed>N/A</fan_speed>
        <performance_state>P0</performance_state>
        <clocks_throttle_reasons>
            <clocks_throttle_reason_gpu_idle>Not Active</clocks_throttle_reason_gpu_idle>
            <clocks_throttle_reason_applications_clocks_setting>Not Active</clocks_throttle_reason_applications_clocks_setting>
            <clocks_throttle_reason_sw_power_cap>Not Active</clocks_throttle_reason_sw_power_cap>
            <clocks_throttle_reason_hw_slowdown>Not Active</clocks_throttle_reason_hw_slowdown>
            <clocks_throttle_reason_hw_thermal_slowdown>Not Active</clocks_throttle_reason_hw_thermal_slowdown>
            <clocks_throttle_reason_hw_power_brake_slowdown>Not Active</clocks_throttle_reason_hw_power_brake_slowdown>
            <clocks_throttle_reason_sync_boost>Not Active</clocks_throttle_reason_sync_boost>
            <clocks_throttle_reason_sw_thermal_slowdown>Not Active</clocks_throttle_reason_sw_thermal_slowdown>
            <clocks_throttle_reason_display_clocks_setting>Not Active</clocks_throttle_reason_display_clocks_setting>
        </clocks_throttle_reasons>
        <fb_memory_usage>
            <total>32510 MiB</total>
            <used>29501 MiB</used>
            <free>3009 MiB</free>
        </fb_memory_usage>
        <bar1_memory_usage>
            <total>32768 MiB</total>
            <used>32 MiB</used>
            <free>32736 MiB</free>
        </bar1_memory_usage>
        <compute_mode>Default</compute_mode>
        <utilization>
            <gpu_util>99 %</gpu_util>
            <memory_util>4 %</memory_util>
            <encoder_util>0 %</encoder_util>
            <decoder_util>0 %</decoder_util>
        </utilization>
        <encoder_stats>
            <session_count>0</session_count>
            <average_fps>0</average_fps>
            <average_latency>0</average_latency>
        </encoder_stats>
        <fbc_stats>
            <session_count>0</session_count>
            <average_fps>0</average_fps>
            <average_latency>0</average_latency>
        </fbc_stats>
        <ecc_mode>
            <current_ecc>Enabled</current_ecc>
            <pending_ecc>Enabled</pending_ecc>
        </ecc_mode>
        <ecc_errors>
            <volatile>
                <single_bit>
                    <device_memory>0</device_memory>
                    <register_file>0</register_file>
                    <l1_cache>0</l1_cache>
                    <l2_cache>0</l2_cache>
                    <texture_memory>N/A</texture_memory>
                    <texture_shm>N/A</texture_shm>
                    <cbu>N/A</cbu>
                    <total>0</total>
                </single_bit>
                <double_bit>
                    <device_memory>0</device_memory>
                    <register_file>0</register_file>
                    <l1_cache>0</l1_cache>
                    <l2_cache>0</l2_cache>
                    <texture_memory>N/A</texture_memory>
                    <texture_shm>N/A</texture_shm>
                    <cbu>0</cbu>
                    <total>0</total>
                </double_bit>
            </volatile>
            <aggregate>
                <single_bit>
                    <device_memory>0</device_memory>
                    <register_file>0</register_file>
                    <l1_cache>0</l1_cache>
                    <l2_cache>0</l2_cache>
                    <texture_memory>N/A</texture_memory>
                    <texture_shm>N/A</texture_shm>
                    <cbu>N/A</cbu>
                    <total>0</total>
                </single_bit>
                <double_bit>
                    <device_memory>0</device_memory>
                    <register_file>0</register_file>
                    <l1_cache>0</l1_cache>
                    <l2_cache>0</l2_cache>
                    <texture_memory>N/A</texture_memory>
                    <texture_shm>N/A</texture_shm>
                    <cbu>0</cbu>
                    <total>0</total>
                </double_bit>
            </aggregate>
        </ecc_errors>
        <retired_pages>
            <multiple_single_bit_retirement>
                <retired_count>0</retired_count>
                <retired_pagelist>
                </retired_pagelist>
            </multiple_single_bit_retirement>
            <double_bit_retirement>
                <retired_count>0</retired_count>
                <retired_pagelist>
                </retired_pagelist>
            </double_bit_retirement>
            <pending_blacklist>No</pending_blacklist>
            <pending_retirement>No</pending_retirement>
        </retired_pages>
        <remapped_rows>N/A</remapped_rows>
        <temperature>
            <gpu_temp>57 C</gpu_temp>
            <gpu_temp_max_threshold>90 C</gpu_temp_max_threshold>
            <gpu_temp_slow_threshold>87 C</gpu_temp_slow_threshold>
            <gpu_temp_max_gpu_threshold>83 C</gpu_temp_max_gpu_threshold>
            <gpu_target_temperature>N/A</gpu_target_temperature>
            <memory_temp>53 C</memory_temp>
            <gpu_temp_max_mem_threshold>85 C</gpu_temp_max_mem_threshold>
        </temperature>
        <supported_gpu_target_temp>
            <gpu_target_temp_min>N/A</gpu_target_temp_min>
            <gpu_target_temp_max>N/A</gpu_target_temp_max>
        </supported_gpu_target_temp>
        <power_readings>
            <power_state>P0</power_state>
            <power_management>Supported</power_management>
            <power_draw>61.47 W</power_draw>
            <power_limit>250.00 W</power_limit>
            <default_power_limit>250.00 W</default_power_limit>
            <enforced_power_limit>250.00 W</enforced_power_limit>
            <min_power_limit>100.00 W</min_power_limit>
            <max_power_limit>250.00 W</max_power_limit>
        </power_readings>
        <clocks>
            <graphics_clock>1575 MHz</graphics_clock>
            <sm_clock>1575 MHz</sm_clock>
            <mem_clock>1107 MHz</mem_clock>
            <video_clock>1417 MHz</video_clock>
        </clocks>
        <applications_clocks>
            <graphics_clock>1245 MHz</graphics_clock>
            <mem_clock>1107 MHz</mem_clock>
        </applications_clocks>
        <default_applications_clocks>
            <graphics_clock>1245 MHz</graphics_clock>
            <mem_clock>1107 MHz</mem_clock>
        </default_applications_clocks>
        <max_clocks>
            <graphics_clock>1597 MHz</graphics_clock>
            <sm_clock>1597 MHz</sm_clock>
            <mem_clock>1107 MHz</mem_clock>
            <video_clock>1432 MHz</video_clock>
        </max_clocks>
        <max_customer_boost_clocks>
            <graphics_clock>1597 MHz</graphics_clock>
        </max_customer_boost_clocks>
        <clock_policy>
            <auto_boost>N/A</auto_boost>
            <auto_boost_default>N/A</auto_boost_default>
        </clock_policy>
        <supported_clocks>
            <supported_mem_clock>
                <value>1107 MHz</value>
                <supported_graphics_clock>1597 MHz</supported_graphics_clock>
                <supported_graphics_clock>1590 MHz</supported_graphics_clock>
                <supported_graphics_clock>1582 MHz</supported_graphics_clock>
                <supported_graphics_clock>1575 MHz</supported_graphics_clock>
                <supported_graphics_clock>1567 MHz</supported_graphics_clock>
                <supported_graphics_clock>1560 MHz</supported_graphics_clock>
                <supported_graphics_clock>1552 MHz</supported_graphics_clock>
                <supported_graphics_clock>1545 MHz</supported_graphics_clock>
                <supported_graphics_clock>1537 MHz</supported_graphics_clock>
                <supported_graphics_clock>1530 MHz</supported_graphics_clock>
                <supported_graphics_clock>1522 MHz</supported_graphics_clock>
                <supported_graphics_clock>1515 MHz</supported_graphics_clock>
                <supported_graphics_clock>1507 MHz</supported_graphics_clock>
                <supported_graphics_clock>1500 MHz</supported_graphics_clock>
                <supported_graphics_clock>1492 MHz</supported_graphics_clock>
                <supported_graphics_clock>1485 MHz</supported_graphics_clock>
                <supported_graphics_clock>1477 MHz</supported_graphics_clock>
                <supported_graphics_clock>1470 MHz</supported_graphics_clock>
                <supported_graphics_clock>1462 MHz</supported_graphics_clock>
                <supported_graphics_clock>1455 MHz</supported_graphics_clock>
                <supported_graphics_clock>1447 MHz</supported_graphics_clock>
                <supported_graphics_clock>1440 MHz</supported_graphics_clock>
                <supported_graphics_clock>1432 MHz</supported_graphics_clock>
                <supported_graphics_clock>1425 MHz</supported_graphics_clock>
                <supported_graphics_clock>1417 MHz</supported_graphics_clock>
                <supported_graphics_clock>1410 MHz</supported_graphics_clock>
                <supported_graphics_clock>1402 MHz</supported_graphics_clock>
                <supported_graphics_clock>1395 MHz</supported_graphics_clock>
                <supported_graphics_clock>1387 MHz</supported_graphics_clock>
                <supported_graphics_clock>1380 MHz</supported_graphics_clock>
                <supported_graphics_clock>1372 MHz</supported_graphics_clock>
                <supported_graphics_clock>1365 MHz</supported_graphics_clock>
                <supported_graphics_clock>1357 MHz</supported_graphics_clock>
                <supported_graphics_clock>1350 MHz</supported_graphics_clock>
                <supported_graphics_clock>1342 MHz</supported_graphics_clock>
                <supported_graphics_clock>1335 MHz</supported_graphics_clock>
                <supported_graphics_clock>1327 MHz</supported_graphics_clock>
                <supported_graphics_clock>1320 MHz</supported_graphics_clock>
                <supported_graphics_clock>1312 MHz</supported_graphics_clock>
                <supported_graphics_clock>1305 MHz</supported_graphics_clock>
                <supported_graphics_clock>1297 MHz</supported_graphics_clock>
                <supported_graphics_clock>1290 MHz</supported_graphics_clock>
                <supported_graphics_clock>1282 MHz</supported_graphics_clock>
                <supported_graphics_clock>1275 MHz</supported_graphics_clock>
                <supported_graphics_clock>1267 MHz</supported_graphics_clock>
                <supported_graphics_clock>1260 MHz</supported_graphics_clock>
                <supported_graphics_clock>1252 MHz</supported_graphics_clock>
                <supported_graphics_clock>1245 MHz</supported_graphics_clock>
                <supported_graphics_clock>1237 MHz</supported_graphics_clock>
                <supported_graphics_clock>1230 MHz</supported_graphics_clock>
                <supported_graphics_clock>1222 MHz</supported_graphics_clock>
                <supported_graphics_clock>1215 MHz</supported_graphics_clock>
                <supported_graphics_clock>1207 MHz</supported_graphics_clock>
                <supported_graphics_clock>1200 MHz</supported_graphics_clock>
                <supported_graphics_clock>1192 MHz</supported_graphics_clock>
                <supported_graphics_clock>1185 MHz</supported_graphics_clock>
                <supported_graphics_clock>1177 MHz</supported_graphics_clock>
                <supported_graphics_clock>1170 MHz</supported_graphics_clock>
                <supported_graphics_clock>1162 MHz</supported_graphics_clock>
                <supported_graphics_clock>1155 MHz</supported_graphics_clock>
                <supported_graphics_clock>1147 MHz</supported_graphics_clock>
                <supported_graphics_clock>1140 MHz</supported_graphics_clock>
                <supported_graphics_clock>1132 MHz</supported_graphics_clock>
                <supported_graphics_clock>1125 MHz</supported_graphics_clock>
                <supported_graphics_clock>1117 MHz</supported_graphics_clock>
                <supported_graphics_clock>1110 MHz</supported_graphics_clock>
                <supported_graphics_clock>1102 MHz</supported_graphics_clock>
                <supported_graphics_clock>1095 MHz</supported_graphics_clock>
                <supported_graphics_clock>1087 MHz</supported_graphics_clock>
                <supported_graphics_clock>1080 MHz</supported_graphics_clock>
                <supported_graphics_clock>1072 MHz</supported_graphics_clock>
                <supported_graphics_clock>1065 MHz</supported_graphics_clock>
                <supported_graphics_clock>1057 MHz</supported_graphics_clock>
                <supported_graphics_clock>1050 MHz</supported_graphics_clock>
                <supported_graphics_clock>1042 MHz</supported_graphics_clock>
                <supported_graphics_clock>1035 MHz</supported_graphics_clock>
                <supported_graphics_clock>1027 MHz</supported_graphics_clock>
                <supported_graphics_clock>1020 MHz</supported_graphics_clock>
                <supported_graphics_clock>1012 MHz</supported_graphics_clock>
                <supported_graphics_clock>1005 MHz</supported_graphics_clock>
                <supported_graphics_clock>997 MHz</supported_graphics_clock>
                <supported_graphics_clock>990 MHz</supported_graphics_clock>
                <supported_graphics_clock>982 MHz</supported_graphics_clock>
                <supported_graphics_clock>975 MHz</supported_graphics_clock>
                <supported_graphics_clock>967 MHz</supported_graphics_clock>
                <supported_graphics_clock>960 MHz</supported_graphics_clock>
                <supported_graphics_clock>952 MHz</supported_graphics_clock>
                <supported_graphics_clock>945 MHz</supported_graphics_clock>
                <supported_graphics_clock>937 MHz</supported_graphics_clock>
                <supported_graphics_clock>930 MHz</supported_graphics_clock>
                <supported_graphics_clock>922 MHz</supported_graphics_clock>
                <supported_graphics_clock>915 MHz</supported_graphics_clock>
                <supported_graphics_clock>907 MHz</supported_graphics_clock>
                <supported_graphics_clock>900 MHz</supported_graphics_clock>
                <supported_graphics_clock>892 MHz</supported_graphics_clock>
                <supported_graphics_clock>885 MHz</supported_graphics_clock>
                <supported_graphics_clock>877 MHz</supported_graphics_clock>
                <supported_graphics_clock>870 MHz</supported_graphics_clock>
                <supported_graphics_clock>862 MHz</supported_graphics_clock>
                <supported_graphics_clock>855 MHz</supported_graphics_clock>
                <supported_graphics_clock>847 MHz</supported_graphics_clock>
                <supported_graphics_clock>840 MHz</supported_graphics_clock>
                <supported_graphics_clock>832 MHz</supported_graphics_clock>
                <supported_graphics_clock>825 MHz</supported_graphics_clock>
                <supported_graphics_clock>817 MHz</supported_graphics_clock>
                <supported_graphics_clock>810 MHz</supported_graphics_clock>
                <supported_graphics_clock>802 MHz</supported_graphics_clock>
                <supported_graphics_clock>795 MHz</supported_graphics_clock>
                <supported_graphics_clock>787 MHz</supported_graphics_clock>
                <supported_graphics_clock>780 MHz</supported_graphics_clock>
                <supported_graphics_clock>772 MHz</supported_graphics_clock>
                <supported_graphics_clock>765 MHz</supported_graphics_clock>
                <supported_graphics_clock>757 MHz</supported_graphics_clock>
                <supported_graphics_clock>750 MHz</supported_graphics_clock>
                <supported_graphics_clock>742 MHz</supported_graphics_clock>
                <supported_graphics_clock>735 MHz</supported_graphics_clock>
                <supported_graphics_clock>727 MHz</supported_graphics_clock>
                <supported_graphics_clock>720 MHz</supported_graphics_clock>
                <supported_graphics_clock>712 MHz</supported_graphics_clock>
                <supported_graphics_clock>705 MHz</supported_graphics_clock>
                <supported_graphics_clock>697 MHz</supported_graphics_clock>
                <supported_graphics_clock>690 MHz</supported_graphics_clock>
                <supported_graphics_clock>682 MHz</supported_graphics_clock>
                <supported_graphics_clock>675 MHz</supported_graphics_clock>
                <supported_graphics_clock>667 MHz</supported_graphics_clock>
                <supported_graphics_clock>660 MHz</supported_graphics_clock>
                <supported_graphics_clock>652 MHz</supported_graphics_clock>
                <supported_graphics_clock>645 MHz</supported_graphics_clock>
                <supported_graphics_clock>637 MHz</supported_graphics_clock>
                <supported_graphics_clock>630 MHz</supported_graphics_clock>
                <supported_graphics_clock>622 MHz</supported_graphics_clock>
                <supported_graphics_clock>615 MHz</supported_graphics_clock>
                <supported_graphics_clock>607 MHz</supported_graphics_clock>
                <supported_graphics_clock>600 MHz</supported_graphics_clock>
                <supported_graphics_clock>592 MHz</supported_graphics_clock>
                <supported_graphics_clock>585 MHz</supported_graphics_clock>
                <supported_graphics_clock>577 MHz</supported_graphics_clock>
                <supported_graphics_clock>570 MHz</supported_graphics_clock>
                <supported_graphics_clock>562 MHz</supported_graphics_clock>
                <supported_graphics_clock>555 MHz</supported_graphics_clock>
                <supported_graphics_clock>547 MHz</supported_graphics_clock>
                <supported_graphics_clock>540 MHz</supported_graphics_clock>
                <supported_graphics_clock>532 MHz</supported_graphics_clock>
                <supported_graphics_clock>525 MHz</supported_graphics_clock>
                <supported_graphics_clock>517 MHz</supported_graphics_clock>
                <supported_graphics_clock>510 MHz</supported_graphics_clock>
                <supported_graphics_clock>502 MHz</supported_graphics_clock>
                <supported_graphics_clock>495 MHz</supported_graphics_clock>
                <supported_graphics_clock>487 MHz</supported_graphics_clock>
                <supported_graphics_clock>480 MHz</supported_graphics_clock>
                <supported_graphics_clock>472 MHz</supported_graphics_clock>
                <supported_graphics_clock>465 MHz</supported_graphics_clock>
                <supported_graphics_clock>457 MHz</supported_graphics_clock>
                <supported_graphics_clock>450 MHz</supported_graphics_clock>
                <supported_graphics_clock>442 MHz</supported_graphics_clock>
                <supported_graphics_clock>435 MHz</supported_graphics_clock>
                <supported_graphics_clock>427 MHz</supported_graphics_clock>
                <supported_graphics_clock>420 MHz</supported_graphics_clock>
                <supported_graphics_clock>412 MHz</supported_graphics_clock>
                <supported_graphics_clock>405 MHz</supported_graphics_clock>
                <supported_graphics_clock>397 MHz</supported_graphics_clock>
                <supported_graphics_clock>390 MHz</supported_graphics_clock>
                <supported_graphics_clock>382 MHz</supported_graphics_clock>
                <supported_graphics_clock>375 MHz</supported_graphics_clock>
                <supported_graphics_clock>367 MHz</supported_graphics_clock>
                <supported_graphics_clock>360 MHz</supported_graphics_clock>
                <supported_graphics_clock>352 MHz</supported_graphics_clock>
                <supported_graphics_clock>345 MHz</supported_graphics_clock>
                <supported_graphics_clock>337 MHz</supported_graphics_clock>
                <supported_graphics_clock>330 MHz</supported_graphics_clock>
                <supported_graphics_clock>322 MHz</supported_graphics_clock>
                <supported_graphics_clock>315 MHz</supported_graphics_clock>
                <supported_graphics_clock>307 MHz</supported_graphics_clock>
                <supported_graphics_clock>300 MHz</supported_graphics_clock>
                <supported_graphics_clock>292 MHz</supported_graphics_clock>
                <supported_graphics_clock>285 MHz</supported_graphics_clock>
                <supported_graphics_clock>277 MHz</supported_graphics_clock>
                <supported_graphics_clock>270 MHz</supported_graphics_clock>
                <supported_graphics_clock>262 MHz</supported_graphics_clock>
                <supported_graphics_clock>255 MHz</supported_graphics_clock>
                <supported_graphics_clock>247 MHz</supported_graphics_clock>
                <supported_graphics_clock>240 MHz</supported_graphics_clock>
                <supported_graphics_clock>232 MHz</supported_graphics_clock>
                <supported_graphics_clock>225 MHz</supported_graphics_clock>
                <supported_graphics_clock>217 MHz</supported_graphics_clock>
                <supported_graphics_clock>210 MHz</supported_graphics_clock>
                <supported_graphics_clock>202 MHz</supported_graphics_clock>
                <supported_graphics_clock>195 MHz</supported_graphics_clock>
                <supported_graphics_clock>187 MHz</supported_graphics_clock>
                <supported_graphics_clock>180 MHz</supported_graphics_clock>
                <supported_graphics_clock>172 MHz</supported_graphics_clock>
                <supported_graphics_clock>165 MHz</supported_graphics_clock>
                <supported_graphics_clock>157 MHz</supported_graphics_clock>
                <supported_graphics_clock>150 MHz</supported_graphics_clock>
                <supported_graphics_clock>142 MHz</supported_graphics_clock>
                <supported_graphics_clock>135 MHz</supported_graphics_clock>
            </supported_mem_clock>
        </supported_clocks>
        <processes>
            <process_info>
                <gpu_instance_id>N/A</gpu_instance_id>
                <compute_instance_id>N/A</compute_instance_id>
                <pid>37305</pid>
                <type>C</type>
                <process_name>python</process_name>
                <used_memory>29495 MiB</used_memory>
            </process_info>
        </processes>
        <accounted_processes>
            <accounted_process_info>
                <pid>54796</pid>
                <gpu_util>0 %</gpu_util>
                <memory_util>0 %</memory_util>
                <max_memory_usage>307 MiB</max_memory_usage>
                <time>17964 ms</time>
                <is_running>0</is_running>
            </accounted_process_info>
            <accounted_process_info>
                <pid>57134</pid>
                <gpu_util>0 %</gpu_util>
                <memory_util>0 %</memory_util>
                <max_memory_usage>31919 MiB</max_memory_usage>
                <time>4907861 ms</time>
                <is_running>0</is_running>
            </accounted_process_info>
            <accounted_process_info>
                <pid>64749</pid>
                <gpu_util>1 %</gpu_util>
                <memory_util>0 %</memory_util>
                <max_memory_usage>31855 MiB</max_memory_usage>
                <time>187955 ms</time>
                <is_running>0</is_running>
            </accounted_process_info>
            <accounted_process_info>
                <pid>2956</pid>
                <gpu_util>54 %</gpu_util>
                <memory_util>18 %</memory_util>
                <max_memory_usage>28395 MiB</max_memory_usage>
                <time>313867920 ms</time>
                <is_running>0</is_running>
            </accounted_process_info>
            <accounted_process_info>
                <pid>45030</pid>
                <gpu_util>0 %</gpu_util>
                <memory_util>0 %</memory_util>
                <max_memory_usage>31463 MiB</max_memory_usage>
                <time>16828353 ms</time>
                <is_running>0</is_running>
            </accounted_process_info>
            <accounted_process_info>
                <pid>37305</pid>
                <gpu_util>58 %</gpu_util>
                <memory_util>22 %</memory_util>
                <max_memory_usage>30615 MiB</max_memory_usage>
                <time>0 ms</time>
                <is_running>1</is_running>
            </accounted_process_info>
        </accounted_processes>
    </gpu>

</nvidia_smi_log>
`)
