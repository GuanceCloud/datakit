package host_process

import (
	"encoding/json"
	"fmt"
	"regexp"
	"runtime"
	"strings"
	"time"

	"github.com/shirou/gopsutil/host"
	pr "github.com/shirou/gopsutil/v3/process"
	"github.com/tweekmonster/luser"

	"gitlab.jiagouyun.com/cloudcare-tools/cliutils/logger"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/config"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/pipeline"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"
)

var (
	l                 = logger.DefaultSLogger(inputName)
	minObjectInterval = time.Minute * 5
	maxObjectInterval = time.Minute * 15
	minMetricInterval = time.Second * 30
	maxMetricInterval = time.Minute
)

func (_ *Input) Catalog() string {
	return category
}

func (_ *Input) SampleConfig() string {
	return sampleConfig
}

func (p *Input) PipelineConfig() map[string]string {
	return map[string]string{
		inputName: pipelineSample,
	}
}

func (_ *Input) AvailableArchs() []string {
	return datakit.AllArch
}

func (_ *Input) SampleMeasurement() []inputs.Measurement {
	return []inputs.Measurement{
		&ProcessMetric{},
		&ProcessObject{},
	}
}

func (p *Input) Run() {
	l = logger.SLogger(inputName)

	l.Info("process start...")

	if p.ProcessName != nil {
		re := strings.Join(p.ProcessName, "|")
		if regexp.MustCompile(re) == nil {
			l.Error("[error] regexp err")
			return
		}
		p.re = re
	}
	p.ObjectInterval.Duration = config.ProtectedInterval(minObjectInterval, maxObjectInterval, p.ObjectInterval.Duration)

	if p.RunTime.Duration < 10*time.Minute {
		p.RunTime.Duration = 10 * time.Minute
	}
	tick := time.NewTicker(p.ObjectInterval.Duration)
	defer tick.Stop()
	if p.OpenMetric {
		go func() {
			p.MetricInterval.Duration = config.ProtectedInterval(minMetricInterval, maxMetricInterval, p.MetricInterval.Duration)

			tick := time.NewTicker(p.MetricInterval.Duration)
			defer tick.Stop()
			for {
				p.WriteMetric()
				select {
				case <-tick.C:
				case <-datakit.Exit.Wait():
					l.Info("process write metric exit")
					return
				}
			}

		}()
	}

	for {
		p.WriteObject()
		select {
		case <-tick.C:
		case <-datakit.Exit.Wait():
			l.Info("process write object exit")
			return
		}
	}

}

func (p *Input) getProcesses() (processList []*pr.Process) {
	pses, err := pr.Processes()
	if err != nil {
		l.Errorf("[error] get process err:%s", err.Error())
		p.lastErr = err
		return
	}

	for _, ps := range pses {
		name, _ := ps.Name()
		if ok, _ := regexp.Match(p.re, []byte(name)); !ok {
			continue
		}
		t, _ := ps.CreateTime()
		tm := time.Unix(0, t*1000000) // 转纳秒
		if time.Now().Sub(tm) > p.RunTime.Duration {
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
			l.Warnf("[warning] process get uid err:%s", err.Error())
			return ""
		}
		u, err := luser.LookupId(fmt.Sprintf("%d", uid[0]))
		if err != nil {
			l.Warnf("[warning] process: pid:%d get username err:%s", ps.Pid, err.Error())
			return ""
		}
		return u.Username
	}
	return username
}

func getStartTime(ps *pr.Process) int64 {
	start, err := ps.CreateTime()
	if err != nil {
		l.Warnf("get start time err:%s", err.Error())
		if bootTime, err := host.BootTime(); err != nil {
			return int64(bootTime)
		}
	}
	return start
}

func (p *Input) Parse(ps *pr.Process) (username, state, name string, fields, message map[string]interface{}) {
	fields = map[string]interface{}{}
	message = map[string]interface{}{}
	name, err := ps.Name()
	if err != nil {
		l.Warnf("[warning] process get name err:%s", err.Error())
	}
	username = getUser(ps)
	if username == "" {
		username = "nobody"
	}
	status, err := ps.Status()
	if err != nil {
		l.Warnf("[warning] process:%s,pid:%d get state err:%s", name, ps.Pid, err.Error())
		state = ""
	} else {
		state = status[0]
	}

	mem, err := ps.MemoryInfo()
	if err != nil {
		l.Warnf("[warning] process:%s,pid:%d get memoryinfo err:%s", name, ps.Pid, err.Error())
	} else {
		message["memory"] = mem
		fields["rss"] = mem.RSS
	}
	memPercent, err := ps.MemoryPercent()
	if err != nil {
		l.Warnf("[warning] process:%s,pid:%d get mempercent err:%s", name, ps.Pid, err.Error())
	} else {
		fields["mem_used_percent"] = memPercent
	}
	cpu, err := ps.Times()
	if err != nil {
		l.Warnf("[warning] process:%s,pid:%d get cpu err:%s", name, ps.Pid, err.Error())
	} else {
		message["cpu"] = cpu
	}
	cpuPercent, _ := ps.CPUPercent()
	if err != nil {
		l.Warnf("[warning] process:%s,pid:%d get cpupercent err:%s", name, ps.Pid, err.Error())
	} else {
		fields["cpu_usage"] = cpuPercent
	}
	Threads, err := ps.NumThreads()
	if err != nil {
		l.Warnf("[warning] process:%s,pid:%d get threads err:%s", name, ps.Pid, err.Error())
	} else {
		fields["threads"] = Threads
	}
	if runtime.GOOS == "linux" {
		OpenFiles, err := ps.OpenFiles()
		if err != nil {
			l.Warnf("[warning] process:%s,pid:%d get openfile err:%s", name, ps.Pid, err.Error())
		} else {
			fields["open_files"] = len(OpenFiles)
			message["open_files"] = OpenFiles
		}
	}

	return username, state, name, fields, message

}

func (p *Input) WriteObject() {
	t := time.Now().UTC()
	var collectCache []inputs.Measurement

	for _, ps := range p.getProcesses() {
		username, state, name, fields, message := p.Parse(ps)
		tags := map[string]string{
			"username":     username,
			"state":        state,
			"name":         fmt.Sprintf("%s_%d", config.Cfg.Hostname, ps.Pid),
			"process_name": name,
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
		fields["start_time"] = getStartTime(ps)
		if runtime.GOOS == "linux" {
			dir, err := ps.Cwd()
			if err != nil {
				l.Warnf("[warning] process:%s,pid:%d get work_directory err:%s", name, ps.Pid, err.Error())
			} else {
				fields["work_directory"] = dir
			}
		}
		cmd, err := ps.Cmdline()
		if err != nil {
			l.Warnf("[warning] process:%s,pid:%d get cmd err:%s", name, ps.Pid, err.Error())
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

		if p.Pipeline != "" {
			pipe, err := pipeline.NewPipelineByScriptPath(p.Pipeline)
			if err == nil {
				pipeMap, err := pipe.Run(string(m)).Result()
				if err == nil {
					for k, v := range pipeMap {
						fields[k] = v
					}
				} else {
					l.Errorf("[error] process run pipeline err:%s", err.Error())
				}

			} else {
				l.Errorf("[error] process new pipeline err:%s", err.Error())
			}
		}
		obj := &ProcessObject{
			name:   inputName,
			tags:   tags,
			fields: fields,
			ts:     t,
		}
		collectCache = append(collectCache, obj)
	}
	if err := inputs.FeedMeasurement(inputName, datakit.Object, collectCache, &io.Option{CollectCost: time.Since(t)}); err != nil {
		l.Errorf("FeedMeasurement err :%s", err.Error())
		p.lastErr = err
	}
	if p.lastErr != nil {
		io.FeedLastError(inputName, p.lastErr.Error())
		p.lastErr = nil
	}

}

func (p *Input) WriteMetric() {
	t := time.Now().UTC()
	var collectCache []inputs.Measurement
	for _, ps := range p.getProcesses() {
		cmd, err := ps.Cmdline() // 无cmd的进程 没有采集指标的意义
		if err != nil || cmd == "" {
			continue
		}
		username, _, name, fields, _ := p.Parse(ps)
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
			ts:     t,
		}
		collectCache = append(collectCache, metric)
	}
	if err := inputs.FeedMeasurement(inputName, datakit.Metric, collectCache, &io.Option{CollectCost: time.Since(t)}); err != nil {
		l.Errorf("FeedMeasurement err :%s", err.Error())
		p.lastErr = err
	}
	if p.lastErr != nil {
		io.FeedLastError(inputName, p.lastErr.Error())
		p.lastErr = nil
	}
}

func init() {
	inputs.Add(inputName, func() inputs.Input {
		return &Input{
			ObjectInterval: datakit.Duration{Duration: 5 * time.Minute},
			MetricInterval: datakit.Duration{Duration: 30 * time.Second},
		}
	})
}
