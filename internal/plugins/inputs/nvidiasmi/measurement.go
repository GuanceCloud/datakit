// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package nvidiasmi

import (
	"github.com/GuanceCloud/cliutils/point"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs"
)

type docMeasurement struct{}

// Info , reflected in the document
//
//nolint:lll
func (docMeasurement) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name:           metricName,
		Cat:            point.Metric,
		MetaDuplicated: true, // collector `nvidia_smi' and `gpu_smi` are the same
		Desc:           "NVIDIA GPU metrics collected from nvidia-smi, including memory, temperature, utilization, PCIe link, encoder/FBC statistics, clocks, and power draw.",
		DescZh:         "通过 nvidia-smi 采集的 NVIDIA GPU 指标，包含显存、温度、利用率、PCIe 链路、编码器/FBC 统计、时钟和功耗。",
		Fields: map[string]interface{}{
			"fan_speed":       &inputs.FieldInfo{Type: inputs.Rate, DataType: inputs.Int, Unit: inputs.RPMPercent, Desc: "GPU fan speed as a percentage of maximum fan speed."},
			"memory_total":    &inputs.FieldInfo{Type: inputs.Gauge, DataType: inputs.Int, Unit: inputs.SizeMB, Desc: "Frame buffer memory total."},
			"memory_used":     &inputs.FieldInfo{Type: inputs.Gauge, DataType: inputs.Int, Unit: inputs.SizeMB, Desc: "Frame buffer memory used."},
			"memory_free":     &inputs.FieldInfo{Type: inputs.Gauge, DataType: inputs.Int, Unit: inputs.SizeMB, Desc: "Frame buffer memory free."},
			"memory_reserved": &inputs.FieldInfo{Type: inputs.Gauge, DataType: inputs.Int, Unit: inputs.SizeMB, Desc: "Frame buffer memory reserved."},
			"ecc_errors_volatile_dram_correctable": &inputs.FieldInfo{
				Type:     inputs.Count,
				DataType: inputs.Int,
				Unit:     inputs.NCount,
				Desc:     "Volatile correctable DRAM ECC error count.",
			},
			"ecc_errors_volatile_dram_uncorrectable": &inputs.FieldInfo{
				Type:     inputs.Count,
				DataType: inputs.Int,
				Unit:     inputs.NCount,
				Desc:     "Volatile uncorrectable DRAM ECC error count.",
			},
			"ecc_errors_volatile_sram_correctable": &inputs.FieldInfo{
				Type:     inputs.Count,
				DataType: inputs.Int,
				Unit:     inputs.NCount,
				Desc:     "Volatile correctable SRAM ECC error count.",
			},
			"ecc_errors_volatile_sram_uncorrectable": &inputs.FieldInfo{
				Type:     inputs.Count,
				DataType: inputs.Int,
				Unit:     inputs.NCount,
				Desc:     "Volatile uncorrectable SRAM ECC error count.",
			},
			"ecc_errors_volatile_sram_uncorrectable_parity": &inputs.FieldInfo{
				Type:     inputs.Count,
				DataType: inputs.Int,
				Unit:     inputs.NCount,
				Desc:     "Volatile uncorrectable parity SRAM ECC error count.",
			},
			"ecc_errors_volatile_sram_uncorrectable_secded": &inputs.FieldInfo{
				Type:     inputs.Count,
				DataType: inputs.Int,
				Unit:     inputs.NCount,
				Desc:     "Volatile uncorrectable SECDED SRAM ECC error count.",
			},
			"ecc_errors_aggregate_dram_correctable": &inputs.FieldInfo{
				Type:     inputs.Count,
				DataType: inputs.Int,
				Unit:     inputs.NCount,
				Desc:     "Aggregate correctable DRAM ECC error count.",
			},
			"ecc_errors_aggregate_dram_uncorrectable": &inputs.FieldInfo{
				Type:     inputs.Count,
				DataType: inputs.Int,
				Unit:     inputs.NCount,
				Desc:     "Aggregate uncorrectable DRAM ECC error count.",
			},
			"ecc_errors_aggregate_sram_correctable": &inputs.FieldInfo{
				Type:     inputs.Count,
				DataType: inputs.Int,
				Unit:     inputs.NCount,
				Desc:     "Aggregate correctable SRAM ECC error count.",
			},
			"ecc_errors_aggregate_sram_uncorrectable": &inputs.FieldInfo{
				Type:     inputs.Count,
				DataType: inputs.Int,
				Unit:     inputs.NCount,
				Desc:     "Aggregate uncorrectable SRAM ECC error count.",
			},
			"ecc_errors_aggregate_sram_uncorrectable_parity": &inputs.FieldInfo{
				Type:     inputs.Count,
				DataType: inputs.Int,
				Unit:     inputs.NCount,
				Desc:     "Aggregate uncorrectable parity SRAM ECC error count.",
			},
			"ecc_errors_aggregate_sram_uncorrectable_secded": &inputs.FieldInfo{
				Type:     inputs.Count,
				DataType: inputs.Int,
				Unit:     inputs.NCount,
				Desc:     "Aggregate uncorrectable SECDED SRAM ECC error count.",
			},
			"ecc_errors_aggregate_sram_uncorrectable_l2": &inputs.FieldInfo{
				Type:     inputs.Count,
				DataType: inputs.Int,
				Unit:     inputs.NCount,
				Desc:     "Aggregate uncorrectable SRAM L2 ECC error count.",
			},
			"ecc_errors_aggregate_sram_uncorrectable_microcontroller": &inputs.FieldInfo{
				Type:     inputs.Count,
				DataType: inputs.Int,
				Unit:     inputs.NCount,
				Desc:     "Aggregate uncorrectable SRAM microcontroller ECC error count.",
			},
			"ecc_errors_aggregate_sram_uncorrectable_other": &inputs.FieldInfo{
				Type:     inputs.Count,
				DataType: inputs.Int,
				Unit:     inputs.NCount,
				Desc:     "Aggregate uncorrectable SRAM other ECC error count.",
			},
			"ecc_errors_aggregate_sram_uncorrectable_pcie": &inputs.FieldInfo{
				Type:     inputs.Count,
				DataType: inputs.Int,
				Unit:     inputs.NCount,
				Desc:     "Aggregate uncorrectable SRAM PCIe ECC error count.",
			},
			"ecc_errors_aggregate_sram_uncorrectable_sm": &inputs.FieldInfo{
				Type:     inputs.Count,
				DataType: inputs.Int,
				Unit:     inputs.NCount,
				Desc:     "Aggregate uncorrectable SRAM SM ECC error count.",
			},
			"retired_pages_multiple_single_bit": &inputs.FieldInfo{
				Type:     inputs.Count,
				DataType: inputs.Int,
				Unit:     inputs.NCount,
				Desc:     "Retired page count due to multiple single-bit errors.",
			},
			"retired_pages_double_bit": &inputs.FieldInfo{
				Type:     inputs.Count,
				DataType: inputs.Int,
				Unit:     inputs.NCount,
				Desc:     "Retired page count due to double-bit errors.",
			},
			"remapped_rows_correctable": &inputs.FieldInfo{
				Type:     inputs.Count,
				DataType: inputs.Int,
				Unit:     inputs.NCount,
				Desc:     "Remapped row count for correctable errors.",
			},
			"remapped_rows_uncorrectable": &inputs.FieldInfo{
				Type:     inputs.Count,
				DataType: inputs.Int,
				Unit:     inputs.NCount,
				Desc:     "Remapped row count for uncorrectable errors.",
			},
			"temperature_gpu":               &inputs.FieldInfo{Type: inputs.Gauge, DataType: inputs.Int, Unit: inputs.Celsius, Desc: "GPU temperature."},
			"utilization_gpu":               &inputs.FieldInfo{Type: inputs.Gauge, DataType: inputs.Int, Unit: inputs.Percent, Desc: "GPU utilization."},
			"utilization_memory":            &inputs.FieldInfo{Type: inputs.Gauge, DataType: inputs.Int, Unit: inputs.Percent, Desc: "Memory utilization."},
			"utilization_encoder":           &inputs.FieldInfo{Type: inputs.Gauge, DataType: inputs.Int, Unit: inputs.Percent, Desc: "Encoder utilization."},
			"utilization_decoder":           &inputs.FieldInfo{Type: inputs.Gauge, DataType: inputs.Int, Unit: inputs.Percent, Desc: "Decoder utilization."},
			"utilization_jpeg":              &inputs.FieldInfo{Type: inputs.Gauge, DataType: inputs.Int, Unit: inputs.Percent, Desc: "JPEG utilization."},
			"utilization_ofa":               &inputs.FieldInfo{Type: inputs.Gauge, DataType: inputs.Int, Unit: inputs.Percent, Desc: "Optical flow accelerator utilization."},
			"pcie_link_gen_current":         &inputs.FieldInfo{Type: inputs.Gauge, DataType: inputs.Int, Unit: inputs.NoUnit, Desc: "Current PCI Express link generation negotiated for the GPU."},
			"pcie_link_width_current":       &inputs.FieldInfo{Type: inputs.Gauge, DataType: inputs.Int, Unit: inputs.NoUnit, Desc: "Current PCI Express link width negotiated for the GPU."},
			"encoder_stats_session_count":   &inputs.FieldInfo{Type: inputs.Count, DataType: inputs.Int, Unit: inputs.NCount, Desc: "Encoder session count."},
			"encoder_stats_average_fps":     &inputs.FieldInfo{Type: inputs.Rate, DataType: inputs.Int, Unit: inputs.FramePerSecond, Desc: "Encoder average fps."},
			"encoder_stats_average_latency": &inputs.FieldInfo{Type: inputs.Count, DataType: inputs.Int, Unit: inputs.DurationMS, Desc: "Encoder average latency."},
			"fbc_stats_session_count":       &inputs.FieldInfo{Type: inputs.Count, DataType: inputs.Int, Unit: inputs.NCount, Desc: "Frame Buffer Cache session count."},
			"fbc_stats_average_fps":         &inputs.FieldInfo{Type: inputs.Count, DataType: inputs.Int, Unit: inputs.FramePerSecond, Desc: "Frame Buffer Cache average fps."},
			"fbc_stats_average_latency":     &inputs.FieldInfo{Type: inputs.Count, DataType: inputs.Int, Unit: inputs.DurationMS, Desc: "Frame Buffer Cache average latency."},
			"clocks_current_graphics":       &inputs.FieldInfo{Type: inputs.Gauge, DataType: inputs.Int, Unit: inputs.FrequencyMHz, Desc: "Graphics clock frequency."},
			"clocks_current_sm":             &inputs.FieldInfo{Type: inputs.Gauge, DataType: inputs.Int, Unit: inputs.FrequencyMHz, Desc: "Streaming Multiprocessor clock frequency."},
			"clocks_current_memory":         &inputs.FieldInfo{Type: inputs.Gauge, DataType: inputs.Int, Unit: inputs.FrequencyMHz, Desc: "Memory clock frequency."},
			"clocks_current_video":          &inputs.FieldInfo{Type: inputs.Gauge, DataType: inputs.Int, Unit: inputs.FrequencyMHz, Desc: "Video clock frequency."},
			"power_draw":                    &inputs.FieldInfo{Type: inputs.Gauge, DataType: inputs.Float, Unit: inputs.Watt, Desc: "Current GPU power draw in watts."},
			"power_limit":                   &inputs.FieldInfo{Type: inputs.Gauge, DataType: inputs.Float, Unit: inputs.Watt, Desc: "Current GPU power limit in watts."},
			"module_power_draw":             &inputs.FieldInfo{Type: inputs.Gauge, DataType: inputs.Float, Unit: inputs.Watt, Desc: "Current module power draw in watts."},
			"sram_uncorrectable":            &inputs.FieldInfo{Type: inputs.Count, DataType: inputs.Int, Unit: inputs.NCount, Desc: "MIG device volatile uncorrectable SRAM error count."},
			"memory_fb_total":               &inputs.FieldInfo{Type: inputs.Gauge, DataType: inputs.Int, Unit: inputs.SizeMB, Desc: "MIG device frame buffer memory total."},
			"memory_fb_reserved":            &inputs.FieldInfo{Type: inputs.Gauge, DataType: inputs.Int, Unit: inputs.SizeMB, Desc: "MIG device frame buffer memory reserved."},
			"memory_fb_used":                &inputs.FieldInfo{Type: inputs.Gauge, DataType: inputs.Int, Unit: inputs.SizeMB, Desc: "MIG device frame buffer memory used."},
			"memory_fb_free":                &inputs.FieldInfo{Type: inputs.Gauge, DataType: inputs.Int, Unit: inputs.SizeMB, Desc: "MIG device frame buffer memory free."},
			"memory_bar1_total":             &inputs.FieldInfo{Type: inputs.Gauge, DataType: inputs.Int, Unit: inputs.SizeMB, Desc: "MIG device BAR1 memory total."},
			"memory_bar1_used":              &inputs.FieldInfo{Type: inputs.Gauge, DataType: inputs.Int, Unit: inputs.SizeMB, Desc: "MIG device BAR1 memory used."},
			"memory_bar1_free":              &inputs.FieldInfo{Type: inputs.Gauge, DataType: inputs.Int, Unit: inputs.SizeMB, Desc: "MIG device BAR1 memory free."},
		},

		Tags: map[string]interface{}{
			"host":           &inputs.TagInfo{Desc: "Host name"},
			"index":          &inputs.TagInfo{Desc: "MIG device index"},
			"gpu_index":      &inputs.TagInfo{Desc: "MIG GPU instance id"},
			"compute_index":  &inputs.TagInfo{Desc: "MIG compute instance id"},
			"pstate":         &inputs.TagInfo{Desc: "GPU performance level"},
			"name":           &inputs.TagInfo{Desc: "GPU card model"},
			"arch":           &inputs.TagInfo{Desc: "GPU architecture"},
			"uuid":           &inputs.TagInfo{Desc: "UUID"},
			"compute_mode":   &inputs.TagInfo{Desc: "Compute mode"},
			"pci_bus_id":     &inputs.TagInfo{Desc: "PCI bus id"},
			"driver_version": &inputs.TagInfo{Desc: "Driver version"},
			"cuda_version":   &inputs.TagInfo{Desc: "CUDA version"},
		},
	}
}
