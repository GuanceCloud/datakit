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
	dkio "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/io"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/metrics"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/ntp"
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
	MatchedProcessNames []string         `toml:"process_name,omitempty"`
	ObjectInterval      datakit.Duration `toml:"object_interval,omitempty"`
	RunTime             datakit.Duration `toml:"min_run_time,omitempty"`

	OpenMetric  bool `toml:"open_metric,omitempty"`
	ListenPorts bool `toml:"enable_listen_ports,omitempty"`

	MetricInterval datakit.Duration  `toml:"metric_interval,omitempty"`
	Tags           map[string]string `toml:"tags"`

	// pipeline on process object removed
	PipelineDeprecated string `toml:"pipeline,omitempty"`

	OnlyContainerProcesses bool `toml:"only_container_processes"`

	lastErr error
	res     []*regexp.Regexp
	isTest  bool

	semStop *cliutils.Sem // start stop signal
	feeder  dkio.Feeder
	Tagger  datakit.GlobalTagger

	metricTime time.Time
	mrec, orec *procRecorder
}

func (*Input) Singleton() {}

func (*Input) Catalog() string { return category }

func (*Input) SampleConfig() string { return sampleConfig }

func (*Input) AvailableArchs() []string { return datakit.AllOS }

func (*Input) SampleMeasurement() []inputs.Measurement {
	return []inputs.Measurement{&processMetric{}, &processObject{}}
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

	ipt.ObjectInterval.Duration = config.ProtectedInterval(
		minObjectInterval,
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

			ipt.mrec = newProcRecorder()
			tick := time.NewTicker(ipt.MetricInterval.Duration)
			defer tick.Stop()

			ipt.metricTime = ntp.Now()
			for {
				collectStart := time.Now()
				processList := ipt.getProcesses(true)
				ipt.collectMetric(processList, collectStart)
				ipt.mrec.flush(processList, ipt.metricTime.UTC())

				select {
				case tt := <-tick.C:
					ipt.metricTime = inputs.AlignTime(tt, ipt.metricTime, ipt.MetricInterval.Duration)
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

	ipt.orec = newProcRecorder()
	for {
		processList := ipt.getProcesses(false)
		tn := time.Now().UTC()
		ipt.collectObject(processList, tn)
		ipt.orec.flush(processList, tn)

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
				l.Debugf("%s not matched", name)
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

func getUser(proc *pr.Process) (string, error) {
	username, err := proc.Username()
	if err != nil {
		uid, err := proc.Uids()
		if err != nil {
			l.Warnf("process %v, proc.Uids(): %s", proc, err.Error())
			return "", fmt.Errorf("proc.Uids(): %w", err)
		}

		u, err := luser.LookupId(fmt.Sprintf("%d", uid[0])) //nolint:stylecheck
		if err != nil {
			// 此处错误极多，故将其 disable 掉，一般的报错是：unknown userid xxx
			if !strings.Contains(err.Error(), ignoreError) {
				l.Debugf("process %v, LookupId(): %s", proc, err.Error())
			}
			return "", fmt.Errorf("luser.LookupId(): %w", err)
		}
		return u.Username, nil
	}

	return username, nil
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

func (ipt *Input) Parse(proc *pr.Process, procRec *procRecorder, tn time.Time) point.KVs {
	var kvs point.KVs

	if containerID := getContainerID(proc); containerID != "" {
		kvs = kvs.AddTag("container_id", containerID)
	} else if ipt.OnlyContainerProcesses {
		l.Debugf("ignore non-container process %v", proc)
		return nil
	}

	if x, err := proc.Name(); err != nil {
		l.Warnf("process: %v, proc.Name(): %s", proc, err.Error())
	} else {
		kvs = kvs.AddTag("process_name", x)
	}

	if username, err := getUser(proc); err != nil {
		kvs = kvs.AddTag("username", username)
	} else {
		kvs = kvs.AddTag("username", "nobody")
	}

	// you may get a null pointer here
	if x, err := proc.MemoryInfo(); err != nil {
		l.Warnf("process: %v, proc.MemoryInfo(): %s", proc, err.Error())
	} else {
		kvs = kvs.Add("rss", x.RSS)
		kvs = kvs.Add("vms", x.VMS)
	}

	if x, err := proc.MemoryPercent(); err != nil {
		l.Warnf("process: %v, proc.MemoryPercent(): %s", proc, err.Error())
	} else {
		kvs = kvs.Add("mem_used_percent", x)
	}

	// you may get a null pointer here
	if _, err := proc.Times(); err != nil {
		l.Warnf("process: %v, proc.Times(): %s", proc, err.Error())
	} else {
		cpuUsage := calculatePercent(proc, tn)
		cpuUsageTop := procRec.calculatePercentTop(proc, tn)

		if runtime.GOOS == "windows" {
			cpuUsage /= float64(runtime.NumCPU())
			cpuUsageTop /= float64(runtime.NumCPU())
		}
		kvs = kvs.Add("cpu_usage", cpuUsage)
		kvs = kvs.Add("cpu_usage_top", cpuUsageTop)
	}

	if x, err := proc.NumThreads(); err != nil {
		l.Warnf("process: %v, proc.NumThreads(): %s", proc, err.Error())
	} else {
		kvs = kvs.Add("threads", x)
	}

	if x, err := proc.IOCounters(); err != nil {
		l.Warnf("process: %v, proc.IOCounters(): %s", proc, err.Error())
	} else {
		kvs = kvs.Add("proc_syscr", x.ReadCount)
		kvs = kvs.Add("proc_syscw", x.WriteCount)
		kvs = kvs.Add("proc_read_bytes", x.ReadBytes)
		kvs = kvs.Add("proc_write_bytes", x.WriteBytes)
	}

	if x, err := proc.NumCtxSwitches(); err != nil {
		l.Warnf("process: %v, proc.NumCtxSwitches(): %s", proc, err.Error())
	} else {
		kvs = kvs.Add("voluntary_ctxt_switches", x.Voluntary)
		kvs = kvs.Add("nonvoluntary_ctxt_switches", x.Involuntary)
	}

	if x, err := proc.PageFaults(); err != nil {
		l.Warnf("process: %v, proc.PageFaults(): %s", proc, err.Error())
	} else {
		kvs = kvs.Add("page_minor_faults", x.MinorFaults)
		kvs = kvs.Add("page_major_faults", x.MajorFaults)
		kvs = kvs.Add("page_children_minor_faults", x.ChildMinorFaults)
		kvs = kvs.Add("page_children_major_faults", x.ChildMajorFaults)
	}

	if runtime.GOOS == "linux" {
		if x, err := proc.NumFDs(); err != nil {
			l.Warnf("process: %v, proc.NumFDs(): %s", proc, err.Error())
		} else {
			kvs = kvs.Add("open_files", x)
		}
	}

	return kvs
}

func (ipt *Input) collectObject(processList []*pr.Process, tn time.Time) {
	var (
		collectCache []*point.Point
		opts         = append(point.DefaultObjectOptions(), point.WithTime(ntp.Now()))
	)

	for _, proc := range processList {
		kvs := ipt.Parse(proc, ipt.orec, tn)
		if kvs == nil {
			continue
		}

		// append object's tag `name': we need a `name' tag for each object.
		kvs = kvs.SetTag("name", fmt.Sprintf("%s_%d", datakit.DKHost, proc.Pid))

		if ipt.ListenPorts {
			if listeningPorts, err := getListeningPortsJSON(proc); err != nil {
				l.Warnf("getListeningPortsJSON(): %s", err)
			} else {
				kvs = kvs.Add("listen_ports", string(listeningPorts))
			}
		}

		for k, v := range ipt.Tags {
			kvs = kvs.AddTag(k, v)
		}

		for k, v := range ipt.Tagger.HostTags() {
			kvs = kvs.AddTag(k, v)
		}

		// Add extra tag/field to process object
		if x, err := proc.Status(); err != nil {
			l.Warnf("process: %v, proc.Status(): %s", proc, err.Error())
		} else {
			kvs = kvs.AddTag("state", x[0])
			if x[0] == pr.Zombie {
				kvs = kvs.Add("state_zombie", true)
			} else {
				kvs = kvs.Add("state_zombie", false)
			}
		}

		kvs = kvs.Add("pid", proc.Pid) // XXX: set pid as int in object

		ct := getCreateTime(proc)
		kvs = kvs.Add("started_duration", int64(time.Since(time.Unix(0, ct*int64(time.Millisecond)))/time.Second))
		kvs = kvs.Add("start_time", ct)

		if runtime.GOOS == "linux" {
			if dir, err := proc.Cwd(); err != nil {
				l.Warnf("process: %v, get work_directory err:%s", proc, err.Error())
			} else {
				kvs = kvs.Add("work_directory", dir)
			}
		}

		if cmd, err := proc.Cmdline(); err != nil {
			l.Warnf("process: %v, proc.Cmdline(): %s", proc, err.Error())

			// use proc-name as cmdline
			kvs = kvs.Add("cmdline", fmt.Sprintf("(%s)", kvs.GetTag("process_name")))
		} else {
			kvs = kvs.Add("cmdline", cmd)
		}

		if ipt.isTest {
			return
		}

		message := map[string]any{}

		// 此处为了全文检索 需要冗余一份数据将 kvs 字段全部塞入 message
		for _, kv := range kvs {
			message[kv.Key] = kv.Raw()
		}

		// get full info of mem info
		if x, err := proc.MemoryInfo(); err != nil {
			l.Warnf("process: %v, proc.MemoryInfo(): %s", proc, err.Error())
		} else {
			message["memory"] = x
		}

		// cpu-time metrics has collected in top kvs. here we get a duplicated for compatibility.
		if x, err := proc.Times(); err != nil {
			l.Warnf("process: %v, proc.Times(): %s", proc, err.Error())
		} else {
			message["cpu"] = x
		}

		if msg, err := json.Marshal(message); err != nil {
			l.Warnf("marshal object message failed: %s", err.Error())
		} else {
			kvs = kvs.Add("message", string(msg))
		}

		collectCache = append(collectCache, point.NewPoint(inputName, kvs, opts...))
	}

	if len(collectCache) == 0 {
		return
	}

	if err := ipt.feeder.Feed(point.Object, collectCache,
		dkio.WithCollectCost(time.Since(tn)),
		dkio.WithSource(objectFeedName),
	); err != nil {
		l.Errorf("Feed object err :%s", err.Error())
		ipt.feeder.FeedLastError(ipt.lastErr.Error(),
			metrics.WithLastErrorInput(inputName),
			metrics.WithLastErrorCategory(point.Object),
		)
	}
}

func (ipt *Input) collectMetric(processList []*pr.Process, tn time.Time) {
	var collectCache []*point.Point

	opts := point.DefaultMetricOptions()

	for _, proc := range processList {
		cmdline, err := proc.Cmdline() // 无cmd的进程 没有采集指标的意义
		if err != nil || cmdline == "" {
			l.Warnf("Cmdline(): %s, err: %v, ignored", proc.String(), err)
			continue
		}

		kvs := ipt.Parse(proc, ipt.mrec, tn)
		if kvs.FieldCount() == 0 {
			l.Warnf("no field on process %v, ignored", proc)
			continue
		}

		// XXX: set pid/cmdline as tag in metric
		kvs = kvs.AddTag("pid", fmt.Sprintf("%d", proc.Pid)).AddTag("cmdline", cmdline)

		for k, v := range ipt.Tags {
			kvs = kvs.AddTag(k, v)
		}

		for k, v := range ipt.Tagger.HostTags() {
			kvs = kvs.AddTag(k, v)
		}

		collectCache = append(collectCache,
			point.NewPoint(inputName, kvs, append(opts, point.WithTime(ipt.metricTime))...))
	}

	if len(collectCache) == 0 {
		return
	}

	if err := ipt.feeder.Feed(point.Metric, collectCache,
		dkio.WithCollectCost(time.Since(tn)),
		dkio.WithSource(dkio.FeedSource(inputName, "metric")),
	); err != nil {
		l.Errorf("Feed() :%s", err.Error())
		ipt.feeder.FeedLastError(err.Error(),
			metrics.WithLastErrorInput(inputName),
			metrics.WithLastErrorCategory(point.Metric),
		)
	}
}

func (ipt *Input) Terminate() {
	if ipt.semStop != nil {
		ipt.semStop.Close()
	}
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
