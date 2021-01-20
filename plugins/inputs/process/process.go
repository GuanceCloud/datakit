package process

import (
	"encoding/json"
	"fmt"
	pr "github.com/shirou/gopsutil/v3/process"
	"gitlab.jiagouyun.com/cloudcare-tools/cliutils/logger"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/pipeline"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"
	"regexp"
	"strings"
	"time"
)

var l = logger.DefaultSLogger(inputName)

func (_ *Processes) Catalog() string {
	return inputName
}

func (_ *Processes) SampleConfig() string {
	return sampleConfig
}

func (p *Processes) PipelineConfig() map[string]string {
	return map[string]string{
		inputName: pipelineSample,
	}
}

func (p *Processes) Test() (*inputs.TestResult, error) {
	result := &inputs.TestResult{}
	return result, nil
}

func (p *Processes) Run() {
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
	if p.Interval.Duration == 0 {
		p.Interval.Duration = 5 * time.Minute
	}
	if p.RunTime.Duration == 0 || p.RunTime.Duration < 10*time.Minute {
		p.RunTime.Duration = 10 * time.Minute
	}
	tick := time.NewTicker(p.Interval.Duration)
	defer tick.Stop()

	for {
		p.run()
		select {
		case <-tick.C:
		case <-datakit.Exit.Wait():
			l.Info("process exit")
			return
		}
	}

}

func (p *Processes) run() {
	pses, err := pr.Processes()
	if err != nil {
		l.Warnf("[error] get process err:%s", err.Error())
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
			p.Parse(ps)
		}
	}
}

func (p *Processes) Parse(ps *pr.Process) {
	username, err := ps.Username()
	if err == nil {
		p.username = username
	} else {
		l.Warnf("[warning] process get username err:%s", err.Error())
	}
	state, err := ps.Status()
	if err == nil {
		p.state = state[0]
	} else {
		l.Warnf("[warning] process get state err:%s", err.Error())
	}
	message, fields := parseField(ps)
	if p.OpenMetric {
		p.WriteMetric(ps, fields)
	}
	p.WriteObject(ps, message, fields)

}

func (p *Processes) WriteObject(ps *pr.Process, message, fields map[string]interface{}) {
	name, _ := ps.Name()
	t, _ := ps.CreateTime()
	tags := map[string]string{
		"username":     p.username,
		"state":        p.state,
		"name":         fmt.Sprintf("%s_%d_%d", name, ps.Pid, t),
		"process_name": name,
	}
	m, _ := json.Marshal(message)
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

	fields["message"] = string(m)
	fields["pid"] = ps.Pid
	fields["start_time"] = t

	dir, err := ps.Cwd()
	if err != nil {
		l.Warnf("[warning] process get work_directory err:%s", err.Error())
	} else {
		fields["work_directory"] = dir
	}
	cmd, err := ps.Cmdline()
	if err != nil {
		l.Warnf("[warning] process get cmd err:%s", err.Error())
	} else {
		fields["cmdline"] = cmd
	}

	io.NamedFeedEx(inputName, io.Object, "process", tags, fields, time.Now().UTC())

}

func parseField(ps *pr.Process) (message, fields map[string]interface{}) {
	message = map[string]interface{}{}
	fields = map[string]interface{}{}

	mem, err := ps.MemoryInfo()
	if err != nil {
		l.Warnf("[warning] process get memory err:%s", err.Error())
	} else {
		message["memory"] = mem
		fields["rss"] = mem.RSS
	}

	memPercent, err := ps.MemoryPercent()
	if err != nil {
		l.Warnf("[warning] process get mempercent err:%s", err.Error())
	} else {
		fields["memory_percent"] = memPercent
	}

	cpu, err := ps.Times()
	if err != nil {
		l.Warnf("[warning] process get cpu err:%s", err.Error())
	} else {
		message["cpu"] = cpu
	}

	cpuPercent, _ := ps.CPUPercent()
	if err != nil {
		l.Warnf("[warning] process get cpu err:%s", err.Error())
	} else {
		fields["cpu_percent"] = cpuPercent
	}

	Threads, err := ps.NumThreads()
	if err != nil {
		l.Warnf("[warning] process get threads err:%s", err.Error())
	} else {
		fields["threads"] = Threads
	}

	OpenFiles, err := ps.OpenFiles()
	if err != nil {
		l.Warnf("[warning] process get openfiles err:%s", err.Error())
	} else {
		fields["open_files"] = len(OpenFiles)
		message["open_files"] = OpenFiles

	}
	return
}

func (p *Processes) WriteMetric(ps *pr.Process, fields map[string]interface{}) {
	name, _ := ps.Name()
	tags := map[string]string{
		"username": p.username,
		"state":    p.state,
		"pid":      fmt.Sprintf("%d", ps.Pid),
		"process_name":     name,
	}
	io.NamedFeedEx(inputName, io.Metric, "process", tags, fields, time.Now().UTC())

}

func init() {
	inputs.Add(inputName, func() inputs.Input {
		return &Processes{}
	})
}
