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
			"fan_speed":                     &inputs.FieldInfo{Type: inputs.Rate, DataType: inputs.Int, Unit: inputs.RPMPercent, Desc: "GPU fan speed as a percentage of maximum fan speed."},
			"memory_total":                  &inputs.FieldInfo{Type: inputs.Gauge, DataType: inputs.Int, Unit: inputs.SizeMB, Desc: "Frame buffer memory total."},
			"memory_used":                   &inputs.FieldInfo{Type: inputs.Gauge, DataType: inputs.Int, Unit: inputs.SizeMB, Desc: "Frame buffer memory used."},
			"temperature_gpu":               &inputs.FieldInfo{Type: inputs.Gauge, DataType: inputs.Int, Unit: inputs.Celsius, Desc: "GPU temperature."},
			"utilization_gpu":               &inputs.FieldInfo{Type: inputs.Gauge, DataType: inputs.Int, Unit: inputs.Percent, Desc: "GPU utilization."},
			"utilization_memory":            &inputs.FieldInfo{Type: inputs.Gauge, DataType: inputs.Int, Unit: inputs.Percent, Desc: "Memory utilization."},
			"utilization_encoder":           &inputs.FieldInfo{Type: inputs.Gauge, DataType: inputs.Int, Unit: inputs.Percent, Desc: "Encoder utilization."},
			"utilization_decoder":           &inputs.FieldInfo{Type: inputs.Gauge, DataType: inputs.Int, Unit: inputs.Percent, Desc: "Decoder utilization."},
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
		},

		Tags: map[string]interface{}{
			"host":           &inputs.TagInfo{Desc: "Host name"},
			"pstate":         &inputs.TagInfo{Desc: "GPU performance level"},
			"name":           &inputs.TagInfo{Desc: "GPU card model"},
			"uuid":           &inputs.TagInfo{Desc: "UUID"},
			"compute_mode":   &inputs.TagInfo{Desc: "Compute mode"},
			"pci_bus_id":     &inputs.TagInfo{Desc: "PCI bus id"},
			"driver_version": &inputs.TagInfo{Desc: "Driver version"},
			"cuda_version":   &inputs.TagInfo{Desc: "CUDA version"},
		},
	}
}
