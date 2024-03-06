// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

// Package process collect host processes metrics/objects
package process

import (
	"context"
	"encoding/json"
	"fmt"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/GuanceCloud/cliutils"
	"github.com/GuanceCloud/cliutils/logger"
	"github.com/GuanceCloud/cliutils/point"
	"github.com/shirou/gopsutil/host"
	pr "github.com/shirou/gopsutil/v3/process"
	"github.com/tweekmonster/luser"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/config"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/export/doc"
	dkio "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/io"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs"
)

var (
	l                                  = logger.DefaultSLogger(inputName)
	minObjectInterval                  = time.Second * 30
	maxObjectInterval                  = time.Minute * 15
	minMetricInterval                  = time.Second * 10
	maxMetricInterval                  = time.Minute
	_                 inputs.ReadEnv   = (*Input)(nil)
	_                 inputs.Singleton = (*Input)(nil)
)

type Input struct {
	MatchedProcessNames []string          `toml:"process_name,omitempty"`
	ObjectInterval      datakit.Duration  `toml:"object_interval,omitempty"`
	RunTime             datakit.Duration  `toml:"min_run_time,omitempty"`
	OpenMetric          bool              `toml:"open_metric,omitempty"`
	OpenFiles           bool              `toml:"enable_open_files,omitempty"`
	ListenPorts         bool              `toml:"enable_listen_ports,omitempty"`
	MetricInterval      datakit.Duration  `toml:"metric_interval,omitempty"`
	Tags                map[string]string `toml:"tags"`

	// pipeline on process object removed
	PipelineDeprecated string `toml:"pipeline,omitempty"`

	lastErr error
	res     []*regexp.Regexp
	isTest  bool

	semStop *cliutils.Sem // start stop signal
	feeder  dkio.Feeder
	Tagger  datakit.GlobalTagger
}

func (*Input) Singleton() {}

func (*Input) Catalog() string { return category }

func (*Input) SampleConfig() string { return sampleConfig }

func (*Input) AvailableArchs() []string { return datakit.AllOS }

func (*Input) SampleMeasurement() []inputs.Measurement {
	return []inputs.Measurement{&ProcessMetric{}, &ProcessObject{}}
}

func (ipt *Input) Run() {
	l = logger.SLogger(inputName)

	l.Info("process start...")
	for _, x := range ipt.MatchedProcessNames {
		if re, err := regexp.Compile(x); err != nil {
			l.Warnf("regexp.Compile(%s): %s, ignored", x, err)
		} else {
			l.Debugf("add regexp %s", x)
			ipt.res = append(ipt.res, re)
		}
	}

	ipt.ObjectInterval.Duration = config.ProtectedInterval(minObjectInterval,
		maxObjectInterval,
		ipt.ObjectInterval.Duration)

	tick := time.NewTicker(ipt.ObjectInterval.Duration)
	defer tick.Stop()
	if ipt.OpenMetric {
		g := datakit.G("inputs_process")
		g.Go(func(ctx context.Context) error {
			ipt.MetricInterval.Duration = config.ProtectedInterval(minMetricInterval,
				maxMetricInterval,
				ipt.MetricInterval.Duration)

			procRecorder := newProcRecorder()
			tick := time.NewTicker(ipt.MetricInterval.Duration)
			defer tick.Stop()
			for {
				processList := ipt.getProcesses(true)
				tn := time.Now().UTC()
				ipt.WriteMetric(processList, procRecorder, tn)
				procRecorder.flush(processList, tn)
				select {
				case <-tick.C:
				case <-datakit.Exit.Wait():
					l.Info("process write metric exit")
					return nil

				case <-ipt.semStop.Wait():
					l.Info("process write metric return")
					return nil
				}
			}
		})
	}

	procRecorder := newProcRecorder()
	for {
		processList := ipt.getProcesses(false)
		tn := time.Now().UTC()
		ipt.WriteObject(processList, procRecorder, tn)
		procRecorder.flush(processList, tn)
		select {
		case <-tick.C:
		case <-datakit.Exit.Wait():
			l.Info("process write object exit")
			return

		case <-ipt.semStop.Wait():
			l.Info("process write object return")
			return
		}
	}
}

func (ipt *Input) Terminate() {
	if ipt.semStop != nil {
		ipt.semStop.Close()
	}
}

func (ipt *Input) GetENVDoc() []*inputs.ENVInfo {
	// nolint:lll
	infos := []*inputs.ENVInfo{
		{FieldName: "OpenMetric", Type: doc.Boolean, Default: `false`, Desc: "Enable process metric collecting", DescZh: "采集处理器指标"},
		{FieldName: "MatchedProcessNames", ENVName: "PROCESS_NAME", Type: doc.List, Example: `.*datakit.*,guance`, Desc: "Whitelist of process", DescZh: "处理器白名单"},
		{FieldName: "RunTime", ENVName: "MIN_RUN_TIME", Type: doc.TimeDuration, Default: `10m`, Desc: "Process minimal run time", DescZh: "处理最短运行时间"},
		{FieldName: "ListenPorts", ENVName: "ENABLE_LISTEN_PORTS", Type: doc.Boolean, Default: `false`, Desc: "Enable listen ports tag", DescZh: "启用监听端口标签"},
		{FieldName: "OpenFiles", ENVName: "ENABLE_OPEN_FILES", Type: doc.Boolean, Default: `false`, Desc: "Enable open files field", DescZh: "启用打开文件字段"},
		{FieldName: "Tags"},
	}

	return doc.SetENVDoc("ENV_INPUT_HOST_PROCESSES_", infos)
}

// ReadEnv support envs：
//
//		ENV_INPUT_OPEN_METRIC : booler   // deprecated
//		ENV_INPUT_HOST_PROCESSES_OPEN_METRIC : booler
//		ENV_INPUT_HOST_PROCESSES_TAGS : "a=b,c=d"
//		ENV_INPUT_HOST_PROCESSES_PROCESS_NAME : []string
//		ENV_INPUT_HOST_PROCESSES_MIN_RUN_TIME : datakit.Duration
//	 ENV_INPUT_HOST_PROCESSES_ENABLE_LISTEN_PORTS : booler
//	 ENV_INPUT_HOST_PROCESSES_ENABLE_OPEN_FILES : booler
func (ipt *Input) ReadEnv(envs map[string]string) {
	// deprecated
	if open, ok := envs["ENV_INPUT_OPEN_METRIC"]; ok {
		b, err := strconv.ParseBool(open)
		if err != nil {
			l.Warnf("parse ENV_INPUT_OPEN_METRIC to bool: %s, ignore", err)
		} else {
			ipt.OpenMetric = b
		}
	}

	if open, ok := envs["ENV_INPUT_HOST_PROCESSES_OPEN_METRIC"]; ok {
		b, err := strconv.ParseBool(open)
		if err != nil {
			l.Warnf("parse ENV_INPUT_HOST_PROCESSES_OPEN_METRIC to bool: %s, ignore", err)
		} else {
			ipt.OpenMetric = b
		}
	}

	if tagsStr, ok := envs["ENV_INPUT_PROCESSES_TAGS"]; ok {
		tags := config.ParseGlobalTags(tagsStr)
		for k, v := range tags {
			ipt.Tags[k] = v
		}
	}

	//   ENV_INPUT_HOST_PROCESSES_PROCESS_NAME : []string
	//   ENV_INPUT_HOST_PROCESSES_MIN_RUN_TIME : datakit.Duration
	if str, ok := envs["ENV_INPUT_HOST_PROCESSES_PROCESS_NAME"]; ok {
		arrays := strings.Split(str, ",")
		l.Debugf("add PROCESS_NAME from ENV: %v", arrays)
		ipt.MatchedProcessNames = append(ipt.MatchedProcessNames, arrays...)
	}

	if str, ok := envs["ENV_INPUT_HOST_PROCESSES_MIN_RUN_TIME"]; ok {
		da, err := time.ParseDuration(str)
		if err != nil {
			l.Warnf("parse ENV_INPUT_HOST_PROCESSES_MIN_RUN_TIME to time.Duration: %s, ignore", err)
		} else {
			ipt.RunTime.Duration = config.ProtectedInterval(minObjectInterval,
				maxObjectInterval,
				da)
		}
	}

	if port, ok := envs["ENV_INPUT_HOST_PROCESSES_ENABLE_LISTEN_PORTS"]; ok {
		b, err := strconv.ParseBool(port)
		if err != nil {
			l.Warnf("parse ENV_INPUT_HOST_PROCESSES_ENABLE_LISTEN_PORTS to bool: %s, ignore", err)
		} else {
			ipt.ListenPorts = b
		}
	}

	if file, ok := envs["ENV_INPUT_HOST_PROCESSES_ENABLE_OPEN_FILES"]; ok {
		b, err := strconv.ParseBool(file)
		if err != nil {
			l.Warnf("parse ENV_INPUT_HOST_PROCESSES_ENABLE_OPEN_FILES to bool: %s, ignore", err)
		} else {
			ipt.OpenFiles = b
		}
	}
}

func (ipt *Input) matched(name string) bool {
	if len(ipt.res) == 0 {
		return true
	}

	for _, re := range ipt.res {
		if re.MatchString(name) {
			return true
		}
	}

	return false
}

func (ipt *Input) getProcesses(match bool) (processList []*pr.Process) {
	pses, err := pr.Processes()
	if err != nil {
		l.Warnf("get process err: %s", err.Error())
		ipt.lastErr = err
		return
	}

	for _, ps := range pses {
		name, err := ps.Name()
		if err != nil {
			l.Warnf("ps.Name: %s", err)
			continue
		}

		if match {
			if !ipt.matched(name) {
				l.Warnf("%s not matched", name)
				continue
			}

			l.Debugf("%s match ok", name)
		}

		t, err := ps.CreateTime()
		if err != nil {
			l.Warnf("ps.CreateTime: %s", err)
			continue
		}

		tm := time.Unix(0, t*1000000) // 转纳秒
		if time.Since(tm) > ipt.RunTime.Duration {
			processList = append(processList, ps)
		}
	}
	return processList
}

const ignoreError = "user: unknown userid"

func getUser(ps *pr.Process) string {
	username, err := ps.Username()
	if err != nil {
		uid, err := ps.Uids()
		if err != nil {
			l.Warnf("process get uid err:%s", err.Error())
			return ""
		}
		u, err := luser.LookupId(fmt.Sprintf("%d", uid[0])) //nolint:stylecheck
		if err != nil {
			// 此处错误极多，故将其 disable 掉，一般的报错是：unknown userid xxx
			if !strings.Contains(err.Error(), ignoreError) {
				l.Debugf("process: pid:%d get username err:%s", ps.Pid, err.Error())
			}
			return ""
		}
		return u.Username
	}
	return username
}

func getCreateTime(ps *pr.Process) int64 {
	start, err := ps.CreateTime()
	if err != nil {
		l.Warnf("get start time err:%s", err.Error())
		if bootTime, err := host.BootTime(); err != nil {
			return int64(bootTime) * 1000
		}
	}
	return start
}

func (ipt *Input) Parse(ps *pr.Process, procRec *procRecorder, tn time.Time) (username, state, name string, fields, message map[string]interface{}) {
	fields = map[string]interface{}{}
	message = map[string]interface{}{}
	name, err := ps.Name()
	if err != nil {
		l.Warnf("process get name err:%s", err.Error())
	}
	username = getUser(ps)
	if username == "" {
		username = "nobody"
	}
	status, err := ps.Status()
	if err != nil {
		l.Warnf("process:%s,pid:%d get state err:%s", name, ps.Pid, err.Error())
		state = ""
	} else {
		state = status[0]
	}

	// you may get a null pointer here
	memInfo, err := ps.MemoryInfo()
	if err != nil {
		l.Warnf("process:%s,pid:%d get memoryinfo err:%s", name, ps.Pid, err.Error())
	} else {
		message["memory"] = memInfo
		fields["rss"] = memInfo.RSS
	}

	memPercent, err := ps.MemoryPercent()
	if err != nil {
		l.Warnf("process:%s,pid:%d get mempercent err:%s", name, ps.Pid, err.Error())
	} else {
		fields["mem_used_percent"] = memPercent
	}

	// you may get a null pointer here
	cpuTime, err := ps.Times()
	if err != nil {
		l.Warnf("process:%s,pid:%d get cpu err:%s", name, ps.Pid, err.Error())
		l.Warnf("process:%s,pid:%d get cpupercent err:%s", name, ps.Pid, err.Error())
	} else {
		message["cpu"] = cpuTime

		cpuUsage := calculatePercent(ps, tn)
		cpuUsageTop := procRec.calculatePercentTop(ps, tn)

		if runtime.GOOS == "windows" {
			cpuUsage /= float64(runtime.NumCPU())
			cpuUsageTop /= float64(runtime.NumCPU())
		}

		fields["cpu_usage"] = cpuUsage
		fields["cpu_usage_top"] = cpuUsageTop
	}

	Threads, err := ps.NumThreads()
	if err != nil {
		l.Warnf("process:%s,pid:%d get threads err:%s", name, ps.Pid, err.Error())
	} else {
		fields["threads"] = Threads
	}

	if runtime.GOOS == "linux" && ipt.OpenFiles {
		openFiles, err := ps.OpenFiles()
		if err != nil {
			l.Warnf("process:%s,pid:%d get openfile err:%s", name, ps.Pid, err.Error())
		} else {
			fields["open_files"] = len(openFiles)
		}
	}

	return username, state, name, fields, message
}

func (ipt *Input) WriteObject(processList []*pr.Process, procRec *procRecorder, tn time.Time) {
	var collectCache []*point.Point

	for _, ps := range processList {
		username, state, name, fields, message := ipt.Parse(ps, procRec, tn)
		tags := map[string]string{
			"username":     username,
			"state":        state,
			"name":         fmt.Sprintf("%s_%d", config.Cfg.Hostname, ps.Pid),
			"process_name": name,
		}
		if ipt.ListenPorts {
			if listeningPorts, err := getListeningPortsJSON(ps); err != nil {
				l.Warnf("getListeningPortsJSON: %v", err)
			} else {
				tags["listen_ports"] = string(listeningPorts)
			}
		}

		for k, v := range ipt.Tags {
			tags[k] = v
		}

		tags = inputs.MergeTags(ipt.Tagger.HostTags(), tags, "")

		stateZombie := false
		if state == "zombie" {
			stateZombie = true
		}
		fields["state_zombie"] = stateZombie

		fields["pid"] = ps.Pid

		ct := getCreateTime(ps)
		fields["started_duration"] = int64(time.Since(time.Unix(0,
			ct*int64(time.Millisecond))) / time.Second)
		fields["start_time"] = ct

		if runtime.GOOS == "linux" {
			dir, err := ps.Cwd()
			if err != nil {
				l.Warnf("process:%s,pid:%d get work_directory err:%s", name, ps.Pid, err.Error())
			} else {
				fields["work_directory"] = dir
			}
		}

		cmd, err := ps.Cmdline()
		if err != nil {
			l.Warnf("process:%s,pid:%d get cmd err:%s", name, ps.Pid, err.Error())
			cmd = ""
		}

		if cmd == "" {
			cmd = fmt.Sprintf("(%s)", name)
		}

		fields["cmdline"] = cmd

		if ipt.isTest {
			return
		}

		// 此处为了全文检索 需要冗余一份数据 将tag field字段全部塞入 message
		for k, v := range tags {
			message[k] = v
		}

		for k, v := range fields {
			message[k] = v
		}

		m, err := json.Marshal(message)
		if err == nil {
			fields["message"] = string(m)
		} else {
			l.Errorf("marshal message err:%s", err.Error())
		}

		if len(fields) == 0 {
			continue
		}
		obj := &ProcessObject{
			name:   inputName,
			tags:   tags,
			fields: fields,
			ts:     tn,
		}
		collectCache = append(collectCache, obj.Point())
	}
	if len(collectCache) == 0 {
		return
	}

	if err := ipt.feeder.Feed(inputName+"/object",
		point.Object,
		collectCache,
		&dkio.Option{CollectCost: time.Since(tn)}); err != nil {
		l.Errorf("FeedMeasurement err :%s", err.Error())
		ipt.lastErr = err
	}

	if ipt.lastErr != nil {
		ipt.feeder.FeedLastError(ipt.lastErr.Error(),
			dkio.WithLastErrorInput(inputName),
			dkio.WithLastErrorCategory(point.Object),
		)
		ipt.lastErr = nil
	}
}

func (ipt *Input) WriteMetric(processList []*pr.Process, procRec *procRecorder, tn time.Time) {
	var collectCache []*point.Point

	for _, ps := range processList {
		cmd, err := ps.Cmdline() // 无cmd的进程 没有采集指标的意义
		if err != nil || cmd == "" {
			continue
		}
		username, _, name, fields, _ := ipt.Parse(ps, procRec, tn)
		tags := map[string]string{
			"username":     username,
			"pid":          fmt.Sprintf("%d", ps.Pid),
			"process_name": name,
		}
		for k, v := range ipt.Tags {
			tags[k] = v
		}
		tags = inputs.MergeTags(ipt.Tagger.HostTags(), tags, "")
		metric := &ProcessMetric{
			name:   inputName,
			tags:   tags,
			fields: fields,
			ts:     tn,
		}
		if len(fields) == 0 {
			continue
		}
		collectCache = append(collectCache, metric.Point())
	}
	if len(collectCache) == 0 {
		return
	}

	if err := ipt.feeder.Feed(inputName+"/metric",
		point.Metric,
		collectCache,
		&dkio.Option{CollectCost: time.Since(tn)}); err != nil {
		l.Errorf("fedder.Feed :%s", err.Error())
		ipt.lastErr = err
	}

	if ipt.lastErr != nil {
		ipt.feeder.FeedLastError(ipt.lastErr.Error(),
			dkio.WithLastErrorInput(inputName),
			dkio.WithLastErrorCategory(point.Metric),
		)
		ipt.lastErr = nil
	}
}

func getListeningPortsJSON(proc *pr.Process) ([]byte, error) {
	connections, err := proc.Connections()
	if err != nil {
		return nil, err
	}
	var listening []uint32
	for _, c := range connections {
		if c.Status == "LISTEN" {
			listening = append(listening, c.Laddr.Port)
		}
	}
	return json.Marshal(listening)
}

func defaultInput() *Input {
	return &Input{
		ObjectInterval: datakit.Duration{Duration: 5 * time.Minute},
		MetricInterval: datakit.Duration{Duration: 30 * time.Second},

		semStop: cliutils.NewSem(),
		Tags:    make(map[string]string),
		feeder:  dkio.DefaultFeeder(),
		Tagger:  datakit.DefaultGlobalTagger(),
	}
}

func init() { //nolint:gochecknoinits
	inputs.Add(inputName, func() inputs.Input {
		return defaultInput()
	})
}
