// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2022-present Guance, Inc.

package nvidiasmi

import (
	"bytes"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/GuanceCloud/cliutils/point"
)

func (ipt *Input) getPts(data []byte, server string) error {
	// ts := time.Now()
	metricOpts := point.DefaultMetricOptions()
	metricOpts = append(metricOpts, point.WithTime(ipt.ptsTime))
	loggingOpts := point.DefaultLoggingOptions()
	loggingOpts = append(loggingOpts, point.WithTime(ipt.ptsTime))

	metrics, metricsLog, err := ipt.parseSMIMetrics(data, server)
	if err != nil {
		l.Errorf("Unmarshal xml data log: %s .", err)
		return err
	}

	for _, metric := range metrics {
		var kvs point.KVs

		for k, v := range metric.fields {
			kvs = kvs.Set(k, v)
		}

		for k, v := range metric.tags {
			kvs = kvs.AddTag(k, v)
		}

		for k, v := range ipt.mergedTags[server] {
			kvs = kvs.AddTag(k, v)
		}

		ipt.collectCache = append(ipt.collectCache, point.NewPoint(inputName, kvs, metricOpts...))
	}

	for i := 0; i < len(metricsLog); i++ {
		var kvs point.KVs

		for k, v := range metricsLog[i].fields {
			kvs = kvs.Set(k, v)
		}

		for k, v := range metricsLog[i].tags {
			kvs = kvs.AddTag(k, v)
		}

		for k, v := range ipt.mergedTags[server] {
			kvs = kvs.AddTag(k, v)
		}

		ipt.collectCacheLog = append(ipt.collectCacheLog, point.NewPoint(inputName, kvs, loggingOpts...))
	}
	return nil
}

func (ipt *Input) parseSMIMetrics(data []byte, server string) ([]metric, []metric, error) {
	schema, err := detectSMISchema(data)
	if err != nil {
		return nil, nil, err
	}

	switch schema {
	case "v12":
		smi := &SMIV12{}
		if err := xml.Unmarshal(data, smi); err != nil {
			return nil, nil, err
		}
		metrics, metricsLog := smi.genTagsFields(ipt, server, ipt.ptsTime)
		return metrics, metricsLog, nil
	case "v13":
		smi := &SMIV13{}
		if err := xml.Unmarshal(data, smi); err != nil {
			return nil, nil, err
		}
		metrics, metricsLog := smi.genTagsFields(ipt, server, ipt.ptsTime)
		return metrics, metricsLog, nil
	case "v10", "v11":
		smi := &SMI{}
		if err := xml.Unmarshal(data, smi); err != nil {
			return nil, nil, err
		}
		metrics, metricsLog := smi.genTagsFields(ipt, server, ipt.ptsTime)
		return metrics, metricsLog, nil
	default:
		smi := &SMIV13{}
		if err := xml.Unmarshal(data, smi); err != nil {
			return nil, nil, err
		}
		metrics, metricsLog := smi.genTagsFields(ipt, server, ipt.ptsTime)
		return metrics, metricsLog, nil
	}
}

func detectSMISchema(data []byte) (string, error) {
	schema := "v11"

	decoder := xml.NewDecoder(bytes.NewBuffer(data))
	for {
		token, err := decoder.Token()
		if err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			return "", fmt.Errorf("reading nvidia-smi XML schema failed: %w", err)
		}
		d, ok := token.(xml.Directive)
		if !ok {
			continue
		}

		directive := string(d)
		if !strings.HasPrefix(directive, "DOCTYPE") {
			continue
		}

		parts := strings.Split(directive, " ")
		s := strings.Trim(parts[len(parts)-1], "\" ")
		if strings.HasPrefix(s, "nvsmi_device_") && strings.HasSuffix(s, ".dtd") {
			schema = strings.TrimSuffix(strings.TrimPrefix(s, "nvsmi_device_"), ".dtd")
		} else {
			l.Warnf("Cannot find schema version in %q", directive)
		}
		break
	}
	return schema, nil
}

type metric struct {
	tags   map[string]string
	fields map[string]interface{}
}

type processMetricContext struct {
	PState      string
	ProdName    string
	UUID        string
	PCIBusID    string
	ComputeMode string
}

type processMetricData struct {
	GpuInstanceID     string
	ComputeInstanceID string
	Pid               string
	Type              string
	ProcessName       string
	UsedMemory        string
}

func setTagIfUsed(m map[string]string, k, v string) {
	if v != "" {
		m[k] = v
	}
}

func setIfUsed(t string, m map[string]interface{}, k, v string) {
	vals := strings.Fields(v)
	if len(vals) < 1 {
		return
	}

	val := vals[0]
	if k == "pcie_link_width_current" {
		val = strings.TrimSuffix(vals[0], "x")
	}

	switch t {
	case "float":
		if val != "" {
			f, err := strconv.ParseFloat(val, 64)
			if err == nil {
				m[k] = f
			}
		}
	case "int":
		if val != "" {
			i, err := strconv.Atoi(val)
			if err == nil {
				m[k] = i
			}
		}
	case "str":
		if val != "" {
			m[k] = val
		}
	}
}

func genProcessMetrics(
	ipt *Input,
	server string,
	gpu processMetricContext,
	processes Processes,
) []metric {
	if ipt.ProcessInfoMaxLen == 0 || len(processes.ProcessInfos) == 0 {
		return nil
	}

	data := make([]processMetricData, 0, len(processes.ProcessInfos))
	for _, process := range processes.ProcessInfos {
		data = append(data, processMetricData{
			GpuInstanceID:     process.GpuInstanceID,
			ComputeInstanceID: process.ComputeInstanceID,
			Pid:               process.Pid,
			Type:              process.Type,
			ProcessName:       process.ProcessName,
			UsedMemory:        process.UsedMemory,
		})
	}

	return genProcessMetricsData(ipt, server, gpu, data)
}

func genProcessMetricsData(
	ipt *Input,
	server string,
	gpu processMetricContext,
	processes []processMetricData,
) []metric {
	if ipt.ProcessInfoMaxLen == 0 || len(processes) == 0 {
		return nil
	}

	metricsLog := []metric{}
	processIndex := make([][2]int, 0, 2) // cap=2, can reduce cpu
	usedMemory := 0
	var err error
	for j := 0; j < len(processes); j++ {
		vals := strings.Fields(processes[j].UsedMemory)
		if len(vals) < 1 {
			processIndex = append(processIndex, [2]int{0, j})
			continue
		}
		val := vals[0]
		if val != "" {
			usedMemory, err = strconv.Atoi(val)
			if err != nil {
				usedMemory = 0
			}
		}
		processIndex = append(processIndex, [2]int{usedMemory, j})
	}
	sort.Slice(processIndex, func(i, j int) bool {
		return processIndex[i][0] > processIndex[j][0]
	})

	endIndex := ipt.ProcessInfoMaxLen
	if endIndex == -1 {
		endIndex = len(processIndex)
	}
	for j := 0; j < endIndex; j++ {
		if j >= len(processIndex) {
			break
		}
		tagsLog := map[string]string{}
		fieldsLog := map[string]interface{}{}

		if server != "" {
			// for ssh remote server
			setTagIfUsed(tagsLog, "host", server)
		}
		setTagIfUsed(tagsLog, "pstate", gpu.PState)
		setTagIfUsed(tagsLog, "name", gpu.ProdName)
		setTagIfUsed(tagsLog, "uuid", gpu.UUID)
		setTagIfUsed(tagsLog, "pci_bus_id", gpu.PCIBusID)
		setTagIfUsed(tagsLog, "compute_mode", gpu.ComputeMode)
		setTagIfUsed(tagsLog, "service", "msi_service")

		process := processes[processIndex[j][1]]
		s := strings.Split(process.ProcessName, "/") // 去除路径名字
		fieldsLog["message"] = fmt.Sprintf("%s:ProcessName=%s,UsedMemory= %s",
			gpu.UUID,
			s[len(s)-1],
			process.UsedMemory)
		fieldsLog["gpu_instance_id"] = process.GpuInstanceID
		fieldsLog["compute_instance_id"] = process.ComputeInstanceID
		setIfUsed("int", fieldsLog, "pid", process.Pid)
		setIfUsed("str", fieldsLog, "type", process.Type)
		setIfUsed("str", fieldsLog, "process_name", process.ProcessName)
		fieldsLog["used_memory"] = processIndex[j][0]
		fieldsLog["status"] = "info"

		metricsLog = append(metricsLog, metric{tagsLog, fieldsLog})
	}
	return metricsLog
}

// SMI defines the structure for the output of _nvidia-smi -q -x_.
type SMI struct {
	GPU           GPU    `xml:"gpu"`
	DriverVersion string `xml:"driver_version"`
	CUDAVersion   string `xml:"cuda_version"`
}

// GPU defines the structure of the GPU portion of the smi output.
type GPU []struct {
	FanSpeed    string           `xml:"fan_speed"` // int
	Memory      MemoryStats      `xml:"fb_memory_usage"`
	PState      string           `xml:"performance_state"`
	Temp        TempStats        `xml:"temperature"`
	ProdName    string           `xml:"product_name"`
	Arch        string           `xml:"product_architecture"`
	UUID        string           `xml:"uuid"`
	ComputeMode string           `xml:"compute_mode"`
	Utilization UtilizationStats `xml:"utilization"`
	Power       PowerReadings    `xml:"power_readings"`
	PCI         PCI              `xml:"pci"`
	Encoder     EncoderStats     `xml:"encoder_stats"`
	FBC         FBCStats         `xml:"fbc_stats"`
	Clocks      ClockStats       `xml:"clocks"`
	Retired     RetiredPages     `xml:"retired_pages"`
	Remapped    RemappedRows     `xml:"remapped_rows"`
	Processes   Processes        `xml:"processes"`
}

// MemoryStats defines the structure of the memory portions in the smi output.
type MemoryStats struct {
	Total    string `xml:"total"`    // int
	Used     string `xml:"used"`     // int
	Free     string `xml:"free"`     // int
	Reserved string `xml:"reserved"` // int
}

// TempStats defines the structure of the temperature portion of the smi output.
type TempStats struct {
	GPUTemp string `xml:"gpu_temp"` // int
}

// UtilizationStats defines the structure of the utilization portion of the smi output.
type UtilizationStats struct {
	GPU     string `xml:"gpu_util"`     // int
	Memory  string `xml:"memory_util"`  // int
	Encoder string `xml:"encoder_util"` // int
	Decoder string `xml:"decoder_util"` // int
}

// PowerReadings defines the structure of the power_readings portion of the smi output.
type PowerReadings struct {
	PowerDraw  string `xml:"power_draw"`  // float
	PowerLimit string `xml:"power_limit"` // float
}

// RetiredPages defines the numeric retired-page counters in the smi output.
type RetiredPages struct {
	MultipleSingleBit struct {
		Count string `xml:"retired_count"` // int
	} `xml:"multiple_single_bit_retirement"`
	DoubleBit struct {
		Count string `xml:"retired_count"` // int
	} `xml:"double_bit_retirement"`
}

// RemappedRows defines the numeric remapped-row counters in the smi output.
type RemappedRows struct {
	Correctable   string `xml:"remapped_row_corr"` // int
	Uncorrectable string `xml:"remapped_row_unc"`  // int
}

// PCI defines the structure of the pci portion of the smi output.
type PCI struct {
	PciBusID string `xml:"pci_bus_id"`
	LinkInfo struct {
		PCIEGen struct {
			CurrentLinkGen string `xml:"current_link_gen"` // int
		} `xml:"pcie_gen"`
		LinkWidth struct {
			CurrentLinkWidth string `xml:"current_link_width"` // int
		} `xml:"link_widths"`
	} `xml:"pci_gpu_link_info"`
}

// EncoderStats defines the structure of the encoder_stats portion of the smi output.
type EncoderStats struct {
	SessionCount   string `xml:"session_count"`   // int
	AverageFPS     string `xml:"average_fps"`     // int
	AverageLatency string `xml:"average_latency"` // int
}

// FBCStats defines the structure of the fbc_stats portion of the smi output.
type FBCStats struct {
	SessionCount   string `xml:"session_count"`   // int
	AverageFPS     string `xml:"average_fps"`     // int
	AverageLatency string `xml:"average_latency"` // int
}

// ClockStats defines the structure of the clocks portion of the smi output.
type ClockStats struct {
	Graphics string `xml:"graphics_clock"` // int
	SM       string `xml:"sm_clock"`       // int
	Memory   string `xml:"mem_clock"`      // int
	Video    string `xml:"video_clock"`    // int
}

type Processes struct {
	ProcessInfos ProcessInfo `xml:"process_info"`
}

type ProcessInfo []struct {
	GpuInstanceID     string `xml:"gpu_instance_id"`
	ComputeInstanceID string `xml:"compute_instance_id"`
	Pid               string `xml:"pid"` // int
	Type              string `xml:"type"`
	ProcessName       string `xml:"process_name"`
	UsedMemory        string `xml:"used_memory"` // int
}

func (s *SMI) genTagsFields(ipt *Input, server string, ptsTime time.Time) ([]metric, []metric) {
	metrics := []metric{}
	metricsLog := []metric{}
	for _, gpu := range s.GPU {
		// handle GPU online info
		ipt.gpuOnlineInfo(gpu.UUID, server, ptsTime)

		tags := map[string]string{}
		fields := map[string]interface{}{}

		setTagIfUsed(tags, "pstate", gpu.PState)
		setTagIfUsed(tags, "name", gpu.ProdName)
		setTagIfUsed(tags, "arch", gpu.Arch)
		setTagIfUsed(tags, "uuid", gpu.UUID)
		setTagIfUsed(tags, "compute_mode", gpu.ComputeMode)
		setTagIfUsed(tags, "pci_bus_id", gpu.PCI.PciBusID)
		setTagIfUsed(tags, "driver_version", s.DriverVersion)
		setTagIfUsed(tags, "cuda_version", s.CUDAVersion)

		setIfUsed("int", fields, "fan_speed", gpu.FanSpeed)
		setIfUsed("int", fields, "memory_total", gpu.Memory.Total)
		setIfUsed("int", fields, "memory_used", gpu.Memory.Used)
		setIfUsed("int", fields, "memory_free", gpu.Memory.Free)
		setIfUsed("int", fields, "memory_reserved", gpu.Memory.Reserved)
		setIfUsed("int", fields, "retired_pages_multiple_single_bit", gpu.Retired.MultipleSingleBit.Count)
		setIfUsed("int", fields, "retired_pages_double_bit", gpu.Retired.DoubleBit.Count)
		setIfUsed("int", fields, "remapped_rows_correctable", gpu.Remapped.Correctable)
		setIfUsed("int", fields, "remapped_rows_uncorrectable", gpu.Remapped.Uncorrectable)

		setIfUsed("int", fields, "temperature_gpu", gpu.Temp.GPUTemp)
		setIfUsed("int", fields, "utilization_gpu", gpu.Utilization.GPU)
		setIfUsed("int", fields, "utilization_memory", gpu.Utilization.Memory)
		setIfUsed("int", fields, "utilization_encoder", gpu.Utilization.Encoder)
		setIfUsed("int", fields, "utilization_decoder", gpu.Utilization.Decoder)
		setIfUsed("int", fields, "pcie_link_gen_current", gpu.PCI.LinkInfo.PCIEGen.CurrentLinkGen)
		setIfUsed("int", fields, "pcie_link_width_current", gpu.PCI.LinkInfo.LinkWidth.CurrentLinkWidth)
		setIfUsed("int", fields, "encoder_stats_session_count", gpu.Encoder.SessionCount)
		setIfUsed("int", fields, "encoder_stats_average_fps", gpu.Encoder.AverageFPS)
		setIfUsed("int", fields, "encoder_stats_average_latency", gpu.Encoder.AverageLatency)
		setIfUsed("int", fields, "fbc_stats_session_count", gpu.FBC.SessionCount)
		setIfUsed("int", fields, "fbc_stats_average_fps", gpu.FBC.AverageFPS)
		setIfUsed("int", fields, "fbc_stats_average_latency", gpu.FBC.AverageLatency)
		setIfUsed("int", fields, "clocks_current_graphics", gpu.Clocks.Graphics)
		setIfUsed("int", fields, "clocks_current_sm", gpu.Clocks.SM)
		setIfUsed("int", fields, "clocks_current_memory", gpu.Clocks.Memory)
		setIfUsed("int", fields, "clocks_current_video", gpu.Clocks.Video)

		setIfUsed("float", fields, "power_draw", gpu.Power.PowerDraw)
		setIfUsed("float", fields, "power_limit", gpu.Power.PowerLimit)
		metrics = append(metrics, metric{tags, fields})

		processMetrics := genProcessMetrics(ipt, server, processMetricContext{
			PState:      gpu.PState,
			ProdName:    gpu.ProdName,
			UUID:        gpu.UUID,
			PCIBusID:    gpu.PCI.PciBusID,
			ComputeMode: gpu.ComputeMode,
		}, gpu.Processes)
		if len(processMetrics) > 0 {
			metricsLog = append(metricsLog, processMetrics...)
		}
	}
	return metrics, metricsLog
}
