// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2022-present Guance, Inc.

// Package nvidiasmi collects host nvidiasmi metrics.
package nvidiasmi

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"os/exec"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/GuanceCloud/cliutils"
	"github.com/GuanceCloud/cliutils/logger"
	"github.com/GuanceCloud/cliutils/point"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/config"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/export/doc"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/getdatassh"
	dkio "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/io"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs"
)

const (
	minInterval = time.Second
	maxInterval = time.Minute
	inputName   = "gpu_smi" // old is "nvidia_smi" , now is "gpu_smi".
	metricName  = inputName

	defaultInterval = time.Second * 10
	defaultTimeout  = time.Second * 5
)

var (
	_ inputs.ReadEnv = (*Input)(nil)
	l                = logger.DefaultSLogger(inputName)

	_ inputs.ElectionInput = (*Input)(nil)
)

type gpuInfo struct {
	uuid            string
	server          string // remote server like "1.1.1.1:22"
	activeTimestamp int64  // alive timestamp nanosecond time.Now().UnixNano()
}

type urlTags map[string]string

type Input struct {
	Interval            time.Duration     // run every "Interval" seconds
	Tags                map[string]string // Indicator name
	platform            string
	BinPaths            []string      `toml:"bin_paths"`              // the file path of "nvidia-smi"
	Timeout             time.Duration `toml:"timeout"`                // "nvidia-smi" timeout
	ProcessInfoMaxLen   int           `toml:"process_info_max_len"`   // Feed how much log data for ProcessInfos. (0: 0 ,-1: all)
	GPUDropWarningDelay time.Duration `toml:"gpu_drop_warning_delay"` // GPU drop card warning delay
	Envs                []string      `toml:"envs"`                   // exec.Command ENV
	getdatassh.SSHServers

	semStop          *cliutils.Sem
	collectCache     []*point.Point
	collectCacheLog  []*point.Point
	collectCacheWarn []*point.Point
	feeder           dkio.Feeder
	mergedTags       map[string]urlTags
	tagger           datakit.GlobalTagger

	Election bool `toml:"election"`
	pause    bool
	pauseCh  chan bool

	gpus   []gpuInfo // online GPU card list
	gpusMu sync.Mutex
}

// Run Start the process of timing acquisition.
// If this indicator is included in the list to be collected, it will only be called once.
// The for{} loops every tick.
func (ipt *Input) Run() {
	if err := ipt.setup(); err != nil {
		l.Errorf("setup err: %v", err)
		return
	}

	tick := time.NewTicker(ipt.Interval)
	defer tick.Stop()

	for {
		start := time.Now()

		if ipt.pause {
			l.Debugf("not leader, %s skipped", inputName)
		} else {
			if err := ipt.collect(); err != nil {
				l.Errorf("collect: %s", err)
				ipt.feeder.FeedLastError(err.Error(),
					dkio.WithLastErrorInput(inputName),
					dkio.WithLastErrorCategory(point.Metric),
				)
			}

			if len(ipt.collectCache) > 0 {
				if err := ipt.feeder.Feed(metricName, point.Metric, ipt.collectCache,
					&dkio.Option{CollectCost: time.Since(start)}); err != nil {
					ipt.feeder.FeedLastError(err.Error(),
						dkio.WithLastErrorInput(inputName),
						dkio.WithLastErrorCategory(point.Metric),
					)
					l.Errorf("feed measurement: %s", err)
				}
			}

			if len(ipt.collectCacheLog) > 0 {
				if err := ipt.feeder.Feed(metricName, point.Logging, ipt.collectCacheLog,
					&dkio.Option{CollectCost: time.Since(start)}); err != nil {
					ipt.feeder.FeedLastError(err.Error(),
						dkio.WithLastErrorInput(inputName),
						dkio.WithLastErrorCategory(point.Metric),
					)
					l.Errorf("feed logging: %s", err)
				}
			}

			if len(ipt.collectCacheWarn) > 0 {
				if err := ipt.feeder.Feed(metricName, point.Logging, ipt.collectCacheWarn,
					&dkio.Option{CollectCost: time.Since(start)}); err != nil {
					ipt.feeder.FeedLastError(err.Error(),
						dkio.WithLastErrorInput(inputName),
						dkio.WithLastErrorCategory(point.Metric),
					)
					l.Errorf("feed warning: %s", err)
				}
			}
		}

		select {
		case <-tick.C:
		case <-datakit.Exit.Wait():
			l.Infof("%s input exit", inputName)
			return
		case <-ipt.semStop.Wait():
			l.Infof("%s input return", inputName)
			return
		case ipt.pause = <-ipt.pauseCh:
		}
	}
}

func (ipt *Input) setup() error {
	l = logger.SLogger(inputName)

	l.Infof("%s input started", inputName)
	ipt.Interval = config.ProtectedInterval(minInterval, maxInterval, ipt.Interval)
	ipt.Timeout = config.ProtectedInterval(minInterval, maxInterval, ipt.Timeout)

	if err := ipt.checkConf(); err != nil {
		return err
	}

	return nil
}

// checkConf check binPath & datakit.SSHServers.
func (ipt *Input) checkConf() error {
	if len(ipt.BinPaths) == 0 && len(ipt.RemoteAddrs) == 0 {
		return fmt.Errorf("remote_addrs & bin_path all be null")
	}
	if len(ipt.RemoteAddrs) > 0 {
		if len(ipt.RemoteCommand) == 0 {
			return fmt.Errorf("remote_command is null")
		}
		if (len(ipt.RemoteUsers) == 0 || len(ipt.RemotePasswords) == 0) && len(ipt.RemoteRsaPaths) == 0 {
			return fmt.Errorf("remote_users & remote_passwords & remote_rsa_paths all be null")
		}
		for _, u := range ipt.RemoteAddrs {
			uu := u
			if !strings.HasPrefix(u, "http") {
				u = "http://" + u
			}
			_, err := url.Parse(u)
			if err != nil {
				return fmt.Errorf("parse remote_addrs : %s, error : %w", u, err)
			}

			if ipt.Election {
				ipt.mergedTags[uu] = inputs.MergeTags(ipt.tagger.ElectionTags(), ipt.Tags, u)
			} else {
				ipt.mergedTags[uu] = inputs.MergeTags(ipt.tagger.HostTags(), ipt.Tags, u)
			}
		}
		l.Debugf("merged tags: %+#v", ipt.mergedTags)
	} else {
		if ipt.Election {
			ipt.mergedTags["localhost"] = inputs.MergeTags(ipt.tagger.ElectionTags(), ipt.Tags, "")
		} else {
			ipt.mergedTags["localhost"] = inputs.MergeTags(ipt.tagger.HostTags(), ipt.Tags, "")
		}
	}
	return nil
}

func (ipt *Input) collect() error {
	ipt.collectCache = make([]*point.Point, 0)
	ipt.collectCacheLog = make([]*point.Point, 0)
	ipt.collectCacheWarn = make([]*point.Point, 0)

	ipt.gpuDropWarning()

	data, err := ipt.getData()
	if err != nil {
		return err
	}

	// convert and calculate GPU metrics
	for _, v := range data {
		if err = ipt.getPts(v.Data, v.Server); err != nil {
			return err
		}
	}

	return nil
}

func (ipt *Input) getData() ([]*getdatassh.SSHData, error) {
	if len(ipt.SSHServers.RemoteAddrs) > 0 {
		// Remote servers
		return getdatassh.GetDataSSH(&ipt.SSHServers, ipt.Timeout)
	} else if len(ipt.BinPaths) > 0 {
		// Local server
		data := make([]*getdatassh.SSHData, 0)
		for _, binPath := range ipt.BinPaths {
			ctx, cancel := context.WithTimeout(context.Background(), ipt.Timeout)
			defer cancel()
			//nolint:gosec
			c := exec.CommandContext(ctx, binPath, "-q", "-x")
			// dd exec.Command ENV
			if len(ipt.Envs) != 0 {
				// in windows here will broken old PATH
				c.Env = ipt.Envs
			}

			var b bytes.Buffer
			c.Stdout = &b
			c.Stderr = &b
			if err := c.Start(); err != nil {
				return nil, fmt.Errorf("c.Start(): %w, %v", err, b.String())
			}
			err := c.Wait()
			if err != nil {
				return nil, fmt.Errorf("c.Wait(): %s, %w, %v", inputName, err, b.String())
			}

			bytes := b.Bytes()
			l.Debugf("get bytes len: %v.", len(bytes))

			data = append(data, &getdatassh.SSHData{
				Server: "localhost",
				Data:   bytes,
			})
		}

		return data, nil
	}

	return nil, fmt.Errorf("%s got no data", inputName)
}

// Handle GPU online info.
func (ipt *Input) gpuOnlineInfo(uuid, server string) {
	ipt.gpusMu.Lock()
	defer ipt.gpusMu.Unlock()

	ts := time.Now()
	opts := point.DefaultMetricOptions()
	opts = append(opts, point.WithTime(ts))

	for i := 0; i < len(ipt.gpus); i++ {
		if ipt.gpus[i].uuid == uuid {
			ipt.gpus[i].activeTimestamp = time.Now().UnixNano() // reset timestamp
			return
		}
	}
	ipt.gpus = append(ipt.gpus, gpuInfo{uuid, server, time.Now().UnixNano()}) // new card

	// send info
	var kvs point.KVs

	kvs = kvs.Add("uuid", uuid, false, true)
	kvs = kvs.Add("service", "msi_service", false, true)

	kvs = kvs.Add("status_gpu", 1, false, true)
	kvs = kvs.Add("message", fmt.Sprintf("Info! GPU online! GPU UUID: %s", uuid), false, true)
	kvs = kvs.Add("status", "info", false, true)

	for k, v := range ipt.mergedTags[server] {
		kvs = kvs.AddTag(k, v)
	}

	ipt.collectCacheWarn = append(ipt.collectCacheWarn, point.NewPointV2(inputName, kvs, opts...))
}

// Handle GPU drop warning.
func (ipt *Input) gpuDropWarning() {
	ipt.gpusMu.Lock()
	defer ipt.gpusMu.Unlock()

	ts := time.Now()
	opts := point.DefaultMetricOptions()
	opts = append(opts, point.WithTime(ts))

	for i := 0; i < len(ipt.gpus); {
		// Several items may be deleted, so i++ is placed later
		if time.Now().UnixNano() > ipt.gpus[i].activeTimestamp+ipt.GPUDropWarningDelay.Nanoseconds() {
			// The survival time stamp of this card exceeds the threshold

			// warning
			var kvs point.KVs

			kvs = kvs.Add("uuid", ipt.gpus[i].uuid, false, true)
			kvs = kvs.Add("service", "msi_service", false, true)

			kvs = kvs.Add("status_gpu", 2, false, true)
			kvs = kvs.Add("message", fmt.Sprintf("Warning! GPU drop! GPU UUID: %s", ipt.gpus[i].uuid), false, true)
			kvs = kvs.Add("status", "warning", false, true)

			for k, v := range ipt.mergedTags[ipt.gpus[i].server] {
				kvs = kvs.AddTag(k, v)
			}

			ipt.collectCacheWarn = append(ipt.collectCacheWarn, point.NewPointV2(inputName, kvs, opts...))
		} else {
			i++
		}
	}
}

func (ipt *Input) Terminate() {
	if ipt.semStop != nil {
		ipt.semStop.Close()
	}
}
func (*Input) Catalog() string          { return inputName }
func (*Input) SampleConfig() string     { return sampleCfg }
func (*Input) AvailableArchs() []string { return datakit.AllOSWithElection }
func (*Input) SampleMeasurement() []inputs.Measurement {
	return []inputs.Measurement{
		&docMeasurement{},
	}
}

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

func (ipt *Input) GetENVDoc() []*inputs.ENVInfo {
	// nolint:lll
	infos := []*inputs.ENVInfo{
		{FieldName: "Interval"},
		{FieldName: "Timeout", Default: `5s`},
		{FieldName: "BinPath", Type: doc.JSON, Example: "`[\"/usr/bin/nvidia-smi\"]`", Desc: "The binPath", DescZh: "执行文件路径"},
		{FieldName: "ProcessInfoMaxLen", Type: doc.Int, Default: `10`, Desc: "Maximum number of GPU processes that consume the most resources", DescZh: "最大收集最耗资源 GPU 进程数"},
		{FieldName: "GPUDropWarningDelay", Type: doc.TimeDuration, ENVName: "DROP_WARNING_DELAY", ConfField: "gpu_drop_warning_delay", Default: `5m`, Desc: "GPU card drop warning delay", DescZh: "掉卡告警延迟"},
		{FieldName: "Envs", Type: doc.JSON, Example: `["LD_LIBRARY_PATH=/usr/local/corex/lib/:$LD_LIBRARY_PATH"]`, Desc: "The envs of LD_LIBRARY_PATH", DescZh: "执行依赖库的路径"},
		{FieldName: "RemoteAddrs", Type: doc.JSON, Example: `["192.168.1.1:22","192.168.1.2:22"]`, Desc: "If use remote GPU servers", DescZh: "远程 GPU 服务器"},
		{FieldName: "RemoteUsers", Type: doc.JSON, Example: `["user_1","user_2"]`, Desc: "Remote login name", DescZh: "远程登录名"},
		{FieldName: "RemotePasswords", Type: doc.JSON, Example: `["pass_1","pass_2"]`, Desc: "Remote password", DescZh: "远程登录密码"},
		{FieldName: "RemoteRsaPaths", Type: doc.JSON, Example: `["/home/your_name/.ssh/id_rsa"]`, Desc: "Remote rsa paths", DescZh: "秘钥文件路径"},
		{FieldName: "RemoteCommand", Type: doc.String, Example: "\"`nvidia-smi -x -q`\"", Desc: "Remote command", DescZh: "远程执行指令"},
		{FieldName: "Election"},
		{FieldName: "Tags"},
	}

	return doc.SetENVDoc("ENV_INPUT_GPUSMI_", infos)
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

	//   ENV_INPUT_NVIDIASMI_INTERVAL : time.Duration
	if str, ok := envs["ENV_INPUT_NVIDIASMI_INTERVAL"]; ok {
		da, err := time.ParseDuration(str)
		if err != nil {
			l.Warnf("parse ENV_INPUT_NVIDIASMI_INTERVAL to time.Duration: %s, ignore", err)
		} else {
			ipt.Interval = config.ProtectedInterval(minInterval,
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
			ipt.Interval = config.ProtectedInterval(minInterval,
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
			ipt.Timeout = config.ProtectedInterval(minInterval,
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
			ipt.GPUDropWarningDelay = config.ProtectedInterval(minInterval,
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

	if str, ok := envs["ENV_INPUT_GPUSMI_ELECTION"]; ok {
		election, err := strconv.ParseBool(str)
		if err != nil {
			l.Warnf("parse ENV_INPUT_GPUSMI_ELECTION: %s, ignore", err)
		} else {
			ipt.Election = election
		}
	}
}

func newDefaultInput() *Input {
	ipt := &Input{
		platform:            runtime.GOOS,
		Interval:            defaultInterval,
		Election:            true,
		semStop:             cliutils.NewSem(),
		Tags:                make(map[string]string),
		BinPaths:            []string{"/usr/bin/nvidia-smi"},
		Timeout:             defaultTimeout,
		ProcessInfoMaxLen:   10,
		GPUDropWarningDelay: time.Second * 300,
		gpus:                make([]gpuInfo, 0, 1),
		Envs:                []string{},
		pauseCh:             make(chan bool, inputs.ElectionPauseChannelLength),
		feeder:              dkio.DefaultFeeder(),
		tagger:              datakit.DefaultGlobalTagger(),
		mergedTags:          make(map[string]urlTags),
	}
	return ipt
}

func init() { //nolint:gochecknoinits
	// TODO 这里 2个，好奇怪啊
	inputs.Add(inputName, func() inputs.Input {
		return newDefaultInput()
	})

	// To adapt to the variable name nvidia_smi --> gpu_smi in the .conf file
	inputs.Add("nvidia_smi", func() inputs.Input {
		return newDefaultInput()
	})
}
