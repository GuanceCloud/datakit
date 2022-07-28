// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

// Package process collect host processes metrics/objects
package process

import (
	"encoding/json"
	"fmt"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/shirou/gopsutil/host"
	pr "github.com/shirou/gopsutil/v3/process"
	"github.com/tweekmonster/luser"
	"gitlab.jiagouyun.com/cloudcare-tools/cliutils"
	"gitlab.jiagouyun.com/cloudcare-tools/cliutils/logger"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/config"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"
)

var (
	l                                = logger.DefaultSLogger(inputName)
	minObjectInterval                = time.Second * 30
	maxObjectInterval                = time.Minute * 15
	minMetricInterval                = time.Second * 10
	maxMetricInterval                = time.Minute
	_                 inputs.ReadEnv = (*Input)(nil)
)

type Input struct {
	MatchedProcessNames []string          `toml:"process_name,omitempty"`
	ObjectInterval      datakit.Duration  `toml:"object_interval,omitempty"`
	RunTime             datakit.Duration  `toml:"min_run_time,omitempty"`
	OpenMetric          bool              `toml:"open_metric,omitempty"`
	MetricInterval      datakit.Duration  `toml:"metric_interval,omitempty"`
	Tags                map[string]string `toml:"tags"`

	// pipeline on process object removed
	PipelineDeprecated string `toml:"pipeline,omitempty"`

	lastErr error
	res     []*regexp.Regexp
	isTest  bool

	semStop *cliutils.Sem // start stop signal
}

func (*Input) Catalog() string { return category }

func (*Input) SampleConfig() string { return sampleConfig }

func (*Input) AvailableArchs() []string { return datakit.AllOS }

func (*Input) SampleMeasurement() []inputs.Measurement {
	return []inputs.Measurement{&ProcessMetric{}, &ProcessObject{}}
}

func (p *Input) Run() {
	l = logger.SLogger(inputName)

	l.Info("process start...")
	for _, x := range p.MatchedProcessNames {
		if re, err := regexp.Compile(x); err != nil {
			l.Warnf("regexp.Compile(%s): %s, ignored", x, err)
		} else {
			l.Debugf("add regexp %s", x)
			p.res = append(p.res, re)
		}
	}

	p.ObjectInterval.Duration = config.ProtectedInterval(minObjectInterval,
		maxObjectInterval,
		p.ObjectInterval.Duration)

	tick := time.NewTicker(p.ObjectInterval.Duration)
	defer tick.Stop()
	if p.OpenMetric {
		go func() {
			p.MetricInterval.Duration = config.ProtectedInterval(minMetricInterval,
				maxMetricInterval,
				p.MetricInterval.Duration)

			procRecorder := newProcRecorder()
			tick := time.NewTicker(p.MetricInterval.Duration)
			defer tick.Stop()
			for {
				processList := p.getProcesses(true)
				tn := time.Now().UTC()
				p.WriteMetric(processList, procRecorder, tn)
				procRecorder.flush(processList, tn)
				select {
				case <-tick.C:
				case <-datakit.Exit.Wait():
					l.Info("process write metric exit")
					return

				case <-p.semStop.Wait():
					l.Info("process write metric return")
					return
				}
			}
		}()
	}

	procRecorder := newProcRecorder()
	for {
		processList := p.getProcesses(false)
		tn := time.Now().UTC()
		p.WriteObject(processList, procRecorder, tn)
		procRecorder.flush(processList, tn)
		select {
		case <-tick.C:
		case <-datakit.Exit.Wait():
			l.Info("process write object exit")
			return

		case <-p.semStop.Wait():
			l.Info("process write object return")
			return
		}
	}
}

func (p *Input) Terminate() {
	if p.semStop != nil {
		p.semStop.Close()
	}
}

// ReadEnv support envs：
//   ENV_INPUT_OPEN_METRIC : booler   // deprecated
//   ENV_INPUT_HOST_PROCESSES_OPEN_METRIC : booler
//   ENV_INPUT_HOST_PROCESSES_TAGS : "a=b,c=d"
//   ENV_INPUT_HOST_PROCESSES_PROCESS_NAME : []string
//   ENV_INPUT_HOST_PROCESSES_MIN_RUN_TIME : datakit.Duration
func (p *Input) ReadEnv(envs map[string]string) {
	// deprecated
	if open, ok := envs["ENV_INPUT_OPEN_METRIC"]; ok {
		b, err := strconv.ParseBool(open)
		if err != nil {
			l.Warnf("parse ENV_INPUT_OPEN_METRIC to bool: %s, ignore", err)
		} else {
			p.OpenMetric = b
		}
	}

	if open, ok := envs["ENV_INPUT_HOST_PROCESSES_OPEN_METRIC"]; ok {
		b, err := strconv.ParseBool(open)
		if err != nil {
			l.Warnf("parse ENV_INPUT_HOST_PROCESSES_OPEN_METRIC to bool: %s, ignore", err)
		} else {
			p.OpenMetric = b
		}
	}

	if tagsStr, ok := envs["ENV_INPUT_PROCESSES_TAGS"]; ok {
		tags := config.ParseGlobalTags(tagsStr)
		for k, v := range tags {
			p.Tags[k] = v
		}
	}

	//   ENV_INPUT_HOST_PROCESSES_PROCESS_NAME : []string
	//   ENV_INPUT_HOST_PROCESSES_MIN_RUN_TIME : datakit.Duration
	if str, ok := envs["ENV_INPUT_HOST_PROCESSES_PROCESS_NAME"]; ok {
		arrays := strings.Split(str, ",")
		l.Debugf("add PROCESS_NAME from ENV: %v", arrays)
		p.MatchedProcessNames = append(p.MatchedProcessNames, arrays...)
	}

	if str, ok := envs["ENV_INPUT_HOST_PROCESSES_MIN_RUN_TIME"]; ok {
		da, err := time.ParseDuration(str)
		if err != nil {
			l.Warnf("parse ENV_INPUT_HOST_PROCESSES_MIN_RUN_TIME to time.Duration: %s, ignore", err)
		} else {
			p.RunTime.Duration = config.ProtectedInterval(minObjectInterval,
				maxObjectInterval,
				da)
		}
	}
}

func (p *Input) matched(name string) bool {
	if len(p.res) == 0 {
		return true
	}

	for _, re := range p.res {
		if re.MatchString(name) {
			return true
		}
	}

	return false
}

func (p *Input) getProcesses(match bool) (processList []*pr.Process) {
	pses, err := pr.Processes()
	if err != nil {
		l.Warnf("get process err: %s", err.Error())
		p.lastErr = err
		return
	}

	for _, ps := range pses {
		name, err := ps.Name()
		if err != nil {
			l.Warnf("ps.Name: %s", err)
			continue
		}

		if match {
			if !p.matched(name) {
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
		if time.Since(tm) > p.RunTime.Duration {
			processList = append(processList, ps)
		}
	}
	return processList
}

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
			l.Debugf("process: pid:%d get username err:%s", ps.Pid, err.Error())
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

func (p *Input) Parse(ps *pr.Process, procRec *procRecorder, tn time.Time) (username, state, name string, fields, message map[string]interface{}) {
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
		fields["cpu_usage"] = calculatePercent(ps, tn)
		fields["cpu_usage_top"] = procRec.calculatePercentTop(ps, tn)
	}

	Threads, err := ps.NumThreads()
	if err != nil {
		l.Warnf("process:%s,pid:%d get threads err:%s", name, ps.Pid, err.Error())
	} else {
		fields["threads"] = Threads
	}
	if runtime.GOOS == "linux" {
		OpenFiles, err := ps.OpenFiles()
		if err != nil {
			l.Warnf("process:%s,pid:%d get openfile err:%s", name, ps.Pid, err.Error())
		} else {
			fields["open_files"] = len(OpenFiles)
			message["open_files"] = OpenFiles
		}
	}

	return username, state, name, fields, message
}

func (p *Input) WriteObject(processList []*pr.Process, procRec *procRecorder, tn time.Time) {
	var collectCache []inputs.Measurement

	for _, ps := range processList {
		username, state, name, fields, message := p.Parse(ps, procRec, tn)
		tags := map[string]string{
			"username":     username,
			"state":        state,
			"name":         fmt.Sprintf("%s_%d", config.Cfg.Hostname, ps.Pid),
			"process_name": name,
			"listen_ports": getListeningPorts(ps),
		}
		for k, v := range p.Tags {
			tags[k] = v
		}

		stateZombie := false
		if state == "zombie" {
			stateZombie = true
		}
		fields["state_zombie"] = stateZombie

		fields["pid"] = ps.Pid
		fields["start_time"] = getCreateTime(ps)
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
		if p.isTest {
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
		collectCache = append(collectCache, obj)
	}
	if len(collectCache) == 0 {
		return
	}
	if err := inputs.FeedMeasurement(inputName+"-object",
		datakit.Object,
		collectCache,
		&io.Option{CollectCost: time.Since(tn)}); err != nil {
		l.Errorf("FeedMeasurement err :%s", err.Error())
		p.lastErr = err
	}

	if p.lastErr != nil {
		io.FeedLastError(inputName, p.lastErr.Error())
		p.lastErr = nil
	}
}

func (p *Input) WriteMetric(processList []*pr.Process, procRec *procRecorder, tn time.Time) {
	var collectCache []inputs.Measurement
	for _, ps := range processList {
		cmd, err := ps.Cmdline() // 无cmd的进程 没有采集指标的意义
		if err != nil || cmd == "" {
			continue
		}
		username, _, name, fields, _ := p.Parse(ps, procRec, tn)
		tags := map[string]string{
			"username":     username,
			"pid":          fmt.Sprintf("%d", ps.Pid),
			"process_name": name,
		}
		for k, v := range p.Tags {
			tags[k] = v
		}
		metric := &ProcessMetric{
			name:   inputName,
			tags:   tags,
			fields: fields,
			ts:     tn,
		}
		if len(fields) == 0 {
			continue
		}
		collectCache = append(collectCache, metric)
	}
	if len(collectCache) == 0 {
		return
	}
	if err := inputs.FeedMeasurement(inputName+"-metric",
		datakit.Metric,
		collectCache,
		&io.Option{CollectCost: time.Since(tn)}); err != nil {
		l.Errorf("FeedMeasurement err :%s", err.Error())
		p.lastErr = err
	}
	if p.lastErr != nil {
		io.FeedLastError(inputName, p.lastErr.Error())
		p.lastErr = nil
	}
}

// getListeningPorts returns ports given process is listening
// in format "[aaa,bbb,ccc]" or "[]" when error occurs.
func getListeningPorts(proc *pr.Process) string {
	connections, err := proc.Connections()
	if err != nil {
		l.Warnf("proc.Connections: %s", err)
		return "[]"
	}
	var listening []string
	for _, c := range connections {
		if c.Status == "LISTEN" {
			listening = append(listening, strconv.FormatInt(int64(c.Laddr.Port), 10))
		}
	}
	return "[" + strings.Join(listening, ",") + "]"
}

func init() { //nolint:gochecknoinits
	inputs.Add(inputName, func() inputs.Input {
		return &Input{
			ObjectInterval: datakit.Duration{Duration: 5 * time.Minute},
			MetricInterval: datakit.Duration{Duration: 30 * time.Second},

			semStop: cliutils.NewSem(),
			Tags:    make(map[string]string),
		}
	})
}
