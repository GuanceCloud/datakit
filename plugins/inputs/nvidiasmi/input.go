// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2022-present Guance, Inc.

// Package nvidiasmi collects host nvidiasmi metrics.
package nvidiasmi

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"os/exec"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"time"

	"gitlab.jiagouyun.com/cloudcare-tools/cliutils"
	"gitlab.jiagouyun.com/cloudcare-tools/cliutils/logger"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/config"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io/point"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"
)

var _ inputs.ReadEnv = (*Input)(nil)

const (
	minInterval = time.Second
	maxInterval = time.Minute
)

const (
	// old is "nvidia_smi" , now is "gpu_smi".
	inputName  = "gpu_smi"
	metricName = "gpu_smi"
	// conf File samples, reflected in the document.
	sampleCfg = `
[[inputs.gpu_smi]]
  ##the binPath of gpu-smi 
  ##if nvidia GPU
  #(example & default) ["/usr/bin/nvidia-smi"]
  ##if lluvatar GPU
  #(example) bin_paths = ["/usr/local/corex/bin/ixsmi"]
  #(example) envs = [ "LD_LIBRARY_PATH=/usr/local/corex/lib/:$LD_LIBRARY_PATH" ]

  ##(optional) exec gpu-smi envs, default is []
  #envs = [ "LD_LIBRARY_PATH=/usr/local/corex/lib/:$LD_LIBRARY_PATH" ]
  ##(optional) exec gpu-smi timeout, default is 5 seconds
  timeout = "5s"
  ##(optional) collect interval, default is 10 seconds
  interval = "10s"
  ##(optional) Feed how much log data for ProcessInfos, default is 10. (0: 0 ,-1: all)
  process_info_max_len = 10
  ##(optional) gpu drop card warning delay, default is 300 seconds
  gpu_drop_warning_delay = "300s"

[inputs.gpu_smi.tags]
  # some_tag = "some_value"
  # more_tag = "some_other_value"`
)

var l = logger.DefaultSLogger(inputName)

type Input struct {
	Interval            datakit.Duration  // run every "Interval" seconds
	Tags                map[string]string // Indicator name
	collectCache        []inputs.Measurement
	pts                 []*point.Point // the log data for ProcessInfos
	platform            string
	BinPaths            []string         `toml:"bin_paths"`              // the file path of "nvidia-smi"
	Timeout             datakit.Duration `toml:"timeout"`                // "nvidia-smi" 超时失败阈值
	ProcessInfoMaxLen   int              `toml:"process_info_max_len"`   // Feed how much log data for ProcessInfos. (0: 0 ,-1: all)
	GPUDropWarningDelay datakit.Duration `toml:"gpu_drop_warning_delay"` // 来自于 nvidia_msi.conf 掉卡后，告警延迟时长。（默认300s）
	Envs                []string         `toml:"envs"`                   // exec.Command 环境变量
	GPUs                []GPUInfo        // 活跃卡的列表，掉卡告警后就会删除那张卡
	semStop             *cliutils.Sem    // start stop signal
}

type GPUInfo struct {
	UUID            string
	ActiveTimestamp int64 // 活跃时间戳 nanosecond time.Now().UnixNano()
}

// nvidiaSmi Measurement structure.
type nvidiaSmiMeasurement struct {
	name   string                 // Indicator set name ，here is "gpu_smi"
	tags   map[string]string      // Indicator name
	fields map[string]interface{} // Indicator measurement results
}

// LineProto data formatting, submit through FeedMeasurement.
func (n *nvidiaSmiMeasurement) LineProto() (*point.Point, error) {
	return point.NewPoint(n.name, n.tags, n.fields, point.MOpt())
}

// Info , reflected in the document
//nolint:lll
func (n *nvidiaSmiMeasurement) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name: metricName,
		Fields: map[string]interface{}{
			"fan_speed":                     NewFieldInfoC("gauge, Fan speed (or N/A)."),
			"memory_total":                  NewFieldInfoC("gauge, Framebuffer memory total (in MiB)."),
			"memory_used":                   NewFieldInfoC("gauge, Framebuffer memory used (in MiB)."),
			"temperature_gpu":               NewFieldInfoC("gauge, GPU temperature (in C)."),
			"utilization_gpu":               NewFieldInfoC("gauge, GPU utilization (in %)."),
			"utilization_memory":            NewFieldInfoC("gauge, Memory utilization (in %)."),
			"utilization_encoder":           NewFieldInfoC("gauge, Encoder utilization (in %)."),
			"utilization_decoder":           NewFieldInfoC("gauge, Decoder utilization (in %)."),
			"pcie_link_gen_current":         NewFieldInfoC("gauge, PCI-Express link gen."),
			"pcie_link_width_current":       NewFieldInfoC("gauge, PCI link width."),
			"encoder_stats_session_count":   NewFieldInfoC("count, Encoder session count."),
			"encoder_stats_average_fps":     NewFieldInfoC("count, Encoder average fps."),
			"encoder_stats_average_latency": NewFieldInfoC("count, Encoder average latency."),
			"fbc_stats_session_count":       NewFieldInfoC("count, Frame Buffer Cache session count."),
			"fbc_stats_average_fps":         NewFieldInfoC("count, Frame Buffer Cache average fps."),
			"fbc_stats_average_latency":     NewFieldInfoC("count, Frame Buffer Cache average latency."),
			"clocks_current_graphics":       NewFieldInfoC("gauge, Graphics clock frequency (in MHz)."),
			"clocks_current_sm":             NewFieldInfoC("gauge, Streaming Multiprocessor clock frequency (in MHz)."),
			"clocks_current_memory":         NewFieldInfoC("gauge, Memory clock frequency (in MHz)."),
			"clocks_current_video":          NewFieldInfoC("gauge, Video clock frequency (in MHz)."),
			"power_draw":                    NewFieldInfoP("gauge, Power draw."),
		},

		Tags: map[string]interface{}{
			"host": &inputs.TagInfo{Desc: "主机名"},
		},
	}
}

// GPUOnlineInfo Add GPU card && GPU online info.
func (ipt *Input) GPUOnlineInfo(uuid string) {
	for i := 0; i < len(ipt.GPUs); i++ {
		if ipt.GPUs[i].UUID == uuid {
			ipt.GPUs[i].ActiveTimestamp = time.Now().UnixNano() // 这个卡原来就在，重置活跃时间戳
			return
		}
	}
	ipt.GPUs = append(ipt.GPUs, GPUInfo{uuid, time.Now().UnixNano()}) // 新卡，加入队列

	// 发 info
	tagsLog := map[string]string{}
	fieldsLog := map[string]interface{}{}

	setTagIfUsed(tagsLog, "uuid", uuid)
	setTagIfUsed(tagsLog, "service", "msi_service")

	fieldsLog["status_gpu"] = 1
	fieldsLog["message"] = fmt.Sprintf("Info! GPU online! GPU UUID: %s", uuid)
	fieldsLog["status"] = "info"

	pt, err := point.NewPoint(
		metricName,
		tagsLog,
		fieldsLog,
		point.LOpt(),
	)
	if err != nil {
		l.Errorf("warning gpu_smi gpu online log: %s .", err)
	} else {
		ipt.pts = append(ipt.pts, pt)
	}
}

// GPUDropWarning handle GPU drop warning.
func (ipt *Input) GPUDropWarning() {
	for i := 0; i < len(ipt.GPUs); {
		// 回头可能删除若干条，这里有大坑，所以 i++ 放在后面
		if time.Now().UnixNano() > ipt.GPUs[i].ActiveTimestamp+ipt.GPUDropWarningDelay.Duration.Nanoseconds() {
			// 本张卡存活时间戳超阈值

			// 发 warning
			tagsLog := map[string]string{}
			fieldsLog := map[string]interface{}{}

			setTagIfUsed(tagsLog, "uuid", ipt.GPUs[i].UUID)
			setTagIfUsed(tagsLog, "service", "msi_service")

			fieldsLog["status_gpu"] = 2
			fieldsLog["message"] = fmt.Sprintf("Warning! GPU drop! GPU UUID: %s", ipt.GPUs[i].UUID)
			fieldsLog["status"] = "warning"

			pt, err := point.NewPoint(
				metricName,
				tagsLog,
				fieldsLog,
				point.LOpt(),
			)
			if err != nil {
				l.Errorf("warning gpu_smi gpu drop log: %s .", err)
			} else {
				ipt.pts = append(ipt.pts, pt)
			}

			// 删除这条gpuInfo，这里有大坑
			ipt.GPUs = append(ipt.GPUs[:i], ipt.GPUs[i+1:]...)
		} else {
			i++
		}
	}
}

// Collect Get, Aggregate Data.
func (ipt *Input) Collect() error {
	ipt.collectCache = make([]inputs.Measurement, 0)
	ipt.pts = make([]*point.Point, 0)

	// handle GPU drop warning
	ipt.GPUDropWarning()

	// cycle exec all binPath (if have multiple "nvidia-smi")
	for _, binPath := range ipt.BinPaths {
		// get data （[]byte）
		data, err := ipt.getBytes(binPath)
		if err != nil {
			l.Errorf("get bytes by binPath log: %s .", err)
			// 这里，出错，data如果!=nil，可能有有价值的信息，需要继续处理。
			if data == nil {
				continue
			}
		}
		// convert xml -> SMI{} struct
		smi := &SMI{}
		err = xml.Unmarshal(data, smi)
		if err != nil {
			l.Errorf("Unmarshal xml data log: %s .", err)
			continue // 不能return，否则，后面的msi运行不了
		}

		// convert to tags + fields
		metrics, metricsLog := smi.genTagsFields(ipt)

		// Append to the cache, the Run() function will handle it
		for _, metric := range metrics {
			ipt.collectCache = append(ipt.collectCache, &nvidiaSmiMeasurement{
				name:   metricName,
				tags:   metric.tags,
				fields: metric.fields,
			})
		}
		for i := 0; i < len(metricsLog); i++ {
			pt, err := point.NewPoint(
				metricName,
				metricsLog[i].tags,
				metricsLog[i].fields,
				point.LOpt(),
			)
			if err != nil {
				l.Errorf("collect gpu_smi prosess log: %s .", err)
			} else {
				ipt.pts = append(ipt.pts, pt)
			}
		}
	}
	return nil
}

// Get the result of binPath execution
// @binPath One of run bin files.
func (ipt *Input) getBytes(binPath string) ([]byte, error) {
	c := exec.Command(binPath, "-q", "-x")
	// 增加 exec.Command 环境变量
	c.Env = ipt.Envs

	var b bytes.Buffer
	c.Stdout = &b
	c.Stderr = &b
	if err := c.Start(); err != nil {
		return nil, err
	}
	err := datakit.WaitTimeout(c, ipt.Timeout.Duration)
	return b.Bytes(), err
}

// Run Start the process of timing acquisition.
// If this indicator is included in the list to be collected, it will only be called once.
// The for{} loops every tick.
func (ipt *Input) Run() {
	l = logger.SLogger(inputName)
	l.Infof("gpuSmi input started")
	ipt.Interval.Duration = config.ProtectedInterval(minInterval, maxInterval, ipt.Interval.Duration)
	tick := time.NewTicker(ipt.Interval.Duration)
	defer tick.Stop()

	for {
		start := time.Now()

		// Collect() to get data
		if err := ipt.Collect(); err != nil {
			l.Errorf("Collect: %s", err)
			io.FeedLastError(inputName, err.Error())
		}

		// If there is data in the cache, submit it
		if len(ipt.collectCache) > 0 {
			if err := inputs.FeedMeasurement(metricName, datakit.Metric, ipt.collectCache,
				&io.Option{CollectCost: time.Since(start)}); err != nil {
				l.Errorf("FeedMeasurement: %s", err)
			}
		}
		if len(ipt.pts) > 0 {
			err := io.Feed(inputName, datakit.Logging, ipt.pts, nil)
			if err != nil {
				l.Errorf("Feed gpu_smi prosess log: %s", err)
			}
		}

		select {
		case <-tick.C:

		case <-datakit.Exit.Wait():
			l.Infof("memory input exit")
			return

		case <-ipt.semStop.Wait():
			l.Infof("memory input return")
			return
		}
	}
}

// Terminate Stop.
func (ipt *Input) Terminate() {
	if ipt.semStop != nil {
		ipt.semStop.Close()
	}
}

// Catalog Catalog.
func (*Input) Catalog() string {
	return "gpu_smi"
}

// SampleConfig : conf File samples, reflected in the document.
func (*Input) SampleConfig() string {
	return sampleCfg
}

// AvailableArchs : OS support, reflected in the document.
func (*Input) AvailableArchs() []string {
	return []string{datakit.OSLabelLinux, datakit.OSLabelWindows, datakit.LabelK8s}
}

func (*Input) AvailableArchsDCGM() []string {
	return []string{datakit.OSLabelLinux, datakit.LabelK8s}
}

// SampleMeasurement Sample measurement results, reflected in the document.
func (*Input) SampleMeasurement() []inputs.Measurement {
	return []inputs.Measurement{
		&nvidiaSmiMeasurement{},
	}
}

// ReadEnv support envs：only for K8S
//   ENV_INPUT_GPUSMI_TAGS : "a=b,c=d"
//   ENV_INPUT_GPUSMI_INTERVAL : datakit.Duration
func (ipt *Input) ReadEnv(envs map[string]string) {
	if tagsStr, ok := envs["ENV_INPUT_NVIDIASMI_TAGS"]; ok {
		tags := config.ParseGlobalTags(tagsStr)
		for k, v := range tags {
			ipt.Tags[k] = v
		}
	}
	// To adapt to the variable name nvidia_smi --> gpu_smi
	if tagsStr, ok := envs["ENV_INPUT_GPUSMI_TAGS"]; ok {
		tags := config.ParseGlobalTags(tagsStr)
		for k, v := range tags {
			ipt.Tags[k] = v
		}
	}

	//   ENV_INPUT_NVIDIASMI_INTERVAL : datakit.Duration
	if str, ok := envs["ENV_INPUT_NVIDIASMI_INTERVAL"]; ok {
		da, err := time.ParseDuration(str)
		if err != nil {
			l.Warnf("parse ENV_INPUT_NVIDIASMI_INTERVAL to time.Duration: %s, ignore", err)
		} else {
			ipt.Interval.Duration = config.ProtectedInterval(minInterval,
				maxInterval,
				da)
		}
	}
	// To adapt to the variable name nvidia_smi --> gpu_smi
	if str, ok := envs["ENV_INPUT_GPUSMI_INTERVAL"]; ok {
		da, err := time.ParseDuration(str)
		if err != nil {
			l.Warnf("parse ENV_INPUT_NVIDIASMI_INTERVAL to time.Duration: %s, ignore", err)
		} else {
			ipt.Interval.Duration = config.ProtectedInterval(minInterval,
				maxInterval,
				da)
		}
	}
}

func newDefaultInput() *Input {
	ipt := &Input{
		platform:            runtime.GOOS,
		Interval:            datakit.Duration{Duration: time.Second * 10},
		semStop:             cliutils.NewSem(),
		Tags:                make(map[string]string),
		BinPaths:            []string{"/usr/bin/nvidia-smi"},
		Timeout:             datakit.Duration{Duration: time.Second * 5},
		ProcessInfoMaxLen:   10,
		GPUDropWarningDelay: datakit.Duration{Duration: time.Second * 300},
		GPUs:                make([]GPUInfo, 0, 1),
		Envs:                []string{},
	}
	return ipt
}

func init() { //nolint:gochecknoinits
	inputs.Add(inputName, func() inputs.Input {
		return newDefaultInput()
	})

	// To adapt to the variable name nvidia_smi --> gpu_smi in the .conf file
	inputs.Add("nvidia_smi", func() inputs.Input {
		return newDefaultInput()
	})
}

type metric struct {
	tags   map[string]string
	fields map[string]interface{}
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
	UUID        string           `xml:"uuid"`
	ComputeMode string           `xml:"compute_mode"`
	Utilization UtilizationStats `xml:"utilization"`
	Power       PowerReadings    `xml:"power_readings"`
	PCI         PCI              `xml:"pci"`
	Encoder     EncoderStats     `xml:"encoder_stats"`
	FBC         FBCStats         `xml:"fbc_stats"`
	Clocks      ClockStats       `xml:"clocks"`
	Processes   Processes        `xml:"processes"`
}

// MemoryStats defines the structure of the memory portions in the smi output.
type MemoryStats struct {
	Total string `xml:"total"` // int
	Used  string `xml:"used"`  // int
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
	PowerDraw string `xml:"power_draw"` // float
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

func (s *SMI) genTagsFields(ipt *Input) ([]metric, []metric) {
	metrics := []metric{}
	metricsLog := []metric{}
	for i, gpu := range s.GPU {
		// handle GPU online info
		ipt.GPUOnlineInfo(gpu.UUID)

		tags := map[string]string{
			"index": strconv.Itoa(i),
		}
		fields := map[string]interface{}{}

		setTagIfUsed(tags, "pstate", gpu.PState)
		setTagIfUsed(tags, "name", gpu.ProdName)
		setTagIfUsed(tags, "uuid", gpu.UUID)
		setTagIfUsed(tags, "compute_mode", gpu.ComputeMode)
		setTagIfUsed(tags, "pci_bus_id", gpu.PCI.PciBusID)

		setTagIfUsed(tags, "driver_version", s.DriverVersion)
		setTagIfUsed(tags, "cuda_version", s.CUDAVersion)

		setIfUsed("int", fields, "fan_speed", gpu.FanSpeed)
		setIfUsed("int", fields, "memory_total", gpu.Memory.Total)
		setIfUsed("int", fields, "memory_used", gpu.Memory.Used)

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
		metrics = append(metrics, metric{tags, fields})

		// Sort and collect the processInfoMaxLen with the largest memory usage, default 10
		if ipt.ProcessInfoMaxLen == 0 {
			continue
		}

		// get gpu processes info for log
		if len(gpu.Processes.ProcessInfos) > 0 {
			// sort by usedMemory
			processIndex := make([][2]int, 0, 2) // cap=2, can reduce cpu
			usedMemory := 0
			var err error
			for j := 0; j < len(gpu.Processes.ProcessInfos); j++ {
				vals := strings.Fields(gpu.Processes.ProcessInfos[j].UsedMemory)
				if len(vals) < 1 {
					usedMemory = 0
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

			// pick processInfoMaxLen data
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
				setTagIfUsed(tagsLog, "pstate", gpu.PState)
				setTagIfUsed(tagsLog, "name", gpu.ProdName)
				setTagIfUsed(tagsLog, "uuid", gpu.UUID)
				setTagIfUsed(tagsLog, "pci_bus_id", gpu.PCI.PciBusID)
				setTagIfUsed(tagsLog, "compute_mode", gpu.ComputeMode)
				setTagIfUsed(tagsLog, "service", "msi_service")

				s := strings.Split(gpu.Processes.ProcessInfos[processIndex[j][1]].ProcessName, "/") // 去除路径名字
				fieldsLog["message"] = fmt.Sprintf("%s:ProcessName=%s,UsedMemory= %s",
					gpu.UUID,
					s[len(s)-1],
					gpu.Processes.ProcessInfos[processIndex[j][1]].UsedMemory)
				fieldsLog["gpu_instance_id"] = gpu.Processes.ProcessInfos[processIndex[j][1]].GpuInstanceID
				fieldsLog["compute_instance_id"] = gpu.Processes.ProcessInfos[processIndex[j][1]].ComputeInstanceID
				setIfUsed("int", fieldsLog, "pid", gpu.Processes.ProcessInfos[processIndex[j][1]].Pid)
				setIfUsed("str", fieldsLog, "type", gpu.Processes.ProcessInfos[processIndex[j][1]].Type)
				setIfUsed("str", fieldsLog, "process_name", gpu.Processes.ProcessInfos[processIndex[j][1]].ProcessName)
				fieldsLog["used_memory"] = processIndex[j][0]
				fieldsLog["status"] = "info"

				metricsLog = append(metricsLog, metric{tagsLog, fieldsLog})
			}
		}
	}
	return metrics, metricsLog
}
