package host_process

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
	return category
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
	if p.ObjectInterval.Duration == 0 {
		p.ObjectInterval.Duration = 5 * time.Minute
	}
	if p.RunTime.Duration == 0 || p.RunTime.Duration < 10*time.Minute {
		p.RunTime.Duration = 10 * time.Minute
	}
	tick := time.NewTicker(p.ObjectInterval.Duration)
	defer tick.Stop()
	if p.OpenMetric {
		go func() {
			if p.MetricInterval.Duration == 0 {
				p.MetricInterval.Duration = 5 * time.Minute
			}
			tick := time.NewTicker(p.MetricInterval.Duration)
			defer tick.Stop()
			for {
				p.WriteMetric()
				select {
				case <-tick.C:
				case <-datakit.Exit.Wait():
					l.Info("process exit")
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

func (p *Processes) getProcesses() (processList []*pr.Process) {
	pses, err := pr.Processes()
	if err != nil {
		l.Errorf("[error] get process err:%s", err.Error())
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

func (p *Processes) Parse(ps *pr.Process) (username, state, name string, fields, message map[string]interface{}) {
	fields = map[string]interface{}{}
	message = map[string]interface{}{}
	username, err := ps.Username()
	if err != nil {
		l.Warnf("[warning] process get username err:%s", err.Error())
	}
	status, err := ps.Status()
	if err != nil {
		l.Warnf("[warning] process get state err:%s", err.Error())
	}
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
		fields["mem"] = memPercent
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
		fields["cpu"] = cpuPercent
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
	name, err = ps.Name()
	if err != nil {
		l.Warnf("[warning] process get name err:%s", err.Error())
	}
	return username, status[0], name, fields, message

}

func (p *Processes) WriteObject() {
	for _, ps := range p.getProcesses() {
		t, _ := ps.CreateTime()
		username, state, name, fields, message := p.Parse(ps)
		tags := map[string]string{
			"username":     username,
			"state":        state,
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
		io.NamedFeedEx(inputName, io.Object, "host_processes", tags, fields, time.Now().UTC())
	}
}

func (p *Processes) WriteMetric() {
	for _, ps := range p.getProcesses() {
		username, _, name, fields, _ := p.Parse(ps)
		tags := map[string]string{
			"username":     username,
			"pid":          fmt.Sprintf("%d", ps.Pid),
			"process_name": name,
		}
		io.NamedFeedEx(inputName, io.Metric, "host_processes", tags, fields, time.Now().UTC())
	}
}

func init() {
	inputs.Add(inputName, func() inputs.Input {
		return &Processes{}
	})
}
