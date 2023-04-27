// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2022-present Guance, Inc.

// Package nvidiasmi collects host nvidiasmi metrics.
package nvidiasmi

import (
	"encoding/json"
	"fmt"
	"runtime"
	"strconv"
	"time"

	"github.com/GuanceCloud/cliutils"
	"github.com/GuanceCloud/cliutils/logger"
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

  ##(Optional) Collect interval, default is 10 seconds
  interval = "10s"

  ##The binPath of gpu-smi 

  ##If nvidia GPU
  #(Example & default) bin_paths = ["/usr/bin/nvidia-smi"]
  #(Example windows) bin_paths = ["nvidia-smi"]

  ##If lluvatar GPU
  #(Example) bin_paths = ["/usr/local/corex/bin/ixsmi"]
  #(Example) envs = [ "LD_LIBRARY_PATH=/usr/local/corex/lib/:$LD_LIBRARY_PATH" ]
  ##(Optional) Exec gpu-smi envs, default is []
  #envs = [ "LD_LIBRARY_PATH=/usr/local/corex/lib/:$LD_LIBRARY_PATH" ]

  ##If remote GPU servers collected
  ##If use remote GPU servers, election must be true
  ##If use remote GPU servers, bin_paths should be shielded
  #(Example) remote_addrs = ["192.168.1.1:22"]
  #(Example) remote_users = ["remote_login_name"]
  ##If use remote_rsa_path, remote_passwords should be shielded
  #(Example) remote_passwords = ["remote_login_password"]
  #(Example) remote_rsa_paths = ["/home/your_name/.ssh/id_rsa"]
  #(Example) remote_command = "nvidia-smi -x -q"

  ##(Optional) Exec gpu-smi timeout, default is 5 seconds
  timeout = "5s"
  ##(Optional) Feed how much log data for ProcessInfos, default is 10. (0: 0 ,-1: all)
  process_info_max_len = 10
  ##(Optional) GPU drop card warning delay, default is 300 seconds
  gpu_drop_warning_delay = "300s"

  ## Set true to enable election
  election = false

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
	Timeout             datakit.Duration `toml:"timeout"`                // "nvidia-smi" timeout
	ProcessInfoMaxLen   int              `toml:"process_info_max_len"`   // Feed how much log data for ProcessInfos. (0: 0 ,-1: all)
	GPUDropWarningDelay datakit.Duration `toml:"gpu_drop_warning_delay"` // GPU drop card warning delay
	Envs                []string         `toml:"envs"`                   // exec.Command ENV
	datakit.SSHServers

	GPUs    []GPUInfo     // online GPU card list
	semStop *cliutils.Sem // start stop signal

	Election bool `toml:"election"`
	pause    bool
	pauseCh  chan bool
}

type GPUInfo struct {
	UUID            string
	Server          string // remote server like "1.1.1.1:22"
	ActiveTimestamp int64  // alive timestamp nanosecond time.Now().UnixNano()
}

// nvidiaSmi Measurement structure.
type nvidiaSmiMeasurement struct {
	name     string                 // Indicator set name ，here is "gpu_smi"
	tags     map[string]string      // Indicator name
	fields   map[string]interface{} // Indicator measurement results
	election bool
}

// LineProto data formatting, submit through FeedMeasurement.
func (n *nvidiaSmiMeasurement) LineProto() (*point.Point, error) {
	return point.NewPoint(n.name, n.tags, n.fields, point.MOptElectionV2(n.election))
}

// Info , reflected in the document.
// nolint:lll
func (n *nvidiaSmiMeasurement) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name: metricName,
		Fields: map[string]interface{}{
			"fan_speed":                     &inputs.FieldInfo{Type: inputs.Rate, DataType: inputs.Int, Unit: inputs.RPMPercent, Desc: "Fan speed."},
			"memory_total":                  &inputs.FieldInfo{Type: inputs.Gauge, DataType: inputs.Int, Unit: inputs.SizeMB, Desc: "Framebuffer memory total."},
			"memory_used":                   &inputs.FieldInfo{Type: inputs.Gauge, DataType: inputs.Int, Unit: inputs.SizeMB, Desc: "Framebuffer memory used."},
			"temperature_gpu":               &inputs.FieldInfo{Type: inputs.Gauge, DataType: inputs.Int, Unit: inputs.Celsius, Desc: "GPU temperature."},
			"utilization_gpu":               &inputs.FieldInfo{Type: inputs.Gauge, DataType: inputs.Int, Unit: inputs.Percent, Desc: "GPU utilization."},
			"utilization_memory":            &inputs.FieldInfo{Type: inputs.Gauge, DataType: inputs.Int, Unit: inputs.Percent, Desc: "Memory utilization."},
			"utilization_encoder":           &inputs.FieldInfo{Type: inputs.Gauge, DataType: inputs.Int, Unit: inputs.Percent, Desc: "Encoder utilization."},
			"utilization_decoder":           &inputs.FieldInfo{Type: inputs.Gauge, DataType: inputs.Int, Unit: inputs.Percent, Desc: "Decoder utilization."},
			"pcie_link_gen_current":         &inputs.FieldInfo{Type: inputs.Gauge, DataType: inputs.Int, Unit: inputs.UnknownUnit, Desc: "PCI-Express link gen."},
			"pcie_link_width_current":       &inputs.FieldInfo{Type: inputs.Gauge, DataType: inputs.Int, Unit: inputs.UnknownUnit, Desc: "PCI link width."},
			"encoder_stats_session_count":   &inputs.FieldInfo{Type: inputs.Count, DataType: inputs.Int, Unit: inputs.NCount, Desc: "Encoder session count."},
			"encoder_stats_average_fps":     &inputs.FieldInfo{Type: inputs.Rate, DataType: inputs.Int, Unit: inputs.UnknownUnit, Desc: "Encoder average fps."},
			"encoder_stats_average_latency": &inputs.FieldInfo{Type: inputs.Count, DataType: inputs.Int, Unit: inputs.UnknownUnit, Desc: "Encoder average latency."},
			"fbc_stats_session_count":       &inputs.FieldInfo{Type: inputs.Count, DataType: inputs.Int, Unit: inputs.UnknownUnit, Desc: "Frame Buffer Cache session count."},
			"fbc_stats_average_fps":         &inputs.FieldInfo{Type: inputs.Count, DataType: inputs.Int, Unit: inputs.UnknownUnit, Desc: "Frame Buffer Cache average fps."},
			"fbc_stats_average_latency":     &inputs.FieldInfo{Type: inputs.Count, DataType: inputs.Int, Unit: inputs.UnknownUnit, Desc: "Frame Buffer Cache average latency."},
			"clocks_current_graphics":       &inputs.FieldInfo{Type: inputs.Gauge, DataType: inputs.Int, Unit: inputs.FrequencyMHz, Desc: "Graphics clock frequency."},
			"clocks_current_sm":             &inputs.FieldInfo{Type: inputs.Gauge, DataType: inputs.Int, Unit: inputs.FrequencyMHz, Desc: "Streaming Multiprocessor clock frequency."},
			"clocks_current_memory":         &inputs.FieldInfo{Type: inputs.Gauge, DataType: inputs.Int, Unit: inputs.FrequencyMHz, Desc: "Memory clock frequency."},
			"clocks_current_video":          &inputs.FieldInfo{Type: inputs.Gauge, DataType: inputs.Int, Unit: inputs.FrequencyMHz, Desc: "Video clock frequency."},
			"power_draw":                    &inputs.FieldInfo{Type: inputs.Gauge, DataType: inputs.Float, Unit: inputs.Watt, Desc: "Power draw."},
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

// GPUOnlineInfo Add GPU card && GPU online info.
func (ipt *Input) GPUOnlineInfo(uuid, server string) {
	for i := 0; i < len(ipt.GPUs); i++ {
		if ipt.GPUs[i].UUID == uuid {
			ipt.GPUs[i].ActiveTimestamp = time.Now().UnixNano() // 这个卡原来就在，重置活跃时间戳
			return
		}
	}
	ipt.GPUs = append(ipt.GPUs, GPUInfo{uuid, server, time.Now().UnixNano()}) // 新卡，加入队列

	// 发 info
	tagsLog := map[string]string{}
	fieldsLog := map[string]interface{}{}

	if server != "" {
		// for ssh remote server
		setTagIfUsed(tagsLog, "host", server)
	}
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
		l.Errorf("warning gpu_smi GPU online log: %s .", err)
	} else {
		ipt.pts = append(ipt.pts, pt)
	}
}

// GPUDropWarning handle GPU drop warning.
func (ipt *Input) GPUDropWarning() {
	for i := 0; i < len(ipt.GPUs); {
		// Several items may be deleted, so i++ is placed later
		if time.Now().UnixNano() > ipt.GPUs[i].ActiveTimestamp+ipt.GPUDropWarningDelay.Duration.Nanoseconds() {
			// The survival time stamp of this card exceeds the threshold

			// warning
			tagsLog := map[string]string{}
			fieldsLog := map[string]interface{}{}

			setTagIfUsed(tagsLog, "uuid", ipt.GPUs[i].UUID)
			setTagIfUsed(tagsLog, "service", "msi_service")
			if ipt.GPUs[i].Server != "" {
				// for ssh remote server
				setTagIfUsed(tagsLog, "host", ipt.GPUs[i].Server)
			}

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
				l.Errorf("warning gpu_smi GPU drop log: %s .", err)
			} else {
				ipt.pts = append(ipt.pts, pt)
			}

			// delete this gpuInfo
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

	if len(ipt.RemoteAddrs) > 0 {
		_ = ipt.collectRemote()
	} else if len(ipt.BinPaths) > 0 {
		// if use remote GPU servers, bin_paths should be shielded
		_ = ipt.collectLocal()
	}

	// handle GPU drop warning
	ipt.GPUDropWarning()

	return nil
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

		if ipt.pause {
			l.Debugf("not leader, gpu_smi skipped")
		} else {
			l.Debugf("is leader, gpu_smi gathering...")
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
					l.Errorf("Feed gpu_smi process log: %s", err)
				}
			}
		}

		select {
		case <-tick.C:
		case ipt.pause = <-ipt.pauseCh:
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
func (*Input) AvailableArchs() []string { return datakit.AllOSWithElection }

func (*Input) AvailableArchsDCGM() []string {
	return []string{datakit.OSLabelLinux, datakit.LabelK8s}
}

// SampleMeasurement Sample measurement results, reflected in the document.
func (*Input) SampleMeasurement() []inputs.Measurement {
	return []inputs.Measurement{
		&nvidiaSmiMeasurement{},
	}
}

// ReadEnv support envs：only for K8S.
func (ipt *Input) ReadEnv(envs map[string]string) {
	if tagsStr, ok := envs["ENV_INPUT_NVIDIASMI_TAGS"]; ok {
		tags := config.ParseGlobalTags(tagsStr)
		for k, v := range tags {
			ipt.Tags[k] = v
		}
	}

	//   ENV_INPUT_GPUSMI_TAGS : "a=b,c=d"
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

	if str, ok := envs["ENV_INPUT_GPUSMI_BIN_PATHS"]; ok {
		var strs []string
		err := json.Unmarshal([]byte(str), &strs)
		if err != nil {
			l.Warnf("parse ENV_INPUT_GPUSMI_BIN_PATHS: %s, ignore", err)
		} else {
			ipt.BinPaths = strs
		}
	}

	if str, ok := envs["ENV_INPUT_GPUSMI_TIMEOUT"]; ok {
		da, err := time.ParseDuration(str)
		if err != nil {
			l.Warnf("parse ENV_INPUT_GPUSMI_TIMEOUT to time.Duration: %s, ignore", err)
		} else {
			ipt.Timeout.Duration = config.ProtectedInterval(minInterval,
				maxInterval,
				da)
		}
	}

	if str, ok := envs["ENV_INPUT_GPUSMI_PROCESS_INFO_MAX_LEN"]; ok {
		i, err := strconv.Atoi(str)
		if err != nil {
			l.Warnf("parse ENV_INPUT_GPUSMI_PROCESS_INFO_MAX_LEN: %s, ignore", err)
		} else {
			ipt.ProcessInfoMaxLen = i
		}
	}

	if str, ok := envs["ENV_INPUT_GPUSMI_DROP_WARNING_DELAY"]; ok {
		da, err := time.ParseDuration(str)
		if err != nil {
			l.Warnf("parse ENV_INPUT_GPUSMI_DROP_WARNING_DELAY to time.Duration: %s, ignore", err)
		} else {
			ipt.GPUDropWarningDelay.Duration = config.ProtectedInterval(minInterval,
				maxInterval,
				da)
		}
	}

	if str, ok := envs["ENV_INPUT_GPUSMI_ENVS"]; ok {
		var strs []string
		err := json.Unmarshal([]byte(str), &strs)
		if err != nil {
			l.Warnf("parse ENV_INPUT_GPUSMI_ENVS: %s, ignore", err)
		} else {
			ipt.Envs = strs
		}
	}

	if str, ok := envs["ENV_INPUT_GPUSMI_REMOTE_ADDRS"]; ok {
		var strs []string
		err := json.Unmarshal([]byte(str), &strs)
		if err != nil {
			l.Warnf("parse ENV_INPUT_GPUSMI_REMOTE_ADDRS: %s, ignore", err)
		} else {
			ipt.RemoteAddrs = strs
		}
	}

	if str, ok := envs["ENV_INPUT_GPUSMI_REMOTE_USERS"]; ok {
		var strs []string
		err := json.Unmarshal([]byte(str), &strs)
		if err != nil {
			l.Warnf("parse ENV_INPUT_GPUSMI_REMOTE_USERS: %s, ignore", err)
		} else {
			ipt.RemoteUsers = strs
		}
	}

	if str, ok := envs["ENV_INPUT_GPUSMI_REMOTE_PASSWORDS"]; ok {
		var strs []string
		err := json.Unmarshal([]byte(str), &strs)
		if err != nil {
			l.Warnf("parse ENV_INPUT_GPUSMI_REMOTE_PASSWORDS: %s, ignore", err)
		} else {
			ipt.RemotePasswords = strs
		}
	}

	if str, ok := envs["ENV_INPUT_GPUSMI_REMOTE_RSA_PATHS"]; ok {
		var strs []string
		err := json.Unmarshal([]byte(str), &strs)
		if err != nil {
			l.Warnf("parse ENV_INPUT_GPUSMI_REMOTE_RSA_PATHS: %s, ignore", err)
		} else {
			ipt.RemoteRsaPaths = strs
		}
	}

	if str, ok := envs["ENV_INPUT_GPUSMI_REMOTE_COMMAND"]; ok {
		ipt.RemoteCommand = str
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
		pauseCh:             make(chan bool, inputs.ElectionPauseChannelLength),
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

// ElectionEnabled election.
func (ipt *Input) ElectionEnabled() bool {
	return ipt.Election
}

func (ipt *Input) Pause() error {
	tick := time.NewTicker(inputs.ElectionPauseTimeout)
	defer tick.Stop()
	select {
	case ipt.pauseCh <- true:
		return nil
	case <-tick.C:
		return fmt.Errorf("pause %s failed", inputName)
	}
}

func (ipt *Input) Resume() error {
	tick := time.NewTicker(inputs.ElectionResumeTimeout)
	defer tick.Stop()
	select {
	case ipt.pauseCh <- false:
		return nil
	case <-tick.C:
		return fmt.Errorf("resume %s failed", inputName)
	}
}
